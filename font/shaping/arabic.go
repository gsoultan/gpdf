// Copyright (c) 2026 Gembit Soultan Shirazi <gembit.soultan@gmail.com>

package shaping

import "unicode"

// JoiningType is the Unicode Joining_Type of a character (UAX #44 / ArabicShaping.txt).
type JoiningType uint8

const (
	// NonJoining characters never connect to a neighbour (U).
	NonJoining JoiningType = iota
	// Transparent characters (combining marks) are skipped when joining (T).
	Transparent
	// RightJoining letters connect only to the preceding letter (R).
	RightJoining
	// DualJoining letters connect on both sides (D).
	DualJoining
	// JoinCausing characters join on both sides without changing form (C):
	// tatweel and the zero width joiner.
	JoinCausing
)

// Form is the contextual form a cursive letter takes.
type Form uint8

const (
	// Isolated is the standalone form (OpenType feature isol).
	Isolated Form = iota
	// Final connects to the preceding letter (fina).
	Final
	// Initial connects to the following letter (init).
	Initial
	// Medial connects on both sides (medi).
	Medial
)

// FeatureTag returns the OpenType feature tag that selects the form.
func (f Form) FeatureTag() string {
	switch f {
	case Final:
		return "fina"
	case Initial:
		return "init"
	case Medial:
		return "medi"
	}
	return "isol"
}

// rightJoining lists the Arabic-block letters that connect only on the right
// (ArabicShaping.txt, Joining_Type=R). Everything else in the Arabic letter
// ranges is dual-joining unless listed in nonJoining.
var rightJoining = map[rune]bool{
	0x0622: true, 0x0623: true, 0x0624: true, 0x0625: true, 0x0627: true,
	0x0629: true, 0x062F: true, 0x0630: true, 0x0631: true, 0x0632: true,
	0x0648: true, 0x0671: true, 0x0672: true, 0x0673: true, 0x0675: true,
	0x0676: true, 0x0677: true, 0x0688: true, 0x0689: true, 0x068A: true,
	0x068B: true, 0x068C: true, 0x068D: true, 0x068E: true, 0x068F: true,
	0x0690: true, 0x0691: true, 0x0692: true, 0x0693: true, 0x0694: true,
	0x0695: true, 0x0696: true, 0x0697: true, 0x0698: true, 0x0699: true,
	0x06C0: true, 0x06C2: true, 0x06C3: true, 0x06C4: true, 0x06C5: true,
	0x06C6: true, 0x06C7: true, 0x06C8: true, 0x06C9: true, 0x06CA: true,
	0x06CB: true, 0x06CD: true, 0x06CF: true, 0x06D2: true, 0x06D3: true,
	0x06D5: true, 0x06EE: true, 0x06EF: true,
	// Arabic Supplement and Extended-A right-joining letters.
	0x0759: true, 0x075A: true, 0x075B: true, 0x076B: true, 0x076C: true,
	0x0771: true, 0x0773: true, 0x0774: true, 0x0778: true, 0x0779: true,
	0x08AA: true, 0x08AB: true, 0x08AC: true, 0x08AE: true, 0x08B1: true,
	0x08B2: true, 0x08B9: true,
}

var nonJoining = map[rune]bool{
	0x0621: true, 0x0674: true, 0x06D4: true, 0x066D: true, 0x06DD: true,
}

// JoiningTypeOf returns the joining type of r. Characters outside the Arabic
// script are NonJoining, except combining marks and format characters, which
// are Transparent so that they do not break a join.
func JoiningTypeOf(r rune) JoiningType {
	switch {
	case r == 0x0640, r == 0x200D: // tatweel, ZWJ
		return JoinCausing
	case r == 0x200C: // ZWNJ
		return NonJoining
	case unicode.Is(unicode.Mn, r), unicode.Is(unicode.Me, r), unicode.Is(unicode.Cf, r):
		return Transparent
	case nonJoining[r]:
		return NonJoining
	case rightJoining[r]:
		return RightJoining
	case r >= 0x0620 && r <= 0x064A, r >= 0x066E && r <= 0x06D3 && r != 0x06D4,
		r == 0x06D5, r >= 0x06EE && r <= 0x06EF, r >= 0x06FA && r <= 0x06FC, r == 0x06FF,
		r >= 0x0750 && r <= 0x077F, r >= 0x08A0 && r <= 0x08C7:
		if r >= 0x0660 && r <= 0x066C || r == 0x066D || r >= 0x06F0 && r <= 0x06F9 || r == 0x06E9 || r == 0x06DE {
			return NonJoining
		}
		return DualJoining
	}
	return NonJoining
}

// ArabicForms returns the contextual form of every rune of logical-order text.
// Non-joining characters and marks report Isolated; callers only act on runes
// whose joining type is RightJoining or DualJoining.
func ArabicForms(runes []rune) []Form {
	forms := make([]Form, len(runes))
	types := make([]JoiningType, len(runes))
	for i, r := range runes {
		types[i] = JoiningTypeOf(r)
	}
	prevVisible := func(i int) JoiningType {
		for j := i - 1; j >= 0; j-- {
			if types[j] != Transparent {
				return types[j]
			}
		}
		return NonJoining
	}
	nextVisible := func(i int) JoiningType {
		for j := i + 1; j < len(types); j++ {
			if types[j] != Transparent {
				return types[j]
			}
		}
		return NonJoining
	}
	for i, t := range types {
		switch t {
		case RightJoining:
			p := prevVisible(i)
			if p == DualJoining || p == JoinCausing {
				forms[i] = Final
			}
		case DualJoining:
			p := prevVisible(i)
			nx := nextVisible(i)
			joinPrev := p == DualJoining || p == JoinCausing
			joinNext := nx == DualJoining || nx == RightJoining || nx == JoinCausing
			switch {
			case joinPrev && joinNext:
				forms[i] = Medial
			case joinNext:
				forms[i] = Initial
			case joinPrev:
				forms[i] = Final
			}
		}
	}
	return forms
}

// presentationForms maps a letter to its Arabic Presentation Forms-B (and a
// few Forms-A) code points in the order isolated, final, initial, medial. A
// zero entry means the letter has no such form.
var presentationForms = map[rune][4]rune{
	0x0621: {0xFE80},
	0x0622: {0xFE81, 0xFE82},
	0x0623: {0xFE83, 0xFE84},
	0x0624: {0xFE85, 0xFE86},
	0x0625: {0xFE87, 0xFE88},
	0x0626: {0xFE89, 0xFE8A, 0xFE8B, 0xFE8C},
	0x0627: {0xFE8D, 0xFE8E},
	0x0628: {0xFE8F, 0xFE90, 0xFE91, 0xFE92},
	0x0629: {0xFE93, 0xFE94},
	0x062A: {0xFE95, 0xFE96, 0xFE97, 0xFE98},
	0x062B: {0xFE99, 0xFE9A, 0xFE9B, 0xFE9C},
	0x062C: {0xFE9D, 0xFE9E, 0xFE9F, 0xFEA0},
	0x062D: {0xFEA1, 0xFEA2, 0xFEA3, 0xFEA4},
	0x062E: {0xFEA5, 0xFEA6, 0xFEA7, 0xFEA8},
	0x062F: {0xFEA9, 0xFEAA},
	0x0630: {0xFEAB, 0xFEAC},
	0x0631: {0xFEAD, 0xFEAE},
	0x0632: {0xFEAF, 0xFEB0},
	0x0633: {0xFEB1, 0xFEB2, 0xFEB3, 0xFEB4},
	0x0634: {0xFEB5, 0xFEB6, 0xFEB7, 0xFEB8},
	0x0635: {0xFEB9, 0xFEBA, 0xFEBB, 0xFEBC},
	0x0636: {0xFEBD, 0xFEBE, 0xFEBF, 0xFEC0},
	0x0637: {0xFEC1, 0xFEC2, 0xFEC3, 0xFEC4},
	0x0638: {0xFEC5, 0xFEC6, 0xFEC7, 0xFEC8},
	0x0639: {0xFEC9, 0xFECA, 0xFECB, 0xFECC},
	0x063A: {0xFECD, 0xFECE, 0xFECF, 0xFED0},
	0x0641: {0xFED1, 0xFED2, 0xFED3, 0xFED4},
	0x0642: {0xFED5, 0xFED6, 0xFED7, 0xFED8},
	0x0643: {0xFED9, 0xFEDA, 0xFEDB, 0xFEDC},
	0x0644: {0xFEDD, 0xFEDE, 0xFEDF, 0xFEE0},
	0x0645: {0xFEE1, 0xFEE2, 0xFEE3, 0xFEE4},
	0x0646: {0xFEE5, 0xFEE6, 0xFEE7, 0xFEE8},
	0x0647: {0xFEE9, 0xFEEA, 0xFEEB, 0xFEEC},
	0x0648: {0xFEED, 0xFEEE},
	0x0649: {0xFEEF, 0xFEF0},
	0x064A: {0xFEF1, 0xFEF2, 0xFEF3, 0xFEF4},
	// Presentation Forms-A: Persian, Urdu and other extended letters.
	0x0671: {0xFB50, 0xFB51},
	0x0679: {0xFB66, 0xFB67, 0xFB68, 0xFB69},
	0x067A: {0xFB5E, 0xFB5F, 0xFB60, 0xFB61},
	0x067B: {0xFB52, 0xFB53, 0xFB54, 0xFB55},
	0x067E: {0xFB56, 0xFB57, 0xFB58, 0xFB59},
	0x067F: {0xFB62, 0xFB63, 0xFB64, 0xFB65},
	0x0680: {0xFB5A, 0xFB5B, 0xFB5C, 0xFB5D},
	0x0683: {0xFB76, 0xFB77, 0xFB78, 0xFB79},
	0x0684: {0xFB72, 0xFB73, 0xFB74, 0xFB75},
	0x0686: {0xFB7A, 0xFB7B, 0xFB7C, 0xFB7D},
	0x0687: {0xFB7E, 0xFB7F, 0xFB80, 0xFB81},
	0x0688: {0xFB88, 0xFB89},
	0x068C: {0xFB84, 0xFB85},
	0x068D: {0xFB82, 0xFB83},
	0x068E: {0xFB86, 0xFB87},
	0x0691: {0xFB8C, 0xFB8D},
	0x0698: {0xFB8A, 0xFB8B},
	0x06A4: {0xFB6A, 0xFB6B, 0xFB6C, 0xFB6D},
	0x06A6: {0xFB6E, 0xFB6F, 0xFB70, 0xFB71},
	0x06A9: {0xFB8E, 0xFB8F, 0xFB90, 0xFB91},
	0x06AD: {0xFBD3, 0xFBD4, 0xFBD5, 0xFBD6},
	0x06AF: {0xFB92, 0xFB93, 0xFB94, 0xFB95},
	0x06B1: {0xFB9A, 0xFB9B, 0xFB9C, 0xFB9D},
	0x06B3: {0xFB96, 0xFB97, 0xFB98, 0xFB99},
	0x06BA: {0xFB9E, 0xFB9F},
	0x06BB: {0xFBA0, 0xFBA1, 0xFBA2, 0xFBA3},
	0x06BE: {0xFBAA, 0xFBAB, 0xFBAC, 0xFBAD},
	0x06C0: {0xFBA4, 0xFBA5},
	0x06C1: {0xFBA6, 0xFBA7, 0xFBA8, 0xFBA9},
	0x06C5: {0xFBE0, 0xFBE1},
	0x06C6: {0xFBD9, 0xFBDA},
	0x06C7: {0xFBD7, 0xFBD8},
	0x06C8: {0xFBDB, 0xFBDC},
	0x06C9: {0xFBE2, 0xFBE3},
	0x06CB: {0xFBDE, 0xFBDF},
	0x06CC: {0xFBFC, 0xFBFD, 0xFBFE, 0xFBFF},
	0x06D0: {0xFBE4, 0xFBE5, 0xFBE6, 0xFBE7},
	0x06D2: {0xFBAE, 0xFBAF},
	0x06D3: {0xFBB0, 0xFBB1},
}

// PresentationForm returns the Arabic presentation form code point of letter r
// in the given form, or 0 when Unicode encodes none.
func PresentationForm(r rune, form Form) rune {
	forms, ok := presentationForms[r]
	if !ok {
		return 0
	}
	return forms[form]
}

// LamAlefLigature returns the presentation-form ligature for lam followed by
// the alef variant alef, in isolated or final form, or 0 when alef is not one
// of the four alefs that form a mandatory ligature.
func LamAlefLigature(alef rune, final bool) rune {
	var base rune
	switch alef {
	case 0x0622:
		base = 0xFEF5
	case 0x0623:
		base = 0xFEF7
	case 0x0625:
		base = 0xFEF9
	case 0x0627:
		base = 0xFEFB
	default:
		return 0
	}
	if final {
		return base + 1
	}
	return base
}

// IsArabicLetter reports whether r is a cursive Arabic-script letter that may
// change form in context.
func IsArabicLetter(r rune) bool {
	t := JoiningTypeOf(r)
	return t == RightJoining || t == DualJoining
}
