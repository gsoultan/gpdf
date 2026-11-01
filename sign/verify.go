package sign

import (
	"bytes"
	"crypto/x509"
	"errors"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/gsoultan/gpdf/model"
	"github.com/gsoultan/gpdf/reader"
)

// VerifyOptions controls signature verification.
type VerifyOptions struct {
	// Roots are the trusted certificates. Nil uses the system roots.
	Roots *x509.CertPool
	// Time is when the certificate chain is checked (default: the signing time from
	// the signature, else now).
	Time time.Time
}

// SignatureInfo is the result of checking one signature field.
type SignatureInfo struct {
	FieldName string
	SubFilter string
	Signer    string // certificate subject common name
	Cert      *x509.Certificate
	// SigningTime is the signature dictionary's /M (or the CMS signing-time).
	SigningTime time.Time
	Reason      string
	Location    string
	ContactInfo string
	// Certification is the DocMDP level for a certification signature, else 0.
	Certification int
	// IntegrityOK is true when the byte range digest matches and the signature over
	// it checks out with the signer certificate.
	IntegrityOK bool
	// ChainOK is true when the signer certificate chains to a trusted root.
	ChainOK bool
	// CoversWholeDocument is false when the file was changed (incremental updates
	// appended) after this signature was made.
	CoversWholeDocument bool
	// HasTimestamp reports an embedded RFC 3161 time-stamp token (PAdES-B-T). The
	// token's own signature is not checked.
	HasTimestamp bool
	// Err explains a failed IntegrityOK or ChainOK.
	Err error
}

// Verify checks every signature field in the document.
func Verify(r io.ReaderAt, size int64, opts VerifyOptions) ([]SignatureInfo, error) {
	doc, err := reader.NewPDFReader().ReadDocument(r, size)
	if err != nil {
		return nil, fmt.Errorf("sign: read: %w", err)
	}
	fields, err := doc.FormFields()
	if err != nil {
		return nil, err
	}
	certLevel, _ := docMDPLevel(doc)
	var out []SignatureInfo
	for _, f := range fields {
		if ft, _ := f[model.Name("FT")].(model.Name); ft != "Sig" {
			continue
		}
		sig, ok := resolveDict(doc, f[model.Name("V")])
		if !ok {
			continue
		}
		info := SignatureInfo{FieldName: textOf(f[model.Name("T")]), Reason: textOf(sig[model.Name("Reason")]),
			Location: textOf(sig[model.Name("Location")]), ContactInfo: textOf(sig[model.Name("ContactInfo")])}
		if sf, _ := sig[model.Name("SubFilter")].(model.Name); sf != "" {
			info.SubFilter = string(sf)
		}
		if m := textOf(sig[model.Name("M")]); m != "" {
			if t, err := model.ParseDate(m); err == nil {
				info.SigningTime = t
			}
		}
		if certLevel > 0 {
			if cat, _ := doc.Catalog(); cat != nil {
				if perms, ok := resolveDict(doc, cat.Dict[model.Name("Perms")]); ok {
					if ref, ok := perms[model.Name("DocMDP")].(model.Ref); ok && f[model.Name("V")] == ref {
						info.Certification = certLevel
					}
				}
			}
		}
		verifyOne(r, size, sig, &info, opts)
		out = append(out, info)
	}
	return out, nil
}

// VerifyFile is Verify for a file on disk.
func VerifyFile(path string, opts VerifyOptions) ([]SignatureInfo, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	st, err := f.Stat()
	if err != nil {
		return nil, err
	}
	return Verify(f, st.Size(), opts)
}

func verifyOne(r io.ReaderAt, size int64, sig model.Dict, info *SignatureInfo, opts VerifyOptions) {
	br, _ := sig[model.Name("ByteRange")].(model.Array)
	if len(br) != 4 {
		info.Err = errors.New("sign: ByteRange must have four numbers")
		return
	}
	var ranges [4]int64
	for i, v := range br {
		n, ok := v.(model.Integer)
		if !ok || n < 0 {
			info.Err = errors.New("sign: invalid ByteRange")
			return
		}
		ranges[i] = int64(n)
	}
	var contents []byte
	switch c := sig[model.Name("Contents")].(type) {
	case model.HexString:
		contents = []byte(c)
	case model.String:
		contents = []byte(c)
	default:
		info.Err = errors.New("sign: missing /Contents")
		return
	}
	if ranges[0] != 0 || ranges[1] > size || ranges[2] < ranges[1] || ranges[2]+ranges[3] > size {
		info.Err = errors.New("sign: ByteRange does not fit the file")
		return
	}
	// The gap between the ranges must be exactly the /Contents string.
	gap := make([]byte, 1)
	r.ReadAt(gap, ranges[1])
	end := make([]byte, 1)
	r.ReadAt(end, ranges[2]-1)
	if gap[0] != '<' || end[0] != '>' {
		info.Err = errors.New("sign: ByteRange gap is not the /Contents string")
		return
	}
	info.CoversWholeDocument = ranges[2]+ranges[3] == size

	cms, err := parseCMS(contents)
	if err != nil {
		info.Err = err
		return
	}
	info.Cert = cms.Signer
	info.Signer = subjectName(cms.Signer.Subject)
	info.HasTimestamp = cms.Timestamp != nil
	if info.SigningTime.IsZero() {
		info.SigningTime = cms.SigningTime
	}
	if !cms.Hash.Available() {
		info.Err = fmt.Errorf("%w: digest not linked in", errUnsupportedAlgorithm)
		return
	}
	h := cms.Hash.New()
	if _, err := io.Copy(h, io.NewSectionReader(r, ranges[0], ranges[1])); err != nil {
		info.Err = err
		return
	}
	if _, err := io.Copy(h, io.NewSectionReader(r, ranges[2], ranges[3])); err != nil {
		info.Err = err
		return
	}
	digest := h.Sum(nil)
	var signed []byte
	if cms.SignedAttrs != nil {
		if !bytes.Equal(cms.MessageDigest, digest) {
			info.Err = errors.New("sign: document digest does not match the signed message-digest: the signed bytes were changed")
			return
		}
		signed = cms.SignedAttrs
	} else {
		// No signed attributes: the signature is over the content itself. x509's
		// CheckSignature hashes the input, so pass the ranges' content; for a
		// detached signature that is the byte range, read again.
		var all bytes.Buffer
		io.Copy(&all, io.NewSectionReader(r, ranges[0], ranges[1]))
		io.Copy(&all, io.NewSectionReader(r, ranges[2], ranges[3]))
		signed = all.Bytes()
	}
	if err := cms.Signer.CheckSignature(cms.Algorithm, signed, cms.Signature); err != nil {
		info.Err = fmt.Errorf("sign: signature check failed: %w", err)
		return
	}
	info.IntegrityOK = true

	inter := x509.NewCertPool()
	for _, c := range cms.Certificates {
		if c != cms.Signer {
			inter.AddCert(c)
		}
	}
	at := opts.Time
	if at.IsZero() {
		at = info.SigningTime
	}
	if at.IsZero() {
		at = time.Now()
	}
	_, err = cms.Signer.Verify(x509.VerifyOptions{Roots: opts.Roots, Intermediates: inter, CurrentTime: at,
		KeyUsages: []x509.ExtKeyUsage{x509.ExtKeyUsageAny}})
	if err != nil {
		info.Err = fmt.Errorf("sign: certificate chain: %w", err)
		return
	}
	info.ChainOK = true
}

func textOf(o model.Object) string {
	switch v := o.(type) {
	case model.String:
		return v.Text()
	case model.HexString:
		return v.Text()
	}
	return ""
}
