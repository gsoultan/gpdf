// Copyright (c) 2026 Gembit Soultan Shirazi <gembit.soultan@gmail.com>

// Package predictor implements the PNG and TIFF predictors that PDF applies
// after FlateDecode and LZWDecode (ISO 32000-1 §7.4.4.4, /DecodeParms /Predictor).
package predictor

import (
	"errors"
	"fmt"
)

// Params holds the predictor-related /DecodeParms entries.
type Params struct {
	Predictor        int // 1 = none, 2 = TIFF, 10..15 = PNG
	Colors           int // default 1
	BitsPerComponent int // default 8
	Columns          int // default 1
}

// Normalize fills in the spec defaults for zero values.
func (p Params) Normalize() Params {
	if p.Predictor == 0 {
		p.Predictor = 1
	}
	if p.Colors <= 0 {
		p.Colors = 1
	}
	if p.BitsPerComponent <= 0 {
		p.BitsPerComponent = 8
	}
	if p.Columns <= 0 {
		p.Columns = 1
	}
	return p
}

// ErrInvalidParams reports predictor parameters that cannot describe a row.
var ErrInvalidParams = errors.New("predictor: invalid parameters")

// maxRowBytes bounds the row size computed from untrusted parameters.
const maxRowBytes = 1 << 28

// Decode reverses the predictor described by p. Predictor 1 (or 0) returns data unchanged.
func Decode(data []byte, p Params) ([]byte, error) {
	p = p.Normalize()
	if p.Predictor == 1 {
		return data, nil
	}
	switch p.BitsPerComponent {
	case 1, 2, 4, 8, 16:
	default:
		return nil, fmt.Errorf("%w: BitsPerComponent %d", ErrInvalidParams, p.BitsPerComponent)
	}
	if p.Colors > 32 {
		return nil, fmt.Errorf("%w: Colors %d", ErrInvalidParams, p.Colors)
	}
	bitsPerPixel := p.Colors * p.BitsPerComponent
	rowBits := int64(p.Columns) * int64(bitsPerPixel)
	if rowBits <= 0 || rowBits/8 > maxRowBytes {
		return nil, fmt.Errorf("%w: Columns %d", ErrInvalidParams, p.Columns)
	}
	rowSize := int((rowBits + 7) / 8)
	bpp := max((bitsPerPixel+7)/8, 1)
	switch {
	case p.Predictor == 2:
		return decodeTIFF(data, rowSize, p), nil
	case p.Predictor >= 10 && p.Predictor <= 15:
		return decodePNG(data, rowSize, bpp)
	default:
		return nil, fmt.Errorf("%w: Predictor %d", ErrInvalidParams, p.Predictor)
	}
}

// decodePNG undoes per-row PNG filters. Each row is prefixed with its filter type byte.
// A trailing partial row is decoded as far as it goes (some producers truncate the last row).
func decodePNG(data []byte, rowSize, bpp int) ([]byte, error) {
	stride := rowSize + 1
	rows := (len(data) + stride - 1) / stride
	out := make([]byte, 0, rows*rowSize)
	prev := make([]byte, rowSize)
	cur := make([]byte, rowSize)
	for off := 0; off < len(data); off += stride {
		end := min(off+stride, len(data))
		src := data[off:end]
		if len(src) < 2 {
			break
		}
		ft := src[0]
		row := src[1:]
		n := len(row)
		clear(cur)
		switch ft {
		case 0:
			copy(cur, row)
		case 1:
			for i := range n {
				var a byte
				if i >= bpp {
					a = cur[i-bpp]
				}
				cur[i] = row[i] + a
			}
		case 2:
			for i := range n {
				cur[i] = row[i] + prev[i]
			}
		case 3:
			for i := range n {
				var a byte
				if i >= bpp {
					a = cur[i-bpp]
				}
				cur[i] = row[i] + byte((int(a)+int(prev[i]))/2)
			}
		case 4:
			for i := range n {
				var a, c byte
				if i >= bpp {
					a = cur[i-bpp]
					c = prev[i-bpp]
				}
				cur[i] = row[i] + paeth(a, prev[i], c)
			}
		default:
			return nil, fmt.Errorf("predictor: unknown PNG filter type %d", ft)
		}
		out = append(out, cur[:n]...)
		prev, cur = cur, prev
	}
	return out, nil
}

func paeth(a, b, c byte) byte {
	p := int(a) + int(b) - int(c)
	pa, pb, pc := abs(p-int(a)), abs(p-int(b)), abs(p-int(c))
	switch {
	case pa <= pb && pa <= pc:
		return a
	case pb <= pc:
		return b
	default:
		return c
	}
}

func abs(v int) int {
	if v < 0 {
		return -v
	}
	return v
}

// decodeTIFF undoes TIFF predictor 2 (horizontal differencing) for every supported bit depth.
func decodeTIFF(data []byte, rowSize int, p Params) []byte {
	out := make([]byte, len(data))
	copy(out, data)
	colors := p.Colors
	bpc := p.BitsPerComponent
	for off := 0; off+rowSize <= len(out); off += rowSize {
		row := out[off : off+rowSize]
		switch bpc {
		case 8:
			for i := colors; i < rowSize; i++ {
				row[i] += row[i-colors]
			}
		case 16:
			for i := 2 * colors; i+1 < rowSize; i += 2 {
				v := uint16(row[i])<<8 | uint16(row[i+1])
				pv := uint16(row[i-2*colors])<<8 | uint16(row[i-2*colors+1])
				v += pv
				row[i], row[i+1] = byte(v>>8), byte(v)
			}
		default:
			mask := uint(1)<<uint(bpc) - 1
			samples := p.Columns * colors
			get := func(i int) uint {
				bit := i * bpc
				return (uint(row[bit/8]) >> (8 - uint(bpc) - uint(bit%8))) & mask
			}
			set := func(i int, v uint) {
				bit := i * bpc
				shift := 8 - uint(bpc) - uint(bit%8)
				row[bit/8] = row[bit/8]&^byte(mask<<shift) | byte((v&mask)<<shift)
			}
			for i := colors; i < samples; i++ {
				set(i, get(i)+get(i-colors))
			}
		}
	}
	return out
}
