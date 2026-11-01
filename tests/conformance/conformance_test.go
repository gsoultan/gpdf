// Copyright (c) 2026 Gembit Soultan Shirazi <gembit.soultan@gmail.com>

// Package conformance_test writes a corpus of documents that exercise every
// output feature and checks them with the external validators when they are
// installed: qpdf (file structure, linearization) and veraPDF (PDF/A and
// PDF/UA conformance). CI installs both and sets GPDF_REQUIRE_TOOLS=1, so a
// missing tool fails there instead of skipping; GPDF_CONFORMANCE_DIR keeps the
// corpus for inspection.
package conformance_test

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"image"
	"image/color"
	"image/png"
	"math/big"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/gsoultan/gpdf/doc"
	"github.com/gsoultan/gpdf/font/system"
	"github.com/gsoultan/gpdf/model"
	"github.com/gsoultan/gpdf/sign"
)

type sample struct {
	name string
	// flavour lists the veraPDF profiles to validate against, space separated
	// ("" for none).
	flavour  string
	password string
	pdf      []byte
}

func squarePNG(t *testing.T) []byte {
	t.Helper()
	img := image.NewNRGBA(image.Rect(0, 0, 16, 16))
	for y := range 16 {
		for x := range 16 {
			img.Set(x, y, color.NRGBA{R: uint8(x * 16), G: uint8(y * 16), B: 120, A: 255})
		}
	}
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

func save(t *testing.T, b *doc.DocumentBuilder) []byte {
	t.Helper()
	d, err := b.Build()
	if err != nil {
		t.Fatal(err)
	}
	var buf bytes.Buffer
	if err := d.Save(&buf); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

func reopen(t *testing.T, pdf []byte) doc.Document {
	t.Helper()
	d, err := doc.OpenReader(bytes.NewReader(pdf), int64(len(pdf)))
	if err != nil {
		t.Fatal(err)
	}
	return d
}

// basic draws text, an image and a link on two pages.
func basic(t *testing.T) *doc.DocumentBuilder {
	b := doc.New().Title("Conformance corpus").Author("gPDF").SetLanguage("en").PageSize(400, 300)
	for i := range 2 {
		b.AddPage()
		b.DrawText("Page text with a link", 20, 260, "Helvetica", 12)
		b.DrawText("Second line of body text", 20, 240, "Helvetica", 12)
		b.DrawPNG(20, 100, 80, 80, squarePNG(t))
		b.AddLinkToURL(i, 20, 255, 160, 272, "https://example.org/")
	}
	return b
}

// tagged draws the structures PDF/UA asks for.
func tagged(t *testing.T, level doc.PDFUALevel) *doc.DocumentBuilder {
	b := doc.New().Title("Accessible corpus").SetLanguage("en").PDFUA(level).PageSize(500, 700).AddPage()
	b.SetFooter(doc.PageText{Text: "Page {page} of {pages}"})
	b.DrawHeading(0, 1, "Heading one", 40, 650, "Helvetica-Bold", 16)
	b.DrawParagraph(0, "A paragraph of body text in the first section.", 40, 620, "Helvetica", 12)
	b.DrawList(0, []string{"first item", "second item"}, 40, 590, 14, true, "Helvetica", 12)
	b.BeginTable(0, 40, 520, 300, 40, 2).Header("Name", "Value").Row("alpha", "1").EndTable()
	b.DrawTaggedPNG(0, 40, 400, 60, 60, squarePNG(t), "A coloured square")
	b.DrawParagraph(0, "See the project site.", 40, 380, "Helvetica", 12)
	b.AddLinkToURL(0, 40, 375, 150, 392, "https://example.org/")
	return b
}

func selfSigned(t *testing.T) (*rsa.PrivateKey, *x509.Certificate) {
	t.Helper()
	key, _ := rsa.GenerateKey(rand.Reader, 2048)
	tmpl := &x509.Certificate{
		SerialNumber:          big.NewInt(1),
		Subject:               pkix.Name{CommonName: "Conformance"},
		NotBefore:             time.Now().Add(-time.Hour),
		NotAfter:              time.Now().Add(time.Hour),
		KeyUsage:              x509.KeyUsageDigitalSignature,
		IsCA:                  true,
		BasicConstraintsValid: true,
	}
	der, err := x509.CreateCertificate(rand.Reader, tmpl, tmpl, &key.PublicKey, key)
	if err != nil {
		t.Fatal(err)
	}
	cert, _ := x509.ParseCertificate(der)
	return key, cert
}

// corpus builds every sample. Samples that need system fonts are left out
// when the fonts are missing.
func corpus(t *testing.T) []sample {
	t.Helper()
	var out []sample
	add := func(name, flavour string, pdf []byte) {
		out = append(out, sample{name: name, flavour: flavour, pdf: pdf})
	}

	plain := save(t, basic(t))
	add("basic", "", plain)
	add("classic", "", save(t, basic(t).NoObjectStreams()))
	add("no-compress", "", save(t, basic(t).NoCompressContent()))
	add("compressed-images", "", save(t, basic(t).CompressImages(doc.ImageCompression{MaxDPI: 50, JPEGQuality: 80})))

	built, err := basic(t).Build()
	if err != nil {
		t.Fatal(err)
	}
	var rc4, aes bytes.Buffer
	if err := built.SaveWithPassword(&rc4, "user", "owner"); err != nil {
		t.Fatal(err)
	}
	if err := built.SaveWithAES256Password(&aes, "user", "owner"); err != nil {
		t.Fatal(err)
	}
	out = append(out, sample{name: "rc4", password: "user", pdf: rc4.Bytes()}, sample{name: "aes256", password: "user", pdf: aes.Bytes()})

	lin, err := os.CreateTemp(t.TempDir(), "lin*.pdf")
	if err != nil {
		t.Fatal(err)
	}
	if err := built.SaveLinearized(lin); err != nil {
		t.Fatal(err)
	}
	lin.Close()
	linBytes, _ := os.ReadFile(lin.Name())
	add("linearized", "", linBytes)

	var streamed bytes.Buffer
	if err := basic(t).BuildTo(&streamed); err != nil {
		t.Fatal(err)
	}
	add("streamed", "", streamed.Bytes())

	merged, err := doc.Merge(reopen(t, plain), reopen(t, linBytes))
	if err != nil {
		t.Fatal(err)
	}
	var mergedBuf bytes.Buffer
	if err := merged.Save(&mergedBuf); err != nil {
		t.Fatal(err)
	}
	add("merged", "", mergedBuf.Bytes())

	key, cert := selfSigned(t)
	var signed bytes.Buffer
	if err := sign.Sign(bytes.NewReader(plain), int64(len(plain)), &signed, sign.NewCMSSigner(key, cert), sign.SignOptions{Reason: "corpus", Rect: [4]float64{200, 20, 380, 60}}); err != nil {
		t.Fatal(err)
	}
	add("signed", "", signed.Bytes())

	upd := reopen(t, signed.Bytes())
	if err := upd.SetInfo("Subject", model.String("updated after signing")); err != nil {
		t.Fatal(err)
	}
	var incremental bytes.Buffer
	if err := upd.SaveIncremental(&incremental); err != nil {
		t.Fatal(err)
	}
	add("incremental", "", incremental.Bytes())

	red := reopen(t, plain)
	if _, err := doc.Redact(red, doc.RedactOptions{Texts: []string{"link"}}); err != nil {
		t.Fatal(err)
	}
	var redacted bytes.Buffer
	if err := red.Save(&redacted); err != nil {
		t.Fatal(err)
	}
	add("redacted", "", redacted.Bytes())

	if _, ok := system.DefaultStandardSubstitutes()["Helvetica"]; !ok {
		t.Log("no Helvetica substitute installed: PDF/A and PDF/UA samples left out")
		return out
	}
	created := time.Date(2026, 11, 1, 12, 0, 0, 0, time.UTC)
	archive := func(level doc.PDFALevel) *doc.DocumentBuilder {
		return basic(t).EmbedStandardFonts().PDFA(level).CreationDate(created)
	}
	add("pdfa-2b", "2b", save(t, archive(doc.PDFA2B)))
	attachment := doc.Attachment{Name: "data.csv", MIME: "text/csv", Data: []byte("a,b\n1,2\n"), Relationship: doc.RelationshipData, Description: "the data"}
	add("pdfa-3b", "3b", save(t, archive(doc.PDFA3B).AttachFile(attachment)))
	add("pdfa-4", "4", save(t, archive(doc.PDFA4)))
	add("pdfa-4f", "4f", save(t, archive(doc.PDFA4F).AttachFile(attachment)))
	add("pdfua-1", "ua1", save(t, tagged(t, doc.PDFUA1)))
	add("pdfua-2", "ua2", save(t, tagged(t, doc.PDFUA2)))
	add("pdfa-2b-pdfua-1", "2b ua1", save(t, tagged(t, doc.PDFUA1).EmbedStandardFonts().PDFA(doc.PDFA2B).CreationDate(created)))
	add("pdfa-4-pdfua-2", "4 ua2", save(t, tagged(t, doc.PDFUA2).EmbedStandardFonts().PDFA(doc.PDFA4).CreationDate(created)))
	return out
}

// corpusDir writes the samples to GPDF_CONFORMANCE_DIR or a temporary directory.
func corpusDir(t *testing.T) (string, []sample) {
	t.Helper()
	dir := os.Getenv("GPDF_CONFORMANCE_DIR")
	if dir == "" {
		dir = t.TempDir()
	} else if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	samples := corpus(t)
	for _, s := range samples {
		if err := os.WriteFile(filepath.Join(dir, s.name+".pdf"), s.pdf, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return dir, samples
}

// tool returns the path of an external validator, or skips (fails under
// GPDF_REQUIRE_TOOLS) when it is not installed.
func tool(t *testing.T, name string) string {
	t.Helper()
	if env := os.Getenv(strings.ToUpper(name) + "_BIN"); env != "" {
		return env
	}
	path, err := exec.LookPath(name)
	if err != nil {
		if os.Getenv("GPDF_REQUIRE_TOOLS") != "" {
			t.Fatalf("%s is required but not installed", name)
		}
		t.Skipf("%s not installed", name)
	}
	return path
}

// Every sample reads back with the reader; the corpus itself must be sound
// before the external tools look at it.
func TestCorpusReadsBack(t *testing.T) {
	_, samples := corpusDir(t)
	for _, s := range samples {
		d, err := doc.OpenReaderWithPassword(bytes.NewReader(s.pdf), int64(len(s.pdf)), s.password)
		if err != nil {
			t.Errorf("%s: open: %v", s.name, err)
			continue
		}
		if pages, err := d.Pages(); err != nil || len(pages) == 0 {
			t.Errorf("%s: pages: %d %v", s.name, len(pages), err)
		}
		if txt, err := d.ReadContent(); err != nil || txt == "" {
			t.Errorf("%s: text: %q %v", s.name, txt, err)
		}
	}
}

// qpdf --check validates the file structure (cross-reference, object streams,
// stream lengths, encryption, linearization hint tables). Exit status 0 means
// no errors and no warnings.
func TestQPDFCheck(t *testing.T) {
	qpdf := tool(t, "qpdf")
	dir, samples := corpusDir(t)
	for _, s := range samples {
		args := []string{"--check"}
		if s.password != "" {
			args = append(args, "--password="+s.password)
		}
		args = append(args, filepath.Join(dir, s.name+".pdf"))
		out, err := exec.Command(qpdf, args...).CombinedOutput()
		if err != nil {
			t.Errorf("%s: qpdf --check: %v\n%s", s.name, err, out)
			continue
		}
		if s.name == "linearized" && !bytes.Contains(out, []byte("is linearized")) {
			t.Errorf("%s: qpdf does not see the file as linearized:\n%s", s.name, out)
		}
	}
}

// veraPDF validates PDF/A and PDF/UA conformance against the named flavour.
func TestVeraPDF(t *testing.T) {
	verapdf := tool(t, "verapdf")
	dir, samples := corpusDir(t)
	checked := 0
	for _, s := range samples {
		if s.flavour == "" {
			continue
		}
		path := filepath.Join(dir, s.name+".pdf")
		for _, flavour := range strings.Fields(s.flavour) {
			checked++
			out, err := exec.Command(verapdf, "--format", "text", "--flavour", flavour, path).CombinedOutput()
			if err != nil {
				if _, isExit := err.(*exec.ExitError); !isExit {
					t.Fatalf("%s: verapdf: %v\n%s", s.name, err, out)
				}
			}
			if !veraPassed(out) {
				t.Errorf("%s: veraPDF flavour %s:\n%s", s.name, flavour, out)
			}
		}
	}
	if checked == 0 {
		t.Skip("no PDF/A or PDF/UA samples (fonts missing)")
	}
}

// veraPassed reports whether veraPDF's text output has a PASS verdict line
// (the Java runtime may print notices of its own first).
func veraPassed(out []byte) bool {
	for _, line := range bytes.Split(out, []byte("\n")) {
		if bytes.HasPrefix(line, []byte("PASS ")) {
			return true
		}
	}
	return false
}
