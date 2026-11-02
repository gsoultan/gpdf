// Copyright (c) 2026 Gembit Soultan Shirazi <gembit.soultan@gmail.com>

package reader

import (
	"strconv"
	"strings"
	"unicode/utf16"

	"github.com/gsoultan/gpdf/model"
)

// Simple-font text decoding (ISO 32000-1 §9.6.5 and Annex D): a single-byte code is mapped
// through the font's base encoding, patched by /Differences glyph names.

// glyphNames maps the glyph names of the standard Latin character set (Annex D.2) and a few
// common extras to Unicode. Other names are resolved by glyphRune (uniXXXX, uXXXX[XX]).
var glyphNames = func() map[string]rune {
	const table = "A 41 AE c6 Aacute c1 Acircumflex c2 Adieresis c4 Agrave c0 Aring c5 Atilde c3 B 42 C 43 Ccedilla c7 " +
		"D 44 E 45 Eacute c9 Ecircumflex ca Edieresis cb Egrave c8 Eth d0 Euro 20ac F 46 G 47 H 48 I 49 Iacute cd " +
		"Icircumflex ce Idieresis cf Igrave cc J 4a K 4b L 4c Lslash 141 M 4d N 4e Ntilde d1 O 4f OE 152 Oacute d3 " +
		"Ocircumflex d4 Odieresis d6 Ograve d2 Oslash d8 Otilde d5 P 50 Q 51 R 52 S 53 Scaron 160 T 54 Thorn de U 55 " +
		"Uacute da Ucircumflex db Udieresis dc Ugrave d9 V 56 W 57 X 58 Y 59 Yacute dd Ydieresis 178 Z 5a Zcaron 17d " +
		"a 61 aacute e1 acircumflex e2 acute b4 adieresis e4 ae e6 agrave e0 ampersand 26 aring e5 asciicircum 5e " +
		"asciitilde 7e asterisk 2a at 40 atilde e3 b 62 backslash 5c bar 7c braceleft 7b braceright 7d bracketleft 5b " +
		"bracketright 5d breve 2d8 brokenbar a6 bullet 2022 c 63 caron 2c7 ccedilla e7 cedilla b8 cent a2 circumflex 2c6 " +
		"colon 3a comma 2c copyright a9 currency a4 d 64 dagger 2020 daggerdbl 2021 degree b0 dieresis a8 divide f7 " +
		"dollar 24 dotaccent 2d9 dotlessi 131 e 65 eacute e9 ecircumflex ea edieresis eb egrave e8 eight 38 ellipsis 2026 " +
		"emdash 2014 endash 2013 equal 3d eth f0 exclam 21 exclamdown a1 f 66 fi fb01 five 35 fl fb02 florin 192 four 34 " +
		"fraction 2044 g 67 germandbls df grave 60 greater 3e guillemotleft ab guillemotright bb guilsinglleft 2039 " +
		"guilsinglright 203a h 68 hungarumlaut 2dd hyphen 2d i 69 iacute ed icircumflex ee idieresis ef igrave ec j 6a " +
		"k 6b l 6c less 3c logicalnot ac lslash 142 m 6d macron af minus 2212 mu b5 multiply d7 n 6e nine 39 ntilde f1 " +
		"numbersign 23 o 6f oacute f3 ocircumflex f4 odieresis f6 oe 153 ogonek 2db ograve f2 one 31 onehalf bd " +
		"onequarter bc onesuperior b9 ordfeminine aa ordmasculine ba oslash f8 otilde f5 p 70 paragraph b6 parenleft 28 " +
		"parenright 29 percent 25 period 2e periodcentered b7 perthousand 2030 plus 2b plusminus b1 q 71 question 3f " +
		"questiondown bf quotedbl 22 quotedblbase 201e quotedblleft 201c quotedblright 201d quoteleft 2018 " +
		"quoteright 2019 quotesinglbase 201a quotesingle 27 r 72 registered ae ring 2da s 73 scaron 161 section a7 " +
		"semicolon 3b seven 37 six 36 slash 2f space 20 sterling a3 t 74 thorn fe three 33 threequarters be " +
		"threesuperior b3 tilde 2dc trademark 2122 two 32 twosuperior b2 u 75 uacute fa ucircumflex fb udieresis fc " +
		"ugrave f9 underscore 5f v 76 w 77 x 78 y 79 yacute fd ydieresis ff yen a5 z 7a zcaron 17e zero 30 " +
		"ff fb00 ffi fb03 ffl fb04 nbspace a0 nonbreakingspace a0 sfthyphen ad softhyphen ad"
	f := strings.Fields(table)
	m := make(map[string]rune, len(f)/2)
	for i := 0; i+1 < len(f); i += 2 {
		v, _ := strconv.ParseUint(f[i+1], 16, 32)
		m[f[i]] = rune(v)
	}
	return m
}()

// glyphRune resolves a glyph name to Unicode text ("" when unknown). Suffixes after '.'
// (a.sc, one.oldstyle) and ligature parts joined with '_' (f_f_i) are handled as in the
// Adobe Glyph List specification.
func glyphRune(name string) string {
	if i := strings.IndexByte(name, '.'); i > 0 {
		name = name[:i]
	}
	if strings.Contains(name, "_") {
		var sb strings.Builder
		for part := range strings.SplitSeq(name, "_") {
			sb.WriteString(glyphRune(part))
		}
		return sb.String()
	}
	if r, ok := glyphNames[name]; ok {
		return string(r)
	}
	if strings.HasPrefix(name, "uni") && len(name) >= 7 && (len(name)-3)%4 == 0 {
		var units []uint16
		for i := 3; i+4 <= len(name); i += 4 {
			v, err := strconv.ParseUint(name[i:i+4], 16, 16)
			if err != nil {
				return ""
			}
			units = append(units, uint16(v))
		}
		return string(utf16.Decode(units))
	}
	if strings.HasPrefix(name, "u") && len(name) >= 5 && len(name) <= 7 {
		if v, err := strconv.ParseUint(name[1:], 16, 32); err == nil && v <= 0x10FFFF {
			return string(rune(v))
		}
	}
	return ""
}

// baseEncoding returns the 256-entry code -> rune table of a named base encoding.
func baseEncoding(name string) *[256]rune {
	var t [256]rune
	for i := range 256 {
		t[i] = rune(i)
	}
	switch name {
	case "WinAnsiEncoding":
		for i := 0x80; i < 0xA0; i++ {
			if r := winAnsiHigh[i-0x80]; r != 0 {
				t[i] = r
			}
		}
		t[0x7F] = '•'
	case "MacRomanEncoding":
		copy(t[0x80:], []rune(macRomanHigh))
	case "StandardEncoding":
		t[0x27], t[0x60] = '’', '‘'
		for i := 0x7F; i < 256; i++ {
			t[i] = 0
		}
		for code, r := range standardHigh {
			t[code] = r
		}
	}
	return &t
}

var winAnsiHigh = [32]rune{
	0x20AC, 0, 0x201A, 0x0192, 0x201E, 0x2026, 0x2020, 0x2021, 0x02C6, 0x2030, 0x0160, 0x2039, 0x0152, 0, 0x017D, 0,
	0, 0x2018, 0x2019, 0x201C, 0x201D, 0x2022, 0x2013, 0x2014, 0x02DC, 0x2122, 0x0161, 0x203A, 0x0153, 0, 0x017E, 0x0178,
}

const macRomanHigh = "ÄÅÇÉÑÖÜáàâäãåçéèêëíìîïñóòôöõúùûü†°¢£§•¶ß®©™´¨≠ÆØ∞±≤≥¥µ∂∑∏π∫ªºΩæø" +
	"¿¡¬√ƒ≈∆«»… ÀÃÕŒœ–—“”‘’÷◊ÿŸ⁄¤‹›ﬁﬂ‡·‚„‰ÂÊÁËÈÍÎÏÌÓÔÒÚÛÙıˆ˜¯˘˙˚¸˝˛ˇ"

var standardHigh = map[int]rune{
	0xA1: '¡', 0xA2: '¢', 0xA3: '£', 0xA4: '⁄', 0xA5: '¥', 0xA6: 'ƒ', 0xA7: '§', 0xA8: '¤', 0xA9: '\'', 0xAA: '“',
	0xAB: '«', 0xAC: '‹', 0xAD: '›', 0xAE: 'ﬁ', 0xAF: 'ﬂ', 0xB1: '–', 0xB2: '†', 0xB3: '‡', 0xB4: '·', 0xB6: '¶',
	0xB7: '•', 0xB8: '‚', 0xB9: '„', 0xBA: '”', 0xBB: '»', 0xBC: '…', 0xBD: '‰', 0xBF: '¿', 0xC1: '`', 0xC2: '´',
	0xC3: 'ˆ', 0xC4: '˜', 0xC5: '¯', 0xC6: '˘', 0xC7: '˙', 0xC8: '¨', 0xCA: '˚', 0xCB: '¸', 0xCD: '˝', 0xCE: '˛',
	0xCF: 'ˇ', 0xD0: '—', 0xE1: 'Æ', 0xE3: 'ª', 0xE8: 'Ł', 0xE9: 'Ø', 0xEA: 'Œ', 0xEB: 'º', 0xF1: 'æ', 0xF5: 'ı',
	0xF8: 'ł', 0xF9: 'ø', 0xFA: 'œ', 0xFB: 'ß',
}

// simpleFontEncoding builds the code -> text table of a simple font from /Encoding
// (a name, or a dictionary with /BaseEncoding and /Differences). It returns nil for
// composite (Type0) fonts and for symbolic standard fonts, whose codes are not text.
func simpleFontEncoding(src contentSource, fontDict model.Dict) *[256]string {
	subtype, _ := fontDict[model.Name("Subtype")].(model.Name)
	if subtype == "Type0" {
		return nil
	}
	baseFont, _ := fontDict[model.Name("BaseFont")].(model.Name)
	if i := strings.IndexByte(string(baseFont), '+'); i == 6 {
		baseFont = baseFont[i+1:] // subset tag
	}
	if baseFont == "Symbol" || baseFont == "ZapfDingbats" {
		return nil
	}
	base := "StandardEncoding"
	if subtype == "TrueType" {
		base = "WinAnsiEncoding"
	}
	var differences model.Array
	encObj := fontDict[model.Name("Encoding")]
	if ref, ok := encObj.(model.Ref); ok && src != nil {
		if resolved, err := src.Resolve(ref); err == nil {
			encObj = resolved
		}
	}
	switch e := encObj.(type) {
	case model.Name:
		base = string(e)
	case model.Dict:
		if b, ok := e[model.Name("BaseEncoding")].(model.Name); ok {
			base = string(b)
		}
		differences, _ = e[model.Name("Differences")].(model.Array)
	}
	runes := baseEncoding(base)
	var table [256]string
	for i, r := range runes {
		if r != 0 {
			table[i] = string(r)
		}
	}
	code := -1
	for _, item := range differences {
		switch v := item.(type) {
		case model.Integer:
			code = int(v)
		case model.Real:
			code = int(v)
		case model.Name:
			if code >= 0 && code < 256 {
				if text := glyphRune(string(v)); text != "" {
					table[code] = text
				}
			}
			code++
		}
	}
	return &table
}
