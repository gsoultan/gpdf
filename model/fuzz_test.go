package model

import (
	"testing"
	"time"
	"unicode/utf8"
)

// FuzzTextString checks that text strings in any encoding decode without
// panicking and that encoding a decoded string reproduces it.
func FuzzTextString(f *testing.F) {
	for _, s := range []string{"plain", "\xfe\xff\x00H\x00i", "\xef\xbb\xbfcaf\xc3\xa9", "\xfe\xff\xd8\x00", "\x1b\xfe\xff", "\xff\xfe\x00"} {
		f.Add([]byte(s))
	}
	f.Fuzz(func(t *testing.T, data []byte) {
		text := DecodeTextString(data)
		_ = IsTextStringBOM(data)
		if !utf8.ValidString(text) {
			return // a UTF-8 BOM passes raw bytes through
		}
		for _, bom := range []bool{false, true} {
			if got := EncodeTextString(text, bom).Text(); got != text {
				t.Fatalf("round trip (utf8=%v): %q -> %q", bom, text, got)
			}
		}
	})
}

// FuzzParseDate checks that date strings never panic the parser and that a
// parsed date formats back to one the parser accepts.
func FuzzParseDate(f *testing.F) {
	for _, s := range []string{"D:20261101120000Z", "D:2026", "D:20261101120000+05'30'", "D:20261101120000+05'30", "20261101", "D:99999999999999Z", "D:2026-11-01"} {
		f.Add(s)
	}
	f.Fuzz(func(t *testing.T, s string) {
		tm, err := ParseDate(s)
		if err != nil {
			return
		}
		for _, pdf20 := range []bool{false, true} {
			back, err := ParseDate(FormatDate(tm, pdf20))
			if err != nil {
				t.Fatalf("formatted %q does not parse: %v", FormatDate(tm, pdf20), err)
			}
			if !back.Equal(tm.Truncate(time.Second)) {
				t.Fatalf("round trip %q: %v -> %v", s, tm, back)
			}
		}
	})
}
