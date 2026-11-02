// Copyright (c) 2026 Gembit Soultan Shirazi <gembit.soultan@gmail.com>

package shaping

import (
	"reflect"
	"testing"
)

func TestReorderMixedDirections(t *testing.T) {
	cases := []struct {
		in   string
		dir  Direction
		want []Run
	}{
		{"abc", Auto, []Run{{"abc", false}}},
		{"Hello مرحبا world", Auto, []Run{{"Hello ", false}, {"مرحبا", true}, {" world", false}}},
		// Numbers inside an RTL paragraph stay left-to-right (rule I2).
		{"שלום 42 עולם", Auto, []Run{{" עולם", true}, {"42", false}, {"שלום ", true}}},
		// Arabic-Indic context: European digits after Arabic letters become AN (W2).
		// Even-level pieces that are visually adjacent merge into one run;
		// the RTL run keeps logical order (space before the word).
		{"price في 2024", Auto, []Run{{"price 2024", false}, {"في ", true}}},
		// A forced RTL base direction puts the Latin run on the right.
		{"abc", RightToLeft, []Run{{"abc", false}}},
		{"abc def", RightToLeft, []Run{{"abc def", false}}},
		{"abc عربي", RightToLeft, []Run{{" عربي", true}, {"abc", false}}},
	}
	for _, c := range cases {
		got := Reorder(c.in, c.dir)
		if !reflect.DeepEqual(got, c.want) {
			t.Errorf("Reorder(%q, %v) = %v, want %v", c.in, c.dir, got, c.want)
		}
	}
}

func TestReorderMirrorsBrackets(t *testing.T) {
	runs := Reorder("(مرحبا)", Auto)
	if len(runs) != 1 || !runs[0].RTL {
		t.Fatalf("got %v", runs)
	}
	// Logical "(" ... ")" becomes ")" ... "(" so that the reversed drawing
	// shows the opening bracket on the right.
	if r := []rune(runs[0].Text); r[0] != ')' || r[len(r)-1] != '(' {
		t.Errorf("brackets not mirrored: %q", runs[0].Text)
	}
	if Mirror('[') != ']' || Mirror('a') != 'a' || Mirror(0xAB) != 0xBB {
		t.Error("Mirror pairs wrong")
	}
}

func TestBaseDirection(t *testing.T) {
	if BaseDirection("123 abc") != LeftToRight || BaseDirection("12 עברית") != RightToLeft || BaseDirection("...") != LeftToRight {
		t.Error("BaseDirection wrong")
	}
	if !IsRTL('ع') || !IsRTL('א') || IsRTL('a') || IsRTL('٣') {
		t.Error("IsRTL wrong")
	}
}

func TestLevelsTrailingWhitespace(t *testing.T) {
	// L1: trailing whitespace takes the paragraph level even after RTL text.
	levels := Levels([]rune("abc عربي  "), Auto)
	if levels[len(levels)-1] != 0 || levels[len(levels)-2] != 0 || levels[4] != 1 {
		t.Errorf("levels %v", levels)
	}
}

func TestArabicForms(t *testing.T) {
	got := ArabicForms([]rune("مرحبا بالعالم"))
	want := []Form{Initial, Final, Initial, Medial, Final, Isolated, Initial, Final, Initial, Medial, Final, Initial, Final}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("forms %v, want %v", got, want)
	}
	// Marks are transparent: the letters around a shadda still join.
	got = ArabicForms([]rune("مّد"))
	if got[0] != Initial || got[2] != Final {
		t.Errorf("forms with mark %v", got)
	}
	// A ZWNJ breaks the join; tatweel keeps it.
	if f := ArabicForms([]rune("ب‌ب")); f[0] != Isolated || f[2] != Isolated {
		t.Errorf("ZWNJ forms %v", f)
	}
	if f := ArabicForms([]rune("بـب")); f[0] != Initial || f[2] != Final {
		t.Errorf("tatweel forms %v", f)
	}
}

func TestJoiningTypes(t *testing.T) {
	cases := map[rune]JoiningType{
		'ا': RightJoining, 'د': RightJoining, 'و': RightJoining, 'ب': DualJoining, 'ي': DualJoining,
		'ء': NonJoining, 'ـ': JoinCausing, 0x200D: JoinCausing, 0x064E: Transparent, 'a': NonJoining, '٣': NonJoining,
		'پ': DualJoining, 'ے': RightJoining, 'ک': DualJoining,
	}
	for r, want := range cases {
		if got := JoiningTypeOf(r); got != want {
			t.Errorf("JoiningTypeOf(%U) = %v, want %v", r, got, want)
		}
	}
}

func TestPresentationForms(t *testing.T) {
	if PresentationForm('ب', Medial) != 0xFE92 || PresentationForm('ا', Final) != 0xFE8E || PresentationForm('ا', Medial) != 0 {
		t.Error("presentation forms wrong")
	}
	if PresentationForm('x', Isolated) != 0 {
		t.Error("non-Arabic letter should have no presentation form")
	}
	if LamAlefLigature('ا', false) != 0xFEFB || LamAlefLigature('أ', true) != 0xFEF8 || LamAlefLigature('ب', false) != 0 {
		t.Error("lam-alef ligature wrong")
	}
}
