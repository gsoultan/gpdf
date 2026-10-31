package builder

import (
	"fmt"

	imgpkg "github.com/gsoultan/gpdf/doc/image"
)

// SetPNG decodes pngData into the run's image data: 8-bit DeviceGray or DeviceRGB samples
// and, when the PNG has transparency, its alpha channel as a soft mask (/SMask), so
// transparent pixels show what is underneath instead of turning black.
func (r *ImageRun) SetPNG(pngData []byte) error {
	d, err := imgpkg.DecodePNG(pngData)
	if err != nil {
		return err
	}
	if d.Width == 0 || d.Height == 0 {
		return fmt.Errorf("invalid dimensions %dx%d", d.Width, d.Height)
	}
	r.Raw, r.WidthPx, r.HeightPx = d.Raw, d.Width, d.Height
	r.BitsPerComponent, r.ColorSpace, r.IsJPEG = 8, d.ColorSpace, false
	r.HasMask, r.Mask, r.MaskWidth, r.MaskHeight = false, nil, 0, 0
	if d.Alpha != nil {
		r.HasMask, r.Mask, r.MaskWidth, r.MaskHeight = true, d.Alpha, d.Width, d.Height
	}
	return nil
}
