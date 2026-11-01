// Copyright (c) 2026 Gembit Soultan Shirazi <gembit.soultan@gmail.com>

package doc_test

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gsoultan/gpdf/doc"
	imgpkg "github.com/gsoultan/gpdf/doc/image"
	"github.com/gsoultan/gpdf/model"
)

// halfTransparentPNG is 2x1: an opaque green pixel and a fully transparent white one.
func halfTransparentPNG(t *testing.T) []byte {
	t.Helper()
	img := image.NewNRGBA(image.Rect(0, 0, 2, 1))
	img.SetNRGBA(0, 0, color.NRGBA{0, 255, 0, 255})
	img.SetNRGBA(1, 0, color.NRGBA{255, 255, 255, 0})
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

func before(t *testing.T, s, a, b string) {
	t.Helper()
	i, j := strings.Index(s, a), strings.Index(s, b)
	if i < 0 || j < 0 || i > j {
		t.Fatalf("want %q before %q in:\n%s", a, b, s)
	}
}

func TestPaintInCallOrderInterleavesRuns(t *testing.T) {
	b := doc.New().PageSize(200, 200).SetPaintOrder(doc.PaintInCallOrder).AddPage()
	b.FillRect(0, 0, 0, 200, 200, doc.ColorBlue) // background
	b.DrawPNG(10, 10, 50, 50, halfTransparentPNG(t))
	b.DrawText("Over", 20, 20, "Helvetica", 12)
	b.FillRect(0, 100, 100, 10, 10, doc.ColorRed) // badge on top
	d, err := b.Build()
	if err != nil {
		t.Fatal(err)
	}
	ops := firstPageContent(t, d)
	before(t, ops, "200 200 re", "/Im1 Do")
	before(t, ops, "/Im1 Do", "(Over)")
	before(t, ops, "(Over)", "100 100 10 10 re")
}

// The default stays layered (graphics, text, images) for existing callers.
func TestDefaultPaintOrderIsLayered(t *testing.T) {
	b := doc.New().PageSize(200, 200).AddPage()
	b.DrawPNG(10, 10, 50, 50, halfTransparentPNG(t))
	b.DrawText("Text", 20, 20, "Helvetica", 12)
	b.FillRect(0, 0, 0, 200, 200, doc.ColorBlue)
	d, err := b.Build()
	if err != nil {
		t.Fatal(err)
	}
	ops := firstPageContent(t, d)
	before(t, ops, "200 200 re", "(Text)")
	before(t, ops, "(Text)", "/Im1 Do")
}

// A graphic between two text runs of the same color may change the fill color, so the
// second run must set its color again.
func TestPaintInCallOrderRestoresTextColorAfterGraphics(t *testing.T) {
	b := doc.New().PageSize(200, 200).SetPaintOrder(doc.PaintInCallOrder).AddPage()
	b.DrawTextColored("One", 10, 150, "Helvetica", 12, doc.ColorRed)
	b.FillRect(0, 0, 0, 50, 50, doc.ColorBlue)
	b.DrawTextColored("Two", 10, 100, "Helvetica", 12, doc.ColorRed)
	d, err := b.Build()
	if err != nil {
		t.Fatal(err)
	}
	ops := firstPageContent(t, d)
	if strings.Count(ops, "1 0 0 rg") != 2 {
		t.Fatalf("second red run does not set its color again:\n%s", ops)
	}
}

func imageXObject(t *testing.T, d doc.Document) *model.Stream {
	t.Helper()
	pages, err := d.Pages()
	if err != nil {
		t.Fatal(err)
	}
	r := d.(resolver)
	res := pages[0].Dict[model.Name("Resources")]
	if ref, ok := res.(model.Ref); ok {
		res, _ = r.Resolve(ref)
	}
	xo := res.(model.Dict)[model.Name("XObject")].(model.Dict)
	obj, err := r.Resolve(xo[model.Name("Im1")].(model.Ref))
	if err != nil {
		t.Fatal(err)
	}
	return obj.(*model.Stream)
}

// Regression: DecodePNGToRaw dropped alpha and premultiplied the color, so transparent
// pixels were painted black. The PNG's alpha is now written as an /SMask, and image
// samples are FlateDecode-compressed.
func TestPNGAlphaBecomesSoftMaskAndImagesAreCompressed(t *testing.T) {
	b := doc.New().PageSize(100, 100).AddPage()
	b.DrawPNG(0, 0, 100, 100, halfTransparentPNG(t))
	built, err := b.Build()
	if err != nil {
		t.Fatal(err)
	}
	var buf bytes.Buffer
	if err := built.Save(&buf); err != nil {
		t.Fatal(err)
	}
	pdf := buf.Bytes()
	d, err := doc.OpenReaderWithPassword(bytes.NewReader(pdf), int64(len(pdf)), "")
	if err != nil {
		t.Fatal(err)
	}
	im := imageXObject(t, d)
	if f, _ := im.Dict[model.Name("Filter")].(model.Name); f != "FlateDecode" {
		t.Fatalf("image /Filter = %v, want FlateDecode", im.Dict[model.Name("Filter")])
	}
	if want := []byte{0, 255, 0, 255, 255, 255}; !bytes.Equal(im.Content, want) {
		t.Fatalf("image samples = %v, want %v (not premultiplied)", im.Content, want)
	}
	maskRef, ok := im.Dict[model.Name("SMask")].(model.Ref)
	if !ok {
		t.Fatalf("image has no /SMask: %v", im.Dict)
	}
	maskObj, err := d.(resolver).Resolve(maskRef)
	if err != nil {
		t.Fatal(err)
	}
	mask := maskObj.(*model.Stream)
	if f, _ := mask.Dict[model.Name("Filter")].(model.Name); f != "FlateDecode" {
		t.Fatalf("mask /Filter = %v, want FlateDecode", mask.Dict[model.Name("Filter")])
	}
	if !bytes.Equal(mask.Content, []byte{255, 0}) {
		t.Fatalf("mask samples = %v, want [255 0]", mask.Content)
	}

	// Interoperability: poppler must show the red background through the transparent
	// pixel instead of black.
	bin, err := exec.LookPath("pdftoppm")
	if err != nil {
		t.Log("pdftoppm not installed; skipping the render check")
		return
	}
	b = doc.New().PageSize(100, 100).AddPage()
	b.FillRect(0, 0, 0, 100, 100, doc.ColorRed)
	b.DrawPNG(0, 0, 100, 100, halfTransparentPNG(t))
	built, err = b.Build()
	if err != nil {
		t.Fatal(err)
	}
	buf.Reset()
	if err := built.Save(&buf); err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	in := filepath.Join(dir, "in.pdf")
	if err := os.WriteFile(in, buf.Bytes(), 0o600); err != nil {
		t.Fatal(err)
	}
	out := filepath.Join(dir, "out")
	if msg, err := exec.Command(bin, "-png", "-r", "72", "-singlefile", in, out).CombinedOutput(); err != nil {
		t.Fatalf("pdftoppm: %v\n%s", err, msg)
	}
	f, err := os.Open(out + ".png")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	rendered, err := png.Decode(f)
	if err != nil {
		t.Fatal(err)
	}
	check := func(x, y int, want color.NRGBA) {
		got := color.NRGBAModel.Convert(rendered.At(x, y)).(color.NRGBA)
		if diff(got.R, want.R) > 8 || diff(got.G, want.G) > 8 || diff(got.B, want.B) > 8 {
			t.Errorf("pixel (%d,%d) = %v, want %v", x, y, got, want)
		}
	}
	check(25, 50, color.NRGBA{0, 255, 0, 255}) // opaque green pixel
	check(75, 50, color.NRGBA{255, 0, 0, 255}) // red shows through the transparent pixel
}

func diff(a, b uint8) int {
	if a > b {
		return int(a - b)
	}
	return int(b - a)
}

func TestDecodePNGKeepsGrayAlphaGray(t *testing.T) {
	img := image.NewNRGBA(image.Rect(0, 0, 1, 2))
	img.SetNRGBA(0, 0, color.NRGBA{100, 100, 100, 128})
	img.SetNRGBA(0, 1, color.NRGBA{200, 200, 200, 255})
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatal(err)
	}
	d, err := imgpkg.DecodePNG(buf.Bytes())
	if err != nil {
		t.Fatal(err)
	}
	if d.ColorSpace != "DeviceGray" || !bytes.Equal(d.Raw, []byte{100, 200}) || !bytes.Equal(d.Alpha, []byte{128, 255}) {
		t.Fatalf("got %s %v alpha %v", d.ColorSpace, d.Raw, d.Alpha)
	}
}

func TestDecodePNGOpaqueHasNoAlpha(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 1, 1))
	img.Set(0, 0, color.RGBA{1, 2, 3, 255})
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatal(err)
	}
	d, err := imgpkg.DecodePNG(buf.Bytes())
	if err != nil {
		t.Fatal(err)
	}
	if d.Alpha != nil || !bytes.Equal(d.Raw, []byte{1, 2, 3}) {
		t.Fatalf("got %v alpha %v", d.Raw, d.Alpha)
	}
}

// Regression: FlowBuilder.Image passed the PNG file bytes as raw samples with no size.
func TestDrawImageWithDecodesPNGFile(t *testing.T) {
	b := doc.New().PageSize(100, 100).AddPage()
	b.DrawImageWith(doc.ImageOptions{Data: halfTransparentPNG(t), X: 0, Y: 0, Width: 50, Height: 50, PageIndex: -1})
	d, err := b.Build()
	if err != nil {
		t.Fatal(err)
	}
	im := imageXObject(t, d)
	if w, _ := im.Dict[model.Name("Width")].(model.Integer); w != 2 {
		t.Fatalf("Width = %v, want 2", im.Dict[model.Name("Width")])
	}
	if _, ok := im.Dict[model.Name("SMask")]; !ok {
		t.Fatal("no /SMask")
	}
}

// Adobe (APP14) CMYK JPEGs store inverted samples and need /Decode [1 0 1 0 1 0 1 0];
// other JPEGs must not get it.
func TestAdobeCMYKJPEGGetsInvertedDecode(t *testing.T) {
	adobe := append([]byte{0xFF, 0xD8, 0xFF, 0xEE, 0, 14}, "Adobe\x00\x64\x00\x00\x00\x00\x02"...)
	adobe = append(adobe, 0xFF, 0xDA, 0, 2)
	plain := []byte{0xFF, 0xD8, 0xFF, 0xE0, 0, 4, 0, 0, 0xFF, 0xDA, 0, 2}
	for _, tc := range []struct {
		name string
		data []byte
		want bool
	}{{"adobe", adobe, true}, {"plain", plain, false}} {
		b := doc.New().PageSize(50, 50).AddPage()
		b.DrawJPEG(0, 0, 10, 10, tc.data, 1, 1, "DeviceCMYK")
		d, err := b.Build()
		if err != nil {
			t.Fatal(err)
		}
		_, got := imageXObject(t, d).Dict[model.Name("Decode")]
		if got != tc.want {
			t.Errorf("%s: /Decode present = %v, want %v", tc.name, got, tc.want)
		}
	}
}
