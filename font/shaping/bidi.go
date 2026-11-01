// Copyright (c) 2026 Gembit Soultan Shirazi <gembit.soultan@gmail.com>

// Package shaping implements the text preparation steps that sit between a
// Unicode string and the glyphs written into a PDF content stream: the Unicode
// bidirectional algorithm (UAX #9) that turns logical text into visual runs,
// bracket mirroring, and Arabic cursive joining.
//
// The bidi implementation covers the implicit part of UAX #9 (rules P2–P3,
// W1–W7, N1–N2, I1–I2, L1–L2 and L4). Explicit embedding and isolate controls
// (LRE, RLE, LRO, RLO, PDF, LRI, RLI, FSI, PDI) are ignored, which is what a
// PDF generator needs for ordinary mixed-script lines; the directional marks
// LRM and RLM are honoured.
package shaping

import "unicode"

// Direction is the base direction of a paragraph.
type Direction int

const (
	// Auto picks the base direction from the first strong character (UAX #9
	// rule P2) and defaults to left-to-right.
	Auto Direction = iota
	// LeftToRight forces a left-to-right paragraph.
	LeftToRight
	// RightToLeft forces a right-to-left paragraph.
	RightToLeft
)

// Run is a maximal visual-order sequence of characters with one direction.
// Text keeps the characters in logical order; a renderer draws an RTL run with
// its characters reversed.
type Run struct {
	Text string
	RTL  bool
}

// bidi character classes (UAX #9 Table 4).
type bidiClass uint8

const (
	classL bidiClass = iota
	classR
	classAL
	classEN
	classES
	classET
	classAN
	classCS
	classNSM
	classBN
	classB
	classS
	classWS
	classON
)

func classify(r rune) bidiClass {
	switch {
	case r == '\n', r == '\r', r == 0x1C, r == 0x1D, r == 0x1E, r == 0x85, r == 0x2029:
		return classB
	case r == '\t', r == 0x0B, r == 0x1F:
		return classS
	case r == ' ', r == 0x0C, r == 0x1680, r == 0x2028, r == 0x3000, r >= 0x2000 && r <= 0x200A:
		return classWS
	case r == 0x200E: // LRM
		return classL
	case r == 0x200F: // RLM
		return classR
	case r == 0x061C: // ALM
		return classAL
	case r >= '0' && r <= '9', r >= 0x06F0 && r <= 0x06F9, r == 0xB2, r == 0xB3, r == 0xB9, r >= 0x2070 && r <= 0x2079 && r != 0x2071, r >= 0x2080 && r <= 0x2089, r >= 0xFF10 && r <= 0xFF19:
		return classEN
	case r >= 0x0660 && r <= 0x0669, r == 0x060C, r == 0x066B, r == 0x066C, r >= 0x0600 && r <= 0x0605, r == 0x06DD, r >= 0x10E60 && r <= 0x10E7E:
		if r == 0x060C {
			return classCS
		}
		return classAN
	case r == '+', r == '-', r == 0x207A, r == 0x207B, r == 0x208A, r == 0x208B, r == 0x2212, r == 0xFB29, r == 0xFF0B, r == 0xFF0D:
		return classES
	case r == '#', r == '$', r == '%', r == 0xB0, r == 0xB1, r == 0x2030, r == 0x2031, r == 0x2032, r == 0x2033, r == 0x2034, r == 0x20A0 || r >= 0x20A1 && r <= 0x20CF, r == 0xA2, r == 0xA3, r == 0xA4, r == 0xA5, r == 0x066A, r == 0x2212:
		return classET
	case r == ',', r == '.', r == '/', r == ':', r == 0xA0, r == 0x2044, r == 0xFE50, r == 0xFE52, r == 0xFE55, r == 0xFF0C, r == 0xFF0E, r == 0xFF0F, r == 0xFF1A, r == 0x060D:
		return classCS
	case r < 0x20 || r == 0x7F, r >= 0x80 && r <= 0x9F, r == 0xAD, r >= 0x200B && r <= 0x200D, r == 0x2060, r == 0xFEFF, r >= 0x2066 && r <= 0x2069, r >= 0x202A && r <= 0x202E:
		return classBN
	case unicode.Is(unicode.Mn, r), unicode.Is(unicode.Me, r):
		return classNSM
	case unicode.Is(unicode.Sc, r):
		return classET
	case isArabicLetterScript(r):
		return classAL
	case unicode.Is(unicode.Hebrew, r), r >= 0x0590 && r <= 0x05FF, r >= 0xFB1D && r <= 0xFB4F, unicode.Is(unicode.Nko, r), unicode.Is(unicode.Samaritan, r), unicode.Is(unicode.Mandaic, r):
		return classR
	case unicode.IsLetter(r), unicode.Is(unicode.Mc, r), unicode.IsDigit(r):
		return classL
	default:
		return classON
	}
}

// isArabicLetterScript reports whether r belongs to a script whose strong
// characters are class AL: Arabic, Syriac, Thaana, and the Arabic
// presentation form blocks.
func isArabicLetterScript(r rune) bool {
	switch {
	case r >= 0x0600 && r <= 0x06FF:
		// Digits, the Arabic comma, percent sign and numeric separators are
		// handled before this function is reached, so the rest of the block is
		// AL except the few ON characters.
		switch r {
		case 0x060C, 0x060D, 0x066A, 0x066B, 0x066C, 0x06DD, 0x06DE, 0x06E9:
			return false
		}
		if r >= 0x0660 && r <= 0x0669 || r >= 0x06F0 && r <= 0x06F9 {
			return false
		}
		return !unicode.Is(unicode.Mn, r)
	case r >= 0x0750 && r <= 0x077F, r >= 0x08A0 && r <= 0x08FF, r >= 0xFB50 && r <= 0xFDFF, r >= 0xFE70 && r <= 0xFEFF:
		return !unicode.Is(unicode.Mn, r)
	case unicode.Is(unicode.Syriac, r), unicode.Is(unicode.Thaana, r):
		return !unicode.Is(unicode.Mn, r)
	}
	return false
}

// BaseDirection returns the paragraph direction implied by the first strong
// character of s (UAX #9 rule P2), or LeftToRight when there is none.
func BaseDirection(s string) Direction {
	for _, r := range s {
		switch classify(r) {
		case classL:
			return LeftToRight
		case classR, classAL:
			return RightToLeft
		}
	}
	return LeftToRight
}

// IsRTL reports whether r is a strong right-to-left character (class R or AL).
func IsRTL(r rune) bool {
	c := classify(r)
	return c == classR || c == classAL
}

// Reorder resolves the bidi embedding levels of one line of logical text and
// returns its characters grouped into visual-order runs, left to right. Within
// an RTL run the text stays in logical order and mirrored characters such as
// parentheses are already swapped for their mirror image.
//
// Lines, not paragraphs, should be passed: UAX #9 reorders each line
// separately after line breaking.
func Reorder(s string, dir Direction) []Run {
	runes := []rune(s)
	if len(runes) == 0 {
		return nil
	}
	levels := Levels(runes, dir)
	order := visualOrder(levels)

	var runs []Run
	for i := 0; i < len(order); {
		rtl := levels[order[i]]%2 == 1
		j := i
		for j < len(order) && levels[order[j]]%2 == 1 == rtl {
			j++
		}
		idx := order[i:j]
		text := make([]rune, len(idx))
		if rtl {
			// Visual order of an RTL run is reversed logical order.
			for k, li := range idx {
				text[len(idx)-1-k] = Mirror(runes[li])
			}
		} else {
			for k, li := range idx {
				text[k] = runes[li]
			}
		}
		runs = append(runs, Run{Text: string(text), RTL: rtl})
		i = j
	}
	return runs
}

// Levels returns the resolved embedding level of each rune of a line.
func Levels(runes []rune, dir Direction) []uint8 {
	n := len(runes)
	types := make([]bidiClass, n)
	for i, r := range runes {
		types[i] = classify(r)
	}

	base := uint8(0)
	switch dir {
	case RightToLeft:
		base = 1
	case Auto:
		if BaseDirection(string(runes)) == RightToLeft {
			base = 1
		}
	}
	sos := classL
	if base == 1 {
		sos = classR
	}

	// Explicit directional controls are not interpreted; they are already
	// class BN and are dropped from the resolution by the rules below.

	// W1: non-spacing marks take the type of the previous character.
	prev := sos
	for i := range n {
		switch types[i] {
		case classNSM:
			types[i] = prev
		case classBN:
			continue
		}
		prev = types[i]
	}
	// W2: European numbers after Arabic letters become Arabic numbers.
	// W3: Arabic letters become R.
	lastStrong := sos
	for i := range n {
		switch types[i] {
		case classL, classR:
			lastStrong = types[i]
		case classAL:
			lastStrong = classAL
			types[i] = classR
		case classEN:
			if lastStrong == classAL {
				types[i] = classAN
			}
		}
	}
	// W4: a single separator between two numbers of the same kind joins them.
	for i := 1; i < n-1; i++ {
		switch types[i] {
		case classES:
			if types[i-1] == classEN && types[i+1] == classEN {
				types[i] = classEN
			}
		case classCS:
			if types[i-1] == classEN && types[i+1] == classEN {
				types[i] = classEN
			} else if types[i-1] == classAN && types[i+1] == classAN {
				types[i] = classAN
			}
		}
	}
	// W5: terminators adjacent to European numbers become European numbers.
	for i := 0; i < n; i++ {
		if types[i] != classET {
			continue
		}
		j := i
		for j < n && types[j] == classET {
			j++
		}
		adjacent := (i > 0 && types[i-1] == classEN) || (j < n && types[j] == classEN)
		if adjacent {
			for k := i; k < j; k++ {
				types[k] = classEN
			}
		}
		i = j - 1
	}
	// W6: remaining separators and terminators are neutral.
	for i := range n {
		switch types[i] {
		case classES, classET, classCS:
			types[i] = classON
		}
	}
	// W7: European numbers after L become L.
	lastStrong = sos
	for i := range n {
		switch types[i] {
		case classL, classR:
			lastStrong = types[i]
		case classEN:
			if lastStrong == classL {
				types[i] = classL
			}
		}
	}

	// N1/N2: neutrals between characters of the same direction take that
	// direction; numbers count as R for this purpose. Other neutrals take the
	// embedding direction.
	eos := sos
	isNeutral := func(c bidiClass) bool {
		return c == classB || c == classS || c == classWS || c == classON || c == classBN
	}
	dirOf := func(c bidiClass) bidiClass {
		if c == classEN || c == classAN || c == classR {
			return classR
		}
		return classL
	}
	for i := 0; i < n; i++ {
		if !isNeutral(types[i]) {
			continue
		}
		j := i
		for j < n && isNeutral(types[j]) {
			j++
		}
		before := sos
		if i > 0 {
			before = dirOf(types[i-1])
		}
		after := eos
		if j < n {
			after = dirOf(types[j])
		}
		resolved := sos
		if before == after {
			resolved = before
		}
		for k := i; k < j; k++ {
			if types[k] == classB || types[k] == classS || types[k] == classWS {
				// Remembered for L1: whitespace keeps its own class marker in
				// levels below; the type only drives I1/I2 here.
				types[k] = resolved | 0x80
				continue
			}
			types[k] = resolved
		}
		i = j - 1
	}

	// I1/I2: implicit levels.
	levels := make([]uint8, n)
	for i := range n {
		t := types[i] &^ 0x80
		lvl := base
		if base%2 == 0 {
			switch t {
			case classR:
				lvl = 1
			case classAN, classEN:
				lvl = 2
			}
		} else {
			switch t {
			case classL, classEN, classAN:
				lvl = 2
			}
		}
		levels[i] = lvl
	}
	// L1: segment separators, paragraph separators, and trailing whitespace
	// (plus whitespace before a separator) reset to the paragraph level.
	trailing := true
	for i := n - 1; i >= 0; i-- {
		c := classify(runes[i])
		switch c {
		case classB, classS:
			levels[i] = base
			trailing = true
		case classWS, classBN:
			if trailing {
				levels[i] = base
			}
		default:
			trailing = false
		}
	}
	return levels
}

// visualOrder applies rule L2: from the highest level down to the lowest odd
// level, reverse every maximal run of characters at that level or higher.
func visualOrder(levels []uint8) []int {
	n := len(levels)
	order := make([]int, n)
	for i := range order {
		order[i] = i
	}
	var maxLevel, minOdd uint8 = 0, 255
	for _, l := range levels {
		maxLevel = max(maxLevel, l)
		if l%2 == 1 && l < minOdd {
			minOdd = l
		}
	}
	if minOdd == 255 {
		return order
	}
	for lvl := maxLevel; lvl >= minOdd; lvl-- {
		for i := 0; i < n; {
			if levels[order[i]] < lvl {
				i++
				continue
			}
			j := i
			for j < n && levels[order[j]] >= lvl {
				j++
			}
			for a, b := i, j-1; a < b; a, b = a+1, b-1 {
				order[a], order[b] = order[b], order[a]
			}
			i = j
		}
		if lvl == 0 {
			break
		}
	}
	return order
}

// mirrorPairs lists the Bidi_Mirroring_Glyph pairs a document generator meets
// in practice: brackets, quotation marks and the common mathematical
// relations.
var mirrorPairs = map[rune]rune{
	'(': ')', ')': '(', '[': ']', ']': '[', '{': '}', '}': '{', '<': '>', '>': '<',
	0xAB: 0xBB, 0xBB: 0xAB, 0x2039: 0x203A, 0x203A: 0x2039,
	0x2045: 0x2046, 0x2046: 0x2045, 0x207D: 0x207E, 0x207E: 0x207D,
	0x208D: 0x208E, 0x208E: 0x208D, 0x2208: 0x220B, 0x220B: 0x2208,
	0x2264: 0x2265, 0x2265: 0x2264, 0x2282: 0x2283, 0x2283: 0x2282,
	0x2286: 0x2287, 0x2287: 0x2286, 0x2329: 0x232A, 0x232A: 0x2329,
	0x27E6: 0x27E7, 0x27E7: 0x27E6, 0x27E8: 0x27E9, 0x27E9: 0x27E8,
	0x27EA: 0x27EB, 0x27EB: 0x27EA, 0x2983: 0x2984, 0x2984: 0x2983,
	0x2985: 0x2986, 0x2986: 0x2985, 0x3008: 0x3009, 0x3009: 0x3008,
	0x300A: 0x300B, 0x300B: 0x300A, 0x300C: 0x300D, 0x300D: 0x300C,
	0x300E: 0x300F, 0x300F: 0x300E, 0x3010: 0x3011, 0x3011: 0x3010,
	0xFE59: 0xFE5A, 0xFE5A: 0xFE59, 0xFE5B: 0xFE5C, 0xFE5C: 0xFE5B,
	0xFE5D: 0xFE5E, 0xFE5E: 0xFE5D, 0xFF08: 0xFF09, 0xFF09: 0xFF08,
	0xFF1C: 0xFF1E, 0xFF1E: 0xFF1C, 0xFF3B: 0xFF3D, 0xFF3D: 0xFF3B,
	0xFF5B: 0xFF5D, 0xFF5D: 0xFF5B, 0xFF5F: 0xFF60, 0xFF60: 0xFF5F,
	0xFF62: 0xFF63, 0xFF63: 0xFF62,
}

// Mirror returns the mirrored counterpart of r for display in right-to-left
// text (UAX #9 rule L4), or r itself when it has no mirror image.
func Mirror(r rune) rune {
	if m, ok := mirrorPairs[r]; ok {
		return m
	}
	return r
}
