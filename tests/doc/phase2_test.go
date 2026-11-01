package doc_test

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gsoultan/gpdf/doc"
	"github.com/gsoultan/gpdf/font/system"
	"github.com/gsoultan/gpdf/font/truetype"
)

// pdfText extracts the text of pdf with pdftotext, or returns "" when it is
// not installed.
func pdfText(t *testing.T, pdf []byte, args ...string) string {
	t.Helper()
	bin, err := exec.LookPath("pdftotext")
	if err != nil {
		return ""
	}
	path := filepath.Join(t.TempDir(), "x.pdf")
	os.WriteFile(path, pdf, 0o600)
	out, err := exec.Command(bin, append(args, path, "-")...).CombinedOutput()
	if err != nil {
		t.Errorf("pdftotext: %v\n%s", err, out)
	}
	return string(out)
}

// loadFontFile parses the first font file that exists, or skips the test.
func loadFontFile(t *testing.T, paths ...string) *truetype.Font {
	t.Helper()
	for _, p := range paths {
		data, err := os.ReadFile(p)
		if err != nil {
			continue
		}
		if strings.HasSuffix(p, ".ttc") {
			fonts, err := truetype.ParseCollection(data)
			if err != nil || len(fonts) == 0 {
				t.Fatalf("parse %s: %v", p, err)
			}
			return fonts[0]
		}
		f, err := truetype.Parse(data)
		if err != nil {
			t.Fatalf("parse %s: %v", p, err)
		}
		return f
	}
	t.Skipf("none of %v installed", paths)
	return nil
}

func TestEmbedStandardFonts(t *testing.T) {
	subs := system.DefaultStandardSubstitutes()
	if _, ok := subs["Helvetica"]; !ok {
		t.Skip("no Helvetica substitute installed")
	}
	b := doc.New().EmbedStandardFonts().PageSize(400, 200).AddPage()
	b.DrawText("Standard fonts embedded", 20, 150, "Helvetica", 14)
	b.DrawText("Bold serif substitute", 20, 110, "Times-Bold", 14)
	b.DrawText("Mono text", 20, 70, "Courier", 12)
	pdf := savePDF(t, b)

	fonts := popplerCheck(t, "pdffonts", pdf)
	if fonts == "" {
		t.Skip("pdffonts not installed")
	}
	if strings.Contains(fonts, "Type 1") {
		t.Errorf("standard fonts still referenced as Type 1:\n%s", fonts)
	}
	for _, name := range []string{subs["Helvetica"].Font.PostScriptName(), subs["Times-Bold"].Font.PostScriptName()} {
		if !strings.Contains(fonts, name) {
			t.Errorf("pdffonts lacks embedded %s:\n%s", name, fonts)
		}
	}
	text := pdfText(t, pdf)
	for _, want := range []string{"Standard fonts embedded", "Bold serif substitute", "Mono text"} {
		if text != "" && !strings.Contains(text, want) {
			t.Errorf("pdftotext lacks %q:\n%s", want, text)
		}
	}
	// Metric-compatible substitutes keep the standard widths, so a document
	// laid out with Helvetica metrics does not reflow.
	if subs["Helvetica"].MetricCompatible {
		std := doc.New().AddPage()
		emb := doc.New().EmbedStandardFonts().AddPage()
		const s = "The quick brown fox jumps over the lazy dog"
		if ws, we := std.TextWidth(s, 12, "Helvetica"), emb.TextWidth(s, 12, "Helvetica"); ws == 0 || we < ws*0.99 || we > ws*1.01 {
			t.Errorf("Helvetica width %.2f, substitute width %.2f", ws, we)
		}
	}
}

func TestEmbedStandardFontsPDFA(t *testing.T) {
	if _, ok := system.DefaultStandardSubstitutes()["Helvetica"]; !ok {
		t.Skip("no Helvetica substitute installed")
	}
	b := doc.New().EmbedStandardFonts().PDFA(doc.PDFA2B).NoObjectStreams().AddPage()
	b.DrawText("PDF/A with Helvetica", 20, 700, "Helvetica", 12)
	pdf := savePDF(t, b)
	if !bytes.Contains(pdf, []byte("/FontFile2")) {
		t.Error("no embedded font program in PDF/A output")
	}
}

func TestEmbedOpenTypeCFF(t *testing.T) {
	f := loadFontFile(t,
		"/usr/share/fonts/opentype/inter/Inter-Regular.otf",
		"/usr/share/fonts/opentype/tlwg/Loma.otf",
		"/usr/share/fonts/opentype/urw-base35/NimbusSans-Regular.otf",
	)
	if !f.IsCFF() {
		t.Fatalf("%s should have CFF outlines", f.PostScriptName())
	}
	b := doc.New().RegisterFont(f).PageSize(300, 150).NoObjectStreams().AddPage()
	b.DrawText("OpenType CFF 123", 20, 80, f.PostScriptName(), 16)
	pdf := savePDF(t, b)

	if !bytes.Contains(pdf, []byte("/CIDFontType0C")) || !bytes.Contains(pdf, []byte("/FontFile3")) {
		t.Error("CFF program should be embedded as /FontFile3 /CIDFontType0C")
	}
	if bytes.Contains(pdf, []byte("/CIDToGIDMap")) {
		t.Error("CIDFontType0 must not carry /CIDToGIDMap")
	}
	if fonts := popplerCheck(t, "pdffonts", pdf); fonts != "" && !strings.Contains(fonts, "CID Type 0C") {
		t.Errorf("pdffonts:\n%s", fonts)
	}
	if text := pdfText(t, pdf); text != "" && !strings.Contains(text, "OpenType CFF 123") {
		t.Errorf("pdftotext:\n%s", text)
	}
}

func TestEmbedCIDKeyedCFF(t *testing.T) {
	f := loadFontFile(t, "/usr/share/fonts/opentype/unifont/unifont.otf", "/usr/share/fonts/opentype/noto/NotoSansCJK-Regular.ttc")
	reg, ord, _ := f.CIDSystemInfo()
	if reg == "" || ord == "" {
		t.Fatal("CIDSystemInfo empty")
	}
	b := doc.New().RegisterFont(f).PageSize(300, 150).AddPage()
	b.DrawText("CID keyed 字", 20, 80, f.PostScriptName(), 16)
	pdf := savePDF(t, b)
	if text := pdfText(t, pdf); text != "" && !strings.Contains(text, "CID keyed") {
		t.Errorf("pdftotext:\n%s", text)
	}
	popplerCheck(t, "pdffonts", pdf)
}

func TestArabicShapingAndBidi(t *testing.T) {
	f := loadTestFont(t) // DejaVu Sans carries init/medi/fina GSUB features
	b := doc.New().RegisterFont(f).PageSize(400, 200).AddPage()
	const mixed = "Hello مرحبا بالعالم (world) 42"
	b.DrawText(mixed, 20, 150, f.PostScriptName(), 14)
	b.DrawTextBox(0, "العربية في فقرة كاملة مع أرقام 123 ونص", 20, 110, f.PostScriptName(), 12, doc.TextLayoutOptions{Width: 360, LineHeight: 16, Align: doc.TextAlignRight})
	pdf := savePDF(t, b)

	// The contextual forms must be real glyphs of the font, so the medial
	// beh differs from the isolated one and appears in the content.
	iso := f.Shape("ب")[0].ID
	medial := f.Shape("ببب")[1].ID
	if iso == medial {
		t.Fatal("medial beh should differ from isolated beh")
	}
	text := pdfText(t, pdf)
	if text == "" {
		t.Skip("pdftotext not installed")
	}
	for _, want := range []string{"Hello", "مرحبا", "بالعالم", "(world)", "42", "فقرة"} {
		if !strings.Contains(text, want) {
			t.Errorf("pdftotext lacks %q:\n%s", want, text)
		}
	}
	popplerCheck(t, "pdffonts", pdf)
}

func TestLigatureGlyphsSurviveSubsetting(t *testing.T) {
	f := loadTestFont(t)
	if len(f.Shape("fi")) != 1 {
		t.Skip("font has no fi ligature")
	}
	b := doc.New().RegisterFont(f).PageSize(300, 150).AddPage()
	b.DrawText("fi fl office", 20, 80, f.PostScriptName(), 16)
	pdf := savePDF(t, b)
	if text := pdfText(t, pdf); text != "" && !strings.Contains(text, "fi fl office") {
		t.Errorf("ligatures must map back to their letters:\n%s", text)
	}
}

func TestSystemFontFallback(t *testing.T) {
	if system.DefaultFallback().FontFor('語') == nil {
		t.Skip("no installed font covers CJK")
	}
	b := doc.New().EnableSystemFontFallback().PageSize(300, 150).NoObjectStreams().AddPage()
	b.DrawText("Fallback 日本語 text", 20, 80, "Helvetica", 14)
	pdf := savePDF(t, b)
	for _, w := range b.Warnings() {
		if strings.Contains(w, "no glyph") {
			t.Errorf("unexpected warning %q", w)
		}
	}
	if !bytes.Contains(pdf, []byte("/CIDFontType")) {
		t.Error("fallback font not embedded")
	}
	if text := pdfText(t, pdf); text != "" && !strings.Contains(text, "日本語") {
		t.Errorf("pdftotext:\n%s", text)
	}
}

func TestBlockFallbackFont(t *testing.T) {
	f := loadTestFont(t)
	b := doc.New().RegisterBlockFallbackFont("Arabic", f).AddPage()
	b.DrawText("abc سلام", 20, 700, "Helvetica", 12)
	savePDF(t, b)
	for _, w := range b.Warnings() {
		if strings.Contains(w, "no glyph") {
			t.Errorf("unexpected warning %q", w)
		}
	}
}
