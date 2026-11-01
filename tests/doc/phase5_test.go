// Copyright (c) 2026 Gembit Soultan Shirazi <gembit.soultan@gmail.com>

package doc_test

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"errors"
	"math/big"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/gsoultan/gpdf/doc"
	"github.com/gsoultan/gpdf/font/system"
	"github.com/gsoultan/gpdf/model"
	"github.com/gsoultan/gpdf/sign"
)

// A tagged document reads in structure order, whatever the drawing order.
func TestReadingOrderTagged(t *testing.T) {
	if _, ok := system.DefaultStandardSubstitutes()["Helvetica"]; !ok {
		t.Skip("no Helvetica substitute installed")
	}
	b := uaBuilder()
	b.SetFooter(doc.PageText{Text: "footer {page}"})
	// Drawn right column first, then left: the structure order is what counts.
	b.DrawHeading(0, 1, "Title", 72, 780, "Helvetica-Bold", 16)
	b.DrawParagraph(0, "Right column text.", 320, 740, "Helvetica", 12)
	b.DrawParagraph(0, "Left column text.", 72, 740, "Helvetica", 12)
	b.DrawList(0, []string{"one", "two"}, 72, 700, 14, true, "Helvetica", 12)
	b.BeginTable(0, 72, 640, 300, 40, 2).Header("A", "B").Row("1", "2").EndTable()
	b.DrawTaggedPNG(0, 72, 500, 40, 40, tinyPNG(t), "Blue square")
	pdf := savePDF(t, b)
	d := reopen(t, pdf)

	elems, err := d.ReadStructuredText()
	if err != nil {
		t.Fatal(err)
	}
	var roles []string
	for _, e := range elems {
		roles = append(roles, e.Role)
	}
	want := "H1 P P Lbl LBody Lbl LBody TH TH TD TD Figure"
	if got := strings.Join(roles, " "); got != want {
		t.Errorf("roles = %q, want %q\n%+v", got, want, elems)
	}
	text, err := d.ReadContentInReadingOrder()
	if err != nil {
		t.Fatal(err)
	}
	wantText := "Title\nRight column text.\nLeft column text.\n1. one\n2. two\nA\tB\n1\t2\nBlue square"
	if text != wantText {
		t.Errorf("reading order text:\n%q\nwant\n%q", text, wantText)
	}
	if strings.Contains(text, "footer") {
		t.Error("artifact text (footer) must not be in the reading order")
	}
	// The in-memory document gives the same answer.
	built, err := b.Build()
	if err != nil {
		t.Fatal(err)
	}
	if got, _ := built.ReadContentInReadingOrder(); got != wantText {
		t.Errorf("built document reading order:\n%q", got)
	}
}

// An untagged two-column page is read column by column, with the title first,
// even though the right column was drawn before the left one.
func TestReadingOrderUntaggedColumns(t *testing.T) {
	b := doc.New().PageSize(500, 400).AddPage()
	b.DrawText("Two column title", 40, 370, "Helvetica", 16)
	for i, line := range []string{"right one", "right two", "right three"} {
		b.DrawText(line, 270, 330-float64(i)*14, "Helvetica", 11)
	}
	for i, line := range []string{"left one", "left two", "left three"} {
		b.DrawText(line, 40, 330-float64(i)*14, "Helvetica", 11)
	}
	// A second paragraph in the left column after a blank line.
	b.DrawText("left para two", 40, 260, "Helvetica", 11)
	pdf := savePDF(t, b)
	text, err := reopen(t, pdf).ReadContentInReadingOrder()
	if err != nil {
		t.Fatal(err)
	}
	want := "Two column title\nleft one left two left three\nleft para two\nright one right two right three"
	if text != want {
		t.Errorf("reading order:\n%q\nwant\n%q", text, want)
	}
	// Plain extraction interleaves the columns line by line.
	plain, _ := reopen(t, pdf).ReadContent()
	if strings.Index(plain, "right one") > strings.Index(plain, "left two") {
		t.Errorf("plain extraction should interleave: %q", plain)
	}
}

func selfSignedCert(t *testing.T, key *rsa.PrivateKey) *x509.Certificate {
	t.Helper()
	tmpl := &x509.Certificate{
		SerialNumber:          big.NewInt(1),
		Subject:               pkix.Name{CommonName: "Updater"},
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
	return cert
}

// SaveIncremental appends only the changed objects after the original bytes, so
// an earlier signature stays intact.
func TestSaveIncremental(t *testing.T) {
	b := doc.New().Title("Original").PageSize(300, 200).AddPage()
	b.DrawText("Hello base", 20, 100, "Helvetica", 12)
	base := savePDF(t, b)

	key, _ := rsa.GenerateKey(rand.Reader, 2048)
	cert := selfSignedCert(t, key)
	var signed bytes.Buffer
	if err := sign.Sign(bytes.NewReader(base), int64(len(base)), &signed, sign.NewCMSSigner(key, cert), sign.SignOptions{Reason: "Approved"}); err != nil {
		t.Fatal(err)
	}
	src := signed.Bytes()

	d := reopen(t, src)
	if err := d.Replace("Hello base", "Hello update"); err != nil {
		t.Fatal(err)
	}
	if err := d.SetInfo("Title", model.String("Updated")); err != nil {
		t.Fatal(err)
	}
	noteRef := d.NewObject(model.Dict{model.Name("Note"): model.String("added")})
	var out bytes.Buffer
	if err := d.SaveIncremental(&out); err != nil {
		t.Fatal(err)
	}
	updated := out.Bytes()
	if !bytes.HasPrefix(updated, src) {
		t.Fatal("original bytes were changed")
	}
	tail := updated[len(src):]
	if !bytes.Contains(tail, []byte("/Prev")) || !bytes.Contains(tail, []byte("(Updated)")) {
		t.Errorf("update section:\n%s", tail)
	}
	if n := bytes.Count(tail, []byte(" obj")); n != 3 {
		t.Errorf("update holds %d objects, want 3 (content, info, note)", n)
	}

	d2 := reopen(t, updated)
	if txt, _ := d2.ReadContent(); !strings.Contains(txt, "Hello update") || strings.Contains(txt, "Hello base") {
		t.Errorf("text = %q", txt)
	}
	if info, _ := d2.Info(); info[model.Name("Title")].(model.String).Text() != "Updated" {
		t.Errorf("info = %v", info)
	}
	if note, err := d2.Resolve(noteRef); err != nil || note.(model.Dict)[model.Name("Note")] != model.String("added") {
		t.Errorf("added object: %v %v", note, err)
	}
	if txt := pdfText(t, updated); txt != "" && !strings.Contains(txt, "Hello update") {
		t.Errorf("pdftotext = %q", txt)
	}
	infos, err := sign.Verify(bytes.NewReader(updated), int64(len(updated)), sign.VerifyOptions{})
	if err != nil || len(infos) != 1 {
		t.Fatalf("verify: %v %d", err, len(infos))
	}
	if !infos[0].IntegrityOK || infos[0].CoversWholeDocument {
		t.Errorf("signature after update: integrity=%v covers=%v", infos[0].IntegrityOK, infos[0].CoversWholeDocument)
	}

	// A full save of the updated document still carries everything.
	var full bytes.Buffer
	if err := d.Save(&full); err != nil {
		t.Fatal(err)
	}
	d3 := reopen(t, full.Bytes())
	if txt, _ := d3.ReadContent(); !strings.Contains(txt, "Hello update") {
		t.Errorf("full save text = %q", txt)
	}
	if _, err := d3.Resolve(noteRef); err != nil {
		t.Errorf("added object missing after full save: %v", err)
	}

	// Nothing changed: the file is copied as it is.
	var same bytes.Buffer
	if err := reopen(t, base).SaveIncremental(&same); err != nil || !bytes.Equal(same.Bytes(), base) {
		t.Errorf("unchanged document: %v, %d bytes vs %d", err, same.Len(), len(base))
	}
	built, _ := doc.New().AddPage().Build()
	if err := built.SaveIncremental(&same); !errors.Is(err, doc.ErrNoBaseFile) {
		t.Errorf("built document: %v", err)
	}
}

// Redact removes the matched text from the content and draws a box over it;
// the rest keeps its place and the removed words are gone from the output.
func TestRedact(t *testing.T) {
	b := doc.New().PageSize(400, 300).NoCompressContent().AddPage()
	b.DrawText("Account 12345 belongs to Alice", 20, 200, "Helvetica", 12)
	b.DrawText("Public line", 20, 180, "Helvetica", 12)
	b.DrawText("Phone 555-0100", 20, 160, "Helvetica", 12)
	b.DrawText("Bottom secret area", 20, 40, "Helvetica", 12)
	d := reopen(t, savePDF(t, b))

	before, _ := d.ReadLayout()
	var publicX float64
	for _, blk := range before[0].Blocks {
		if blk.Text == "Public line" {
			publicX = blk.X
		}
	}

	report, err := doc.Redact(d, doc.RedactOptions{
		Texts:   []string{"Alice"},
		Pattern: regexp.MustCompile(`\d{3}-\d{4}`),
		Areas:   []doc.RedactArea{{Page: 0, X: 0, Y: 0, Width: 400, Height: 60}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if report.Removed != 3 || report.Boxes != 3 || len(report.Pages) != 1 {
		t.Errorf("report = %+v", report)
	}
	var out bytes.Buffer
	if err := d.Save(&out); err != nil {
		t.Fatal(err)
	}
	pdf := out.Bytes()
	for _, gone := range []string{"Alice", "12345", "555-0100", "Bottom secret"} {
		if bytes.Contains(pdf, []byte(gone)) {
			t.Errorf("%q is still in the file", gone)
		}
	}
	d2 := reopen(t, pdf)
	txt, _ := d2.ReadContent()
	if !strings.Contains(txt, "Public line") || strings.Contains(txt, "Alice") || strings.Contains(txt, "Phone") {
		t.Errorf("text after redaction = %q", txt)
	}
	if s := pdfText(t, pdf); s != "" && (strings.Contains(s, "Alice") || !strings.Contains(s, "Public line")) {
		t.Errorf("pdftotext = %q", s)
	}
	after, _ := d2.ReadLayout()
	for _, blk := range after[0].Blocks {
		if blk.Text == "Public line" && blk.X != publicX {
			t.Errorf("remaining text moved from %g to %g", publicX, blk.X)
		}
	}
	content := firstPageContent(t, d2)
	if n := len(regexp.MustCompile(`re\s+f\b`).FindAllString(content, -1)); n != 3 {
		t.Errorf("%d boxes drawn, want 3:\n%s", n, content)
	}
	if err := d.SaveIncremental(&out); !errors.Is(err, doc.ErrRedacted) && err == nil {
		t.Error("incremental save after redaction must be refused")
	}
	if _, err := doc.Redact(reopen(t, pdf), doc.RedactOptions{Texts: []string{"nowhere"}}); err == nil {
		t.Error("no match should be an error")
	}
}
