// Copyright (c) 2026 Gembit Soultan Shirazi <gembit.soultan@gmail.com>

package doc

import (
	"strings"
	"unicode"

	"github.com/gsoultan/gpdf/font"
	"github.com/gsoultan/gpdf/font/shaping"
)

var standardSubstitutions = map[string]string{
	"Arial":           "Helvetica",
	"Times New Roman": "Times-Roman",
	"Courier New":     "Courier",
}

// fontConfig holds font-related state for the builder.
type fontConfig struct {
	fonts          map[string]font.Font
	embeddedFonts  map[string]*embeddedFontUsage
	fallbackChain  []string
	blockFallbacks map[string]string
	// systemFallback, when set, finds an installed font for a rune no
	// registered font covers; the font it returns is registered on the fly.
	systemFallback func(r rune) font.EmbeddableFont
	direction      shaping.Direction
	onWarning      func(string)
}

// aliasedFont registers an embeddable font under another name, so that a
// standard font name such as Helvetica resolves to an embedded program.
type aliasedFont struct {
	font.EmbeddableFont
	name string
}

func (a *aliasedFont) PostScriptName() string { return a.name }

// EmbeddedFontName returns the PostScript name of the program that is embedded.
func (a *aliasedFont) EmbeddedFontName() string { return a.EmbeddableFont.PostScriptName() }

// unwrapFont returns the font behind an alias.
func unwrapFont(f font.Font) font.Font {
	if a, ok := f.(*aliasedFont); ok {
		return a.EmbeddableFont
	}
	return f
}

func (fc *fontConfig) registerFont(f font.Font) {
	if fc.fonts == nil {
		fc.fonts = make(map[string]font.Font)
	}
	fc.fonts[f.PostScriptName()] = f
	if ef, ok := f.(font.EmbeddableFont); ok {
		if fc.embeddedFonts == nil {
			fc.embeddedFonts = make(map[string]*embeddedFontUsage)
		}
		fc.embeddedFonts[ef.PostScriptName()] = newEmbeddedFontUsage(ef)
	}
}

// registerFontAs registers f under name instead of its own PostScript name.
func (fc *fontConfig) registerFontAs(f font.EmbeddableFont, name string) {
	fc.registerFont(&aliasedFont{EmbeddableFont: f, name: name})
}

func (fc *fontConfig) addFallback(fontName string) {
	fc.fallbackChain = append(fc.fallbackChain, fontName)
}

func (fc *fontConfig) setBlockFallback(block string, fontName string) {
	if fc.blockFallbacks == nil {
		fc.blockFallbacks = make(map[string]string)
	}
	fc.blockFallbacks[block] = fontName
}

// hasFallbacks reports whether text may be drawn with more than its primary
// font, in which case widths must be measured per resolved segment.
func (fc *fontConfig) hasFallbacks() bool {
	return len(fc.fallbackChain) > 0 || len(fc.blockFallbacks) > 0 || fc.systemFallback != nil
}

func (fc *fontConfig) resolveFontName(name string) string {
	if f, ok := fc.fonts[name]; ok {
		return f.PostScriptName()
	}
	if sub, ok := standardSubstitutions[name]; ok {
		return sub
	}
	return name
}

// shaperFor returns the shaper behind a registered font name, if any.
func (fc *fontConfig) shaperFor(name string) font.Shaper {
	f, ok := fc.fonts[name]
	if !ok {
		return nil
	}
	s, _ := unwrapFont(f).(font.Shaper)
	return s
}

// resolveFont prepares one line of text for drawing: it reorders the line
// into visual runs with the Unicode bidirectional algorithm, splits each run
// by font support (the primary font first, then block fallbacks, the
// fallback chain and the system fallback), and shapes every segment whose
// font can shape, so that Arabic letters get their contextual forms and
// right-to-left runs come out in visual order.
func (fc *fontConfig) resolveFont(text string, primaryFontName string) []fontSegment {
	if text == "" {
		return nil
	}
	primaryFontName = fc.resolveFontName(primaryFontName)

	var segments []fontSegment
	for _, run := range shaping.Reorder(text, fc.direction) {
		segments = append(segments, fc.segmentRun(run, primaryFontName)...)
	}
	return segments
}

// segmentRun splits one bidi run by font and shapes the segments.
func (fc *fontConfig) segmentRun(run shaping.Run, primaryFontName string) []fontSegment {
	var segments []fontSegment
	var currentSegment strings.Builder
	var currentFont string

	flush := func() {
		if currentSegment.Len() == 0 {
			return
		}
		seg := fontSegment{text: currentSegment.String(), fontName: currentFont, isRTL: run.RTL}
		if shaper := fc.shaperFor(currentFont); shaper != nil {
			seg.glyphs = shaper.Shape(seg.text)
			if run.RTL {
				seg.glyphs = reverseGlyphs(seg.glyphs)
			}
		}
		segments = append(segments, seg)
		currentSegment.Reset()
	}

	for _, r := range run.Text {
		resolved := fc.findFontForRune(r, primaryFontName)
		if resolved == "" {
			resolved = primaryFontName // Fallback to primary if nothing found (will render placeholder)
		}
		if resolved != currentFont {
			flush()
		}
		currentFont = resolved
		currentSegment.WriteRune(r)
	}
	flush()
	return segments
}

// reverseGlyphs returns the glyphs in reverse order, keeping each kern
// adjustment between the same pair of glyphs.
func reverseGlyphs(gs []font.Glyph) []font.Glyph {
	n := len(gs)
	out := make([]font.Glyph, n)
	for i, g := range gs {
		out[n-1-i] = g
		out[n-1-i].Kern = 0
	}
	for i := 1; i < n; i++ {
		out[i].Kern = gs[n-i].Kern
	}
	return out
}

func (fc *fontConfig) findFontForRune(r rune, primary string) string {
	// Try primary
	if f, ok := fc.fonts[primary]; ok {
		if f.Contains(r) {
			return primary
		}
	} else {
		// For standard fonts, we don't have a Font object, but we check if it's a known standard font.
		if font.GetStandardWidth(primary, r) > 0 || r == ' ' || r == '\n' || r == '\r' || r == '\t' {
			return primary
		}
	}

	// Try block-specific fallback
	block := getRuneBlock(r)
	if name, ok := fc.blockFallbacks[block]; ok {
		if f, ok := fc.fonts[name]; ok && f.Contains(r) {
			return name
		}
	}

	// Try fallback chain
	for _, name := range fc.fallbackChain {
		if name == primary {
			continue
		}
		if f, ok := fc.fonts[name]; ok && f.Contains(r) {
			return name
		}
	}

	// Try the fonts installed on the system.
	if fc.systemFallback != nil {
		if f := fc.systemFallback(r); f != nil {
			name := f.PostScriptName()
			if _, registered := fc.fonts[name]; !registered {
				fc.registerFont(f)
			}
			return name
		}
	}

	if fc.onWarning != nil {
		fc.onWarning(strings.Join([]string{"no glyph for", string(r), "in font", primary, "or fallbacks"}, " "))
	}
	return ""
}
func getRuneBlock(r rune) string {
	switch {
	case unicode.Is(unicode.Han, r), unicode.Is(unicode.Hiragana, r), unicode.Is(unicode.Katakana, r), unicode.Is(unicode.Hangul, r):
		return "CJK"
	case unicode.Is(unicode.Arabic, r):
		return "Arabic"
	case unicode.Is(unicode.Hebrew, r):
		return "Hebrew"
	case unicode.Is(unicode.Thai, r):
		return "Thai"
	case unicode.Is(unicode.Bengali, r), unicode.Is(unicode.Devanagari, r), unicode.Is(unicode.Gujarati, r), unicode.Is(unicode.Gurmukhi, r), unicode.Is(unicode.Kannada, r), unicode.Is(unicode.Malayalam, r), unicode.Is(unicode.Oriya, r), unicode.Is(unicode.Tamil, r), unicode.Is(unicode.Telugu, r):
		return "Indic"
	default:
		return "Default"
	}
}
