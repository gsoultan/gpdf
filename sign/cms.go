// Copyright (c) 2026 Gembit Soultan Shirazi <gembit.soultan@gmail.com>

package sign

import (
	"bytes"
	"crypto"
	"crypto/ecdsa"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/asn1"
	"errors"
	"fmt"
	"math/big"
	"time"
)

// CMS (RFC 5652) SignedData with the CAdES attributes PAdES baseline signatures need
// (ETSI EN 319 122-1 / 319 142-1): content-type, message-digest and
// ESS signing-certificate-v2 as signed attributes, an optional RFC 3161 time-stamp
// token as an unsigned attribute. The standard library's encoding/asn1 does the DER.

var (
	oidData                 = asn1.ObjectIdentifier{1, 2, 840, 113549, 1, 7, 1}
	oidSignedData           = asn1.ObjectIdentifier{1, 2, 840, 113549, 1, 7, 2}
	oidAttrContentType      = asn1.ObjectIdentifier{1, 2, 840, 113549, 1, 9, 3}
	oidAttrMessageDigest    = asn1.ObjectIdentifier{1, 2, 840, 113549, 1, 9, 4}
	oidAttrSigningTime      = asn1.ObjectIdentifier{1, 2, 840, 113549, 1, 9, 5}
	oidAttrSigningCertV2    = asn1.ObjectIdentifier{1, 2, 840, 113549, 1, 9, 16, 2, 47}
	oidAttrTimeStampToken   = asn1.ObjectIdentifier{1, 2, 840, 113549, 1, 9, 16, 2, 14}
	oidSHA1                 = asn1.ObjectIdentifier{1, 3, 14, 3, 2, 26}
	oidSHA256               = asn1.ObjectIdentifier{2, 16, 840, 1, 101, 3, 4, 2, 1}
	oidSHA384               = asn1.ObjectIdentifier{2, 16, 840, 1, 101, 3, 4, 2, 2}
	oidSHA512               = asn1.ObjectIdentifier{2, 16, 840, 1, 101, 3, 4, 2, 3}
	oidRSAEncryption        = asn1.ObjectIdentifier{1, 2, 840, 113549, 1, 1, 1}
	oidSHA256WithRSA        = asn1.ObjectIdentifier{1, 2, 840, 113549, 1, 1, 11}
	oidSHA384WithRSA        = asn1.ObjectIdentifier{1, 2, 840, 113549, 1, 1, 12}
	oidSHA512WithRSA        = asn1.ObjectIdentifier{1, 2, 840, 113549, 1, 1, 13}
	oidSHA1WithRSA          = asn1.ObjectIdentifier{1, 2, 840, 113549, 1, 1, 5}
	oidECDSAWithSHA256      = asn1.ObjectIdentifier{1, 2, 840, 10045, 4, 3, 2}
	oidECDSAWithSHA384      = asn1.ObjectIdentifier{1, 2, 840, 10045, 4, 3, 3}
	oidECDSAWithSHA512      = asn1.ObjectIdentifier{1, 2, 840, 10045, 4, 3, 4}
	oidECDSAWithSHA1        = asn1.ObjectIdentifier{1, 2, 840, 10045, 4, 1}
	oidECPublicKey          = asn1.ObjectIdentifier{1, 2, 840, 10045, 2, 1}
	asn1NullBytes           = []byte{5, 0}
	errUnsupportedAlgorithm = errors.New("sign: unsupported algorithm")
)

type algorithmIdentifier struct {
	Algorithm  asn1.ObjectIdentifier
	Parameters asn1.RawValue `asn1:"optional"`
}

type contentInfo struct {
	ContentType asn1.ObjectIdentifier
	Content     asn1.RawValue `asn1:"explicit,tag:0"`
}

// contentInfoOut is contentInfo for marshaling: encoding/asn1 writes a RawValue as
// is, so the [0] EXPLICIT wrapper is built by hand.
type contentInfoOut struct {
	ContentType asn1.ObjectIdentifier
	Content     asn1.RawValue
}

func explicit0(der []byte) asn1.RawValue {
	return asn1.RawValue{Class: asn1.ClassContextSpecific, Tag: 0, IsCompound: true, Bytes: der}
}

type encapsulatedContentInfo struct {
	ContentType asn1.ObjectIdentifier
	Content     asn1.RawValue `asn1:"optional,explicit,tag:0"`
}

type issuerAndSerialNumber struct {
	Issuer       asn1.RawValue
	SerialNumber *big.Int
}

type attribute struct {
	Type   asn1.ObjectIdentifier
	Values []asn1.RawValue `asn1:"set"`
}

type signerInfo struct {
	Version            int
	SID                issuerAndSerialNumber
	DigestAlgorithm    algorithmIdentifier
	SignedAttrs        asn1.RawValue `asn1:"optional,tag:0"`
	SignatureAlgorithm algorithmIdentifier
	Signature          []byte
	UnsignedAttrs      asn1.RawValue `asn1:"optional,tag:1"`
}

type signedData struct {
	Version          int
	DigestAlgorithms []algorithmIdentifier `asn1:"set"`
	EncapContentInfo encapsulatedContentInfo
	Certificates     asn1.RawValue `asn1:"optional,tag:0"`
	CRLs             asn1.RawValue `asn1:"optional,tag:1"`
	SignerInfos      []signerInfo  `asn1:"set"`
}

// essCertIDv2 is RFC 5035 ESSCertIDv2 with the default (SHA-256) hash algorithm omitted.
type essCertIDv2 struct {
	CertHash []byte
}

type signingCertificateV2 struct {
	Certs []essCertIDv2
}

func digestOID(h crypto.Hash) (asn1.ObjectIdentifier, error) {
	switch h {
	case crypto.SHA256:
		return oidSHA256, nil
	case crypto.SHA384:
		return oidSHA384, nil
	case crypto.SHA512:
		return oidSHA512, nil
	case crypto.SHA1:
		return oidSHA1, nil
	}
	return nil, fmt.Errorf("%w: digest %v", errUnsupportedAlgorithm, h)
}

func hashFromOID(oid asn1.ObjectIdentifier) (crypto.Hash, error) {
	switch {
	case oid.Equal(oidSHA256):
		return crypto.SHA256, nil
	case oid.Equal(oidSHA384):
		return crypto.SHA384, nil
	case oid.Equal(oidSHA512):
		return crypto.SHA512, nil
	case oid.Equal(oidSHA1):
		return crypto.SHA1, nil
	}
	return 0, fmt.Errorf("%w: digest %v", errUnsupportedAlgorithm, oid)
}

// signatureAlgorithm picks the SignerInfo signatureAlgorithm for a key type and
// digest, and the x509 algorithm used to check it.
func signatureAlgorithm(pub crypto.PublicKey, h crypto.Hash) (algorithmIdentifier, x509.SignatureAlgorithm, error) {
	switch pub.(type) {
	case *rsa.PublicKey:
		// CAdES names the key algorithm here; the digest comes from digestAlgorithm.
		ai := algorithmIdentifier{Algorithm: oidRSAEncryption, Parameters: asn1.RawValue{FullBytes: asn1NullBytes}}
		switch h {
		case crypto.SHA256:
			return ai, x509.SHA256WithRSA, nil
		case crypto.SHA384:
			return ai, x509.SHA384WithRSA, nil
		case crypto.SHA512:
			return ai, x509.SHA512WithRSA, nil
		case crypto.SHA1:
			return ai, x509.SHA1WithRSA, nil
		}
	case *ecdsa.PublicKey:
		switch h {
		case crypto.SHA256:
			return algorithmIdentifier{Algorithm: oidECDSAWithSHA256}, x509.ECDSAWithSHA256, nil
		case crypto.SHA384:
			return algorithmIdentifier{Algorithm: oidECDSAWithSHA384}, x509.ECDSAWithSHA384, nil
		case crypto.SHA512:
			return algorithmIdentifier{Algorithm: oidECDSAWithSHA512}, x509.ECDSAWithSHA512, nil
		case crypto.SHA1:
			return algorithmIdentifier{Algorithm: oidECDSAWithSHA1}, x509.ECDSAWithSHA1, nil
		}
	}
	return algorithmIdentifier{}, 0, fmt.Errorf("%w: key %T with %v", errUnsupportedAlgorithm, pub, h)
}

// verifyAlgorithm maps a SignerInfo's algorithms back to the x509 check.
func verifyAlgorithm(sig, digest asn1.ObjectIdentifier) (x509.SignatureAlgorithm, error) {
	switch {
	case sig.Equal(oidSHA256WithRSA):
		return x509.SHA256WithRSA, nil
	case sig.Equal(oidSHA384WithRSA):
		return x509.SHA384WithRSA, nil
	case sig.Equal(oidSHA512WithRSA):
		return x509.SHA512WithRSA, nil
	case sig.Equal(oidSHA1WithRSA):
		return x509.SHA1WithRSA, nil
	case sig.Equal(oidECDSAWithSHA256):
		return x509.ECDSAWithSHA256, nil
	case sig.Equal(oidECDSAWithSHA384):
		return x509.ECDSAWithSHA384, nil
	case sig.Equal(oidECDSAWithSHA512):
		return x509.ECDSAWithSHA512, nil
	case sig.Equal(oidECDSAWithSHA1):
		return x509.ECDSAWithSHA1, nil
	case sig.Equal(oidRSAEncryption), sig.Equal(oidECPublicKey):
		h, err := hashFromOID(digest)
		if err != nil {
			return 0, err
		}
		rsa := sig.Equal(oidRSAEncryption)
		switch h {
		case crypto.SHA256:
			return pick(rsa, x509.SHA256WithRSA, x509.ECDSAWithSHA256), nil
		case crypto.SHA384:
			return pick(rsa, x509.SHA384WithRSA, x509.ECDSAWithSHA384), nil
		case crypto.SHA512:
			return pick(rsa, x509.SHA512WithRSA, x509.ECDSAWithSHA512), nil
		case crypto.SHA1:
			return pick(rsa, x509.SHA1WithRSA, x509.ECDSAWithSHA1), nil
		}
	}
	return 0, fmt.Errorf("%w: signature %v", errUnsupportedAlgorithm, sig)
}

func pick(c bool, a, b x509.SignatureAlgorithm) x509.SignatureAlgorithm {
	if c {
		return a
	}
	return b
}

// CMSSigner builds CAdES-compliant CMS signatures with a private key and its
// certificate chain. It implements Signer.
type CMSSigner struct {
	Key   crypto.Signer
	Cert  *x509.Certificate
	Chain []*x509.Certificate
	// Timestamp, when set, is called with the signature value's digest and must return
	// an RFC 3161 TimeStampToken (the DER ContentInfo a time-stamping authority sends
	// back), which is embedded as the signature-time-stamp attribute (PAdES-B-T).
	Timestamp func(hash crypto.Hash, digest []byte) ([]byte, error)
	// IncludeSigningTime adds the signing-time signed attribute. PAdES baseline
	// signatures carry the time in the signature dictionary's /M instead, so it is
	// off by default; turn it on for plain adbe.pkcs7.detached signatures.
	IncludeSigningTime bool
}

// NewCMSSigner returns a CMSSigner for key and cert; chain holds the intermediate
// certificates to embed (the root may be omitted).
func NewCMSSigner(key crypto.Signer, cert *x509.Certificate, chain ...*x509.Certificate) *CMSSigner {
	return &CMSSigner{Key: key, Cert: cert, Chain: chain}
}

// Sign returns a DER CMS SignedData (detached: the signed content is the PDF byte
// range whose digest is given).
func (s *CMSSigner) Sign(digest []byte, hash crypto.Hash, signingTime time.Time) ([]byte, error) {
	if s.Key == nil || s.Cert == nil {
		return nil, errors.New("sign: CMSSigner needs a key and a certificate")
	}
	if hash == 0 {
		hash = crypto.SHA256
	}
	digestAlg, err := digestOID(hash)
	if err != nil {
		return nil, err
	}
	sigAlg, _, err := signatureAlgorithm(s.Key.Public(), hash)
	if err != nil {
		return nil, err
	}

	// ESS signing-certificate-v2 binds the signer certificate (SHA-256 of its DER).
	certHash := crypto.SHA256.New()
	certHash.Write(s.Cert.Raw)
	essDER, err := asn1.Marshal(signingCertificateV2{Certs: []essCertIDv2{{CertHash: certHash.Sum(nil)}}})
	if err != nil {
		return nil, err
	}
	attrs := []attribute{
		{Type: oidAttrContentType, Values: []asn1.RawValue{rawOID(oidData)}},
		{Type: oidAttrMessageDigest, Values: []asn1.RawValue{{Tag: asn1.TagOctetString, Bytes: digest}}},
		{Type: oidAttrSigningCertV2, Values: []asn1.RawValue{{FullBytes: essDER}}},
	}
	if s.IncludeSigningTime {
		if signingTime.IsZero() {
			signingTime = time.Now()
		}
		t, err := asn1.Marshal(signingTime.UTC())
		if err != nil {
			return nil, err
		}
		attrs = append(attrs, attribute{Type: oidAttrSigningTime, Values: []asn1.RawValue{{FullBytes: t}}})
	}
	signedAttrsDER, err := asn1.MarshalWithParams(attrs, "set")
	if err != nil {
		return nil, err
	}
	// The attributes are signed as a SET (tag 0x31); in the SignerInfo they are
	// written with the implicit [0] tag.
	h := hash.New()
	h.Write(signedAttrsDER)
	signature, err := s.Key.Sign(nil, h.Sum(nil), hash)
	if err != nil {
		return nil, fmt.Errorf("sign: %w", err)
	}
	implicitAttrs := append([]byte{0xA0}, signedAttrsDER[1:]...)

	si := signerInfo{
		Version: 1,
		SID: issuerAndSerialNumber{
			Issuer:       asn1.RawValue{FullBytes: s.Cert.RawIssuer},
			SerialNumber: s.Cert.SerialNumber,
		},
		DigestAlgorithm:    algorithmIdentifier{Algorithm: digestAlg, Parameters: asn1.RawValue{FullBytes: asn1NullBytes}},
		SignedAttrs:        asn1.RawValue{FullBytes: implicitAttrs},
		SignatureAlgorithm: sigAlg,
		Signature:          signature,
	}
	if s.Timestamp != nil {
		th := hash.New()
		th.Write(signature)
		token, err := s.Timestamp(hash, th.Sum(nil))
		if err != nil {
			return nil, fmt.Errorf("sign: time-stamp: %w", err)
		}
		if _, err := asn1.Unmarshal(token, &contentInfo{}); err != nil {
			return nil, fmt.Errorf("sign: time-stamp token is not a ContentInfo: %w", err)
		}
		unsigned, err := asn1.MarshalWithParams([]attribute{{Type: oidAttrTimeStampToken, Values: []asn1.RawValue{{FullBytes: token}}}}, "set")
		if err != nil {
			return nil, err
		}
		si.UnsignedAttrs = asn1.RawValue{FullBytes: append([]byte{0xA1}, unsigned[1:]...)}
	}

	var certs []byte
	certs = append(certs, s.Cert.Raw...)
	for _, c := range s.Chain {
		if c != nil {
			certs = append(certs, c.Raw...)
		}
	}
	sd := signedData{
		Version:          1,
		DigestAlgorithms: []algorithmIdentifier{si.DigestAlgorithm},
		EncapContentInfo: encapsulatedContentInfo{ContentType: oidData},
		Certificates:     asn1.RawValue{Class: asn1.ClassContextSpecific, Tag: 0, IsCompound: true, Bytes: certs},
		SignerInfos:      []signerInfo{si},
	}
	sdDER, err := asn1.Marshal(sd)
	if err != nil {
		return nil, err
	}
	return asn1.Marshal(contentInfoOut{ContentType: oidSignedData, Content: explicit0(sdDER)})
}

func rawOID(oid asn1.ObjectIdentifier) asn1.RawValue {
	b, _ := asn1.Marshal(oid)
	return asn1.RawValue{FullBytes: b}
}

// parsedCMS is what verification needs from a SignedData.
type parsedCMS struct {
	Certificates []*x509.Certificate
	Signer       *x509.Certificate
	Hash         crypto.Hash
	// MessageDigest is the signed message-digest attribute (nil when the signer did
	// not use signed attributes, in which case the signature is over the content).
	MessageDigest []byte
	SignedAttrs   []byte // DER with the SET tag, as signed
	Signature     []byte
	Algorithm     x509.SignatureAlgorithm
	SigningTime   time.Time
	Timestamp     []byte // RFC 3161 token, if any
}

func parseCMS(der []byte) (*parsedCMS, error) {
	var ci contentInfo
	rest, err := asn1.Unmarshal(der, &ci)
	if err != nil {
		return nil, fmt.Errorf("sign: CMS: %w", err)
	}
	_ = rest // /Contents is zero-padded, so trailing bytes are expected
	if !ci.ContentType.Equal(oidSignedData) {
		return nil, fmt.Errorf("sign: CMS content type %v is not SignedData", ci.ContentType)
	}
	var sd signedData
	if _, err := asn1.Unmarshal(ci.Content.Bytes, &sd); err != nil {
		return nil, fmt.Errorf("sign: SignedData: %w", err)
	}
	if len(sd.SignerInfos) == 0 {
		return nil, errors.New("sign: SignedData has no SignerInfo")
	}
	out := &parsedCMS{}
	if len(sd.Certificates.Bytes) > 0 {
		out.Certificates, err = parseCertificateSet(sd.Certificates.Bytes)
		if err != nil {
			return nil, err
		}
	}
	si := sd.SignerInfos[0]
	for _, c := range out.Certificates {
		if c.SerialNumber.Cmp(si.SID.SerialNumber) == 0 && bytes.Equal(c.RawIssuer, si.SID.Issuer.FullBytes) {
			out.Signer = c
			break
		}
	}
	if out.Signer == nil {
		return nil, errors.New("sign: signer certificate is not in the CMS")
	}
	out.Hash, err = hashFromOID(si.DigestAlgorithm.Algorithm)
	if err != nil {
		return nil, err
	}
	out.Algorithm, err = verifyAlgorithm(si.SignatureAlgorithm.Algorithm, si.DigestAlgorithm.Algorithm)
	if err != nil {
		return nil, err
	}
	out.Signature = si.Signature
	if len(si.SignedAttrs.FullBytes) > 0 {
		out.SignedAttrs = append([]byte{0x31}, si.SignedAttrs.FullBytes[1:]...)
		var attrs []attribute
		if _, err := asn1.UnmarshalWithParams(out.SignedAttrs, &attrs, "set"); err != nil {
			return nil, fmt.Errorf("sign: signed attributes: %w", err)
		}
		for _, a := range attrs {
			if len(a.Values) == 0 {
				continue
			}
			switch {
			case a.Type.Equal(oidAttrMessageDigest):
				out.MessageDigest = a.Values[0].Bytes
			case a.Type.Equal(oidAttrSigningTime):
				var t time.Time
				if _, err := asn1.Unmarshal(a.Values[0].FullBytes, &t); err == nil {
					out.SigningTime = t
				}
			}
		}
		if out.MessageDigest == nil {
			return nil, errors.New("sign: signed attributes lack message-digest")
		}
	}
	if len(si.UnsignedAttrs.FullBytes) > 0 {
		var attrs []attribute
		if _, err := asn1.UnmarshalWithParams(append([]byte{0x31}, si.UnsignedAttrs.FullBytes[1:]...), &attrs, "set"); err == nil {
			for _, a := range attrs {
				if a.Type.Equal(oidAttrTimeStampToken) && len(a.Values) > 0 {
					out.Timestamp = a.Values[0].FullBytes
				}
			}
		}
	}
	return out, nil
}

// parseCertificateSet reads the concatenated certificates of the [0] IMPLICIT set.
func parseCertificateSet(b []byte) ([]*x509.Certificate, error) {
	var out []*x509.Certificate
	for len(b) > 0 {
		var raw asn1.RawValue
		rest, err := asn1.Unmarshal(b, &raw)
		if err != nil {
			return nil, fmt.Errorf("sign: certificates: %w", err)
		}
		// Only plain certificates are supported (not extended or attribute certificates).
		if raw.Class == asn1.ClassUniversal && raw.Tag == asn1.TagSequence {
			c, err := x509.ParseCertificate(raw.FullBytes)
			if err != nil {
				return nil, fmt.Errorf("sign: certificate: %w", err)
			}
			out = append(out, c)
		}
		b = rest
	}
	return out, nil
}

// subjectName renders a certificate subject for reports.
func subjectName(n pkix.Name) string {
	if n.CommonName != "" {
		return n.CommonName
	}
	return n.String()
}
