// Copyright (c) 2026 Gembit Soultan Shirazi <gembit.soultan@gmail.com>

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
	"github.com/gsoultan/gpdf/font/truetype"
	"github.com/gsoultan/gpdf/model"
)

func loadTestFont(t *testing.T) *truetype.Font {
	t.Helper()
	for _, p := range []string{
		"/usr/share/fonts/truetype/dejavu/DejaVuSans.ttf",
		"/usr/share/fonts/TTF/DejaVuSans.ttf",
		"/Library/Fonts/Arial.ttf",
		"C:\\Windows\\Fonts\\arial.ttf",
	} {
		data, err := os.ReadFile(p)
		if err != nil {
			continue
		}
		f, err := truetype.Parse(data)
		if err != nil {
			t.Fatalf("parse %s: %v", p, err)
		}
		return f
	}
	t.Skip("no TrueType font found on this machine")
	return nil
}

func TestPDFAOutput(t *testing.T) {
	f := loadTestFont(t)
	created := time.Date(2026, 10, 31, 12, 0, 0, 0, time.UTC)
	b := doc.New().PDFA(doc.PDFA2B).Title("Archive <test>").Author("gPDF").CreationDate(created).
		RegisterFont(f).PageSize(300, 200).AddPage()
	b.DrawText("Hello PDF/A", 20, 100, f.PostScriptName(), 14)
	b.AddLinkToURL(0, 20, 90, 100, 110, "https://example.com")
	built, err := b.Build()
	if err != nil {
		t.Fatal(err)
	}
	var buf bytes.Buffer
	if err := built.Save(&buf); err != nil {
		t.Fatal(err)
	}
	pdf := buf.Bytes()
	if !bytes.HasPrefix(pdf, []byte("%PDF-1.7\n")) {
		t.Errorf("header = %q, want PDF 1.7", pdf[:9])
	}
	if err := built.SaveWithPassword(&bytes.Buffer{}, "u", "o"); !errors.Is(err, doc.ErrPDFAEncryption) {
		t.Errorf("SaveWithPassword err = %v", err)
	}

	d, err := doc.OpenReaderWithPassword(bytes.NewReader(pdf), int64(len(pdf)), "")
	if err != nil {
		t.Fatal(err)
	}
	r := d.(resolver)
	xmp, err := d.MetadataStream()
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"<pdfaid:part>2</pdfaid:part>", "<pdfaid:conformance>B</pdfaid:conformance>",
		"Archive &lt;test&gt;", "<xmp:CreateDate>2026-10-31T12:00:00Z</xmp:CreateDate>"} {
		if !strings.Contains(string(xmp), want) {
			t.Errorf("XMP lacks %q:\n%s", want, xmp)
		}
	}
	info, err := d.Info()
	if err != nil {
		t.Fatal(err)
	}
	if cd, _ := info[model.Name("CreationDate")].(model.String); cd != "D:20261031120000+00'00'" {
		t.Errorf("CreationDate = %q", cd)
	}

	cat, err := d.Catalog()
	if err != nil {
		t.Fatal(err)
	}
	intents, _ := cat.Dict[model.Name("OutputIntents")].(model.Array)
	if len(intents) != 1 {
		t.Fatalf("OutputIntents = %v", cat.Dict[model.Name("OutputIntents")])
	}
	intentObj, _ := r.Resolve(intents[0].(model.Ref))
	intent := intentObj.(model.Dict)
	if s, _ := intent[model.Name("S")].(model.Name); s != "GTS_PDFA1" {
		t.Errorf("intent /S = %v", intent[model.Name("S")])
	}
	profObj, _ := r.Resolve(intent[model.Name("DestOutputProfile")].(model.Ref))
	prof := profObj.(*model.Stream)
	if n, _ := prof.Dict[model.Name("N")].(model.Integer); n != 3 {
		t.Errorf("profile /N = %v", prof.Dict[model.Name("N")])
	}
	icc := prof.Content
	if len(icc) < 132 || string(icc[36:40]) != "acsp" || string(icc[16:20]) != "RGB " || string(icc[12:16]) != "mntr" {
		t.Fatalf("output profile is not an RGB monitor ICC profile (%d bytes)", len(icc))
	}
	if size := int(icc[0])<<24 | int(icc[1])<<16 | int(icc[2])<<8 | int(icc[3]); size != len(icc) {
		t.Errorf("ICC size field %d, want %d", size, len(icc))
	}

	// Fonts are embedded with a CIDSet; link annotations carry the Print flag.
	pages, _ := d.Pages()
	resObj := pages[0].Dict[model.Name("Resources")]
	if ref, ok := resObj.(model.Ref); ok {
		resObj, _ = r.Resolve(ref)
	}
	fonts := resObj.(model.Dict)[model.Name("Font")].(model.Dict)
	for _, fo := range fonts {
		fd, _ := r.Resolve(fo.(model.Ref))
		type0 := fd.(model.Dict)
		desc, _ := r.Resolve(type0[model.Name("DescendantFonts")].(model.Array)[0].(model.Ref))
		fdesc, _ := r.Resolve(desc.(model.Dict)[model.Name("FontDescriptor")].(model.Ref))
		fdd := fdesc.(model.Dict)
		if _, ok := fdd[model.Name("FontFile2")]; !ok {
			t.Errorf("font %v has no FontFile2", type0[model.Name("BaseFont")])
		}
		if _, ok := fdd[model.Name("CIDSet")]; !ok {
			t.Errorf("font %v has no CIDSet", type0[model.Name("BaseFont")])
		}
	}
	annots, _ := pages[0].Dict[model.Name("Annots")].(model.Array)
	if len(annots) != 1 {
		t.Fatalf("Annots = %v", pages[0].Dict[model.Name("Annots")])
	}
	a, _ := r.Resolve(annots[0].(model.Ref))
	if fl, _ := a.(model.Dict)[model.Name("F")].(model.Integer); fl&4 == 0 {
		t.Errorf("link /F = %v, want Print bit", a.(model.Dict)[model.Name("F")])
	}

	if text, _ := d.ReadContent(); !strings.Contains(text, "Hello PDF/A") {
		t.Errorf("text = %q", text)
	}
	if bin, err := exec.LookPath("pdfinfo"); err == nil {
		path := filepath.Join(t.TempDir(), "a.pdf")
		os.WriteFile(path, pdf, 0o600)
		out, err := exec.Command(bin, path).CombinedOutput()
		if err != nil || strings.Contains(string(out), "Error") {
			t.Errorf("pdfinfo: %v\n%s", err, out)
		}
	}
}

func TestPDFARejectsStandardFontsAndCMYK(t *testing.T) {
	b := doc.New().PDFA(doc.PDFA3B).PageSize(100, 100).AddPage()
	b.DrawText("x", 10, 10, "Helvetica", 12)
	if _, err := b.Build(); !errors.Is(err, doc.ErrPDFA) || !strings.Contains(err.Error(), "Helvetica") {
		t.Fatalf("err = %v", err)
	}
	f := loadTestFont(t)
	b = doc.New().PDFA(doc.PDFA2B).RegisterFont(f).PageSize(100, 100).AddPage()
	b.DrawImage(0, 0, 10, 10, []byte{0, 0, 0, 0}, 1, 1, 8, "DeviceCMYK")
	if _, err := b.Build(); !errors.Is(err, doc.ErrPDFA) {
		t.Fatalf("err = %v", err)
	}
	if _, err := doc.New().PDFA("1A").Build(); !errors.Is(err, doc.ErrPDFA) {
		t.Fatalf("err = %v", err)
	}
}
