package truetype

import (
	"encoding/binary"
	"testing"
)

// Fonts such as Carlito kern only through GPOS (no legacy 'kern' table).
func TestGPOSPairKerning(t *testing.T) {
	f := loadFont(t, "/usr/share/fonts/truetype/crosextra/Carlito-Regular.ttf")
	if f.gpos == nil {
		t.Fatal("GPOS kerning not parsed")
	}
	if k := f.Kern('A', 'V'); k >= 0 {
		t.Fatalf("Kern(A, V) = %d, want negative", k)
	}
	if f.TextWidth("AV", 12) >= f.TextWidth("A", 12)+f.TextWidth("V", 12) {
		t.Fatal("TextWidth ignores GPOS kerning")
	}
}

// A hand-built PairPos format 2 subtable: glyph 5 (class 1) before glyph 9
// (class 2) moves by -80 units.
func TestParsePairPosClassBased(t *testing.T) {
	be := binary.BigEndian
	var d []byte
	u16 := func(v int) { d = be.AppendUint16(d, uint16(v)) }
	// header: format 2, coverage, vf1 = XAdvance, vf2 = 0, classDef1, classDef2, class1Count 2, class2Count 3
	u16(2)
	u16(0) // coverage offset, patched below
	u16(4)
	u16(0)
	u16(0) // classDef1
	u16(0) // classDef2
	u16(2)
	u16(3)
	for c1 := range 2 { // records [class1][class2]: one int16 each
		for c2 := range 3 {
			v := 0
			if c1 == 1 && c2 == 2 {
				v = -80
			}
			u16(v & 0xFFFF)
		}
	}
	cov := len(d)
	u16(1)
	u16(1)
	u16(5) // coverage format 1: glyph 5
	cd1 := len(d)
	u16(1)
	u16(5)
	u16(1)
	u16(1) // classDef format 1: glyph 5 -> class 1
	cd2 := len(d)
	u16(2)
	u16(1)
	u16(9)
	u16(9)
	u16(2) // classDef format 2: glyph 9 -> class 2
	be.PutUint16(d[2:], uint16(cov))
	be.PutUint16(d[8:], uint16(cd1))
	be.PutUint16(d[10:], uint16(cd2))
	st, ok := parsePairPos(d)
	if !ok {
		t.Fatal("subtable rejected")
	}
	k := &gposKern{lookups: [][]pairSubtable{{st}}}
	if got := k.adjust(5, 9); got != -80 {
		t.Fatalf("adjust(5, 9) = %d, want -80", got)
	}
	if got := k.adjust(5, 7); got != 0 {
		t.Fatalf("adjust(5, 7) = %d, want 0", got)
	}
	if got := k.adjust(6, 9); got != 0 {
		t.Fatalf("uncovered first glyph: adjust(6, 9) = %d, want 0", got)
	}
}
