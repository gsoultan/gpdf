package doc_test

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	"github.com/gsoultan/gpdf/doc"
	"github.com/gsoultan/gpdf/internal/pdftest"
)

// threePagePDF has pages 3, 4, 5 with content streams 6, 7, 8. Page 4 has a link
// annotation (9) whose /Dest points at page 3 and whose /P points back to page 4.
func threePagePDF() []byte {
	b := pdftest.New("1.7")
	b.Obj(1, 0, "<< /Type /Catalog /Pages 2 0 R >>")
	b.Obj(2, 0, "<< /Type /Pages /Kids [3 0 R 4 0 R 5 0 R] /Count 3 /MediaBox [0 0 200 200] /Resources << /Font << /F1 10 0 R >> >> >>")
	for i := range 3 {
		extra := ""
		if i == 1 {
			extra = " /Annots [9 0 R]"
		}
		b.Obj(3+i, 0, fmt.Sprintf("<< /Type /Page /Parent 2 0 R /Contents %d 0 R /StructParents %d%s >>", 6+i, i, extra))
		b.Stream(6+i, 0, "", []byte(fmt.Sprintf("BT /F1 12 Tf 20 100 Td (Page number %d) Tj ET", i+1)))
	}
	b.Obj(9, 0, "<< /Type /Annot /Subtype /Link /Rect [0 0 50 50] /P 4 0 R /Dest [3 0 R /Fit] >>")
	b.Obj(10, 0, "<< /Type /Font /Subtype /Type1 /BaseFont /Helvetica >>")
	return b.XRefTable("/Root 1 0 R")
}

// Regression: Extract/Split/Merge followed the page's /Parent and copied every page of
// the source document into the output.
func TestExtractCopiesOnlySelectedPage(t *testing.T) {
	pdf := threePagePDF()
	src, err := doc.OpenReader(bytes.NewReader(pdf), int64(len(pdf)))
	if err != nil {
		t.Fatal(err)
	}
	one, err := doc.Extract(src, 1, 2)
	if err != nil {
		t.Fatal(err)
	}
	var out bytes.Buffer
	if err := one.Save(&out); err != nil {
		t.Fatal(err)
	}
	reopened, err := doc.OpenReader(bytes.NewReader(out.Bytes()), int64(out.Len()))
	if err != nil {
		t.Fatalf("reopen: %v", err)
	}
	pages, err := reopened.Pages()
	if err != nil || len(pages) != 1 {
		t.Fatalf("pages = %d, err = %v", len(pages), err)
	}
	text, err := reopened.ReadContent()
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(text, "Page number 2") {
		t.Fatalf("text = %q", text)
	}
	for _, other := range []string{"Page number 1", "Page number 3"} {
		if strings.Contains(text, other) {
			t.Fatalf("output still contains %q", other)
		}
	}
	// The other pages' (uncompressed) content streams must not be in the file at all.
	for _, other := range []string{"Page number 1", "Page number 3"} {
		if bytes.Contains(out.Bytes(), []byte(other)) {
			t.Fatalf("saved file still contains the content stream with %q", other)
		}
	}
}

func TestMergeKeepsEachPageOnce(t *testing.T) {
	pdf := threePagePDF()
	a, err := doc.OpenReader(bytes.NewReader(pdf), int64(len(pdf)))
	if err != nil {
		t.Fatal(err)
	}
	first, err := doc.Extract(a, 0, 1)
	if err != nil {
		t.Fatal(err)
	}
	last, err := doc.Extract(a, 2, 3)
	if err != nil {
		t.Fatal(err)
	}
	merged, err := doc.Merge(first, last)
	if err != nil {
		t.Fatal(err)
	}
	pages, err := merged.Pages()
	if err != nil || len(pages) != 2 {
		t.Fatalf("pages = %d, err = %v", len(pages), err)
	}
	perPage, err := merged.ReadContentPerPage()
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(perPage[0], "Page number 1") || !strings.Contains(perPage[1], "Page number 3") {
		t.Fatalf("per page = %q", perPage)
	}
}
