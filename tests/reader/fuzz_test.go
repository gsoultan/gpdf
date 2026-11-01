// Copyright (c) 2026 Gembit Soultan Shirazi <gembit.soultan@gmail.com>

package reader_test

import (
	"bytes"
	"testing"

	"github.com/gsoultan/gpdf/internal/pdftest"
	"github.com/gsoultan/gpdf/reader"
)

func fuzzSeeds() [][]byte {
	content := []byte("BT /F1 12 Tf 20 100 Td (Seed) Tj ET")
	table := pdftest.New("1.4").OnePage()
	table.Stream(4, 0, "", content)
	stream := pdftest.New("1.5").OnePage()
	stream.Stream(4, 0, "", content)
	hybrid := pdftest.New("1.5")
	hybrid.Obj(1, 0, "<< /Type /Catalog /Pages 2 0 R >>")
	hybrid.Obj(2, 0, "<< /Type /Pages /Kids [3 0 R] /Count 1 >>")
	hybrid.Stream(4, 0, "", content)
	hybrid.ObjStm(6, map[int]string{
		3: "<< /Type /Page /Parent 2 0 R /MediaBox [0 0 200 200] /Resources << /Font << /F1 5 0 R >> >> /Contents 4 0 R >>",
		5: "<< /Type /Font /Subtype /Type1 /BaseFont /Helvetica >>",
	})
	predicted := pdftest.New("1.5").OnePage()
	predicted.Stream(4, 0, "/Filter /FlateDecode", pdftest.Flate(content))
	return [][]byte{
		table.XRefTable("/Root 1 0 R"),
		stream.XRefStream("/Root 1 0 R"),
		hybrid.Hybrid("/Root 1 0 R", map[int][2]int{3: {6, 0}, 5: {6, 1}}),
		predicted.XRefStream("/Root 1 0 R"),
		[]byte("%PDF-1.4\n1 0 obj << /Type /Catalog /Pages 2 0 R >> endobj\ntrailer << /Root 1 0 R >>"),
	}
}

// FuzzReadDocument checks that a hostile file never panics or hangs the reader: opening,
// walking the pages and extracting text must return or fail cleanly.
func FuzzReadDocument(f *testing.F) {
	for _, s := range fuzzSeeds() {
		f.Add(s)
	}
	f.Fuzz(func(t *testing.T, data []byte) {
		d, err := reader.NewPDFReader().ReadDocument(bytes.NewReader(data), int64(len(data)))
		if err != nil {
			return
		}
		if _, err := d.Pages(); err != nil {
			return
		}
		_, _ = d.Content()
		_, _ = d.Images()
	})
}

func BenchmarkReadDocument(b *testing.B) {
	pdf := fuzzSeeds()[3]
	b.SetBytes(int64(len(pdf)))
	b.ReportAllocs()
	for b.Loop() {
		d, err := reader.NewPDFReader().ReadDocument(bytes.NewReader(pdf), int64(len(pdf)))
		if err != nil {
			b.Fatal(err)
		}
		if _, err := d.Content(); err != nil {
			b.Fatal(err)
		}
	}
}
