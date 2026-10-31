package doc

import (
	"bytes"
	"fmt"
	"image/color"
	"image/jpeg"

	imgpkg "github.com/gsoultan/gpdf/doc/image"
	taggedpkg "github.com/gsoultan/gpdf/doc/tagged"
)

// ImageOptions is a type alias for image.Options.
type ImageOptions = imgpkg.Options

// DrawImageWith queues an image using the options struct.
func (b *DocumentBuilder) DrawImageWith(opts ImageOptions) *DocumentBuilder {
	if b.err != nil {
		return b
	}
	if len(b.pc.pages) == 0 {
		return b
	}

	colorSpace := opts.ColorSpace
	if colorSpace == "" {
		colorSpace = "DeviceRGB"
	}
	bpc := opts.BitsPerComponent
	if bpc <= 0 {
		bpc = 8
	}
	pageIndex := opts.PageIndex
	if pageIndex < 0 {
		pageIndex = len(b.pc.pages) - 1
	}
	if !b.pc.validPageIndex(pageIndex) {
		return b
	}

	run := imageRun{
		X: opts.X, Y: opts.Y,
		WidthPt: opts.Width, HeightPt: opts.Height,
		Matrix: opts.Matrix,
		Raw:    opts.Data, WidthPx: opts.PixelWidth, HeightPx: opts.PixelHeight,
		BitsPerComponent: bpc, ColorSpace: colorSpace,
		IsJPEG: opts.IsJPEG, Opacity: opts.Opacity, RotateDeg: opts.RotateDeg,
		ClipCircle: opts.ClipCircle, ClipCX: opts.ClipCX, ClipCY: opts.ClipCY, ClipR: opts.ClipRadius,
		IsArtifact: opts.IsArtifact,
		HasMask:    opts.HasMask, Mask: opts.Mask, MaskWidth: opts.MaskWidth, MaskHeight: opts.MaskHeight,
	}
	if err := decodeImageFile(&run); err != nil {
		b.setErr(fmt.Errorf("DrawImageWith: %w", err))
		return b
	}
	b.pc.pages[pageIndex].AddImage(run)
	return b
}

// decodeImageFile handles a run whose data is a PNG or JPEG file rather than raw samples
// (no pixel size given): a PNG is decoded with its alpha channel as a soft mask, and a
// JPEG is passed through with DCTDecode.
func decodeImageFile(run *imageRun) error {
	if run.WidthPx > 0 && run.HeightPx > 0 {
		return nil
	}
	switch {
	case bytes.HasPrefix(run.Raw, []byte("\x89PNG\r\n\x1a\n")):
		mask, maskW, maskH, hasMask := run.Mask, run.MaskWidth, run.MaskHeight, run.HasMask
		if err := run.SetPNG(run.Raw); err != nil {
			return err
		}
		if hasMask && !run.HasMask {
			run.HasMask, run.Mask, run.MaskWidth, run.MaskHeight = true, mask, maskW, maskH
		}
	case bytes.HasPrefix(run.Raw, []byte("\xff\xd8")):
		w, h, isJPEG, _, err := imgpkg.DetectDimensions(run.Raw)
		if err != nil {
			return err
		}
		if !isJPEG {
			return fmt.Errorf("unsupported image data")
		}
		cs := "DeviceRGB"
		if cfg, err := jpegColorComponents(run.Raw); err == nil {
			cs = cfg
		}
		run.WidthPx, run.HeightPx, run.BitsPerComponent, run.ColorSpace, run.IsJPEG = w, h, 8, cs, true
	}
	return nil
}

// DrawTaggedImageWith queues a tagged image using the options struct.
func (b *DocumentBuilder) DrawTaggedImageWith(opts ImageOptions) *DocumentBuilder {
	if b.err != nil {
		return b
	}
	pageIndex := opts.PageIndex
	if !b.pc.validPageIndex(pageIndex) {
		return b
	}

	colorSpace := opts.ColorSpace
	if colorSpace == "" {
		colorSpace = "DeviceRGB"
	}
	bpc := opts.BitsPerComponent
	if bpc <= 0 {
		bpc = 8
	}

	b.useTagged = true
	ps := &b.pc.pages[pageIndex]
	mcid := ps.NextMCID
	ps.NextMCID++
	b.tagging.Figures = append(b.tagging.Figures, taggedpkg.Figure{PageIndex: pageIndex, MCID: mcid, Alt: opts.AltText})
	b.tagging.RecordSectionFigure(len(b.tagging.Figures) - 1)

	run := imageRun{
		X: opts.X, Y: opts.Y,
		WidthPt: opts.Width, HeightPt: opts.Height,
		Raw: opts.Data, WidthPx: opts.PixelWidth, HeightPx: opts.PixelHeight,
		BitsPerComponent: bpc, ColorSpace: colorSpace,
		IsJPEG: opts.IsJPEG, MCID: mcid, HasMCID: true,
	}
	if err := decodeImageFile(&run); err != nil {
		b.setErr(fmt.Errorf("DrawTaggedImageWith: %w", err))
		return b
	}
	ps.AddImage(run)
	return b
}

// jpegColorComponents returns the PDF color space of a JPEG from its component count.
func jpegColorComponents(data []byte) (string, error) {
	cfg, err := jpeg.DecodeConfig(bytes.NewReader(data))
	if err != nil {
		return "", err
	}
	switch cfg.ColorModel {
	case color.GrayModel:
		return "DeviceGray", nil
	case color.CMYKModel:
		return "DeviceCMYK", nil
	}
	return "DeviceRGB", nil
}
