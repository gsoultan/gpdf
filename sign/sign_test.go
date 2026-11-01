package sign

import (
	"bytes"
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"errors"
	"math/big"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/gsoultan/gpdf/doc"
	"github.com/gsoultan/gpdf/model"
)

func testCert(t *testing.T, key crypto.Signer, cn string) *x509.Certificate {
	t.Helper()
	tmpl := &x509.Certificate{
		SerialNumber: big.NewInt(time.Now().UnixNano()),
		Subject:      pkix.Name{CommonName: cn, Organization: []string{"gPDF tests"}},
		NotBefore:    time.Now().Add(-time.Hour),
		NotAfter:     time.Now().Add(24 * time.Hour),
		KeyUsage:     x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageAny},
		IsCA:         true, BasicConstraintsValid: true,
	}
	der, err := x509.CreateCertificate(rand.Reader, tmpl, tmpl, key.Public(), key)
	if err != nil {
		t.Fatal(err)
	}
	c, err := x509.ParseCertificate(der)
	if err != nil {
		t.Fatal(err)
	}
	return c
}

func samplePDF(t *testing.T) []byte {
	t.Helper()
	b := doc.New().Title("Contract").PageSize(300, 200).NoCompressContent().AddPage()
	b.DrawText("Please sign here", 20, 100, "Helvetica", 12)
	built, err := b.Build()
	if err != nil {
		t.Fatal(err)
	}
	var buf bytes.Buffer
	if err := built.Save(&buf); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

func TestSignAndVerify(t *testing.T) {
	rsaKey, _ := rsa.GenerateKey(rand.Reader, 2048)
	ecKey, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	for _, tc := range []struct {
		name string
		key  crypto.Signer
		hash crypto.Hash
		sub  SubFilter
	}{
		{"rsa-sha256-cades", rsaKey, crypto.SHA256, SubFilterCAdES},
		{"ecdsa-sha384-pkcs7", ecKey, crypto.SHA384, SubFilterPKCS7},
	} {
		t.Run(tc.name, func(t *testing.T) {
			cert := testCert(t, tc.key, "Alice Example")
			pdf := samplePDF(t)
			signer := NewCMSSigner(tc.key, cert)
			signer.IncludeSigningTime = tc.sub == SubFilterPKCS7
			when := time.Date(2026, 11, 1, 10, 0, 0, 0, time.UTC)
			var out bytes.Buffer
			err := Sign(bytes.NewReader(pdf), int64(len(pdf)), &out, signer, SignOptions{
				Reason: "Approved", Location: "Jakarta", Hash: tc.hash, SubFilter: tc.sub, SigningTime: when,
				Rect: [4]float64{20, 20, 200, 70}, Certify: 2,
			})
			if err != nil {
				t.Fatal(err)
			}
			signed := out.Bytes()
			if !bytes.HasPrefix(signed, pdf) {
				t.Fatal("original bytes were changed")
			}
			roots := x509.NewCertPool()
			roots.AddCert(cert)
			infos, err := Verify(bytes.NewReader(signed), int64(len(signed)), VerifyOptions{Roots: roots})
			if err != nil {
				t.Fatal(err)
			}
			if len(infos) != 1 {
				t.Fatalf("signatures = %+v", infos)
			}
			s := infos[0]
			if !s.IntegrityOK || !s.ChainOK || !s.CoversWholeDocument || s.Err != nil {
				t.Fatalf("signature not valid: %+v", s)
			}
			if s.Signer != "Alice Example" || s.Reason != "Approved" || s.Location != "Jakarta" || s.FieldName != "Signature1" ||
				!s.SigningTime.Equal(when) || s.SubFilter != string(tc.sub) || s.Certification != 2 {
				t.Errorf("info = %+v", s)
			}
			// Without the root the chain fails but integrity holds.
			infos, _ = Verify(bytes.NewReader(signed), int64(len(signed)), VerifyOptions{Roots: x509.NewCertPool()})
			if !infos[0].IntegrityOK || infos[0].ChainOK {
				t.Errorf("untrusted root: %+v", infos[0])
			}
			// Tampering with the signed bytes is detected.
			bad := bytes.Replace(signed, []byte("(Please sign here)"), []byte("(Please sign HERE)"), 1)
			if bytes.Equal(bad, signed) {
				t.Fatal("tamper target not found")
			}
			infos, _ = Verify(bytes.NewReader(bad), int64(len(bad)), VerifyOptions{Roots: roots})
			if infos[0].IntegrityOK || infos[0].Err == nil {
				t.Errorf("tampered document verified: %+v", infos[0])
			}
			// The signed file still opens and reads.
			d, err := doc.OpenReader(bytes.NewReader(signed), int64(len(signed)))
			if err != nil {
				t.Fatal(err)
			}
			if text, _ := d.ReadContent(); !strings.Contains(text, "Please sign here") {
				t.Errorf("text = %q", text)
			}
			popplerSig(t, signed, cert)
		})
	}
}

func popplerSig(t *testing.T, pdf []byte, cert *x509.Certificate) {
	t.Helper()
	bin, err := exec.LookPath("pdfsig")
	if err != nil {
		return
	}
	dir := t.TempDir()
	path := filepath.Join(dir, "signed.pdf")
	os.WriteFile(path, pdf, 0o600)
	out, _ := exec.Command(bin, path).CombinedOutput()
	s := string(out)
	if !strings.Contains(s, "Signature is Valid") || strings.Contains(s, "Digest Mismatch") || strings.Contains(s, "Error") {
		t.Errorf("pdfsig:\n%s", s)
	}
}

func TestSecondSignatureKeepsFirstValid(t *testing.T) {
	key, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	cert := testCert(t, key, "Signer")
	pdf := samplePDF(t)
	var once bytes.Buffer
	if err := Sign(bytes.NewReader(pdf), int64(len(pdf)), &once, NewCMSSigner(key, cert), SignOptions{}); err != nil {
		t.Fatal(err)
	}
	var twice bytes.Buffer
	if err := Sign(bytes.NewReader(once.Bytes()), int64(once.Len()), &twice, NewCMSSigner(key, cert), SignOptions{FieldName: "Countersign"}); err != nil {
		t.Fatal(err)
	}
	roots := x509.NewCertPool()
	roots.AddCert(cert)
	infos, err := Verify(bytes.NewReader(twice.Bytes()), int64(twice.Len()), VerifyOptions{Roots: roots})
	if err != nil {
		t.Fatal(err)
	}
	if len(infos) != 2 {
		t.Fatalf("got %d signatures", len(infos))
	}
	if !infos[0].IntegrityOK || infos[0].CoversWholeDocument || infos[0].FieldName != "Signature1" {
		t.Errorf("first: %+v", infos[0])
	}
	if !infos[1].IntegrityOK || !infos[1].CoversWholeDocument || infos[1].FieldName != "Countersign" {
		t.Errorf("second: %+v", infos[1])
	}
	// A certified (no changes) document refuses further signatures.
	var cert1 bytes.Buffer
	if err := Sign(bytes.NewReader(pdf), int64(len(pdf)), &cert1, NewCMSSigner(key, cert), SignOptions{Certify: 1}); err != nil {
		t.Fatal(err)
	}
	if err := Sign(bytes.NewReader(cert1.Bytes()), int64(cert1.Len()), &bytes.Buffer{}, NewCMSSigner(key, cert), SignOptions{}); !errors.Is(err, ErrCertified) {
		t.Errorf("signing a locked document: %v", err)
	}
}

func TestTimestampAndValidationData(t *testing.T) {
	key, _ := rsa.GenerateKey(rand.Reader, 2048)
	cert := testCert(t, key, "Stamped")
	tsaKey, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	tsaCert := testCert(t, tsaKey, "TSA")
	// A fake TSA: any CMS SignedData stands in for the TimeStampToken ContentInfo.
	signer := NewCMSSigner(key, cert)
	called := false
	signer.Timestamp = func(h crypto.Hash, digest []byte) ([]byte, error) {
		called = true
		return NewCMSSigner(tsaKey, tsaCert).Sign(digest, h, time.Now())
	}
	pdf := samplePDF(t)
	var out bytes.Buffer
	if err := Sign(bytes.NewReader(pdf), int64(len(pdf)), &out, signer, SignOptions{}); err != nil {
		t.Fatal(err)
	}
	if !called {
		t.Fatal("Timestamp not called")
	}
	var lt bytes.Buffer
	if err := AddValidationData(bytes.NewReader(out.Bytes()), int64(out.Len()), &lt, ValidationData{Certs: [][]byte{cert.Raw, tsaCert.Raw}}); err != nil {
		t.Fatal(err)
	}
	roots := x509.NewCertPool()
	roots.AddCert(cert)
	infos, err := Verify(bytes.NewReader(lt.Bytes()), int64(lt.Len()), VerifyOptions{Roots: roots})
	if err != nil || len(infos) != 1 {
		t.Fatalf("%v %+v", err, infos)
	}
	if !infos[0].IntegrityOK || !infos[0].HasTimestamp || infos[0].CoversWholeDocument {
		t.Errorf("%+v", infos[0])
	}
	d, err := doc.OpenReader(bytes.NewReader(lt.Bytes()), int64(lt.Len()))
	if err != nil {
		t.Fatal(err)
	}
	cat, _ := d.Catalog()
	dssObj, err := d.Resolve(cat.Dict[model.Name("DSS")].(model.Ref))
	if err != nil {
		t.Fatal(err)
	}
	certs := dssObj.(model.Dict)[model.Name("Certs")].(model.Array)
	if len(certs) != 2 {
		t.Errorf("DSS /Certs = %v", certs)
	}
	first, _ := d.Resolve(certs[0].(model.Ref))
	if got := first.(*model.Stream).Content; !bytes.Equal(got, cert.Raw) {
		t.Errorf("DSS certificate differs (%d bytes)", len(got))
	}
}
