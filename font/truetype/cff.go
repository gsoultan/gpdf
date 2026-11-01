package truetype

import (
	"encoding/binary"
	"fmt"
)

// cffInfo holds what a PDF writer needs from the Compact Font Format program
// of an OpenType font with PostScript outlines (an "OTTO" font): whether the
// program is CID-keyed and, if so, its Registry/Ordering/Supplement and the
// glyph-to-CID charset.
type cffInfo struct {
	data       []byte
	cidKeyed   bool
	registry   string
	ordering   string
	supplement int
	numGlyphs  int
	gidToCID   []uint16
	cidToGID   map[uint16]uint16
}

// parseCFF reads the Top DICT and charset of the "CFF " table.
func (f *Font) parseCFF() error {
	d, err := f.tableData("CFF ")
	if err != nil {
		return nil // TrueType outlines
	}
	if len(d) < 4 {
		return fmt.Errorf("truetype: CFF table too short")
	}
	hdrSize := int(d[2])
	pos := hdrSize
	// Name INDEX.
	_, pos, err = cffIndex(d, pos)
	if err != nil {
		return err
	}
	topDicts, pos, err := cffIndex(d, pos)
	if err != nil {
		return err
	}
	if len(topDicts) == 0 {
		return fmt.Errorf("truetype: CFF has no Top DICT")
	}
	strings, _, err := cffIndex(d, pos)
	if err != nil {
		return err
	}
	top := parseCFFDict(topDicts[0])

	info := &cffInfo{data: d, cidToGID: map[uint16]uint16{}}
	if cs, ok := top[17]; ok && len(cs) == 1 {
		charStrings, _, err := cffIndexCount(d, int(cs[0]))
		if err != nil {
			return err
		}
		info.numGlyphs = charStrings
	}
	if ros, ok := top[12<<8|30]; ok && len(ros) == 3 {
		info.cidKeyed = true
		info.registry = cffString(strings, int(ros[0]))
		info.ordering = cffString(strings, int(ros[1]))
		info.supplement = int(ros[2])
	}
	charsetOff := 0
	if v, ok := top[15]; ok && len(v) == 1 {
		charsetOff = int(v[0])
	}
	n := info.numGlyphs
	if n == 0 {
		n = f.numGlyphs
	}
	info.gidToCID = make([]uint16, n)
	for i := range info.gidToCID {
		info.gidToCID[i] = uint16(i) // identity for the predefined charsets
	}
	if charsetOff > 2 && charsetOff < len(d) && n > 0 {
		parseCFFCharset(d[charsetOff:], info.gidToCID)
	}
	for gid, cid := range info.gidToCID {
		if _, dup := info.cidToGID[cid]; !dup {
			info.cidToGID[cid] = uint16(gid)
		}
	}
	f.cff = info
	return nil
}

// cffIndex returns the items of the INDEX structure at pos and the position
// just past it.
func cffIndex(d []byte, pos int) ([][]byte, int, error) {
	if pos+2 > len(d) {
		return nil, 0, fmt.Errorf("truetype: CFF INDEX out of bounds")
	}
	count := int(binary.BigEndian.Uint16(d[pos:]))
	if count == 0 {
		return nil, pos + 2, nil
	}
	if pos+3 > len(d) {
		return nil, 0, fmt.Errorf("truetype: CFF INDEX out of bounds")
	}
	offSize := int(d[pos+2])
	if offSize < 1 || offSize > 4 {
		return nil, 0, fmt.Errorf("truetype: CFF INDEX has invalid offSize %d", offSize)
	}
	offArray := pos + 3
	dataStart := offArray + (count+1)*offSize - 1
	if dataStart > len(d) {
		return nil, 0, fmt.Errorf("truetype: CFF INDEX offsets out of bounds")
	}
	readOff := func(i int) int {
		p := offArray + i*offSize
		v := 0
		for k := range offSize {
			v = v<<8 | int(d[p+k])
		}
		return v
	}
	items := make([][]byte, count)
	for i := range count {
		start, end := dataStart+readOff(i), dataStart+readOff(i+1)
		if start < 0 || end > len(d) || start > end {
			return nil, 0, fmt.Errorf("truetype: CFF INDEX item %d out of bounds", i)
		}
		items[i] = d[start:end]
	}
	return items, dataStart + readOff(count), nil
}

// cffIndexCount returns only the item count of an INDEX (its items are not needed).
func cffIndexCount(d []byte, pos int) (int, int, error) {
	if pos < 0 || pos+2 > len(d) {
		return 0, 0, fmt.Errorf("truetype: CFF INDEX out of bounds")
	}
	return int(binary.BigEndian.Uint16(d[pos:])), pos, nil
}

// parseCFFDict decodes a DICT into operator -> operands. Two-byte operators
// (escape 12) are keyed as 12<<8|op.
func parseCFFDict(d []byte) map[int][]float64 {
	out := map[int][]float64{}
	var operands []float64
	for i := 0; i < len(d); {
		b0 := int(d[i])
		switch {
		case b0 <= 21:
			op := b0
			i++
			if b0 == 12 && i < len(d) {
				op = 12<<8 | int(d[i])
				i++
			}
			out[op] = append([]float64(nil), operands...)
			operands = operands[:0]
		case b0 == 28:
			if i+3 > len(d) {
				return out
			}
			operands = append(operands, float64(int16(binary.BigEndian.Uint16(d[i+1:]))))
			i += 3
		case b0 == 29:
			if i+5 > len(d) {
				return out
			}
			operands = append(operands, float64(int32(binary.BigEndian.Uint32(d[i+1:]))))
			i += 5
		case b0 == 30: // real number, nibble encoded; only its presence matters here
			i++
			for i < len(d) {
				b := d[i]
				i++
				if b&0x0F == 0x0F || b>>4 == 0x0F {
					break
				}
			}
			operands = append(operands, 0)
		case b0 >= 32 && b0 <= 246:
			operands = append(operands, float64(b0-139))
			i++
		case b0 >= 247 && b0 <= 250:
			if i+2 > len(d) {
				return out
			}
			operands = append(operands, float64((b0-247)*256+int(d[i+1])+108))
			i += 2
		case b0 >= 251 && b0 <= 254:
			if i+2 > len(d) {
				return out
			}
			operands = append(operands, float64(-(b0-251)*256-int(d[i+1])-108))
			i += 2
		default:
			i++
		}
	}
	return out
}

// parseCFFCharset fills gidToCID (glyph 0 is always CID 0) from a charset
// table in format 0, 1 or 2.
func parseCFFCharset(d []byte, gidToCID []uint16) {
	if len(d) < 1 {
		return
	}
	gidToCID[0] = 0
	gid := 1
	switch d[0] {
	case 0:
		for p := 1; gid < len(gidToCID) && p+2 <= len(d); p += 2 {
			gidToCID[gid] = binary.BigEndian.Uint16(d[p:])
			gid++
		}
	case 1, 2:
		p := 1
		for gid < len(gidToCID) {
			if d[0] == 1 {
				if p+3 > len(d) {
					return
				}
				first := int(binary.BigEndian.Uint16(d[p:]))
				nLeft := int(d[p+2])
				p += 3
				for k := 0; k <= nLeft && gid < len(gidToCID); k++ {
					gidToCID[gid] = uint16(first + k)
					gid++
				}
			} else {
				if p+4 > len(d) {
					return
				}
				first := int(binary.BigEndian.Uint16(d[p:]))
				nLeft := int(binary.BigEndian.Uint16(d[p+2:]))
				p += 4
				for k := 0; k <= nLeft && gid < len(gidToCID); k++ {
					gidToCID[gid] = uint16(first + k)
					gid++
				}
			}
		}
	}
}

// cffString resolves a string identifier against the standard strings and the
// font's String INDEX.
func cffString(custom [][]byte, sid int) string {
	if sid < len(cffStandardStrings) {
		return cffStandardStrings[sid]
	}
	if i := sid - len(cffStandardStrings); i < len(custom) {
		return string(custom[i])
	}
	return ""
}

// IsCFF reports whether the font's glyphs are PostScript (CFF) outlines, i.e.
// the font is an OpenType "OTTO" font rather than TrueType.
func (f *Font) IsCFF() bool { return f.cff != nil }

// CFF returns the bare CFF font program of an OpenType CFF font, suitable for
// a /FontFile3 stream of /Subtype /CIDFontType0C, or nil for TrueType fonts.
func (f *Font) CFF() []byte {
	if f.cff == nil {
		return nil
	}
	return f.cff.data
}

// CIDSystemInfo returns the Registry, Ordering and Supplement of a CID-keyed
// CFF program. For TrueType fonts and non-CID-keyed CFF fonts, whose glyphs
// are addressed directly by glyph index, it returns Adobe-Identity-0.
func (f *Font) CIDSystemInfo() (registry, ordering string, supplement int) {
	if f.cff != nil && f.cff.cidKeyed {
		return f.cff.registry, f.cff.ordering, f.cff.supplement
	}
	return "Adobe", "Identity", 0
}

// glyphCode returns the code written into an Identity-H content stream for a
// glyph index: the CID from the charset for CID-keyed CFF fonts, the glyph
// index itself otherwise.
func (f *Font) glyphCode(gid uint16) uint16 {
	if f.cff != nil && f.cff.cidKeyed && int(gid) < len(f.cff.gidToCID) {
		return f.cff.gidToCID[gid]
	}
	return gid
}

// glyphIndex is the inverse of glyphCode.
func (f *Font) glyphIndex(code uint16) uint16 {
	if f.cff != nil && f.cff.cidKeyed {
		if gid, ok := f.cff.cidToGID[code]; ok {
			return gid
		}
		return 0
	}
	return code
}

// cffStandardStrings are the 391 predefined CFF strings (CFF spec Appendix A).
var cffStandardStrings = []string{".notdef", "space", "exclam", "quotedbl", "numbersign", "dollar", "percent", "ampersand", "quoteright", "parenleft", "parenright", "asterisk", "plus", "comma", "hyphen", "period", "slash", "zero", "one", "two", "three", "four", "five", "six", "seven", "eight", "nine", "colon", "semicolon", "less", "equal", "greater", "question", "at", "A", "B", "C", "D", "E", "F", "G", "H", "I", "J", "K", "L", "M", "N", "O", "P", "Q", "R", "S", "T", "U", "V", "W", "X", "Y", "Z", "bracketleft", "backslash", "bracketright", "asciicircum", "underscore", "quoteleft", "a", "b", "c", "d", "e", "f", "g", "h", "i", "j", "k", "l", "m", "n", "o", "p", "q", "r", "s", "t", "u", "v", "w", "x", "y", "z", "braceleft", "bar", "braceright", "asciitilde", "exclamdown", "cent", "sterling", "fraction", "yen", "florin", "section", "currency", "quotesingle", "quotedblleft", "guillemotleft", "guilsinglleft", "guilsinglright", "fi", "fl", "endash", "dagger", "daggerdbl", "periodcentered", "paragraph", "bullet", "quotesinglbase", "quotedblbase", "quotedblright", "guillemotright", "ellipsis", "perthousand", "questiondown", "grave", "acute", "circumflex", "tilde", "macron", "breve", "dotaccent", "dieresis", "ring", "cedilla", "hungarumlaut", "ogonek", "caron", "emdash", "AE", "ordfeminine", "Lslash", "Oslash", "OE", "ordmasculine", "ae", "dotlessi", "lslash", "oslash", "oe", "germandbls", "onesuperior", "logicalnot", "mu", "trademark", "Eth", "onehalf", "plusminus", "Thorn", "onequarter", "divide", "brokenbar", "degree", "thorn", "threequarters", "twosuperior", "registered", "minus", "eth", "multiply", "threesuperior", "copyright", "Aacute", "Acircumflex", "Adieresis", "Agrave", "Aring", "Atilde", "Ccedilla", "Eacute", "Ecircumflex", "Edieresis", "Egrave", "Iacute", "Icircumflex", "Idieresis", "Igrave", "Ntilde", "Oacute", "Ocircumflex", "Odieresis", "Ograve", "Otilde", "Scaron", "Uacute", "Ucircumflex", "Udieresis", "Ugrave", "Yacute", "Ydieresis", "Zcaron", "aacute", "acircumflex", "adieresis", "agrave", "aring", "atilde", "ccedilla", "eacute", "ecircumflex", "edieresis", "egrave", "iacute", "icircumflex", "idieresis", "igrave", "ntilde", "oacute", "ocircumflex", "odieresis", "ograve", "otilde", "scaron", "uacute", "ucircumflex", "udieresis", "ugrave", "yacute", "ydieresis", "zcaron", "exclamsmall", "Hungarumlautsmall", "dollaroldstyle", "dollarsuperior", "ampersandsmall", "Acutesmall", "parenleftsuperior", "parenrightsuperior", "twodotenleader", "onedotenleader", "zerooldstyle", "oneoldstyle", "twooldstyle", "threeoldstyle", "fouroldstyle", "fiveoldstyle", "sixoldstyle", "sevenoldstyle", "eightoldstyle", "nineoldstyle", "commasuperior", "threequartersemdash", "periodsuperior", "questionsmall", "asuperior", "bsuperior", "centsuperior", "dsuperior", "esuperior", "isuperior", "lsuperior", "msuperior", "nsuperior", "osuperior", "rsuperior", "ssuperior", "tsuperior", "ff", "ffi", "ffl", "parenleftinferior", "parenrightinferior", "Circumflexsmall", "hyphensuperior", "Gravesmall", "Asmall", "Bsmall", "Csmall", "Dsmall", "Esmall", "Fsmall", "Gsmall", "Hsmall", "Ismall", "Jsmall", "Ksmall", "Lsmall", "Msmall", "Nsmall", "Osmall", "Psmall", "Qsmall", "Rsmall", "Ssmall", "Tsmall", "Usmall", "Vsmall", "Wsmall", "Xsmall", "Ysmall", "Zsmall", "colonmonetary", "onefitted", "rupiah", "Tildesmall", "exclamdownsmall", "centoldstyle", "Lslashsmall", "Scaronsmall", "Zcaronsmall", "Dieresissmall", "Brevesmall", "Caronsmall", "Dotaccentsmall", "Macronsmall", "figuredash", "hypheninferior", "Ogoneksmall", "Ringsmall", "Cedillasmall", "questiondownsmall", "oneeighth", "threeeighths", "fiveeighths", "seveneighths", "onethird", "twothirds", "zerosuperior", "foursuperior", "fivesuperior", "sixsuperior", "sevensuperior", "eightsuperior", "ninesuperior", "zeroinferior", "oneinferior", "twoinferior", "threeinferior", "fourinferior", "fiveinferior", "sixinferior", "seveninferior", "eightinferior", "nineinferior", "centinferior", "dollarinferior", "periodinferior", "commainferior", "Agravesmall", "Aacutesmall", "Acircumflexsmall", "Atildesmall", "Adieresissmall", "Aringsmall", "AEsmall", "Ccedillasmall", "Egravesmall", "Eacutesmall", "Ecircumflexsmall", "Edieresissmall", "Igravesmall", "Iacutesmall", "Icircumflexsmall", "Idieresissmall", "Ethsmall", "Ntildesmall", "Ogravesmall", "Oacutesmall", "Ocircumflexsmall", "Otildesmall", "Odieresissmall", "OEsmall", "Oslashsmall", "Ugravesmall", "Uacutesmall", "Ucircumflexsmall", "Udieresissmall", "Yacutesmall", "Thornsmall", "Ydieresissmall", "001.000", "001.001", "001.002", "001.003", "Black", "Bold", "Book", "Light", "Medium", "Regular", "Roman", "Semibold"}
