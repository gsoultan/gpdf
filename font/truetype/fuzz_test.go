package truetype

import (
	"os"
	"testing"
)

// FuzzParse checks that a corrupt font program never panics the parser,
// shaping, metrics or subsetting.
func FuzzParse(f *testing.F) {
	for _, p := range []string{
		"/usr/share/fonts/truetype/dejavu/DejaVuSans.ttf",
		"/usr/share/fonts/truetype/liberation/LiberationSans-Regular.ttf",
		"/usr/share/fonts/opentype/urw-base35/NimbusSans-Regular.otf",
	} {
		if data, err := os.ReadFile(p); err == nil && len(data) < 1<<20 {
			f.Add(data)
		}
	}
	f.Add([]byte("\x00\x01\x00\x00\x00\x01\x00\x00\x00\x00\x00\x00glyf\x00\x00\x00\x00\x00\x00\x00\x1c\x00\x00\x00\x00"))
	f.Add([]byte("OTTO\x00\x01\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00CFF \x00\x00\x00\x00\x00\x00\x00\x1c\x00\x00\x00\x04\x01\x00\x04\x01"))
	f.Add([]byte("ttcf\x00\x01\x00\x00\x00\x00\x00\x01\x00\x00\x00\x0c"))
	f.Fuzz(func(t *testing.T, data []byte) {
		if fonts, err := ParseCollection(data); err == nil {
			for _, ft := range fonts {
				exercise(ft)
			}
		}
		ft, err := Parse(data)
		if err != nil {
			return
		}
		exercise(ft)
	})
}

func exercise(ft *Font) {
	_ = ft.PostScriptName()
	_ = ft.Metrics()
	_, _, _ = ft.CIDSystemInfo()
	used := map[uint16]bool{}
	for _, g := range ft.Shape("Hello, مرحبا 字 fi") {
		used[g.ID] = true
	}
	_ = ft.GlyphWidth('H')
	if ft.IsCFF() {
		_ = ft.CFF()
		_ = ft.ProgramCIDs()
		return
	}
	_ = ft.SubsetGlyphs(used)
	_, _ = ft.Subset(used)
}
