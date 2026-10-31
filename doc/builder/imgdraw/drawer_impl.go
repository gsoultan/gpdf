package imgdraw

import (
	"fmt"

	"github.com/gsoultan/gpdf/doc/builder"
	"github.com/gsoultan/gpdf/doc/tagged"
)

type drawer struct{}

func NewDrawer() Drawer {
	return &drawer{}
}

func lastPage(pa builder.PageAccess) *builder.Page {
	n := pa.PageCount()
	if n == 0 {
		return nil
	}
	return pa.PageAt(n - 1)
}

func defaultColorSpace(cs string) string {
	if cs == "" {
		return "DeviceRGB"
	}
	return cs
}

func defaultBPC(bpc int) int {
	if bpc <= 0 {
		return 8
	}
	return bpc
}

func (d *drawer) DrawImage(pa builder.PageAccess, x, y, widthPt, heightPt float64, raw []byte, widthPx, heightPx, bitsPerComponent int, colorSpace string) {
	p := lastPage(pa)
	if p == nil {
		return
	}
	p.AddImage(builder.ImageRun{
		X: x, Y: y, WidthPt: widthPt, HeightPt: heightPt,
		Raw: raw, WidthPx: widthPx, HeightPx: heightPx,
		BitsPerComponent: defaultBPC(bitsPerComponent), ColorSpace: defaultColorSpace(colorSpace),
	})
}

func (d *drawer) DrawJPEG(pa builder.PageAccess, x, y, widthPt, heightPt float64, jpegData []byte, widthPx, heightPx int, colorSpace string) {
	p := lastPage(pa)
	if p == nil || len(jpegData) == 0 {
		return
	}
	p.AddImage(builder.ImageRun{
		X: x, Y: y, WidthPt: widthPt, HeightPt: heightPt,
		Raw: jpegData, WidthPx: widthPx, HeightPx: heightPx,
		BitsPerComponent: 8, ColorSpace: defaultColorSpace(colorSpace), IsJPEG: true,
	})
}

func (d *drawer) DrawPNG(pa builder.PageAccess, x, y, widthPt, heightPt float64, pngData []byte) error {
	p := lastPage(pa)
	if p == nil || len(pngData) == 0 {
		return nil
	}
	var run builder.ImageRun
	if err := run.SetPNG(pngData); err != nil {
		return fmt.Errorf("DrawPNG: %w", err)
	}
	run.X, run.Y, run.WidthPt, run.HeightPt = x, y, widthPt, heightPt
	p.AddImage(run)
	return nil
}

func (d *drawer) DrawCircularImage(pa builder.PageAccess, x, y, widthPt, heightPt float64, raw []byte, widthPx, heightPx, bitsPerComponent int, colorSpace string, cx, cy, radius float64) {
	p := lastPage(pa)
	if p == nil {
		return
	}
	p.AddImage(builder.ImageRun{
		X: x, Y: y, WidthPt: widthPt, HeightPt: heightPt,
		Raw: raw, WidthPx: widthPx, HeightPx: heightPx,
		BitsPerComponent: defaultBPC(bitsPerComponent), ColorSpace: defaultColorSpace(colorSpace),
		ClipCircle: true, ClipCX: cx, ClipCY: cy, ClipR: radius,
	})
}

func (d *drawer) DrawCircularJPEG(pa builder.PageAccess, x, y, widthPt, heightPt float64, jpegData []byte, widthPx, heightPx int, colorSpace string, cx, cy, radius float64) {
	p := lastPage(pa)
	if p == nil || len(jpegData) == 0 {
		return
	}
	p.AddImage(builder.ImageRun{
		X: x, Y: y, WidthPt: widthPt, HeightPt: heightPt,
		Raw: jpegData, WidthPx: widthPx, HeightPx: heightPx,
		BitsPerComponent: 8, ColorSpace: defaultColorSpace(colorSpace), IsJPEG: true,
		ClipCircle: true, ClipCX: cx, ClipCY: cy, ClipR: radius,
	})
}

func (d *drawer) DrawCircularPNG(pa builder.PageAccess, x, y, widthPt, heightPt float64, pngData []byte, cx, cy, radius float64) error {
	p := lastPage(pa)
	if p == nil || len(pngData) == 0 {
		return nil
	}
	var run builder.ImageRun
	if err := run.SetPNG(pngData); err != nil {
		return fmt.Errorf("DrawCircularPNG: %w", err)
	}
	run.X, run.Y, run.WidthPt, run.HeightPt = x, y, widthPt, heightPt
	run.ClipCircle, run.ClipCX, run.ClipCY, run.ClipR = true, cx, cy, radius
	p.AddImage(run)
	return nil
}

func (d *drawer) DrawImageWithOpacity(pa builder.PageAccess, x, y, widthPt, heightPt float64, raw []byte, widthPx, heightPx, bitsPerComponent int, colorSpace string, opacity float64) {
	p := lastPage(pa)
	if p == nil {
		return
	}
	p.AddImage(builder.ImageRun{
		X: x, Y: y, WidthPt: widthPt, HeightPt: heightPt,
		Raw: raw, WidthPx: widthPx, HeightPx: heightPx,
		BitsPerComponent: defaultBPC(bitsPerComponent), ColorSpace: defaultColorSpace(colorSpace),
		Opacity: opacity,
	})
}

func (d *drawer) DrawJPEGWithOpacity(pa builder.PageAccess, x, y, widthPt, heightPt float64, jpegData []byte, widthPx, heightPx int, colorSpace string, opacity float64) {
	p := lastPage(pa)
	if p == nil || len(jpegData) == 0 {
		return
	}
	p.AddImage(builder.ImageRun{
		X: x, Y: y, WidthPt: widthPt, HeightPt: heightPt,
		Raw: jpegData, WidthPx: widthPx, HeightPx: heightPx,
		BitsPerComponent: 8, ColorSpace: defaultColorSpace(colorSpace), IsJPEG: true,
		Opacity: opacity,
	})
}

func (d *drawer) DrawPNGWithOpacity(pa builder.PageAccess, x, y, widthPt, heightPt float64, pngData []byte, opacity float64) error {
	p := lastPage(pa)
	if p == nil || len(pngData) == 0 {
		return nil
	}
	var run builder.ImageRun
	if err := run.SetPNG(pngData); err != nil {
		return fmt.Errorf("DrawPNGWithOpacity: %w", err)
	}
	run.X, run.Y, run.WidthPt, run.HeightPt = x, y, widthPt, heightPt
	run.Opacity = opacity
	p.AddImage(run)
	return nil
}

func (d *drawer) DrawJPEGRotated(pa builder.PageAccess, x, y, widthPt, heightPt float64, jpegData []byte, widthPx, heightPx int, colorSpace string, rotateDeg float64) {
	p := lastPage(pa)
	if p == nil || len(jpegData) == 0 {
		return
	}
	p.AddImage(builder.ImageRun{
		X: x, Y: y, WidthPt: widthPt, HeightPt: heightPt,
		Raw: jpegData, WidthPx: widthPx, HeightPx: heightPx,
		BitsPerComponent: 8, ColorSpace: defaultColorSpace(colorSpace), IsJPEG: true,
		RotateDeg: rotateDeg,
	})
}

func (d *drawer) DrawPNGRotated(pa builder.PageAccess, x, y, widthPt, heightPt float64, pngData []byte, rotateDeg float64) error {
	p := lastPage(pa)
	if p == nil || len(pngData) == 0 {
		return nil
	}
	var run builder.ImageRun
	if err := run.SetPNG(pngData); err != nil {
		return fmt.Errorf("DrawPNGRotated: %w", err)
	}
	run.X, run.Y, run.WidthPt, run.HeightPt = x, y, widthPt, heightPt
	run.RotateDeg = rotateDeg
	p.AddImage(run)
	return nil
}

func (d *drawer) DrawTaggedFigure(pa builder.PageAccess, pageIndex int, x, y, widthPt, heightPt float64, raw []byte, widthPx, heightPx, bitsPerComponent int, colorSpace, alt string) tagged.Figure {
	if !pa.ValidPageIndex(pageIndex) || len(raw) == 0 {
		return tagged.Figure{}
	}
	mcid := pa.NextMCID(pageIndex)
	p := pa.PageAt(pageIndex)
	p.AddImage(builder.ImageRun{
		X: x, Y: y, WidthPt: widthPt, HeightPt: heightPt,
		Raw: raw, WidthPx: widthPx, HeightPx: heightPx,
		BitsPerComponent: defaultBPC(bitsPerComponent), ColorSpace: defaultColorSpace(colorSpace),
		MCID: mcid, HasMCID: true,
	})
	return tagged.Figure{PageIndex: pageIndex, MCID: mcid, Alt: alt}
}

func (d *drawer) DrawTaggedJPEG(pa builder.PageAccess, pageIndex int, x, y, widthPt, heightPt float64, jpegData []byte, widthPx, heightPx int, colorSpace, alt string) tagged.Figure {
	if !pa.ValidPageIndex(pageIndex) || len(jpegData) == 0 {
		return tagged.Figure{}
	}
	mcid := pa.NextMCID(pageIndex)
	p := pa.PageAt(pageIndex)
	p.AddImage(builder.ImageRun{
		X: x, Y: y, WidthPt: widthPt, HeightPt: heightPt,
		Raw: jpegData, WidthPx: widthPx, HeightPx: heightPx,
		BitsPerComponent: 8, ColorSpace: defaultColorSpace(colorSpace), IsJPEG: true,
		MCID: mcid, HasMCID: true,
	})
	return tagged.Figure{PageIndex: pageIndex, MCID: mcid, Alt: alt}
}

func (d *drawer) DrawTaggedPNG(pa builder.PageAccess, pageIndex int, x, y, widthPt, heightPt float64, pngData []byte, alt string) (tagged.Figure, error) {
	if !pa.ValidPageIndex(pageIndex) || len(pngData) == 0 {
		return tagged.Figure{}, nil
	}
	var run builder.ImageRun
	if err := run.SetPNG(pngData); err != nil {
		return tagged.Figure{}, fmt.Errorf("DrawTaggedPNG: %w", err)
	}
	mcid := pa.NextMCID(pageIndex)
	run.X, run.Y, run.WidthPt, run.HeightPt = x, y, widthPt, heightPt
	run.MCID, run.HasMCID = mcid, true
	pa.PageAt(pageIndex).AddImage(run)
	return tagged.Figure{PageIndex: pageIndex, MCID: mcid, Alt: alt}, nil
}
