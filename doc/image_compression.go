// Copyright (c) 2026 Gembit Soultan Shirazi <gembit.soultan@gmail.com>

package doc

import (
	"bytes"
	"fmt"
	"image"
	"image/jpeg"
	"math"

	"github.com/gsoultan/gpdf/model"
)

// ImageCompression controls how the images drawn on pages are stored. Raw samples
// are always Flate-compressed with a PNG predictor (lossless); these options trade
// quality for size on top of that.
type ImageCompression struct {
	// MaxDPI downsamples an image whose pixel density at its placed size is
	// higher than this (a 3000-pixel-wide photo placed 3 inches wide is 1000 dpi;
	// 150 to 300 dpi is enough for print). 0 keeps every pixel.
	MaxDPI float64
	// JPEGQuality, from 1 to 100, re-encodes 8-bit RGB and gray images that have no
	// transparency as JPEG at that quality (lossy; 75 to 90 is a usual range).
	// 0 keeps them lossless. Images that arrive as JPEG are kept as they are
	// unless MaxDPI makes them smaller, in which case they are re-encoded at this
	// quality, or at 85 when it is 0.
	JPEGQuality int
}

// CompressImages sets how images are stored. Without it images keep their pixels
// and are compressed losslessly.
func (b *DocumentBuilder) CompressImages(c ImageCompression) *DocumentBuilder {
	if c.JPEGQuality < 0 || c.JPEGQuality > 100 {
		b.setErr(fmt.Errorf("CompressImages: JPEGQuality %d is not in 0..100", c.JPEGQuality))
		return b
	}
	if c.MaxDPI < 0 {
		b.setErr(fmt.Errorf("CompressImages: MaxDPI %g is negative", c.MaxDPI))
		return b
	}
	b.imageCompression = c
	return b
}

// components returns the number of samples per pixel of a device colour space, or 0.
func components(colorSpace string) int {
	switch colorSpace {
	case "DeviceGray", "CalGray", "G":
		return 1
	case "DeviceRGB", "CalRGB", "RGB":
		return 3
	case "DeviceCMYK", "CMYK":
		return 4
	}
	return 0
}

// prepareImage applies the image compression settings to a run: downsampling to
// MaxDPI and JPEG re-encoding. It is called once, when the image is added.
func (b *DocumentBuilder) prepareImage(run *imageRun) error {
	c := b.imageCompression
	if run.WidthPx <= 0 || run.HeightPx <= 0 {
		return nil
	}
	if c.MaxDPI > 0 {
		if err := downsampleRun(run, c.MaxDPI, c.JPEGQuality); err != nil {
			return err
		}
	}
	if c.JPEGQuality > 0 && !run.IsJPEG && run.BitsPerComponent == 8 && !run.HasMask {
		if n := components(run.ColorSpace); n == 1 || n == 3 {
			data, err := encodeJPEG(run.Raw, run.WidthPx, run.HeightPx, n, c.JPEGQuality)
			if err != nil {
				return err
			}
			run.Raw, run.IsJPEG = data, true
		}
	}
	return nil
}

// placedSize returns the size on the page the image is drawn at, in points.
func placedSize(run *imageRun) (w, h float64) {
	if run.Matrix != [6]float64{} {
		return math.Hypot(run.Matrix[0], run.Matrix[1]), math.Hypot(run.Matrix[2], run.Matrix[3])
	}
	w, h = run.WidthPt, run.HeightPt
	if w <= 0 {
		w = float64(run.WidthPx)
	}
	if h <= 0 {
		h = float64(run.HeightPx)
	}
	return w, h
}

// downsampleRun reduces the image to at most maxDPI at its placed size.
func downsampleRun(run *imageRun, maxDPI float64, quality int) error {
	wPt, hPt := placedSize(run)
	if wPt <= 0 || hPt <= 0 {
		return nil
	}
	scale := math.Min(wPt/72*maxDPI/float64(run.WidthPx), hPt/72*maxDPI/float64(run.HeightPx))
	if scale >= 1 {
		return nil
	}
	tw := max(int(math.Ceil(float64(run.WidthPx)*scale)), 1)
	th := max(int(math.Ceil(float64(run.HeightPx)*scale)), 1)
	if tw >= run.WidthPx || th >= run.HeightPx {
		return nil
	}
	if run.IsJPEG {
		raw, cs, err := decodeJPEGToRaw(run.Raw)
		if err != nil {
			return fmt.Errorf("downsample: %w", err)
		}
		n := components(cs)
		raw = resampleBox(raw, run.WidthPx, run.HeightPx, n, tw, th)
		run.WidthPx, run.HeightPx, run.ColorSpace, run.BitsPerComponent = tw, th, cs, 8
		if n == 4 {
			// Go's JPEG encoder writes no CMYK: keep the samples lossless.
			run.Raw, run.IsJPEG = raw, false
		} else {
			if quality == 0 {
				quality = 85
			}
			data, err := encodeJPEG(raw, tw, th, n, quality)
			if err != nil {
				return err
			}
			run.Raw = data
		}
	} else {
		n := components(run.ColorSpace)
		if n == 0 || run.BitsPerComponent != 8 || len(run.Raw) < run.WidthPx*run.HeightPx*n {
			return nil
		}
		run.Raw = resampleBox(run.Raw, run.WidthPx, run.HeightPx, n, tw, th)
		run.WidthPx, run.HeightPx = tw, th
	}
	if run.HasMask && run.MaskWidth > 0 && run.MaskHeight > 0 && len(run.Mask) >= run.MaskWidth*run.MaskHeight {
		mw := max(int(math.Ceil(float64(run.MaskWidth)*scale)), 1)
		mh := max(int(math.Ceil(float64(run.MaskHeight)*scale)), 1)
		if mw < run.MaskWidth && mh < run.MaskHeight {
			run.Mask = resampleBox(run.Mask, run.MaskWidth, run.MaskHeight, 1, mw, mh)
			run.MaskWidth, run.MaskHeight = mw, mh
		}
	}
	return nil
}

// resampleBox scales 8-bit samples (n per pixel) from w×h to tw×th, averaging the
// source pixels that fall into each target pixel.
func resampleBox(src []byte, w, h, n, tw, th int) []byte {
	out := make([]byte, tw*th*n)
	sum := make([]float64, n)
	for ty := range th {
		y0 := ty * h / th
		y1 := max((ty+1)*h/th, y0+1)
		for tx := range tw {
			x0 := tx * w / tw
			x1 := max((tx+1)*w/tw, x0+1)
			clear(sum)
			for y := y0; y < y1; y++ {
				row := src[(y*w+x0)*n : (y*w+x1)*n]
				for i, v := range row {
					sum[i%n] += float64(v)
				}
			}
			count := float64((y1 - y0) * (x1 - x0))
			o := (ty*tw + tx) * n
			for i := range n {
				out[o+i] = byte(sum[i]/count + 0.5)
			}
		}
	}
	return out
}

// encodeJPEG encodes 8-bit gray (n=1) or RGB (n=3) samples as a baseline JPEG.
func encodeJPEG(raw []byte, w, h, n, quality int) ([]byte, error) {
	var img image.Image
	switch n {
	case 1:
		g := image.NewGray(image.Rect(0, 0, w, h))
		copy(g.Pix, raw)
		img = g
	case 3:
		rgba := image.NewRGBA(image.Rect(0, 0, w, h))
		for i := 0; i < w*h && i*3+2 < len(raw); i++ {
			rgba.Pix[i*4], rgba.Pix[i*4+1], rgba.Pix[i*4+2], rgba.Pix[i*4+3] = raw[i*3], raw[i*3+1], raw[i*3+2], 255
		}
		img = rgba
	default:
		return nil, fmt.Errorf("encodeJPEG: %d components", n)
	}
	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: quality}); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// decodeJPEGToRaw decodes a JPEG to 8-bit gray, RGB or CMYK samples.
func decodeJPEGToRaw(data []byte) ([]byte, string, error) {
	img, err := jpeg.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, "", err
	}
	bounds := img.Bounds()
	w, h := bounds.Dx(), bounds.Dy()
	switch im := img.(type) {
	case *image.Gray:
		raw := make([]byte, 0, w*h)
		for y := range h {
			raw = append(raw, im.Pix[y*im.Stride:y*im.Stride+w]...)
		}
		return raw, "DeviceGray", nil
	case *image.CMYK:
		raw := make([]byte, 0, w*h*4)
		for y := range h {
			raw = append(raw, im.Pix[y*im.Stride:y*im.Stride+w*4]...)
		}
		return raw, "DeviceCMYK", nil
	}
	raw := make([]byte, 0, w*h*3)
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			r, g, b, _ := img.At(x, y).RGBA()
			raw = append(raw, byte(r>>8), byte(g>>8), byte(b>>8))
		}
	}
	return raw, "DeviceRGB", nil
}

// predictorParms returns the /DecodeParms that make Flate apply a PNG predictor
// to image samples, which compresses photographs and gradients far better than
// plain Flate, or nil when the samples do not fit a predictor row.
func predictorParms(colorSpace string, bpc, width int) model.Object {
	n := components(colorSpace)
	if n == 0 || width <= 0 {
		return nil
	}
	switch bpc {
	case 1, 2, 4, 8, 16:
	default:
		return nil
	}
	return model.Dict{
		model.Name("Predictor"):        model.Integer(15),
		model.Name("Colors"):           model.Integer(int64(n)),
		model.Name("BitsPerComponent"): model.Integer(int64(bpc)),
		model.Name("Columns"):          model.Integer(int64(width)),
	}
}

// imageStreams builds the image XObject of a run and, when the run has a soft
// mask, the mask's XObject (not yet linked: the caller sets /SMask).
func (b *DocumentBuilder) imageStreams(im imageRun) (img, mask *model.Stream) {
	if im.IsJPEG {
		img = b.jpegXObjectStream(im)
	} else {
		img = b.imageXObjectStream(im)
	}
	if im.HasMask && len(im.Mask) > 0 && im.MaskWidth > 0 && im.MaskHeight > 0 {
		maskDict := model.Dict{
			model.Name("Type"):             model.Name("XObject"),
			model.Name("Subtype"):          model.Name("Image"),
			model.Name("Width"):            model.Integer(int64(im.MaskWidth)),
			model.Name("Height"):           model.Integer(int64(im.MaskHeight)),
			model.Name("BitsPerComponent"): model.Integer(8),
			model.Name("ColorSpace"):       model.Name("DeviceGray"),
			model.Name("Filter"):           model.Name("FlateDecode"),
			model.Name("Length"):           model.Integer(int64(len(im.Mask))),
		}
		if p := predictorParms("DeviceGray", 8, im.MaskWidth); p != nil {
			maskDict[model.Name("DecodeParms")] = p
		}
		mask = &model.Stream{Dict: maskDict, Content: im.Mask}
	}
	return img, mask
}
