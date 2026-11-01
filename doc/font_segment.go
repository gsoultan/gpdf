package doc

import "github.com/gsoultan/gpdf/font"

// fontSegment is a run of text drawn with one font. text is in logical order;
// for a right-to-left segment drawn with a font that cannot shape, the
// characters are reversed when the run is emitted. glyphs, when set, is the
// shaped glyph sequence in visual order and takes precedence over text in
// the content stream.
type fontSegment struct {
	text     string
	fontName string
	isRTL    bool
	glyphs   []font.Glyph
}

// drawText returns the text to store in a text run for the segment.
func (s fontSegment) drawText() string {
	if s.isRTL && s.glyphs == nil {
		return reverseRunes(s.text)
	}
	return s.text
}

// textRun returns a text run for the segment at (x, y) with the given size.
func (s fontSegment) textRun(x, y, size float64) textRun {
	return textRun{Text: s.drawText(), Glyphs: s.glyphs, X: x, Y: y, FontName: s.fontName, FontSize: size}
}
