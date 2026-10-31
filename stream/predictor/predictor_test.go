package predictor

import (
	"bytes"
	"testing"
)

func TestDecodePNGUpAndSub(t *testing.T) {
	// 2 rows, 3 columns, 1 colour, 8 bpc. Row 1 uses Sub, row 2 uses Up.
	data := []byte{
		1, 10, 1, 1, // Sub: 10, 11, 12
		2, 1, 1, 1, // Up: 11, 12, 13
	}
	got, err := Decode(data, Params{Predictor: 12, Columns: 3})
	if err != nil {
		t.Fatal(err)
	}
	want := []byte{10, 11, 12, 11, 12, 13}
	if !bytes.Equal(got, want) {
		t.Fatalf("got %v want %v", got, want)
	}
}

func TestDecodePNGAverageAndPaeth(t *testing.T) {
	data := []byte{
		0, 100, 50,
		3, 10, 10, // avg: 10+(0+100)/2=60, 10+(60+50)/2=65
		4, 1, 1, // paeth(a=0,b=60,c=0)=60 ->61 ; paeth(a=61,b=65,c=60)=65 -> 66
	}
	got, err := Decode(data, Params{Predictor: 15, Columns: 2})
	if err != nil {
		t.Fatal(err)
	}
	want := []byte{100, 50, 60, 65, 61, 66}
	if !bytes.Equal(got, want) {
		t.Fatalf("got %v want %v", got, want)
	}
}

func TestDecodeTIFF8And1Bit(t *testing.T) {
	got, err := Decode([]byte{5, 1, 1, 7, 2, 2}, Params{Predictor: 2, Columns: 3})
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, []byte{5, 6, 7, 7, 9, 11}) {
		t.Fatalf("8-bit TIFF: got %v", got)
	}
	// 1-bit, 8 columns: differences 1,0,0,0,1,0,0,0 -> running xor-sum mod 2: 1,1,1,1,0,0,0,0
	got, err = Decode([]byte{0b10001000}, Params{Predictor: 2, Columns: 8, BitsPerComponent: 1})
	if err != nil {
		t.Fatal(err)
	}
	if got[0] != 0b11110000 {
		t.Fatalf("1-bit TIFF: got %08b", got[0])
	}
}

func TestDecodeRejectsBadParams(t *testing.T) {
	if _, err := Decode([]byte{0, 1}, Params{Predictor: 12, BitsPerComponent: 3}); err == nil {
		t.Fatal("expected error for BitsPerComponent 3")
	}
	if _, err := Decode([]byte{9, 1}, Params{Predictor: 12}); err == nil {
		t.Fatal("expected error for unknown PNG filter type")
	}
}
