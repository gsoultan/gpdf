package lzw

import (
	"bytes"
	clzw "compress/lzw"
	"math/rand/v2"
	"testing"

	"github.com/gsoultan/gpdf/model"
)

// The example from ISO 32000-1 §7.4.4.2: "-----A---B" encodes to 80 0B 60 50 22 0C 0C 85 01.
func TestDecodeSpecExample(t *testing.T) {
	var out bytes.Buffer
	if err := (Filter{}).Decode(&out, bytes.NewReader([]byte{0x80, 0x0B, 0x60, 0x50, 0x22, 0x0C, 0x0C, 0x85, 0x01}), "LZWDecode"); err != nil {
		t.Fatal(err)
	}
	if out.String() != "-----A---B" {
		t.Fatalf("got %q", out.String())
	}
	if got := Encode([]byte("-----A---B"), 1); !bytes.Equal(got, []byte{0x80, 0x0B, 0x60, 0x50, 0x22, 0x0C, 0x0C, 0x85, 0x01}) {
		t.Fatalf("encode: got % X", got)
	}
}

func sample(n int) []byte {
	r := rand.New(rand.NewPCG(1, 2))
	b := make([]byte, n)
	for i := range b {
		// Low-entropy data so the table fills and resets several times.
		b[i] = byte('a' + r.IntN(6))
	}
	return b
}

func TestRoundTripAcrossWidthsAndResets(t *testing.T) {
	for _, early := range []int{0, 1} {
		data := sample(200_000)
		enc := Encode(data, early)
		var out bytes.Buffer
		if err := (Filter{}).DecodeWithParams(&out, bytes.NewReader(enc), "LZWDecode", model.Dict{"EarlyChange": model.Integer(early)}); err != nil {
			t.Fatalf("early=%d: %v", early, err)
		}
		if !bytes.Equal(out.Bytes(), data) {
			t.Fatalf("early=%d: round trip mismatch", early)
		}
	}
}

// EarlyChange 0 is the GIF-style variant that Go's compress/lzw produces.
func TestDecodeEarlyChangeZeroMatchesStdlib(t *testing.T) {
	data := sample(50_000)
	var enc bytes.Buffer
	w := clzw.NewWriter(&enc, clzw.MSB, 8)
	w.Write(data)
	w.Close()
	var out bytes.Buffer
	if err := (Filter{}).DecodeWithParams(&out, bytes.NewReader(enc.Bytes()), "LZWDecode", model.Dict{"EarlyChange": model.Integer(0)}); err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(out.Bytes(), data) {
		t.Fatal("mismatch decoding compress/lzw output with EarlyChange 0")
	}
}
