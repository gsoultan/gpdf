package doc_test

import (
	"bytes"
	"errors"
	"image"
	"image/color"
	"image/png"
	"strings"
	"testing"

	"github.com/gsoultan/gpdf/doc"
	"github.com/gsoultan/gpdf/font/system"
	"github.com/gsoultan/gpdf/model"
)

func tinyPNG(t *testing.T) []byte {
	t.Helper()
	img := image.NewNRGBA(image.Rect(0, 0, 4, 4))
	for i := range 16 {
		img.Set(i%4, i/4, color.NRGBA{R: 30, G: 90, B: 200, A: 255})
	}
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

// resolveDict follows a reference (or returns a direct dict).
func resolveDict(t *testing.T, d doc.Document, o model.Object) model.Dict {
	t.Helper()
	if ref, ok := o.(model.Ref); ok {
		obj, err := d.Resolve(ref)
		if err != nil {
			t.Fatalf("resolve %v: %v", ref, err)
		}
		o = obj
	}
	dict, _ := o.(model.Dict)
	return dict
}

func roles(t *testing.T, d doc.Document, kids model.Array) []string {
	t.Helper()
	var out []string
	for _, k := range kids {
		if e := resolveDict(t, d, k); e != nil {
			if s, ok := e[model.Name("S")].(model.Name); ok {
				out = append(out, string(s))
			}
		}
	}
	return out
}

func structRoot(t *testing.T, d doc.Document) model.Dict {
	t.Helper()
	cat, err := d.Catalog()
	if err != nil {
		t.Fatal(err)
	}
	ref := cat.StructTreeRootRef()
	if ref == nil {
		t.Fatal("no StructTreeRoot")
	}
	return resolveDict(t, d, *ref)
}

func uaBuilder() *doc.DocumentBuilder {
	return doc.New().Title("Accessible report").SetLanguage("en-GB").PDFUA(doc.PDFUA2).AddPage()
}

func TestPDFUA2Document(t *testing.T) {
	if _, ok := system.DefaultStandardSubstitutes()["Helvetica"]; !ok {
		t.Skip("no Helvetica substitute installed; PDF/UA needs embedded fonts")
	}
	b := uaBuilder().NoCompressContent()
	b.SetHeader(doc.PageText{Text: "Report {page}/{pages}"})
	b.FillRect(0, 50, 50, 100, 10, doc.ColorBlack)
	b.DrawHeading(0, 1, "Title", 72, 780, "Helvetica-Bold", 0)
	b.DrawParagraph(0, "First paragraph.", 72, 750, "Helvetica", 12)
	b.DrawList(0, []string{"one", "two"}, 72, 720, 14, true, "Helvetica", 12)
	b.BeginTable(0, 72, 660, 300, 40, 2).Header("A", "B").Row("1", "2").EndTable()
	b.DrawTaggedPNG(0, 72, 500, 40, 40, tinyPNG(t), "Blue square")
	b.DrawParagraph(0, "See the site.", 72, 480, "Helvetica", 12)
	b.AddLinkToURL(0, 72, 475, 150, 490, "https://example.org/")
	pdf := savePDF(t, b)

	if !bytes.HasPrefix(pdf, []byte("%PDF-2.0")) {
		t.Errorf("PDF/UA-2 must be PDF 2.0: %q", pdf[:8])
	}
	for _, want := range []string{"<pdfuaid:part>2</pdfuaid:part>", "<pdfuaid:rev>2024</pdfuaid:rev>", "/Artifact BMC"} {
		if !bytes.Contains(pdf, []byte(want)) {
			t.Errorf("output lacks %q", want)
		}
	}
	d := reopen(t, pdf)
	cat, _ := d.Catalog()
	if vp := resolveDict(t, d, cat.Dict[model.Name("ViewerPreferences")]); vp[model.Name("DisplayDocTitle")] != model.Boolean(true) {
		t.Errorf("ViewerPreferences = %v", vp)
	}
	if pages, _ := d.Pages(); len(pages) == 0 || pages[0].Dict[model.Name("Tabs")] != model.Name("S") {
		t.Errorf("page lacks /Tabs /S")
	}
	if lang, _ := cat.Dict[model.Name("Lang")].(model.String); lang.Text() != "en-GB" {
		t.Errorf("Lang = %v", cat.Dict[model.Name("Lang")])
	}
	root := structRoot(t, d)
	ns := root[model.Name("Namespaces")].(model.Array)
	if nsDict := resolveDict(t, d, ns[0]); nsDict[model.Name("NS")].(model.String).Text() != "http://iso.org/pdf2/ssn" {
		t.Errorf("namespace = %v", nsDict)
	}
	if _, ok := root[model.Name("RoleMap")]; ok {
		t.Error("standard types must not be role-mapped")
	}
	document := resolveDict(t, d, root[model.Name("K")].(model.Array)[0])
	if document[model.Name("S")] != model.Name("Document") || document[model.Name("NS")] == nil {
		t.Errorf("root element %v", document)
	}
	got := roles(t, d, document[model.Name("K")].(model.Array))
	want := []string{"H1", "P", "L", "Table", "Figure", "P", "Link"}
	if strings.Join(got, " ") != strings.Join(want, " ") {
		t.Errorf("reading order %v, want %v", got, want)
	}
	kids := document[model.Name("K")].(model.Array)
	list := resolveDict(t, d, kids[2])
	li := resolveDict(t, d, list[model.Name("K")].(model.Array)[0])
	if r := roles(t, d, li[model.Name("K")].(model.Array)); strings.Join(r, " ") != "Lbl LBody" {
		t.Errorf("LI kids %v", r)
	}
	link := resolveDict(t, d, kids[6])
	objr := link[model.Name("K")].(model.Array)[0].(model.Dict)
	if objr[model.Name("Type")] != model.Name("OBJR") {
		t.Errorf("Link kid %v", objr)
	}
	annot := resolveDict(t, d, objr[model.Name("Obj")])
	if _, ok := annot[model.Name("StructParent")].(model.Integer); !ok {
		t.Errorf("annotation lacks StructParent: %v", annot)
	}
	if c, _ := annot[model.Name("Contents")].(model.String); c.Text() != "https://example.org/" {
		t.Errorf("annotation Contents %v", annot[model.Name("Contents")])
	}
	if fonts := popplerCheck(t, "pdffonts", pdf); fonts != "" && strings.Contains(fonts, "Type 1") {
		t.Errorf("unembedded font:\n%s", fonts)
	}
	if text := pdfText(t, pdf); text != "" && !strings.Contains(text, "First paragraph.") {
		t.Errorf("pdftotext:\n%s", text)
	}
}

func TestPDFUA1Header(t *testing.T) {
	if _, ok := system.DefaultStandardSubstitutes()["Helvetica"]; !ok {
		t.Skip("no Helvetica substitute installed")
	}
	b := doc.New().Title("UA-1").SetLanguage("en").PDFUA(doc.PDFUA1).AddPage()
	b.DrawParagraph(0, "Hello", 72, 700, "Helvetica", 12)
	pdf := savePDF(t, b)
	if !bytes.HasPrefix(pdf, []byte("%PDF-1.7")) {
		t.Errorf("PDF/UA-1 must be PDF 1.7: %q", pdf[:8])
	}
	if !bytes.Contains(pdf, []byte("<pdfuaid:part>1</pdfuaid:part>")) || bytes.Contains(pdf, []byte("pdfuaid:rev")) {
		t.Error("pdfuaid schema wrong for part 1")
	}
	d := reopen(t, pdf)
	if root := structRoot(t, d); root[model.Name("Namespaces")] != nil {
		t.Error("PDF/UA-1 structure must not use namespaces")
	}
}

func TestPDFUAChecks(t *testing.T) {
	if _, ok := system.DefaultStandardSubstitutes()["Helvetica"]; !ok {
		t.Skip("no Helvetica substitute installed")
	}
	cases := map[string]func() *doc.DocumentBuilder{
		"missing title": func() *doc.DocumentBuilder {
			return doc.New().SetLanguage("en").PDFUA(doc.PDFUA2).AddPage()
		},
		"missing language": func() *doc.DocumentBuilder {
			return doc.New().Title("T").PDFUA(doc.PDFUA2).AddPage()
		},
		"untagged text": func() *doc.DocumentBuilder {
			b := uaBuilder()
			return b.DrawText("loose", 72, 700, "Helvetica", 12)
		},
		"figure without alt": func() *doc.DocumentBuilder {
			b := uaBuilder()
			return b.DrawTaggedPNG(0, 72, 500, 40, 40, tinyPNG(t), "")
		},
		"heading skip": func() *doc.DocumentBuilder {
			b := uaBuilder()
			b.DrawHeading(0, 1, "A", 72, 780, "Helvetica", 0)
			return b.DrawHeading(0, 3, "B", 72, 740, "Helvetica", 0)
		},
		"UA-2 with PDF/A-2": func() *doc.DocumentBuilder {
			return uaBuilder().PDFA(doc.PDFA2B)
		},
	}
	for name, mk := range cases {
		_, err := mk().Build()
		if !errors.Is(err, doc.ErrPDFUA) {
			t.Errorf("%s: err = %v, want ErrPDFUA", name, err)
		}
	}
	// A heading sequence that only descends one level at a time is fine.
	b := uaBuilder()
	b.DrawHeading(0, 1, "A", 72, 780, "Helvetica", 0).DrawHeading(0, 2, "B", 72, 740, "Helvetica", 0).DrawHeading(0, 2, "C", 72, 700, "Helvetica", 0)
	if _, err := b.Build(); err != nil {
		t.Errorf("valid headings: %v", err)
	}
}

func TestTaggedReadingOrderWithoutUA(t *testing.T) {
	b := doc.New().SetTagged().AddPage()
	b.DrawParagraph(0, "one", 72, 700, "Helvetica", 12)
	b.DrawList(0, []string{"a"}, 72, 680, 14, false, "Helvetica", 12)
	b.BeginSection().DrawParagraph(0, "two", 72, 640, "Helvetica", 12).EndSection()
	b.DrawParagraph(0, "three", 72, 620, "Helvetica", 12)
	pdf := savePDF(t, b)
	d := reopen(t, pdf)
	document := resolveDict(t, d, structRoot(t, d)[model.Name("K")].(model.Array)[0])
	got := roles(t, d, document[model.Name("K")].(model.Array))
	if strings.Join(got, " ") != "P L Sect P" {
		t.Errorf("reading order %v", got)
	}
}
