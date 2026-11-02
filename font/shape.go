// Copyright (c) 2026 Gembit Soultan Shirazi <gembit.soultan@gmail.com>

package font

// Glyph is one shaped glyph: the code to write into an Identity-H content
// stream, the text it stands for, and its advance.
type Glyph struct {
	// ID is the glyph code written into the content stream: the glyph index
	// for TrueType and non-CID-keyed CFF fonts, the CID for CID-keyed CFF
	// fonts.
	ID uint16
	// Text is the source text the glyph renders (one character, or several
	// for a ligature), used for the ToUnicode CMap.
	Text string
	// Advance is the glyph's advance width in font units.
	Advance int
	// Kern is the kerning adjustment to apply before the glyph, in font
	// units; negative values tighten the pair.
	Kern int
}

// Shaper is implemented by fonts that turn logical-order text into glyphs,
// applying Arabic contextual forms, ligatures and kerning. Callers draw a
// right-to-left run by reversing the returned glyphs.
type Shaper interface {
	Shape(text string) []Glyph
}

// GlyphEncoder is implemented by embeddable fonts that can encode shaped
// glyphs for a content stream and map glyph codes back to text.
type GlyphEncoder interface {
	// EncodeGlyphs serialises glyph codes as 2-byte big-endian values.
	EncodeGlyphs(ids []uint16) []byte
	// ToUnicodeCMapByGlyph generates a ToUnicode CMap from glyph codes to the
	// text they render.
	ToUnicodeCMapByGlyph(glyphs map[uint16]string) []byte
}

// CFFFont is implemented by OpenType fonts with PostScript (CFF) outlines.
// Their program is embedded as /FontFile3 /Subtype /CIDFontType0C under a
// CIDFontType0 font instead of /FontFile2 under CIDFontType2.
type CFFFont interface {
	// CFF returns the bare CFF font program, or nil for TrueType outlines.
	CFF() []byte
	// CIDSystemInfo returns the Registry, Ordering and Supplement the glyph
	// codes belong to (Adobe-Identity-0 unless the CFF program is CID-keyed).
	CIDSystemInfo() (registry, ordering string, supplement int)
}
