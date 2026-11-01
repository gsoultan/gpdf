package truetype

import (
	"encoding/binary"
	"fmt"
	"sort"
	"strings"

	"github.com/gsoultan/gpdf/font"
	"github.com/gsoultan/gpdf/font/shaping"
)

// Shape maps logical-order text to glyphs. Arabic-script letters receive
// their contextual form, through the font's init/medi/fina GSUB features when
// it has them and through Unicode presentation-form glyphs otherwise; then
// ligatures (including the mandatory lam-alef) and kern pairs are applied.
func (f *Font) Shape(text string) []font.Glyph {
	runes := []rune(text)
	if len(runes) == 0 {
		return nil
	}
	type pending struct {
		gid  uint16
		text string
	}
	items := make([]pending, 0, len(runes))
	var forms []shaping.Form
	for _, r := range runes {
		if shaping.IsArabicLetter(r) {
			forms = shaping.ArabicForms(runes)
			break
		}
	}
	useGSUB := f.gsub.HasArabicForms()
	for i := 0; i < len(runes); i++ {
		r := runes[i]
		gid := f.runeToGlyph[r]
		if forms == nil || !shaping.IsArabicLetter(r) {
			items = append(items, pending{gid, string(r)})
			continue
		}
		form := forms[i]
		if useGSUB {
			if sub, ok := f.gsub.Forms[form.FeatureTag()][gid]; ok {
				gid = sub
			}
			items = append(items, pending{gid, string(r)})
			continue
		}
		// Presentation forms: lam + alef becomes one ligature glyph.
		if r == 0x0644 && (form == shaping.Initial || form == shaping.Medial) {
			if j := nextBase(runes, i); j >= 0 {
				if lig := shaping.LamAlefLigature(runes[j], form == shaping.Medial); lig != 0 {
					if lg, ok := f.runeToGlyph[lig]; ok {
						items = append(items, pending{lg, string(runes[i : j+1])})
						i = j
						continue
					}
				}
			}
		}
		if pf := shaping.PresentationForm(r, form); pf != 0 {
			if pg, ok := f.runeToGlyph[pf]; ok {
				gid = pg
			}
		}
		items = append(items, pending{gid, string(r)})
	}

	// Ligatures from GSUB, longest match first.
	var out []font.Glyph
	for i := 0; i < len(items); {
		matched := false
		if f.gsub != nil && len(f.gsub.Ligatures) > 0 {
			for n := 3; n >= 2; n-- {
				if i+n > len(items) {
					continue
				}
				key := make([]byte, n*2)
				var sb strings.Builder
				for k := range n {
					binary.BigEndian.PutUint16(key[k*2:], items[i+k].gid)
					sb.WriteString(items[i+k].text)
				}
				if lig, ok := f.gsub.Ligatures[string(key)]; ok {
					out = append(out, font.Glyph{ID: lig, Text: sb.String()})
					i += n
					matched = true
					break
				}
			}
		}
		if !matched {
			out = append(out, font.Glyph{ID: items[i].gid, Text: items[i].text})
			i++
		}
	}

	// Advances, kerning and the content-stream code.
	for i := range out {
		gid := out[i].ID
		if int(gid) < len(f.glyphWidths) {
			out[i].Advance = int(f.glyphWidths[gid])
		}
		if i > 0 && f.kernPairs != nil {
			out[i].Kern = int(f.kernPairs[uint32(out[i-1].ID)<<16|uint32(gid)])
		}
	}
	for i := range out {
		out[i].ID = f.glyphCode(out[i].ID)
	}
	return out
}

// nextBase returns the index of the next non-transparent rune after i, or -1.
func nextBase(runes []rune, i int) int {
	for j := i + 1; j < len(runes); j++ {
		if shaping.JoiningTypeOf(runes[j]) != shaping.Transparent {
			return j
		}
	}
	return -1
}

// EncodeGlyphs serialises glyph codes as 2-byte big-endian values for an
// Identity-H content stream.
func (f *Font) EncodeGlyphs(ids []uint16) []byte {
	out := make([]byte, 0, len(ids)*2)
	for _, id := range ids {
		out = append(out, byte(id>>8), byte(id))
	}
	return out
}

// ToUnicodeCMapByGlyph generates a ToUnicode CMap from glyph codes to the
// text they render; a ligature maps to all of its characters.
func (f *Font) ToUnicodeCMapByGlyph(glyphs map[uint16]string) []byte {
	ids := make([]int, 0, len(glyphs))
	for id, text := range glyphs {
		if id == 0 || text == "" {
			continue
		}
		ids = append(ids, int(id))
	}
	sort.Ints(ids)

	var b strings.Builder
	b.WriteString("/CIDInit /ProcSet findresource begin\n")
	b.WriteString("12 dict begin\n")
	b.WriteString("begincmap\n")
	b.WriteString("/CIDSystemInfo << /Registry (Adobe) /Ordering (UCS) /Supplement 0 >> def\n")
	b.WriteString("/CMapName /Adobe-Identity-UCS def\n")
	b.WriteString("/CMapType 2 def\n")
	b.WriteString("1 begincodespacerange\n<0000> <FFFF>\nendcodespacerange\n")
	for start := 0; start < len(ids); start += 100 {
		end := min(start+100, len(ids))
		fmt.Fprintf(&b, "%d beginbfchar\n", end-start)
		for _, id := range ids[start:end] {
			fmt.Fprintf(&b, "<%04X> <", id)
			for _, r := range glyphs[uint16(id)] {
				writeUTF16BE(&b, r)
			}
			b.WriteString(">\n")
		}
		b.WriteString("endbfchar\n")
	}
	b.WriteString("endcmap\nCMapName currentdict /CMap defineresource pop\nend\nend\n")
	return []byte(b.String())
}

func writeUTF16BE(b *strings.Builder, r rune) {
	if r >= 0x10000 {
		r -= 0x10000
		fmt.Fprintf(b, "%04X%04X", 0xD800+(r>>10), 0xDC00+(r&0x3FF))
		return
	}
	fmt.Fprintf(b, "%04X", r)
}

var _ font.Shaper = (*Font)(nil)
var _ font.GlyphEncoder = (*Font)(nil)
var _ font.CFFFont = (*Font)(nil)
