// Copyright (c) 2026 Gembit Soultan Shirazi <gembit.soultan@gmail.com>

package model

import (
	"testing"
	"time"
)

func TestTextStringRoundTrip(t *testing.T) {
	for _, s := range []string{"plain", "Grüße", "日本語", "emoji 😀", ""} {
		for _, u8 := range []bool{false, true} {
			enc := EncodeTextString(s, u8)
			if got := enc.Text(); got != s {
				t.Errorf("EncodeTextString(%q, utf8=%v) -> %q", s, u8, got)
			}
		}
	}
	if EncodeTextString("plain", true) != "plain" {
		t.Error("ASCII must stay unencoded")
	}
	if got := EncodeTextString("é", true); string(got) != "\xEF\xBB\xBFé" {
		t.Errorf("utf8 bom: %q", got)
	}
	if got := EncodeTextString("é", false); string(got) != "\xFE\xFF\x00\xE9" {
		t.Errorf("utf16 bom: %q", got)
	}
}

func TestDecodePDFDocEncoding(t *testing.T) {
	if got := DecodeTextString([]byte{0x80, 0x20, 0xA0, 0x20, 0xE9}); got != "• € é" {
		t.Errorf("got %q", got)
	}
	if got := DecodeTextString([]byte{0xFE, 0xFF, 0x00, 0x48, 0x00, 0x69}); got != "Hi" {
		t.Errorf("utf16: %q", got)
	}
}

func TestParseDate(t *testing.T) {
	cases := map[string]string{
		"D:20261101120000Z":       "2026-11-01T12:00:00Z",
		"D:20261101120000+05'30'": "2026-11-01T12:00:00+05:30",
		"D:20261101120000+05'30":  "2026-11-01T12:00:00+05:30",
		"D:20261101120000-0800":   "2026-11-01T12:00:00-08:00",
		"D:20261101":              "2026-11-01T00:00:00Z",
		"D:2026":                  "2026-01-01T00:00:00Z",
		"20261101120000":          "2026-11-01T12:00:00Z",
		"D:20261101120000Z00'00'": "2026-11-01T12:00:00Z",
	}
	for in, want := range cases {
		got, err := ParseDate(in)
		if err != nil {
			t.Errorf("%q: %v", in, err)
			continue
		}
		if got.Format(time.RFC3339) != want {
			t.Errorf("%q = %s, want %s", in, got.Format(time.RFC3339), want)
		}
	}
	for _, bad := range []string{"", "D:", "D:20", "D:20261301", "D:20261101120000X"} {
		if _, err := ParseDate(bad); err == nil {
			t.Errorf("%q parsed", bad)
		}
	}
	tm := time.Date(2026, 11, 1, 12, 0, 0, 0, time.UTC)
	if got := FormatDate(tm, true); got != "D:20261101120000Z" {
		t.Errorf("2.0 utc: %s", got)
	}
	if got := FormatDate(tm, false); got != "D:20261101120000+00'00'" {
		t.Errorf("1.x utc: %s", got)
	}
	ist := tm.In(time.FixedZone("IST", 5*3600+1800))
	if got := FormatDate(ist, true); got != "D:20261101173000+05'30" {
		t.Errorf("2.0 ist: %s", got)
	}
	if back, _ := ParseDate(FormatDate(ist, true)); !back.Equal(tm) {
		t.Errorf("round trip: %v", back)
	}
}
