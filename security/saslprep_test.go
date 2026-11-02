// Copyright (c) 2026 Gembit Soultan Shirazi <gembit.soultan@gmail.com>

package security

import (
	"errors"
	"testing"
)

func TestSASLprep(t *testing.T) {
	ok := map[string]string{
		"secret":   "secret",
		"I­X":      "IX",       // soft hyphen removed (RFC 4013 example)
		"a b":      "a b",      // non-ASCII space
		"​x":       " x",       // zero width space maps to space
		"pässwörd": "pässwörd", // already NFC/NFKC
		"אב":       "אב",
		"الع":      "الع",
	}
	for in, want := range ok {
		got, err := saslprep(in)
		if err != nil || got != want {
			t.Errorf("saslprep(%q) = %q, %v; want %q", in, got, err, want)
		}
	}
	for _, bad := range []string{"a\x00b", "\u0007", "x", "�", "a‎b", "אx", "xא", "אaב", "\xff"} {
		if _, err := saslprep(bad); !errors.Is(err, ErrPasswordPrep) {
			t.Errorf("saslprep(%q) accepted: %v", bad, err)
		}
	}
}
