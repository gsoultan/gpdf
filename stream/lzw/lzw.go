// Copyright (c) 2026 Gembit Soultan Shirazi <gembit.soultan@gmail.com>

// Package lzw implements the PDF LZWDecode filter (ISO 32000-1 §7.4.4).
//
// Go's compress/lzw cannot be used: it has no support for /EarlyChange, and PDF's
// default (EarlyChange 1) switches code width one code earlier than GIF does.
package lzw

import (
	"bufio"
	"errors"
	"fmt"
	"io"

	"github.com/gsoultan/gpdf/model"
	"github.com/gsoultan/gpdf/stream"
)

const (
	clearCode = 256
	eodCode   = 257
	firstCode = 258
	maxWidth  = 12
	tableSize = 1 << maxWidth
)

// Filter implements stream.Filter for LZWDecode. It honours /EarlyChange via DecodeWithParams.
type Filter struct{}

// Decode decodes LZW data with the default EarlyChange 1.
func (Filter) Decode(dst io.Writer, src io.Reader, name string) error {
	if name != "LZWDecode" {
		return nil
	}
	return decode(dst, src, 1)
}

// DecodeWithParams decodes LZW data honouring /EarlyChange from params.
func (Filter) DecodeWithParams(dst io.Writer, src io.Reader, name string, params model.Dict) error {
	if name != "LZWDecode" {
		return nil
	}
	early := 1
	if v, ok := params[model.Name("EarlyChange")].(model.Integer); ok && v == 0 {
		early = 0
	}
	return decode(dst, src, early)
}

// Encode encodes data as LZW with EarlyChange 1 (the PDF default, so no /DecodeParms is needed).
func (Filter) Encode(dst io.Writer, src io.Reader, name string) error {
	if name != "LZWDecode" {
		return nil
	}
	data, err := io.ReadAll(src)
	if err != nil {
		return err
	}
	_, err = dst.Write(Encode(data, 1))
	return err
}

// NewFilter returns an LZWDecode stream filter.
func NewFilter() stream.Filter { return Filter{} }

var _ stream.ParamFilter = Filter{}

// ErrCorrupt reports an LZW code that does not exist in the table.
var ErrCorrupt = errors.New("lzw: invalid code")

type bitReader struct {
	r    *bufio.Reader
	bits uint32
	n    uint
}

func (b *bitReader) read(width uint) (int, error) {
	for b.n < width {
		c, err := b.r.ReadByte()
		if err != nil {
			return 0, err
		}
		b.bits = b.bits<<8 | uint32(c)
		b.n += 8
	}
	b.n -= width
	return int(b.bits>>b.n) & (1<<width - 1), nil
}

func decode(dst io.Writer, src io.Reader, early int) error {
	br := &bitReader{r: bufio.NewReader(src)}
	out := bufio.NewWriter(dst)
	table := make([][]byte, tableSize)
	for i := range 256 {
		table[i] = []byte{byte(i)}
	}
	width := uint(9)
	next := firstCode
	var prev []byte
	for {
		code, err := br.read(width)
		if err != nil {
			// Missing EOD is common; treat end of data as end of stream.
			if errors.Is(err, io.EOF) {
				break
			}
			return err
		}
		if code == clearCode {
			width, next, prev = 9, firstCode, nil
			continue
		}
		if code == eodCode {
			break
		}
		var entry []byte
		switch {
		case prev == nil:
			if code > 255 {
				return fmt.Errorf("%w %d after clear", ErrCorrupt, code)
			}
			entry = table[code]
		case code < next && table[code] != nil:
			entry = table[code]
		case code == next:
			entry = append(append([]byte{}, prev...), prev[0])
		default:
			return fmt.Errorf("%w %d (next %d)", ErrCorrupt, code, next)
		}
		if _, err := out.Write(entry); err != nil {
			return err
		}
		if prev != nil && next < tableSize {
			e := make([]byte, len(prev)+1)
			copy(e, prev)
			e[len(prev)] = entry[0]
			table[next] = e
			next++
			if next+early >= 1<<width && width < maxWidth {
				width++
			}
		}
		prev = entry
	}
	return out.Flush()
}

type bitWriter struct {
	out  []byte
	bits uint32
	n    uint
}

func (b *bitWriter) write(code int, width uint) {
	b.bits = b.bits<<width | uint32(code)
	b.n += width
	for b.n >= 8 {
		b.n -= 8
		b.out = append(b.out, byte(b.bits>>b.n))
	}
}

func (b *bitWriter) flush() []byte {
	if b.n > 0 {
		b.out = append(b.out, byte(b.bits<<(8-b.n)))
		b.n = 0
	}
	return b.out
}

// Encode LZW-encodes data. early is the /EarlyChange value (0 or 1) the decoder will use.
func Encode(data []byte, early int) []byte {
	bw := &bitWriter{}
	// Decoder-side state mirrored so the code width matches what the decoder expects.
	width := uint(9)
	decNext := firstCode
	decHasPrev := false
	emit := func(code int) {
		bw.write(code, width)
		if code == clearCode {
			width, decNext, decHasPrev = 9, firstCode, false
			return
		}
		if decHasPrev && decNext < tableSize {
			decNext++
			if decNext+early >= 1<<width && width < maxWidth {
				width++
			}
		}
		decHasPrev = true
	}
	type key struct {
		prefix int
		c      byte
	}
	dict := make(map[key]int)
	encNext := firstCode
	emit(clearCode)
	w := -1
	for _, c := range data {
		if w < 0 {
			w = int(c)
			continue
		}
		if code, ok := dict[key{w, c}]; ok {
			w = code
			continue
		}
		emit(w)
		dict[key{w, c}] = encNext
		encNext++
		if encNext >= tableSize-1 {
			emit(clearCode)
			clear(dict)
			encNext = firstCode
		}
		w = int(c)
	}
	if w >= 0 {
		emit(w)
	}
	emit(eodCode)
	return bw.flush()
}
