package model

import (
	"bytes"
	"unicode/utf16"
	"unicode/utf8"
)

// Text string encodings (ISO 32000-2 §7.9.2.2): a text string is PDFDocEncoding,
// UTF-16BE with a byte order mark, or, since PDF 2.0, UTF-8 with a byte order mark.

var (
	bomUTF16BE = []byte{0xFE, 0xFF}
	bomUTF8    = []byte{0xEF, 0xBB, 0xBF}
)

// pdfDocHigh maps PDFDocEncoding bytes 0x18–0x1F and 0x80–0xA0 to Unicode
// (ISO 32000-2 Annex D.3). Other bytes map to themselves (Latin-1), 0xAD is undefined.
var pdfDocHigh = map[byte]rune{
	0x18: 0x02D8, 0x19: 0x02C7, 0x1A: 0x02C6, 0x1B: 0x02D9, 0x1C: 0x02DD, 0x1D: 0x02DB, 0x1E: 0x02DA, 0x1F: 0x02DC,
	0x80: 0x2022, 0x81: 0x2020, 0x82: 0x2021, 0x83: 0x2026, 0x84: 0x2014, 0x85: 0x2013, 0x86: 0x0192, 0x87: 0x2044,
	0x88: 0x2039, 0x89: 0x203A, 0x8A: 0x2212, 0x8B: 0x2030, 0x8C: 0x201E, 0x8D: 0x201C, 0x8E: 0x201D, 0x8F: 0x2018,
	0x90: 0x2019, 0x91: 0x201A, 0x92: 0x2122, 0x93: 0xFB01, 0x94: 0xFB02, 0x95: 0x0141, 0x96: 0x0152, 0x97: 0x0160,
	0x98: 0x0178, 0x99: 0x017D, 0x9A: 0x0131, 0x9B: 0x0142, 0x9C: 0x0153, 0x9D: 0x0161, 0x9E: 0x017E, 0xA0: 0x20AC,
}

// DecodeTextString converts the bytes of a PDF text string to Go (UTF-8) text. It
// understands the UTF-16BE and UTF-8 byte order marks and otherwise reads
// PDFDocEncoding. Undefined PDFDocEncoding bytes decode to U+FFFD.
func DecodeTextString(b []byte) string {
	switch {
	case bytes.HasPrefix(b, bomUTF16BE):
		b = b[2:]
		units := make([]uint16, 0, len(b)/2)
		for i := 0; i+1 < len(b); i += 2 {
			units = append(units, uint16(b[i])<<8|uint16(b[i+1]))
		}
		return string(utf16.Decode(units))
	case bytes.HasPrefix(b, bomUTF8):
		return string(b[3:])
	}
	ascii := true
	for _, c := range b {
		if c >= 0x80 || (c < 0x20 && c != '\t' && c != '\n' && c != '\r') {
			ascii = false
			break
		}
	}
	if ascii {
		return string(b)
	}
	out := make([]rune, 0, len(b))
	for _, c := range b {
		if r, ok := pdfDocHigh[c]; ok {
			out = append(out, r)
		} else if c == 0x7F || c == 0x9F || c == 0xAD || (c < 0x18 && c != '\t' && c != '\n' && c != '\r') {
			out = append(out, utf8.RuneError)
		} else {
			out = append(out, rune(c))
		}
	}
	return string(out)
}

// Text returns the string decoded as a PDF text string.
func (s String) Text() string { return DecodeTextString([]byte(s)) }

// Text returns the string decoded as a PDF text string.
func (s HexString) Text() string { return DecodeTextString([]byte(s)) }

// IsTextStringBOM reports whether b starts with a UTF-16BE or UTF-8 byte order mark.
func IsTextStringBOM(b []byte) bool {
	return bytes.HasPrefix(b, bomUTF16BE) || bytes.HasPrefix(b, bomUTF8)
}

// EncodeTextString returns s as a PDF text string. Pure ASCII is written as is (it is
// the same in every text encoding). Other text is written as UTF-8 with a byte order
// mark when utf8BOM is set (PDF 2.0) and as UTF-16BE with a byte order mark otherwise,
// which every PDF reader understands.
func EncodeTextString(s string, utf8BOM bool) String {
	ascii := true
	for i := 0; i < len(s); i++ {
		if s[i] >= 0x80 {
			ascii = false
			break
		}
	}
	if ascii {
		return String(s)
	}
	if utf8BOM {
		return String(append(append([]byte(nil), bomUTF8...), s...))
	}
	units := utf16.Encode([]rune(s))
	out := make([]byte, 2, 2+2*len(units))
	copy(out, bomUTF16BE)
	for _, u := range units {
		out = append(out, byte(u>>8), byte(u))
	}
	return String(out)
}
