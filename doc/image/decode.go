package image

import (
	"bytes"
	stdimage "image"
	"image/color"
	_ "image/jpeg"
	_ "image/png"
)

// DetectDimensions attempts to get pixel dimensions from PNG/JPEG data without full decode.
func DetectDimensions(data []byte) (w, h int, isJPEG, isPNG bool, err error) {
	config, format, err := stdimage.DecodeConfig(bytes.NewReader(data))
	if err != nil {
		return 0, 0, false, false, err
	}
	return config.Width, config.Height, format == "jpeg", format == "png", nil
}

// HasAlpha returns true if any pixel in img has an alpha value less than fully opaque.
func HasAlpha(img stdimage.Image) bool {
	bounds := img.Bounds()
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			_, _, _, a := img.At(x, y).RGBA()
			if a < 0xFFFF {
				return true
			}
		}
	}
	return false
}

// Decoded is a PNG decoded into the sample data of a PDF image XObject.
type Decoded struct {
	// Raw holds 8-bit samples, one per pixel for DeviceGray and three for DeviceRGB.
	// Colors are not premultiplied: a transparent pixel keeps its own color.
	Raw           []byte
	Width, Height int
	ColorSpace    string // DeviceGray or DeviceRGB
	// Alpha holds one 8-bit opacity sample per pixel, for use as a DeviceGray /SMask.
	// It is nil when every pixel is fully opaque.
	Alpha []byte
}

// DecodePNG decodes PNG data, splitting it into color samples and an alpha channel.
// Gray and gray-with-alpha images stay DeviceGray; everything else becomes DeviceRGB.
func DecodePNG(pngData []byte) (*Decoded, error) {
	img, _, err := stdimage.Decode(bytes.NewReader(pngData))
	if err != nil {
		return nil, err
	}
	bounds := img.Bounds()
	d := &Decoded{Width: bounds.Dx(), Height: bounds.Dy(), ColorSpace: "DeviceRGB"}
	if d.Width == 0 || d.Height == 0 {
		return d, nil
	}
	opaque := true
	if o, ok := img.(interface{ Opaque() bool }); ok {
		opaque = o.Opaque()
	} else {
		opaque = !HasAlpha(img)
	}
	gray := isGrayImage(img)
	bpp := 3
	if gray {
		d.ColorSpace = "DeviceGray"
		bpp = 1
	}
	n := d.Width * d.Height
	d.Raw = make([]byte, n*bpp)
	if !opaque {
		d.Alpha = make([]byte, n)
	}
	for py := range d.Height {
		for px := range d.Width {
			i := py*d.Width + px
			c := nrgbaAt(img, bounds.Min.X+px, bounds.Min.Y+py)
			if gray {
				// For a gray source R, G and B are equal.
				d.Raw[i] = byte(c.R >> 8)
			} else {
				d.Raw[i*3] = byte(c.R >> 8)
				d.Raw[i*3+1] = byte(c.G >> 8)
				d.Raw[i*3+2] = byte(c.B >> 8)
			}
			if d.Alpha != nil {
				d.Alpha[i] = byte(c.A >> 8)
			}
		}
	}
	return d, nil
}

// nrgbaAt returns the non-premultiplied color at (x, y). Converting through RGBA() would
// premultiply and lose the color of fully transparent pixels.
func nrgbaAt(img stdimage.Image, x, y int) color.NRGBA64 {
	switch m := img.(type) {
	case *stdimage.NRGBA:
		c := m.NRGBAAt(x, y)
		return color.NRGBA64{uint16(c.R) * 0x101, uint16(c.G) * 0x101, uint16(c.B) * 0x101, uint16(c.A) * 0x101}
	case *stdimage.NRGBA64:
		return m.NRGBA64At(x, y)
	case *stdimage.Paletted:
		if c, ok := m.Palette[m.ColorIndexAt(x, y)].(color.NRGBA); ok {
			return color.NRGBA64{uint16(c.R) * 0x101, uint16(c.G) * 0x101, uint16(c.B) * 0x101, uint16(c.A) * 0x101}
		}
	}
	return color.NRGBA64Model.Convert(img.At(x, y)).(color.NRGBA64)
}

// isGrayImage reports whether every pixel of img is a shade of gray: the gray image types,
// and gray-with-alpha PNGs, which Go decodes as NRGBA or NRGBA64 with equal R, G and B.
func isGrayImage(img stdimage.Image) bool {
	switch img.(type) {
	case *stdimage.Gray, *stdimage.Gray16:
		return true
	case *stdimage.NRGBA, *stdimage.NRGBA64:
	default:
		return false
	}
	b := img.Bounds()
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			c := nrgbaAt(img, x, y)
			if c.R != c.G || c.G != c.B {
				return false
			}
		}
	}
	return true
}

// DecodePNGToRaw decodes PNG data into raw pixel bytes suitable for a PDF image XObject.
// Returns the raw bytes, pixel dimensions, and the PDF color space name.
//
// Deprecated: it drops the alpha channel. Use DecodePNG and write Alpha as the image's
// /SMask; the DocumentBuilder PNG methods do this.
func DecodePNGToRaw(pngData []byte) (raw []byte, w, h int, colorSpace string, err error) {
	d, err := DecodePNG(pngData)
	if err != nil {
		return nil, 0, 0, "", err
	}
	return d.Raw, d.Width, d.Height, d.ColorSpace, nil
}

// ImageInfo holds processed image data.
type ImageInfo struct {
	Raw               []byte
	WidthPx, HeightPx int
	ColorSpace        string
	IsJPEG            bool
	BitsPerComponent  int
	// Alpha is the PNG's alpha channel (one byte per pixel), or nil if it is opaque.
	Alpha []byte
}

// ProcessImage detects format and decodes if needed.
func ProcessImage(data []byte) (*ImageInfo, error) {
	if bytes.HasPrefix(data, []byte("\x89PNG\r\n\x1a\n")) {
		d, err := DecodePNG(data)
		if err != nil {
			return nil, err
		}
		return &ImageInfo{
			Raw:              d.Raw,
			WidthPx:          d.Width,
			HeightPx:         d.Height,
			ColorSpace:       d.ColorSpace,
			Alpha:            d.Alpha,
			BitsPerComponent: 8,
			IsJPEG:           false,
		}, nil
	}

	if bytes.HasPrefix(data, []byte("\xff\xd8")) {
		w, h, isJPEG, _, err := DetectDimensions(data)
		if err != nil {
			return nil, err
		}
		if isJPEG {
			return &ImageInfo{
				Raw:              data,
				WidthPx:          w,
				HeightPx:         h,
				ColorSpace:       "DeviceRGB",
				BitsPerComponent: 8,
				IsJPEG:           true,
			}, nil
		}
	}

	// Fallback/unknown
	return &ImageInfo{
		Raw:              data,
		BitsPerComponent: 8,
		ColorSpace:       "DeviceRGB",
	}, nil
}
