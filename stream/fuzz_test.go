package stream_test

import (
	"bytes"
	"testing"

	"github.com/gsoultan/gpdf/model"
	"github.com/gsoultan/gpdf/stream"
	"github.com/gsoultan/gpdf/stream/defaults"
	"github.com/gsoultan/gpdf/stream/predictor"
)

var fuzzFilters = []string{"FlateDecode", "LZWDecode", "ASCII85Decode", "ASCIIHexDecode", "RunLengthDecode"}

// FuzzFilters decodes hostile input through every standard filter, with and
// without a predictor, and checks that encoding and decoding round-trips.
func FuzzFilters(f *testing.F) {
	f.Add([]byte("x\x9c\xcbH\xcd\xc9\xc9\a\x00\x06,\x02\x15"), uint8(0), uint8(1), uint8(8), uint8(4))
	f.Add([]byte("\x80\x0b\x60\x50\x22\x0c\x0c\x85\x01"), uint8(1), uint8(2), uint8(3), uint8(2))
	f.Add([]byte("87cURD]i,\"Ebo80~>"), uint8(2), uint8(12), uint8(8), uint8(3))
	f.Add([]byte("48656C6C6F>"), uint8(3), uint8(15), uint8(1), uint8(8))
	f.Add([]byte("\x04Hello\xfeA\x80"), uint8(4), uint8(10), uint8(16), uint8(1))
	f.Fuzz(func(t *testing.T, data []byte, which, pred, bpc, cols uint8) {
		name := fuzzFilters[int(which)%len(fuzzFilters)]
		reg := defaults.NewStandardRegistry()
		filter := reg.Get(name)
		if filter == nil {
			t.Fatalf("no %s filter", name)
		}
		params := model.Dict{
			model.Name("Predictor"):        model.Integer(int64(pred)),
			model.Name("Colors"):           model.Integer(int64(cols%4 + 1)),
			model.Name("BitsPerComponent"): model.Integer(int64(bpc)),
			model.Name("Columns"):          model.Integer(int64(cols) + 1),
		}
		var decoded bytes.Buffer
		if pf, ok := filter.(stream.ParamFilter); ok {
			_ = pf.DecodeWithParams(&decoded, bytes.NewReader(data), name, params)
		} else {
			_ = filter.Decode(&decoded, bytes.NewReader(data), name)
		}
		p := predictor.Params{Predictor: int(pred), Colors: int(cols%4 + 1), BitsPerComponent: int(bpc), Columns: int(cols) + 1}
		_, _ = predictor.Decode(data, p)
		if enc, err := predictor.Encode(data, p); err == nil {
			if dec, err := predictor.Decode(enc, p); err != nil || !bytes.Equal(dec, data) {
				t.Fatalf("predictor %+v round trip: %v", p, err)
			}
		}
		var encoded bytes.Buffer
		if err := filter.Encode(&encoded, bytes.NewReader(data), name); err != nil {
			return
		}
		var back bytes.Buffer
		if err := filter.Decode(&back, bytes.NewReader(encoded.Bytes()), name); err != nil || !bytes.Equal(back.Bytes(), data) {
			t.Fatalf("%s round trip: %v", name, err)
		}
	})
}
