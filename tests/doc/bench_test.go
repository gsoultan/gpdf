package doc_test

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/gsoultan/gpdf/doc"
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
