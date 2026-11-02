// Copyright (c) 2026 Gembit Soultan Shirazi <gembit.soultan@gmail.com>

package form

import (
	"fmt"
	"strings"

	"github.com/gsoultan/gpdf/model"
)

// Appearance streams (ISO 32000-2 §12.5.5, §12.7.4.3). PDF 2.0 deprecates
// /NeedAppearances, so every widget the builder writes carries its own /AP.

const (
	helvName = "Helv"
	zadbName = "ZaDb"
)

// fontResources returns the /DR font dictionary objects: Helvetica for text and
// ZapfDingbats for check marks and radio dots.
func fontResources(objs map[int]model.Object, nextNum int) (helvRef, zadbRef model.Ref, outNext int) {
	helv := nextNum
	zadb := nextNum + 1
	objs[helv] = model.Dict{
		model.Name("Type"):     model.Name("Font"),
		model.Name("Subtype"):  model.Name("Type1"),
		model.Name("BaseFont"): model.Name("Helvetica"),
		model.Name("Encoding"): model.Name("WinAnsiEncoding"),
	}
	objs[zadb] = model.Dict{
		model.Name("Type"):     model.Name("Font"),
		model.Name("Subtype"):  model.Name("Type1"),
		model.Name("BaseFont"): model.Name("ZapfDingbats"),
	}
	return model.Ref{ObjectNumber: helv}, model.Ref{ObjectNumber: zadb}, nextNum + 2
}

// formXObject wraps content as an appearance stream for a widget of the given size.
func formXObject(content string, w, h float64, fonts model.Dict) *model.Stream {
	d := model.Dict{
		model.Name("Type"):    model.Name("XObject"),
		model.Name("Subtype"): model.Name("Form"),
		model.Name("BBox"):    model.Array{model.Integer(0), model.Integer(0), model.Real(w), model.Real(h)},
		model.Name("Length"):  model.Integer(len(content)),
	}
	if len(fonts) > 0 {
		d[model.Name("Resources")] = model.Dict{model.Name("Font"): fonts}
	}
	return &model.Stream{Dict: d, Content: []byte(content)}
}

// textFieldAppearance draws value with Helvetica inside a text field.
func textFieldAppearance(value string, w, h float64, helv model.Ref) (*model.Stream, float64) {
	size := 10.0
	if h < 14 {
		size = max(h-4, 4)
	}
	y := (h - size*0.72) / 2
	var sb strings.Builder
	sb.WriteString("/Tx BMC\nq\n")
	fmt.Fprintf(&sb, "1 1 %s %s re W n\n", num(w-2), num(h-2))
	if value != "" {
		fmt.Fprintf(&sb, "BT\n/%s %s Tf\n0 g\n2 %s Td\n(%s) Tj\nET\n", helvName, num(size), num(y), pdfLiteral(value))
	}
	sb.WriteString("Q\nEMC\n")
	return formXObject(sb.String(), w, h, model.Dict{model.Name(helvName): helv}), size
}

// symbolAppearance draws the on state of a check box (glyph "4": a check mark) or a
// radio button (glyph "l": a dot) as vector paths, so it looks the same whether or
// not the viewer has a ZapfDingbats font.
func symbolAppearance(glyph string, w, h float64) *model.Stream {
	s := min(w, h)
	cx, cy := w/2, h/2
	content := frameOps(glyph, w, h)
	if glyph == "l" {
		r := s * 0.3
		const k = 0.5523 // Bézier circle constant
		content += fmt.Sprintf("q\n0 g\n%s %s m\n%s %s %s %s %s %s c\n%s %s %s %s %s %s c\n%s %s %s %s %s %s c\n%s %s %s %s %s %s c\nf\nQ\n",
			num(cx+r), num(cy),
			num(cx+r), num(cy+r*k), num(cx+r*k), num(cy+r), num(cx), num(cy+r),
			num(cx-r*k), num(cy+r), num(cx-r), num(cy+r*k), num(cx-r), num(cy),
			num(cx-r), num(cy-r*k), num(cx-r*k), num(cy-r), num(cx), num(cy-r),
			num(cx+r*k), num(cy-r), num(cx+r), num(cy-r*k), num(cx+r), num(cy))
	} else {
		content += fmt.Sprintf("q\n0 G\n%s w\n1 J\n1 j\n%s %s m\n%s %s l\n%s %s l\nS\nQ\n",
			num(max(s*0.12, 1)),
			num(cx-s*0.3), num(cy), num(cx-s*0.08), num(cy-s*0.25), num(cx+s*0.32), num(cy+s*0.28))
	}
	return formXObject(content, w, h, nil)
}

// emptyAppearance is the Off state of a check box ("4") or radio button ("l"): just
// the frame.
func emptyAppearance(glyph string, w, h float64) *model.Stream {
	return formXObject(frameOps(glyph, w, h), w, h, nil)
}

// frameOps draws a thin grey square (check box) or circle (radio button) so the
// control is visible when off.
func frameOps(glyph string, w, h float64) string {
	s := min(w, h)
	cx, cy := w/2, h/2
	if glyph == "l" {
		r := s/2 - 0.5
		const k = 0.5523
		return fmt.Sprintf("q\n0.5 G\n1 w\n%s %s m\n%s %s %s %s %s %s c\n%s %s %s %s %s %s c\n%s %s %s %s %s %s c\n%s %s %s %s %s %s c\nS\nQ\n",
			num(cx+r), num(cy),
			num(cx+r), num(cy+r*k), num(cx+r*k), num(cy+r), num(cx), num(cy+r),
			num(cx-r*k), num(cy+r), num(cx-r), num(cy+r*k), num(cx-r), num(cy),
			num(cx-r), num(cy-r*k), num(cx-r*k), num(cy-r), num(cx), num(cy-r),
			num(cx+r*k), num(cy-r), num(cx+r), num(cy-r*k), num(cx+r), num(cy))
	}
	return fmt.Sprintf("q\n0.5 G\n1 w\n%s %s %s %s re\nS\nQ\n", num(cx-s/2+0.5), num(cy-s/2+0.5), num(s-1), num(s-1))
}

// buttonAppearance draws a push button: light grey face, dark border, centred caption.
func buttonAppearance(caption string, w, h float64, helv model.Ref) *model.Stream {
	size := 10.0
	if h < 14 {
		size = max(h-4, 4)
	}
	var sb strings.Builder
	fmt.Fprintf(&sb, "q\n0.75 g\n0 0 %s %s re f\n0.25 G\n1 w\n0.5 0.5 %s %s re S\n", num(w), num(h), num(w-1), num(h-1))
	if caption != "" {
		// Helvetica averages about 0.5 em per character.
		tw := float64(len(caption)) * size * 0.5
		x := max((w-tw)/2, 2)
		y := (h - size*0.72) / 2
		fmt.Fprintf(&sb, "BT\n/%s %s Tf\n0 g\n%s %s Td\n(%s) Tj\nET\n", helvName, num(size), num(x), num(y), pdfLiteral(caption))
	}
	sb.WriteString("Q\n")
	return formXObject(sb.String(), w, h, model.Dict{model.Name(helvName): helv})
}

func num(v float64) string { return model.FormatReal(v) }

// pdfLiteral encodes s for a literal string in a content stream using WinAnsiEncoding
// for Latin-1 text; other characters become '?'.
func pdfLiteral(s string) string {
	var sb strings.Builder
	for _, r := range s {
		var c byte
		switch {
		case r < 0x80:
			c = byte(r)
		case r >= 0xA0 && r <= 0xFF:
			c = byte(r)
		case r == '€':
			c = 0x80
		case r == '•':
			c = 0x95
		case r == '–':
			c = 0x96
		case r == '—':
			c = 0x97
		default:
			c = '?'
		}
		switch c {
		case '(', ')', '\\':
			sb.WriteByte('\\')
			sb.WriteByte(c)
		case '\r', '\n':
			sb.WriteByte(' ')
		default:
			if c < 0x20 || c >= 0x80 {
				fmt.Fprintf(&sb, "\\%03o", c)
			} else {
				sb.WriteByte(c)
			}
		}
	}
	return sb.String()
}
