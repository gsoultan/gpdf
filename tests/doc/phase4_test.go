package doc_test

import (
	"bytes"
	"errors"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gsoultan/gpdf/doc"
	"github.com/gsoultan/gpdf/model"
	"github.com/gsoultan/gpdf/reader"
)

// gradientPNG returns a w×h RGB(A) PNG with a smooth gradient, which predictors
// compress well.
func gradientPNG(t *testing.T, w, h int, alpha bool) []byte {
	t.Helper()
	img := image.NewNRGBA(image.Rect(0, 0, w, h))
	for y := range h {
		for x := range w {
			a := uint8(255)
			if alpha {
				a = uint8(x * 255 / max(w-1, 1))
			}
			img.Set(x, y, color.NRGBA{R: uint8(x * 255 / max(w-1, 1)), G: uint8(y * 255 / max(h-1, 1)), B: 128, A: a})
		}
	}
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

func gradientJPEG(t *testing.T, w, h int) []byte {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := range h {
		for x := range w {
			img.Set(x, y, color.RGBA{R: uint8(x * 255 / max(w-1, 1)), G: uint8(y * 255 / max(h-1, 1)), B: 128, A: 255})
		}
	}
	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: 90}); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

// pdfImages returns the number of images pdfimages lists, or -1 when it is not installed.
func pdfImages(t *testing.T, pdf []byte) int {
	t.Helper()
	bin, err := exec.LookPath("pdfimages")
	if err != nil {
		return -1
	}
	path := filepath.Join(t.TempDir(), "x.pdf")
	os.WriteFile(path, pdf, 0o600)
	out, err := exec.Command(bin, "-list", path).CombinedOutput()
	if err != nil {
		t.Errorf("pdfimages: %v\n%s", err, out)
		return -1
	}
	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	return max(len(lines)-2, 0)
}

func firstImage(t *testing.T, pdf []byte) reader.ImageInfo {
	t.Helper()
	d := reopen(t, pdf)
	imgs, err := d.ReadImages()
	if err != nil {
		t.Fatal(err)
	}
	if len(imgs) == 0 {
		t.Fatal("no image found")
	}
	return imgs[0]
}

// Object streams and a cross-reference stream are the default layout; the
// classic table is available on request. Both read back with the reader and poppler.
func TestObjectStreamsByDefault(t *testing.T) {
	build := func(classic bool) []byte {
		b := doc.New().Title("Layout").PageSize(300, 300)
		if classic {
			b.NoObjectStreams()
		}
		for i := range 3 {
			b.AddPage()
			b.DrawText("Object streams page", 20, 250, "Helvetica", 12)
			b.DrawPNG(20, 100, 40, 40, tinyPNG(t))
			b.AddLinkToURL(i, 20, 20, 100, 40, "https://example.org/")
		}
		return savePDF(t, b)
	}
	compact, classic := build(false), build(true)
	if !bytes.Contains(compact, []byte("/ObjStm")) || !bytes.Contains(compact, []byte("/XRef")) || bytes.Contains(compact, []byte("trailer")) {
		t.Error("default output should use object streams and a cross-reference stream")
	}
	if bytes.Contains(classic, []byte("/ObjStm")) || bytes.Contains(classic, []byte("/XRef")) || !bytes.Contains(classic, []byte("trailer")) {
		t.Error("NoObjectStreams output should use a classic table")
	}
	if len(compact) >= len(classic) {
		t.Errorf("object streams did not shrink the file: %d >= %d", len(compact), len(classic))
	}
	for name, pdf := range map[string][]byte{"compact": compact, "classic": classic} {
		d := reopen(t, pdf)
		pages, err := d.Pages()
		if err != nil || len(pages) != 3 {
			t.Fatalf("%s: pages = %d, %v", name, len(pages), err)
		}
		if txt, _ := d.ReadContent(); !strings.Contains(txt, "Object streams page") {
			t.Errorf("%s: text = %q", name, txt)
		}
		if txt := pdfText(t, pdf); txt != "" && !strings.Contains(txt, "Object streams page") {
			t.Errorf("%s: pdftotext = %q", name, txt)
		}
		links := 0
		for _, p := range pages {
			if annots, _ := p.Dict[model.Name("Annots")].(model.Array); len(annots) > 0 {
				links += len(annots)
			}
		}
		if links != 3 {
			t.Errorf("%s: %d link annotations, want 3", name, links)
		}
	}
}

// Encrypted output keeps the object streams; strings inside them are not
// encrypted on their own, which poppler and the reader both expect.
func TestEncryptedObjectStreams(t *testing.T) {
	b := doc.New().Title("Secret layout").PageSize(300, 200).AddPage()
	b.DrawText("Encrypted object streams", 20, 100, "Helvetica", 12)
	built, err := b.Build()
	if err != nil {
		t.Fatal(err)
	}
	for name, save := range map[string]func(*bytes.Buffer) error{
		"rc4":    func(w *bytes.Buffer) error { return built.SaveWithPassword(w, "user", "owner") },
		"aes256": func(w *bytes.Buffer) error { return built.SaveWithAES256Password(w, "user", "owner") },
	} {
		var out bytes.Buffer
		if err := save(&out); err != nil {
			t.Fatal(err)
		}
		pdf := out.Bytes()
		if !bytes.Contains(pdf, []byte("/ObjStm")) {
			t.Errorf("%s: no object streams", name)
		}
		if bytes.Contains(pdf, []byte("Secret layout")) {
			t.Errorf("%s: title is readable in the output", name)
		}
		d, err := doc.OpenReaderWithPassword(bytes.NewReader(pdf), int64(len(pdf)), "user")
		if err != nil {
			t.Fatalf("%s: open: %v", name, err)
		}
		if info, _ := d.Info(); info == nil || model.DecodeTextString([]byte(info[model.Name("Title")].(model.String))) != "Secret layout" {
			t.Errorf("%s: info = %v", name, info)
		}
		if txt, _ := d.ReadContent(); !strings.Contains(txt, "Encrypted object streams") {
			t.Errorf("%s: text = %q", name, txt)
		}
		if txt := pdfText(t, pdf, "-upw", "user"); txt != "" && !strings.Contains(txt, "Encrypted object streams") {
			t.Errorf("%s: pdftotext = %q", name, txt)
		}
	}
}

// Raw image samples are stored with a PNG predictor, which the reader undoes;
// the samples come back unchanged and the file is smaller than without it.
func TestImagePredictor(t *testing.T) {
	b := doc.New().PageSize(300, 300).AddPage()
	b.DrawPNG(20, 20, 200, 200, gradientPNG(t, 64, 64, true))
	pdf := savePDF(t, b)
	if !bytes.Contains(pdf, []byte("/Predictor 15")) {
		t.Fatal("image stream has no PNG predictor")
	}
	img := firstImage(t, pdf)
	if img.Width != 64 || img.Height != 64 || img.ColorSpace != "DeviceRGB" {
		t.Fatalf("image = %dx%d %s", img.Width, img.Height, img.ColorSpace)
	}
	if len(img.Data) != 64*64*3 {
		t.Fatalf("decoded %d bytes, want %d", len(img.Data), 64*64*3)
	}
	// Top-left pixel is (0, 0, 128), bottom-right is (255, 255, 128).
	if img.Data[0] != 0 || img.Data[1] != 0 || img.Data[2] != 128 || img.Data[len(img.Data)-3] != 255 || img.Data[len(img.Data)-1] != 128 {
		t.Errorf("samples changed: first %v last %v", img.Data[:3], img.Data[len(img.Data)-3:])
	}
	if !img.HasSMask || !img.SMaskDecoded || len(img.SMaskData) != 64*64 || img.SMaskData[0] != 0 || img.SMaskData[63] != 255 {
		t.Errorf("soft mask: has=%v decoded=%v len=%d", img.HasSMask, img.SMaskDecoded, len(img.SMaskData))
	}
	if n := pdfImages(t, pdf); n >= 0 && n != 2 {
		t.Errorf("pdfimages lists %d images, want image and mask", n)
	}
}

// MaxDPI downsamples raw and JPEG images (and their soft masks) to the placed
// size; JPEGQuality re-encodes raw RGB images as JPEG.
func TestCompressImages(t *testing.T) {
	// 400 px drawn 1 inch wide: 400 dpi, limited to 100 dpi.
	b := doc.New().PageSize(300, 300).CompressImages(doc.ImageCompression{MaxDPI: 100}).AddPage()
	b.DrawPNG(20, 20, 72, 72, gradientPNG(t, 400, 400, true))
	img := firstImage(t, savePDF(t, b))
	if img.Width != 100 || img.Height != 100 || img.SMaskWidth != 100 || img.SMaskHeight != 100 {
		t.Errorf("downsampled raw image = %dx%d, mask %dx%d; want 100x100", img.Width, img.Height, img.SMaskWidth, img.SMaskHeight)
	}
	if img.Filter != "FlateDecode" || img.SMaskData[0] > 4 || img.SMaskData[99] < 250 {
		t.Errorf("mask not resampled: filter %s, mask row %v", img.Filter, img.SMaskData[:100])
	}

	b = doc.New().PageSize(300, 300).CompressImages(doc.ImageCompression{MaxDPI: 100, JPEGQuality: 80}).AddPage()
	b.DrawJPEG(20, 20, 72, 72, gradientJPEG(t, 400, 300), 400, 300, "DeviceRGB")
	pdf := savePDF(t, b)
	img = firstImage(t, pdf)
	if img.Width != 100 || img.Height != 75 || img.Filter != "DCTDecode" {
		t.Errorf("downsampled JPEG = %dx%d %s; want 100x75 DCTDecode", img.Width, img.Height, img.Filter)
	}
	if decoded, err := jpeg.Decode(bytes.NewReader(img.Data)); err != nil || decoded.Bounds().Dx() != 100 {
		t.Errorf("stored JPEG: %v", err)
	}

	// Below the limit nothing changes.
	b = doc.New().PageSize(300, 300).CompressImages(doc.ImageCompression{MaxDPI: 300}).AddPage()
	b.DrawJPEG(20, 20, 72, 72, gradientJPEG(t, 200, 200), 200, 200, "DeviceRGB")
	if img = firstImage(t, savePDF(t, b)); img.Width != 200 {
		t.Errorf("image under the limit was resampled to %d px", img.Width)
	}

	// JPEGQuality turns raw RGB into JPEG, but leaves images with transparency lossless.
	b = doc.New().PageSize(300, 300).CompressImages(doc.ImageCompression{JPEGQuality: 75}).AddPage()
	b.DrawPNG(20, 20, 72, 72, gradientPNG(t, 64, 64, false))
	b.DrawPNG(120, 20, 72, 72, gradientPNG(t, 64, 64, true))
	d := reopen(t, savePDF(t, b))
	imgs, _ := d.ReadImages()
	if len(imgs) != 2 || imgs[0].Filter != "DCTDecode" || imgs[1].Filter != "FlateDecode" {
		t.Errorf("filters = %v", []string{imgs[0].Filter, imgs[1].Filter})
	}

	if err := doc.New().CompressImages(doc.ImageCompression{JPEGQuality: 101}).Err(); err == nil {
		t.Error("JPEGQuality 101 accepted")
	}
}

// StreamTo writes images as they are drawn and pages as they are built; the
// result reads like the in-memory one.
func TestStreamTo(t *testing.T) {
	build := func(b *doc.DocumentBuilder) {
		b.Title("Streamed").PDFUA(doc.PDFUA2).SetLanguage("en").PageSize(300, 300)
		for i := range 3 {
			b.AddPage()
			b.DrawHeading(i, 1, "Streamed page", 20, 270, "Helvetica", 14)
			b.DrawParagraph(i, "Content written as it is built.", 20, 240, "Helvetica", 11)
			b.DrawTaggedPNG(i, 20, 100, 60, 60, gradientPNG(t, 32, 32, true), "Gradient")
			b.AddLinkToURL(i, 20, 20, 100, 40, "https://example.org/")
		}
	}
	var streamed bytes.Buffer
	sb := doc.New().StreamTo(&streamed)
	build(sb)
	if err := sb.Finish(); err != nil {
		t.Fatal(err)
	}
	mb := doc.New()
	build(mb)
	inMemory := savePDF(t, mb)

	pdf := streamed.Bytes()
	if !bytes.HasPrefix(pdf, []byte("%PDF-2.0")) {
		t.Errorf("header %q", pdf[:8])
	}
	// The image streams were written when drawn, long before the object streams
	// that hold the page dictionaries.
	if bytes.Index(pdf, []byte("/Subtype /Image")) > bytes.Index(pdf, []byte("/Type /ObjStm")) {
		t.Error("images were not written before the pages")
	}
	for name, p := range map[string][]byte{"streamed": pdf, "in-memory": inMemory} {
		d := reopen(t, p)
		if txt, _ := d.ReadContent(); strings.Count(txt, "Streamed page") != 3 {
			t.Errorf("%s: text = %q", name, txt)
		}
		imgs, _ := d.ReadImages()
		if len(imgs) != 3 || !imgs[0].HasSMask {
			t.Errorf("%s: %d images", name, len(imgs))
		}
		if txt := pdfText(t, p); txt != "" && strings.Count(txt, "Streamed page") != 3 {
			t.Errorf("%s: pdftotext = %q", name, txt)
		}
		root := structRoot(t, d)
		if kids := roles(t, d, root[model.Name("K")].(model.Array)); strings.Join(kids, " ") != "Document" {
			t.Errorf("%s: structure root kids = %v", name, kids)
		}
	}
	if len(pdf) > len(inMemory)+len(inMemory)/10 {
		t.Errorf("streamed file is much larger: %d vs %d", len(pdf), len(inMemory))
	}

	// BuildTo is the one-shot form; Build is refused after StreamTo.
	var out bytes.Buffer
	b := doc.New().PageSize(200, 200).AddPage()
	b.DrawText("one shot", 20, 100, "Helvetica", 12)
	if err := b.BuildTo(&out); err != nil {
		t.Fatal(err)
	}
	if txt, _ := reopen(t, out.Bytes()).ReadContent(); !strings.Contains(txt, "one shot") {
		t.Errorf("BuildTo text = %q", txt)
	}
	if _, err := doc.New().StreamTo(&out).Build(); !errors.Is(err, doc.ErrStreaming) {
		t.Errorf("Build after StreamTo: %v", err)
	}
	// The version is fixed by the first image written.
	b = doc.New().StreamTo(&out).AddPage()
	b.DrawPNG(10, 10, 50, 50, tinyPNG(t))
	b.PDFA(doc.PDFA2B)
	if err := b.Finish(); !errors.Is(err, doc.ErrStreaming) {
		t.Errorf("PDF/A after streaming started: %v", err)
	}
}
