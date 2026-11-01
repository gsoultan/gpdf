package security

import (
	"errors"
	"fmt"
	"unicode"
	"unicode/utf8"
)

// ErrPasswordPrep is returned when a password contains characters that SASLprep
// (RFC 4013) prohibits.
var ErrPasswordPrep = errors.New("security: password contains characters SASLprep prohibits")

// saslprep applies the SASLprep profile of stringprep to a password, as ISO 32000-2
// §7.6.4.3.3 requires before hashing an R 6 password: non-ASCII spaces map to U+0020,
// "commonly mapped to nothing" characters are removed, prohibited characters are
// rejected, and the bidirectional rule is checked. Unicode NFKC normalisation is not
// applied (the standard library has no normaliser), so passwords are hashed in the
// form they were typed; text from ordinary keyboards is already in NFKC form.
func saslprep(s string) (string, error) {
	if !utf8.ValidString(s) {
		return "", fmt.Errorf("%w: invalid UTF-8", ErrPasswordPrep)
	}
	out := make([]rune, 0, len(s))
	for _, r := range s {
		switch {
		case isNonASCIISpace(r):
			out = append(out, ' ')
		case isMappedToNothing(r):
		default:
			out = append(out, r)
		}
	}
	for _, r := range out {
		if isProhibited(r) {
			return "", fmt.Errorf("%w: U+%04X", ErrPasswordPrep, r)
		}
	}
	if err := checkBidi(out); err != nil {
		return "", err
	}
	return string(out), nil
}

// isNonASCIISpace is stringprep table C.1.2.
func isNonASCIISpace(r rune) bool {
	return r == 0x00A0 || r == 0x1680 || (r >= 0x2000 && r <= 0x200B) || r == 0x202F || r == 0x205F || r == 0x3000
}

// isMappedToNothing is stringprep table B.1.
func isMappedToNothing(r rune) bool {
	return r == 0x00AD || r == 0x034F || r == 0x1806 || (r >= 0x180B && r <= 0x180D) ||
		(r >= 0x200B && r <= 0x200D) || r == 0x2060 || (r >= 0xFE00 && r <= 0xFE0F) || r == 0xFEFF
}

// isProhibited covers stringprep tables C.1.2, C.2.1, C.2.2, C.3, C.4, C.5, C.6, C.7,
// C.8 and C.9, which SASLprep prohibits.
func isProhibited(r rune) bool {
	switch {
	case isNonASCIISpace(r):
		return true
	case r < 0x20 || r == 0x7F: // C.2.1
		return true
	case r >= 0x80 && r <= 0x9F: // C.2.2
		return true
	case r == 0x06DD || r == 0x070F || r == 0x180E || r == 0x200C || r == 0x200D ||
		r == 0x2028 || r == 0x2029 || (r >= 0x2060 && r <= 0x2063) || (r >= 0x206A && r <= 0x206F) ||
		r == 0xFEFF || (r >= 0xFFF9 && r <= 0xFFFC) || (r >= 0x1D173 && r <= 0x1D17A): // C.2.2
		return true
	case unicode.Is(unicode.Co, r): // C.3 private use
		return true
	case (r >= 0xFDD0 && r <= 0xFDEF) || r&0xFFFE == 0xFFFE: // C.4 non-characters
		return true
	case r >= 0xD800 && r <= 0xDFFF: // C.5 surrogates
		return true
	case r == 0xFFFD: // C.6
		return true
	case r >= 0x2FF0 && r <= 0x2FFB: // C.7
		return true
	case r == 0x0340 || r == 0x0341 || r == 0x200E || r == 0x200F || (r >= 0x202A && r <= 0x202E): // C.8
		return true
	case r == 0xE0001 || (r >= 0xE0020 && r <= 0xE007F): // C.9
		return true
	}
	return false
}

// isRandAL approximates the stringprep D.1 table (characters with bidirectional
// class R or AL) by script: Hebrew, Arabic, Syriac, Thaana and N'Ko.
func isRandAL(r rune) bool {
	return unicode.In(r, unicode.Hebrew, unicode.Arabic, unicode.Syriac, unicode.Thaana, unicode.Nko)
}

// isLCat approximates table D.2 (bidirectional class L): letters and digits that are
// not right-to-left.
func isLCat(r rune) bool {
	return !isRandAL(r) && (unicode.IsLetter(r) || (unicode.IsDigit(r) && r > 0x7F))
}

// checkBidi applies RFC 3454 §6: a string with RandALCat characters must not contain
// LCat characters and must start and end with a RandALCat character.
func checkBidi(rs []rune) error {
	hasRAL, hasL := false, false
	for _, r := range rs {
		if isRandAL(r) {
			hasRAL = true
		} else if isLCat(r) {
			hasL = true
		}
	}
	if !hasRAL {
		return nil
	}
	if hasL || !isRandAL(rs[0]) || !isRandAL(rs[len(rs)-1]) {
		return fmt.Errorf("%w: mixed bidirectional text", ErrPasswordPrep)
	}
	return nil
}
