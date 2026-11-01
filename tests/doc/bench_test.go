package doc_test

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/gsoultan/gpdf/doc"
	"github.com/gsoultan/gpdf/model"
)

func benchDocument() doc.Document {
	b := doc.New().SetFooter(doc.PageText{Text: "Page {page} of {pages}", Align: doc.TextAlignCenter})
	for p := range 20 {
		b.AddPage()
		b.FillRect(p, 72, 700, 450, 20, doc.ColorBlue)
		for line := range 40 {
			b.DrawText(fmt.Sprintf("Line %d of page %d, some running text to measure.", line, p), 72, 680-float64(line)*14, "Helvetica", 11)
		}
	}
	d, err := b.Build()
	if err != nil {
		panic(err)
	}
	return d
}

func BenchmarkBuildAndSave(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		var buf bytes.Buffer
		if err := benchDocument().Save(&buf); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkOpenAndExtract(b *testing.B) {
	var buf bytes.Buffer
	if err := benchDocument().Save(&buf); err != nil {
		b.Fatal(err)
	}
	pdf := buf.Bytes()
	b.SetBytes(int64(len(pdf)))
	b.ReportAllocs()
	for b.Loop() {
		d, err := doc.OpenReaderWithPassword(bytes.NewReader(pdf), int64(len(pdf)), "")
		if err != nil {
			b.Fatal(err)
		}
		if _, err := d.ReadContent(); err != nil {
			b.Fatal(err)
		}
	}
}

func benchBuilder(pages int, images bool) *doc.DocumentBuilder {
	b := doc.New().SetFooter(doc.PageText{Text: "Page {page} of {pages}", Align: doc.TextAlignCenter})
	for p := range pages {
		b.AddPage()
		b.FillRect(p, 72, 700, 450, 20, doc.ColorBlue)
		for line := range 40 {
			b.DrawText(fmt.Sprintf("Line %d of page %d, some running text to measure.", line, p), 72, 680-float64(line)*14, "Helvetica", 11)
		}
		if images {
			b.DrawPNG(300, 100, 120, 120, benchPNG)
		}
	}
	return b
}

var benchPNG = func() []byte {
	img := image.NewNRGBA(image.Rect(0, 0, 256, 256))
	for y := range 256 {
		for x := range 256 {
			img.Set(x, y, color.NRGBA{R: uint8(x), G: uint8(y), B: uint8(x ^ y), A: 255})
		}
	}
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		panic(err)
	}
	return buf.Bytes()
}()

func benchPDF(b *testing.B, builder *doc.DocumentBuilder) []byte {
	b.Helper()
	d, err := builder.Build()
	if err != nil {
		b.Fatal(err)
	}
	var buf bytes.Buffer
	if err := d.Save(&buf); err != nil {
		b.Fatal(err)
	}
	return buf.Bytes()
}

// BenchmarkBuildAndSaveClassic writes the same document with a classic
// cross-reference table and no object streams, for comparison with the
// compressed default measured by BenchmarkBuildAndSave.
func BenchmarkBuildAndSaveClassic(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		var buf bytes.Buffer
		d, err := benchBuilder(20, false).NoObjectStreams().Build()
		if err != nil {
			b.Fatal(err)
		}
		if err := d.Save(&buf); err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkStreamTo writes a 200-page document through the streaming writer,
// which releases each page as it goes.
func BenchmarkStreamTo(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		if err := benchBuilder(200, false).BuildTo(io.Discard); err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkBuild200Pages is the in-memory counterpart of BenchmarkStreamTo.
func BenchmarkBuild200Pages(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		d, err := benchBuilder(200, false).Build()
		if err != nil {
			b.Fatal(err)
		}
		if err := d.Save(io.Discard); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkCompressImages(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		d, err := benchBuilder(5, true).CompressImages(doc.ImageCompression{MaxDPI: 72, JPEGQuality: 75}).Build()
		if err != nil {
			b.Fatal(err)
		}
		if err := d.Save(io.Discard); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkLinearize(b *testing.B) {
	d, err := benchBuilder(20, false).Build()
	if err != nil {
		b.Fatal(err)
	}
	path := filepath.Join(b.TempDir(), "lin.pdf")
	b.ReportAllocs()
	for b.Loop() {
		f, err := os.Create(path)
		if err != nil {
			b.Fatal(err)
		}
		if err := d.SaveLinearized(f); err != nil {
			b.Fatal(err)
		}
		f.Close()
	}
}

func openBench(b *testing.B, pdf []byte) doc.Document {
	b.Helper()
	d, err := doc.OpenReader(bytes.NewReader(pdf), int64(len(pdf)))
	if err != nil {
		b.Fatal(err)
	}
	return d
}

func BenchmarkReadingOrder(b *testing.B) {
	pdf := benchPDF(b, benchBuilder(20, false))
	b.SetBytes(int64(len(pdf)))
	b.ReportAllocs()
	for b.Loop() {
		if _, err := openBench(b, pdf).ReadContentInReadingOrder(); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkLayout(b *testing.B) {
	pdf := benchPDF(b, benchBuilder(20, false))
	b.SetBytes(int64(len(pdf)))
	b.ReportAllocs()
	for b.Loop() {
		if _, err := openBench(b, pdf).ReadLayout(); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkRedact(b *testing.B) {
	pdf := benchPDF(b, benchBuilder(20, false))
	b.SetBytes(int64(len(pdf)))
	b.ReportAllocs()
	for b.Loop() {
		d := openBench(b, pdf)
		if _, err := doc.Redact(d, doc.RedactOptions{Texts: []string{"running text"}}); err != nil {
			b.Fatal(err)
		}
		if err := d.Save(io.Discard); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkSaveIncremental(b *testing.B) {
	pdf := benchPDF(b, benchBuilder(20, false))
	b.SetBytes(int64(len(pdf)))
	b.ReportAllocs()
	for b.Loop() {
		d := openBench(b, pdf)
		if err := d.SetInfo("Title", model.String("Updated")); err != nil {
			b.Fatal(err)
		}
		if err := d.SaveIncremental(io.Discard); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkMerge(b *testing.B) {
	pdf := benchPDF(b, benchBuilder(20, false))
	b.ReportAllocs()
	for b.Loop() {
		merged, err := doc.Merge(openBench(b, pdf), openBench(b, pdf), openBench(b, pdf))
		if err != nil {
			b.Fatal(err)
		}
		if err := merged.Save(io.Discard); err != nil {
			b.Fatal(err)
		}
	}
}
