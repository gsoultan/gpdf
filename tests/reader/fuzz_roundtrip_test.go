// Copyright (c) 2026 Gembit Soultan Shirazi <gembit.soultan@gmail.com>

package reader_test

import (
	"bytes"
	"testing"

	"github.com/gsoultan/gpdf/doc"
	"github.com/gsoultan/gpdf/internal/pdftest"
)

func roundTripSeeds(t testing.TB) [][]byte {
	seeds := fuzzSeeds()
	tagged := pdftest.New("1.7")
	tagged.Obj(1, 0, "<< /Type /Catalog /Pages 2 0 R /MarkInfo << /Marked true >> /StructTreeRoot 6 0 R >>")
	tagged.Obj(2, 0, "<< /Type /Pages /Kids [3 0 R] /Count 1 >>")
	tagged.Obj(3, 0, "<< /Type /Page /Parent 2 0 R /MediaBox [0 0 200 200] /StructParents 0 /Resources << /Font << /F1 5 0 R >> >> /Contents 4 0 R >>")
	tagged.Stream(4, 0, "", []byte("/P << /MCID 0 >> BDC BT /F1 12 Tf 20 100 Td (Secret word) Tj ET EMC"))
	tagged.Obj(5, 0, "<< /Type /Font /Subtype /Type1 /BaseFont /Helvetica >>")
	tagged.Obj(6, 0, "<< /Type /StructTreeRoot /K [7 0 R] /ParentTree << /Nums [0 [8 0 R]] >> >>")
	tagged.Obj(7, 0, "<< /Type /StructElem /S /Document /P 6 0 R /K [8 0 R] >>")
	tagged.Obj(8, 0, "<< /Type /StructElem /S /P /P 7 0 R /Pg 3 0 R /K [0] >>")
	seeds = append(seeds, tagged.XRefTable("/Root 1 0 R"))
	b := doc.New().PageSize(200, 200).AddPage()
	b.DrawText("Built seed", 20, 100, "Helvetica", 12)
	built, err := b.Build()
	if err != nil {
		t.Fatal(err)
	}
	var buf bytes.Buffer
	if err := built.Save(&buf); err != nil {
		t.Fatal(err)
	}
	return append(seeds, buf.Bytes())
}

// FuzzRoundTrip opens a hostile file through the public API, exercises the
// extractors over it, saves it again (classic and incremental form), redacts
// it, and checks that everything the writer produced opens again.
func FuzzRoundTrip(f *testing.F) {
	for _, s := range roundTripSeeds(f) {
		f.Add(s)
	}
	f.Fuzz(func(t *testing.T, data []byte) {
		d, err := doc.OpenReader(bytes.NewReader(data), int64(len(data)))
		if err != nil {
			return
		}
		if pages, err := d.Pages(); err != nil || len(pages) == 0 {
			return
		}
		_, _ = d.ReadContentInReadingOrder()
		_, _ = d.ReadStructuredText()
		_, _ = d.ReadLayout()
		_, _ = d.ReadTables()
		var saved bytes.Buffer
		if err := d.Save(&saved); err != nil {
			return
		}
		reopened, err := doc.OpenReader(bytes.NewReader(saved.Bytes()), int64(saved.Len()))
		if err != nil {
			t.Fatalf("saved file does not open: %v", err)
		}
		if _, err := reopened.Pages(); err != nil {
			t.Fatalf("saved file has no page tree: %v", err)
		}
		var incremental bytes.Buffer
		if err := d.SaveIncremental(&incremental); err == nil {
			if _, err := doc.OpenReader(bytes.NewReader(incremental.Bytes()), int64(incremental.Len())); err != nil {
				t.Fatalf("incremental update does not open: %v", err)
			}
		}
		if _, err := doc.Redact(d, doc.RedactOptions{Texts: []string{"Secret"}}); err != nil {
			return
		}
		var redacted bytes.Buffer
		if err := d.Save(&redacted); err != nil {
			return
		}
		if _, err := doc.OpenReader(bytes.NewReader(redacted.Bytes()), int64(redacted.Len())); err != nil {
			t.Fatalf("redacted file does not open: %v", err)
		}
	})
}
