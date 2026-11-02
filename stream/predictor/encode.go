// Copyright (c) 2026 Gembit Soultan Shirazi <gembit.soultan@gmail.com>

package predictor

import "fmt"

// Encode applies the predictor described by p before compression, the inverse of
// Decode. PNG predictors 10 to 14 use the fixed filter type they name for every row;
// predictor 15 picks, for each row, the filter whose output has the smallest sum of
// absolute values, the heuristic PNG encoders use. The TIFF predictor is applied for
// 8- and 16-bit samples. Predictor 1 (or 0) returns data unchanged.
func Encode(data []byte, p Params) ([]byte, error) {
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
		return encodeTIFF(data, rowSize, p)
	case p.Predictor >= 10 && p.Predictor <= 15:
		return encodePNG(data, rowSize, bpp, p.Predictor), nil
	default:
		return nil, fmt.Errorf("%w: Predictor %d", ErrInvalidParams, p.Predictor)
	}
}

// encodePNG prefixes every row with its filter type byte and filters it. A trailing
// partial row is filtered as far as it goes.
func encodePNG(data []byte, rowSize, bpp, predictor int) []byte {
	rows := (len(data) + rowSize - 1) / rowSize
	out := make([]byte, 0, rows*(rowSize+1))
	prev := make([]byte, rowSize)
	filtered := make([]byte, rowSize)
	best := make([]byte, rowSize)
	for off := 0; off < len(data); off += rowSize {
		row := data[off:min(off+rowSize, len(data))]
		n := len(row)
		if predictor == 15 {
			bestType, bestScore := 0, -1
			for ft := 0; ft <= 4; ft++ {
				pngFilter(ft, row, prev[:n], bpp, filtered[:n])
				score := 0
				for _, v := range filtered[:n] {
					score += int(int8(v)) * sign(int8(v))
				}
				if bestScore < 0 || score < bestScore {
					bestType, bestScore = ft, score
					copy(best[:n], filtered[:n])
				}
			}
			out = append(out, byte(bestType))
			out = append(out, best[:n]...)
		} else {
			ft := predictor - 10
			pngFilter(ft, row, prev[:n], bpp, filtered[:n])
			out = append(out, byte(ft))
			out = append(out, filtered[:n]...)
		}
		clear(prev)
		copy(prev, row)
	}
	return out
}

func sign(v int8) int {
	if v < 0 {
		return -1
	}
	return 1
}

// pngFilter writes the PNG filter ft of row (with the previous, unfiltered row prev) to dst.
func pngFilter(ft int, row, prev []byte, bpp int, dst []byte) {
	for i := range row {
		var a, c byte
		if i >= bpp {
			a = row[i-bpp]
			c = prev[i-bpp]
		}
		b := prev[i]
		switch ft {
		case 0:
			dst[i] = row[i]
		case 1:
			dst[i] = row[i] - a
		case 2:
			dst[i] = row[i] - b
		case 3:
			dst[i] = row[i] - byte((int(a)+int(b))/2)
		case 4:
			dst[i] = row[i] - paeth(a, b, c)
		}
	}
}

// encodeTIFF applies horizontal differencing (predictor 2) for 8- and 16-bit samples.
func encodeTIFF(data []byte, rowSize int, p Params) ([]byte, error) {
	out := make([]byte, len(data))
	copy(out, data)
	colors := p.Colors
	for off := 0; off+rowSize <= len(out); off += rowSize {
		row := out[off : off+rowSize]
		src := data[off : off+rowSize]
		switch p.BitsPerComponent {
		case 8:
			for i := colors; i < rowSize; i++ {
				row[i] = src[i] - src[i-colors]
			}
		case 16:
			for i := 2 * colors; i+1 < rowSize; i += 2 {
				v := uint16(src[i])<<8 | uint16(src[i+1])
				pv := uint16(src[i-2*colors])<<8 | uint16(src[i-2*colors+1])
				v -= pv
				row[i], row[i+1] = byte(v>>8), byte(v)
			}
		default:
			return nil, fmt.Errorf("%w: TIFF predictor with BitsPerComponent %d", ErrInvalidParams, p.BitsPerComponent)
		}
	}
	return out, nil
}
