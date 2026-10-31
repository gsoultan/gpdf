package impl

import (
	"bytes"
	"strings"
	"testing"

	"github.com/gsoultan/gpdf/model"
	"github.com/gsoultan/gpdf/syntax"
)

func parseInline(t *testing.T, src string) model.Object {
	t.Helper()
	p := NewParser(bytes.NewReader([]byte(src)), int64(len(src)))
	_, obj, err := p.ParseObject()
	if err != nil {
		t.Fatalf("ParseObject(%q): %v", src, err)
	}
	return obj
}

func TestTokenizerStringFixes(t *testing.T) {
	cases := map[string]string{
		`(a\5b)`:      "a\x05b", // 1-digit octal escape
		`(a\53b)`:     "a+b",    // 2-digit octal escape
		"(a\\\r\nb)":  "ab",     // backslash + CRLF continuation
		"(a\r\nb)":    "a\nb",   // bare CRLF reads as LF
		`<414>`:       "A@",     // odd hex digit count pads with 0
		`<41 42\n43>`: "ABC",    // whitespace inside hex
		`(\q)`:        "q",      // unknown escape drops the backslash
		"(x\fy)":      "x\fy",
	}
	for src, want := range cases {
		got, ok := parseInline(t, src).(model.String)
		if !ok || string(got) != want {
			t.Errorf("%q: got %q want %q", src, got, want)
		}
	}
}

func TestTokenizerNameAndWhitespace(t *testing.T) {
	if got := parseInline(t, "/#41BC"); got != model.Name("ABC") {
		t.Errorf("#xx as first name char: got %v", got)
	}
	d, ok := parseInline(t, "<<\f/A\f1\f>>").(model.Dict)
	if !ok || d["A"] != model.Integer(1) {
		t.Errorf("form feed as whitespace: got %#v", d)
	}
}

func TestTokenizerCommentAdvancesOffset(t *testing.T) {
	src := "% comment\n42 "
	tok := NewTokenizer(bytes.NewReader([]byte(src)), 0, int64(len(src)))
	tk, _ := tok.Next()
	if tk.Kind != syntax.TokenInteger || tk.Int != 42 {
		t.Fatalf("got %+v", tk)
	}
	if off := tok.CurrentOffset(); off != int64(strings.Index(src, "42")+2) {
		t.Fatalf("CurrentOffset = %d", off)
	}
}

func TestTokenizerStrayDelimitersMakeProgress(t *testing.T) {
	for _, src := range []string{">", ")", "{", "}"} {
		tok := NewTokenizer(bytes.NewReader([]byte(src+" 1")), 0, int64(len(src)+2))
		first, _ := tok.Next()
		if first.Kind == syntax.TokenEOF {
			t.Errorf("%q: stray delimiter reported as EOF", src)
		}
		second, _ := tok.Next()
		if second.Kind != syntax.TokenInteger {
			t.Errorf("%q: tokenizer did not advance past the delimiter: %+v", src, second)
		}
	}
}

func TestMalformedNumbers(t *testing.T) {
	cases := map[string]float64{"--5": 5, "-": 0, "1.2.3": 1.2, "+-3": -3, "4.": 4}
	for src, want := range cases {
		var got float64
		switch v := parseInline(t, "["+src+"]").(model.Array)[0].(type) {
		case model.Integer:
			got = float64(v)
		case model.Real:
			got = float64(v)
		}
		if got != want {
			t.Errorf("%q: got %v want %v", src, got, want)
		}
	}
}

func TestTwoIntegersNotObjOrRefKeepsGeneration(t *testing.T) {
	src := "12 34 true"
	p := NewParser(bytes.NewReader([]byte(src)), int64(len(src)))
	var got []model.Object
	for range 3 {
		_, obj, err := p.ParseObject()
		if err != nil {
			t.Fatal(err)
		}
		got = append(got, obj)
	}
	if got[0] != model.Integer(12) || got[1] != model.Integer(34) || got[2] != model.Boolean(true) {
		t.Fatalf("got %#v", got)
	}
}

func TestDeepNestingRejected(t *testing.T) {
	src := strings.Repeat("[", 100000) + strings.Repeat("]", 100000)
	p := NewParser(bytes.NewReader([]byte(src)), int64(len(src)))
	if _, _, err := p.ParseObject(); err == nil {
		t.Fatal("expected nesting error")
	}
}

// FuzzParseObject checks that hostile input never panics, hangs or over-allocates.
func FuzzParseObject(f *testing.F) {
	for _, s := range []string{
		"1 0 obj << /Length 99999999999 >> stream\nabc\nendstream endobj",
		"1 0 obj << /Length 5 0 R >> stream\nabc",
		"<< /A [1 2 R (x\\5) <4> /#41] >>",
		"1 0 obj\n> ) { } endobj",
		"%comment\n[1 0 R 2 0 3]",
	} {
		f.Add([]byte(s))
	}
	f.Fuzz(func(t *testing.T, data []byte) {
		p := NewParser(bytes.NewReader(data), int64(len(data)))
		for range 64 {
			if _, _, err := p.ParseObject(); err != nil {
				return
			}
		}
	})
}
