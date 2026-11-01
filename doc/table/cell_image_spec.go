// Copyright (c) 2026 Gembit Soultan Shirazi <gembit.soultan@gmail.com>

package table

import "github.com/gsoultan/gpdf/doc/style"

// CellImageSpec describes an image to be placed inside a table cell.
type CellImageSpec struct {
	Raw               []byte
	WidthPx, HeightPx int
	BitsPerComponent  int
	ColorSpace        string
	WidthPt, HeightPt float64
	IsJPEG            bool
	// Alpha is an optional soft mask: one 8-bit opacity sample per pixel. It is filled in
	// when a PNG is decoded from Raw.
	Alpha []byte

	Wrap style.ImageWrap
	Side style.ImageSide

	// PaddingPt defines extra space around the image.
	PaddingPt float64
}
