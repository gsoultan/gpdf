package doc_test

import (
	"bytes"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/gsoultan/gpdf/doc"
	"github.com/gsoultan/gpdf/internal/pdftest"
	"github.com/gsoultan/gpdf/model"
	"github.com/gsoultan/gpdf/reader"
)

func savePDF(t *testing.T, b *doc.DocumentBuilder) []byte {
	t.Helper()
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

func reopen(t *testing.T, pdf []byte) doc.Document {
	t.Helper()
	d, err := doc.OpenReader(bytes.NewReader(pdf), int64(len(pdf)))
	if err != nil {
		t.Fatal(err)
	}
	return d
}

func popplerCheck(t *testing.T, tool string, pdf []byte, args ...string) string {
	t.Helper()
	bin, err := exec.LookPath(tool)
	if err != nil {
		return ""
	}
	path := filepath.Join(t.TempDir(), "x.pdf")
	os.WriteFile(path, pdf, 0o600)
	out, err := exec.Command(bin, append(args, path)...).CombinedOutput()
	if err != nil || strings.Contains(string(out), "Error") || strings.Contains(string(out), "Syntax") {
		t.Errorf("%s: %v\n%s", tool, err, out)
	}
	return string(out)
}

func TestPDF20TextStringsAndDates(t *testing.T) {
	created := time.Date(2026, 11, 1, 12, 0, 0, 0, time.UTC)
	title := "Grüße 日本語 – test"
	b := doc.New().Title(title).Author("Ünal").CreationDate(created).PageSize(200, 100).AddPage()
	b.AddOutline("Kapitel Ü", 0)
	pdf := savePDF(t, b)
	if !bytes.HasPrefix(pdf, []byte("%PDF-2.0\n")) {
		t.Errorf("header = %q", pdf[:9])
	}
	d := reopen(t, pdf)
	info, _ := d.Info()
	raw := info[model.Name("Title")].(model.String)
	if !strings.HasPrefix(string(raw), "\xEF\xBB\xBF") {
		t.Errorf("Title is not a UTF-8 text string: %q", raw)
	}
	if raw.Text() != title {
		t.Errorf("Title = %q", raw.Text())
	}
	if cd, _ := info[model.Name("CreationDate")].(model.String); cd != "D:20261101120000Z" {
		t.Errorf("CreationDate = %q, want PDF 2.0 form", cd)
	}
	if tm, err := model.ParseDate(string(info[model.Name("CreationDate")].(model.String))); err != nil || !tm.Equal(created) {
		t.Errorf("ParseDate: %v %v", tm, err)
	}
	// PDF 2.0 deprecates the Info entries, so the XMP packet carries them too.
	xmp, err := d.MetadataStream()
	if err != nil || !strings.Contains(string(xmp), "Grüße 日本語") || !strings.Contains(string(xmp), "<dc:creator><rdf:Seq><rdf:li>Ünal") {
		t.Errorf("XMP: %v\n%s", err, xmp)
	}
	if out := popplerCheck(t, "pdfinfo", pdf); out != "" && !strings.Contains(out, title) {
		t.Errorf("pdfinfo does not show the UTF-8 title:\n%s", out)
	}
	cat, _ := d.Catalog()
	outlines, _ := d.Resolve(*cat.OutlinesRef())
	first, _ := d.Resolve(outlines.(model.Dict)[model.Name("First")].(model.Ref))
	if got := (&model.OutlineItem{Dict: first.(model.Dict)}).Title(); got != "Kapitel Ü" {
		t.Errorf("outline title = %q", got)
	}
}

func TestPDFA2KeepsPDF1xDatesAndUTF16(t *testing.T) {
	f := loadTestFont(t)
	created := time.Date(2026, 11, 1, 12, 0, 0, 0, time.FixedZone("", 5*3600+1800))
	b := doc.New().PDFA(doc.PDFA2B).Title("Résumé").CreationDate(created).RegisterFont(f).PageSize(200, 100).AddPage()
	b.DrawText("x", 10, 10, f.PostScriptName(), 10)
	pdf := savePDF(t, b)
	d := reopen(t, pdf)
	info, _ := d.Info()
	if cd, _ := info[model.Name("CreationDate")].(model.String); cd != "D:20261101120000+05'30'" {
		t.Errorf("CreationDate = %q", cd)
	}
	raw := info[model.Name("Title")].(model.String)
	if !strings.HasPrefix(string(raw), "\xFE\xFF") || raw.Text() != "Résumé" {
		t.Errorf("Title = %q", raw)
	}
}

func TestAttachments(t *testing.T) {
	mod := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)
	b := doc.New().PageSize(200, 100).AddPage()
	b.AttachFile(doc.Attachment{Name: "data.xml", MIME: "application/xml", Description: "Rechnung",
		Relationship: doc.RelationshipAlternative, Modified: mod, Data: []byte("<invoice/>")})
	b.CurrentPage().AttachFile(doc.Attachment{Name: "page-ü.csv", MIME: "text/csv", Data: []byte("a,b\n")})
	pdf := savePDF(t, b)
	d := reopen(t, pdf)
	files, err := d.Attachments()
	if err != nil {
		t.Fatal(err)
	}
	if len(files) != 2 {
		t.Fatalf("attachments = %+v", files)
	}
	byName := map[string]reader.EmbeddedFile{}
	for _, f := range files {
		byName[f.Name] = f
	}
	x := byName["data.xml"]
	if string(x.Data) != "<invoice/>" || x.MIME != "application/xml" || x.Description != "Rechnung" ||
		x.Relationship != "Alternative" || !x.Modified.Equal(mod) {
		t.Errorf("data.xml = %+v", x)
	}
	if c := byName["page-ü.csv"]; string(c.Data) != "a,b\n" || c.Relationship != "Unspecified" {
		t.Errorf("csv = %+v", c)
	}
	cat, _ := d.Catalog()
	if af := cat.AssociatedFilesArray(); len(af) != 1 {
		t.Errorf("catalog /AF = %v", af)
	}
	pages, _ := d.Pages()
	if af := pages[0].AssociatedFiles(); len(af) != 1 {
		t.Errorf("page /AF = %v", af)
	}
	if out := popplerCheck(t, "pdfdetach", pdf, "-list"); out != "" && (!strings.Contains(out, "data.xml") || !strings.Contains(out, "page-ü.csv")) {
		t.Errorf("pdfdetach -list:\n%s", out)
	}

	// Level checks.
	f := loadTestFont(t)
	a2 := doc.New().PDFA(doc.PDFA2B).RegisterFont(f).PageSize(100, 100).AddPage()
	a2.AttachFile(doc.Attachment{Name: "x", Data: []byte("x")})
	if _, err := a2.Build(); !errors.Is(err, doc.ErrPDFA) {
		t.Errorf("PDF/A-2 with attachment: err = %v", err)
	}
	a4f := doc.New().PDFA(doc.PDFA4F).RegisterFont(f).PageSize(100, 100).AddPage()
	if _, err := a4f.Build(); !errors.Is(err, doc.ErrPDFA) {
		t.Errorf("PDF/A-4f without attachment: err = %v", err)
	}
}

func TestPDFA4(t *testing.T) {
	f := loadTestFont(t)
	created := time.Date(2026, 11, 1, 12, 0, 0, 0, time.UTC)
	b := doc.New().PDFA(doc.PDFA4F).Title("Archive 4").Author("A").CreationDate(created).RegisterFont(f).PageSize(200, 100).AddPage()
	b.DrawText("PDF/A-4", 10, 50, f.PostScriptName(), 12)
	b.AttachFile(doc.Attachment{Name: "payload.json", MIME: "application/json", Relationship: doc.RelationshipData, Data: []byte("{}")})
	pdf := savePDF(t, b)
	if !bytes.HasPrefix(pdf, []byte("%PDF-2.0\n")) {
		t.Errorf("header = %q, want 2.0", pdf[:9])
	}
	d := reopen(t, pdf)
	xmp, _ := d.MetadataStream()
	for _, want := range []string{"<pdfaid:part>4</pdfaid:part>", "<pdfaid:rev>2020</pdfaid:rev>", "<pdfaid:conformance>F</pdfaid:conformance>", "<dc:title><rdf:Alt><rdf:li xml:lang=\"x-default\">Archive 4"} {
		if !strings.Contains(string(xmp), want) {
			t.Errorf("XMP lacks %q", want)
		}
	}
	plain := doc.New().PDFA(doc.PDFA4).RegisterFont(f).PageSize(100, 100).AddPage()
	plain.DrawText("x", 10, 10, f.PostScriptName(), 10)
	pd := reopen(t, savePDF(t, plain))
	if x, _ := pd.MetadataStream(); strings.Contains(string(x), "pdfaid:conformance") {
		t.Error("plain PDF/A-4 has no conformance level")
	}
	info, _ := d.Info()
	if _, ok := info[model.Name("Title")]; ok {
		t.Errorf("PDF/A-4 Info must only carry the dates: %v", info)
	}
	if cd, _ := info[model.Name("CreationDate")].(model.String); cd != "D:20261101120000Z" {
		t.Errorf("CreationDate = %q", cd)
	}
	// CIDSet is deprecated in PDF 2.0 and must not be written for PDF/A-4.
	pages, _ := d.Pages()
	res := pages[0].Dict[model.Name("Resources")]
	if ref, ok := res.(model.Ref); ok {
		res, _ = d.Resolve(ref)
	}
	for _, fo := range res.(model.Dict)[model.Name("Font")].(model.Dict) {
		fd, _ := d.Resolve(fo.(model.Ref))
		desc, _ := d.Resolve(fd.(model.Dict)[model.Name("DescendantFonts")].(model.Array)[0].(model.Ref))
		fdesc, _ := d.Resolve(desc.(model.Dict)[model.Name("FontDescriptor")].(model.Ref))
		if _, ok := fdesc.(model.Dict)[model.Name("CIDSet")]; ok {
			t.Error("CIDSet written for PDF/A-4")
		}
	}
	if text, _ := d.ReadContent(); !strings.Contains(text, "PDF/A-4") {
		t.Errorf("text = %q", text)
	}
	popplerCheck(t, "pdfinfo", pdf)
}

func TestFormsHaveAppearanceStreams(t *testing.T) {
	b := doc.New().PageSize(300, 200).AddPage()
	b.AddTextField(0, 20, 150, 200, 170, "name", "Zoë", "Your name", true)
	b.AddCheckBox(0, 20, 120, 34, 134, "agree", true, "", false)
	b.AddRadioButton(0, 20, 90, 34, 104, "colour", "red", true, "")
	b.AddRadioButton(0, 50, 90, 64, 104, "colour", "blue", false, "")
	pdf := savePDF(t, b)
	d := reopen(t, pdf)
	cat, _ := d.Catalog()
	acro, _ := d.Resolve(*cat.AcroFormRef())
	ad := acro.(model.Dict)
	if _, ok := ad[model.Name("NeedAppearances")]; ok {
		t.Error("NeedAppearances is deprecated in PDF 2.0")
	}
	if _, ok := ad[model.Name("DR")]; !ok {
		t.Error("AcroForm has no /DR")
	}
	pages, _ := d.Pages()
	annots := pages[0].Dict[model.Name("Annots")].(model.Array)
	if len(annots) != 4 {
		t.Fatalf("annots = %v", annots)
	}
	for _, a := range annots {
		w, _ := d.Resolve(a.(model.Ref))
		wd := w.(model.Dict)
		ap, ok := wd[model.Name("AP")].(model.Dict)
		if !ok {
			t.Errorf("widget without /AP: %v", wd)
			continue
		}
		n := ap[model.Name("N")]
		switch v := n.(type) {
		case model.Ref:
			if s, _ := d.Resolve(v); s == nil {
				t.Errorf("AP /N %v missing", v)
			}
		case model.Dict:
			if len(v) != 2 {
				t.Errorf("AP /N states = %v", v)
			}
		default:
			t.Errorf("AP /N = %T", n)
		}
	}
	popplerCheck(t, "pdftoppm", pdf, "-r", "20", "-png", "-singlefile", "-o")
}

func TestDocumentPartsAndPageOutputIntent(t *testing.T) {
	b := doc.New().PageSize(100, 100)
	for range 5 {
		b.AddPage()
	}
	b.DocumentPart(1, 2, map[string]string{"Recipient": "Müller"}).DocumentPart(4, 4, nil)
	prof := doc.SRGBProfile()
	prof.Data = []byte("not a real profile, but data")
	b.CurrentPage().OutputIntent(prof, "FOGRA39")
	b.FillRectWithState(0, 10, 10, 50, 50, doc.Color{R: 1}, doc.GraphicsState{BlackPointCompensation: doc.BlackPointOn})
	pdf := savePDF(t, b)
	d := reopen(t, pdf)
	cat, _ := d.Catalog()
	rootObj, err := d.Resolve(*cat.DPartRootRef())
	if err != nil {
		t.Fatal(err)
	}
	rootNode, _ := d.Resolve(rootObj.(model.Dict)[model.Name("DPartRootNode")].(model.Ref))
	parts := rootNode.(model.Dict)[model.Name("DParts")].(model.Array)[0].(model.Array)
	if len(parts) != 4 { // [0], [1-2], [3], [4]
		t.Fatalf("parts = %v", parts)
	}
	pages, _ := d.Pages()
	for i, p := range pages {
		if p.DPartRef() == nil {
			t.Errorf("page %d has no /DPart", i)
		}
	}
	second, _ := d.Resolve(parts[1].(model.Ref))
	dpm := second.(model.Dict)[model.Name("DPM")].(model.Dict)
	if dpm[model.Name("Recipient")].(model.String).Text() != "Müller" {
		t.Errorf("DPM = %v", dpm)
	}
	if end := second.(model.Dict)[model.Name("End")].(model.Ref); end != pages[2].Ref {
		t.Errorf("part 2 /End = %v, want page 3 (%v)", end, pages[2].Ref)
	}
	if oi := pages[4].OutputIntents(); len(oi) != 1 {
		t.Errorf("page 5 OutputIntents = %v", oi)
	} else {
		intent, _ := d.Resolve(oi[0].(model.Ref))
		if id := intent.(model.Dict)[model.Name("OutputConditionIdentifier")].(model.String); id != "FOGRA39" {
			t.Errorf("intent = %v", intent)
		}
	}
	res := pages[0].Dict[model.Name("Resources")]
	if ref, ok := res.(model.Ref); ok {
		res, _ = d.Resolve(ref)
	}
	gs := res.(model.Dict)[model.Name("ExtGState")].(model.Dict)
	found := false
	for _, g := range gs {
		if ref, ok := g.(model.Ref); ok {
			g, _ = d.Resolve(ref)
		}
		if v, _ := g.(model.Dict)[model.Name("UseBlackPtComp")].(model.Name); v == "ON" {
			found = true
		}
	}
	if !found {
		t.Errorf("no ExtGState with /UseBlackPtComp /ON: %v", gs)
	}
	if _, err := doc.New().AddPage().AddPage().DocumentPart(0, 1, nil).DocumentPart(1, 1, nil).Build(); err == nil {
		t.Error("overlapping parts accepted")
	}
	popplerCheck(t, "pdfinfo", pdf)
}

func TestEffectiveVersionFromCatalog(t *testing.T) {
	b := pdftest.New("1.4")
	b.Obj(1, 0, "<< /Type /Catalog /Pages 2 0 R /Version /1.7 >>")
	b.Obj(2, 0, "<< /Type /Pages /Kids [3 0 R] /Count 1 >>")
	b.Obj(3, 0, "<< /Type /Page /Parent 2 0 R /MediaBox [0 0 200 200] >>")
	pdf := b.XRefTable("<< /Root 1 0 R >>")
	rd, err := reader.NewPDFReader().ReadDocument(bytes.NewReader(pdf), int64(len(pdf)))
	if err != nil {
		t.Fatal(err)
	}
	if v := rd.PDFVersion(); v.String() != "1.4" {
		t.Errorf("header version = %s", v)
	}
	if v := rd.EffectiveVersion(); v.String() != "1.7" {
		t.Errorf("effective version = %s", v)
	}
}

func TestAES256PasswordIsSASLprepped(t *testing.T) {
	b := doc.New().PageSize(100, 100).AddPage()
	b.DrawText("secret", 10, 50, "Helvetica", 12)
	built, err := b.Build()
	if err != nil {
		t.Fatal(err)
	}
	var buf bytes.Buffer
	// A soft hyphen is "commonly mapped to nothing" (RFC 4013 §2.1 example).
	if err := built.SaveWithAES256Password(&buf, "pa­ss", "owner"); err != nil {
		t.Fatal(err)
	}
	pdf := buf.Bytes()
	d, err := doc.OpenReaderWithPassword(bytes.NewReader(pdf), int64(len(pdf)), "pass")
	if err != nil {
		t.Fatalf("open with prepared password: %v", err)
	}
	if text, _ := d.ReadContent(); !strings.Contains(text, "secret") {
		t.Errorf("text = %q", text)
	}
	if _, err := doc.OpenReaderWithPassword(bytes.NewReader(pdf), int64(len(pdf)), "wrong"); err == nil {
		t.Error("wrong password accepted")
	}
	if err := built.SaveWithAES256Password(&bytes.Buffer{}, "bad\x01", "o"); err == nil {
		t.Error("control character in password accepted")
	}
}
