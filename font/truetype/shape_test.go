package truetype

import (
	"os"
	"strings"
	"testing"
)

func loadFont(t *testing.T, paths ...string) *Font {
	t.Helper()
	for _, p := range paths {
		data, err := os.ReadFile(p)
		if err != nil {
			continue
		}
		f, err := Parse(data)
		if err != nil {
			t.Fatalf("parse %s: %v", p, err)
		}
		return f
	}
	t.Skipf("none of %v installed", paths)
	return nil
}

func TestShapeArabicWithGSUB(t *testing.T) {
	f := loadFont(t, "/usr/share/fonts/truetype/dejavu/DejaVuSans.ttf", "/usr/share/fonts/truetype/freefont/FreeSerif.ttf")
	if !f.gsub.HasArabicForms() {
		t.Skip("font has no Arabic GSUB features")
	}
	glyphs := f.Shape("ببب")
	if len(glyphs) != 3 {
		t.Fatalf("got %d glyphs", len(glyphs))
	}
	iso := f.GlyphID('ب')
	if glyphs[0].ID == iso || glyphs[1].ID == iso || glyphs[2].ID == iso {
		t.Errorf("contextual forms should differ from the isolated glyph: %v (isolated %d)", glyphs, iso)
	}
	if glyphs[0].ID == glyphs[1].ID || glyphs[1].ID == glyphs[2].ID {
		t.Errorf("initial, medial and final forms should differ: %v", glyphs)
	}
	for _, g := range glyphs {
		if g.Text != "ب" || g.Advance <= 0 {
			t.Errorf("glyph %+v should keep its source text and advance", g)
		}
	}
	// Width after shaping is what the content stream advances by.
	var sum int
	for _, g := range glyphs {
		sum += g.Advance + g.Kern
	}
	if w := f.TextWidth("ببب", float64(f.UnitsPerEm())); int(w+0.5) != sum {
		t.Errorf("TextWidth %v, glyph sum %d", w, sum)
	}
}

func TestShapePresentationFormsFallback(t *testing.T) {
	f := loadFont(t, "/usr/share/fonts/truetype/dejavu/DejaVuSans.ttf")
	// Force the presentation-form path by hiding the GSUB features.
	saved := f.gsub
	f.gsub = &GSUB{Ligatures: saved.Ligatures}
	defer func() { f.gsub = saved }()
	glyphs := f.Shape("لا بب")
	if glyphs[0].Text != "لا" || glyphs[0].ID != f.runeToGlyph[0xFEFB] {
		t.Errorf("lam-alef should become the FEFB ligature: %+v", glyphs[0])
	}
	if glyphs[2].ID != f.runeToGlyph[0xFE91] || glyphs[3].ID != f.runeToGlyph[0xFE90] {
		t.Errorf("beh forms should be FE91 (initial) and FE90 (final): %+v", glyphs[2:])
	}
}

func TestShapeLigaturesAndToUnicode(t *testing.T) {
	f := loadFont(t, "/usr/share/fonts/truetype/dejavu/DejaVuSans.ttf")
	glyphs := f.Shape("fi")
	if len(glyphs) != 1 || glyphs[0].Text != "fi" {
		t.Skipf("no fi ligature: %v", glyphs)
	}
	cmap := string(f.ToUnicodeCMapByGlyph(map[uint16]string{glyphs[0].ID: glyphs[0].Text}))
	if !strings.Contains(cmap, "<00660069>") {
		t.Errorf("ligature should map to both letters:\n%s", cmap)
	}
	if string(f.EncodeGlyphs([]uint16{0x1234})) != "\x12\x34" {
		t.Error("EncodeGlyphs wrong")
	}
}

func TestParseCFF(t *testing.T) {
	f := loadFont(t, "/usr/share/fonts/opentype/inter/Inter-Regular.otf", "/usr/share/fonts/opentype/tlwg/Loma.otf")
	if !f.IsCFF() || len(f.CFF()) == 0 {
		t.Fatal("expected CFF outlines")
	}
	if f.cff.numGlyphs != f.NumGlyphs() {
		t.Errorf("CharStrings count %d, maxp %d", f.cff.numGlyphs, f.NumGlyphs())
	}
	reg, ord, sup := f.CIDSystemInfo()
	if reg != "Adobe" || ord != "Identity" || sup != 0 {
		t.Errorf("non-CID-keyed CFF should report Adobe-Identity-0, got %s-%s-%d", reg, ord, sup)
	}
	if f.GlyphID('A') == 0 || f.GlyphWidth('A') == 0 {
		t.Error("cmap/hmtx of the OTTO font not read")
	}
	sub, err := f.Subset(map[uint16]bool{f.GlyphID('A'): true})
	if err != nil || len(sub) != len(f.CFF()) {
		t.Errorf("Subset of a CFF font should return the CFF program: %v, %d bytes", err, len(sub))
	}
}

func TestParseCFFCIDKeyed(t *testing.T) {
	f := loadFont(t, "/usr/share/fonts/opentype/unifont/unifont.otf")
	if !f.IsCFF() || !f.cff.cidKeyed {
		t.Skip("not a CID-keyed CFF font")
	}
	reg, ord, _ := f.CIDSystemInfo()
	if reg == "" || ord == "" {
		t.Error("ROS not read")
	}
	code := f.GlyphID('A')
	gid := f.glyphIndex(code)
	if f.cff.gidToCID[gid] != code {
		t.Errorf("charset round trip: gid %d cid %d", gid, code)
	}
	if w := f.CIDWidths(map[uint16]bool{code: true})[code]; w != int(f.glyphWidths[gid]) {
		t.Errorf("width by CID %d, by GID %d", w, f.glyphWidths[gid])
	}
}

func TestCFFDictAndCharset(t *testing.T) {
	// 28 0x01 0x00 -> 256 ; 139 -> 0 ; operator 15 (charset)
	d := parseCFFDict([]byte{28, 0x01, 0x00, 139, 15, 12, 30})
	if v := d[15]; len(v) != 2 || v[0] != 256 || v[1] != 0 {
		t.Errorf("dict operands %v", d)
	}
	if _, ok := d[12<<8|30]; !ok {
		t.Error("escaped operator not keyed")
	}
	gidToCID := make([]uint16, 5)
	parseCFFCharset([]byte{1, 0x00, 0x0A, 3}, gidToCID) // format 1: first 10, nLeft 3
	if gidToCID[1] != 10 || gidToCID[4] != 13 {
		t.Errorf("charset %v", gidToCID)
	}
	parseCFFCharset([]byte{0, 0x00, 0x20, 0x00, 0x21}, gidToCID[:3]) // format 0
	if gidToCID[1] != 0x20 || gidToCID[2] != 0x21 {
		t.Errorf("charset format 0 %v", gidToCID)
	}
}
