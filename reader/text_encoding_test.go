// Copyright (c) 2026 Gembit Soultan Shirazi <gembit.soultan@gmail.com>

package reader

import (
	"testing"
	"unicode/utf8"

	"github.com/gsoultan/gpdf/model"
)

func TestEncodingTablesComplete(t *testing.T) {
	if n := utf8.RuneCountInString(macRomanHigh); n != 128 {
		t.Fatalf("MacRoman high half has %d runes, want 128", n)
	}
	if got := baseEncoding("MacRomanEncoding")[0x8E]; got != 'é' {
		t.Fatalf("MacRoman 0x8E = %q", got)
	}
	if got := baseEncoding("StandardEncoding")[0x27]; got != '’' {
		t.Fatalf("Standard 0x27 = %q", got)
	}
}

// Regression: /Encoding and /Differences were ignored for simple fonts; text was guessed
// as UTF-8 or WinAnsi, so remapped codes came out wrong.
func TestSimpleFontDifferences(t *testing.T) {
	font := model.Dict{
		"Type": model.Name("Font"), "Subtype": model.Name("Type1"), "BaseFont": model.Name("ABCDEF+Custom"),
		"Encoding": model.Dict{
			"BaseEncoding": model.Name("WinAnsiEncoding"),
			"Differences":  model.Array{model.Integer(1), model.Name("H"), model.Name("i"), model.Integer(3), model.Name("eacute"), model.Name("uni20AC"), model.Name("f_i")},
		},
	}
	d := &toUnicodeDecoder{simple: simpleFontEncoding(nil, font)}
	if got := d.decode([]byte{1, 2, 3, 4, 5, 0x80, 'A'}); got != "Hié€fi€A" {
		t.Fatalf("got %q", got)
	}
}

// Regression: the ToUnicode parser was line based: compact or CR-only CMaps lost
// mappings, codespace ranges were ignored and bfrange with a multi-character base
// incremented the first character.
func TestToUnicodeParserRobustness(t *testing.T) {
	compact := "/CIDInit /ProcSet findresource begin 12 dict begin begincmap 1 begincodespacerange <0000> <FFFF> endcodespacerange " +
		"2 beginbfchar <0001> <0048> <0002> <0069> endbfchar 1 beginbfrange <0010> <0012> <00660066> endbfrange " +
		"1 beginbfrange <0020> <0021> [<0041> <0042>] endbfrange endcmap"
	crOnly := "1 begincodespacerange\r<0000> <FFFF>\rendcodespacerange\r1 beginbfchar\r<0001> <0048>\rendbfchar\r"
	for name, cmap := range map[string]string{"compact": compact, "crOnly": crOnly} {
		d := parseToUnicodeDecoder([]byte(cmap))
		if d == nil {
			t.Fatalf("%s: no decoder", name)
		}
		if got := d.decode([]byte{0, 1}); got != "H" {
			t.Fatalf("%s: got %q", name, got)
		}
	}
	d := parseToUnicodeDecoder([]byte(compact))
	if got := d.decode([]byte{0, 0x10, 0, 0x12, 0, 0x21, 0, 2}); got != "fffhBi" {
		t.Fatalf("ranges: got %q", got)
	}
	// With a 2-byte codespace, an unmapped code is skipped whole instead of being read as
	// two single bytes that shift every following character.
	if got := d.decode([]byte{0x7F, 0x7F, 0, 1}); got != "H" {
		t.Fatalf("codespace: got %q", got)
	}
}
