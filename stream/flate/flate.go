package flate

import (
	"compress/zlib"
	"io"
	"sync"

	"github.com/gsoultan/gpdf/stream"
)

// Filter implements stream.Filter for FlateDecode (zlib/deflate).
type Filter struct{}

// Decode decodes zlib-compressed data.
func (Filter) Decode(dst io.Writer, src io.Reader, name string) error {
	if name != "FlateDecode" {
		return nil
	}
	r, err := zlib.NewReader(src)
	if err != nil {
		return err
	}
	defer r.Close()
	_, err = io.Copy(dst, r)
	return err
}

// Encode encodes data with zlib.
func (Filter) Encode(dst io.Writer, src io.Reader, name string) error {
	if name != "FlateDecode" {
		return nil
	}
	w := writers.Get().(*zlib.Writer)
	w.Reset(dst)
	defer writers.Put(w)
	_, err := io.Copy(w, src)
	if err != nil {
		w.Close()
		return err
	}
	return w.Close()
}

// writers recycles zlib writers: each one carries ~1 MB of compressor state,
// and a document compresses one stream per page, font and image.
var writers = sync.Pool{New: func() any { return zlib.NewWriter(io.Discard) }}

// NewFilter returns a FlateDecode filter.
func NewFilter() stream.Filter {
	return Filter{}
}
