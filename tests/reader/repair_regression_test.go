// Copyright (c) 2026 Gembit Soultan Shirazi <gembit.soultan@gmail.com>

package reader_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/gsoultan/gpdf/internal/pdftest"
)

// Regression: a hybrid file's /XRefStm was ignored, so objects stored only in object
// streams (here the page content's font) could not be resolved.
func TestHybridFileReadsXRefStm(t *testing.T) {
	b := pdftest.New("1.5")
	b.Obj(1, 0, "<< /Type /Catalog /Pages 2 0 R >>")
	b.Obj(2, 0, "<< /Type /Pages /Kids [3 0 R] /Count 1 >>")
	b.Stream(4, 0, "", []byte("BT /F1 12 Tf 20 100 Td (Hybrid text) Tj ET"))
	b.ObjStm(6, map[int]string{
		3: "<< /Type /Page /Parent 2 0 R /MediaBox [0 0 200 200] /Resources << /Font << /F1 5 0 R >> >> /Contents 4 0 R >>",
		5: "<< /Type /Font /Subtype /Type1 /BaseFont /Helvetica >>",
	})
	pdf := b.Hybrid("/Root 1 0 R", map[int][2]int{3: {6, 0}, 5: {6, 1}})
	if got := contentOf(t, readBytes(t, pdf)); !strings.Contains(got, "Hybrid text") {
		t.Fatalf("content = %q", got)
	}
}

// Regression: the repair scan kept the FIRST definition of each object (incremental
// updates need the last) and returned a trailer without /Root, so repaired files still
// failed to open.
func TestRepairUsesLastDefinitionAndFindsTrailer(t *testing.T) {
	b := onePageWithContentObj("<< /Length 41 >>\nstream\nBT /F1 12 Tf 20 100 Td (Old text) Tj ET\nendstream")
	pdf := b.XRefTable("/Root 1 0 R")
	// Incremental update redefines object 4; then break startxref.
	update := "\n4 0 obj\n<< >>\nstream\nBT /F1 12 Tf 20 100 Td (New text) Tj ET\nendstream\nendobj\n"
	pdf = append(pdf, update...)
	pdf = bytes.Replace(pdf, []byte("xref\n0 "), []byte("xxxx\n0 "), 1)
	got := contentOf(t, readBytes(t, pdf))
	if !strings.Contains(got, "New text") || strings.Contains(got, "Old text") {
		t.Fatalf("content = %q", got)
	}
}

// With no trailer at all (a PDF 1.5 file whose xref stream offset is wrong), repair takes
// /Root from the cross-reference stream dictionary and registers object-stream objects.
func TestRepairWithoutTrailerUsesXRefStreamAndObjectStreams(t *testing.T) {
	b := pdftest.New("1.5")
	b.Obj(1, 0, "<< /Type /Catalog /Pages 2 0 R >>")
	b.Obj(2, 0, "<< /Type /Pages /Kids [3 0 R] /Count 1 >>")
	b.Stream(4, 0, "", []byte("BT /F1 12 Tf 20 100 Td (Object stream text) Tj ET"))
	b.ObjStm(6, map[int]string{
		3: "<< /Type /Page /Parent 2 0 R /MediaBox [0 0 200 200] /Resources << /Font << /F1 5 0 R >> >> /Contents 4 0 R >>",
		5: "<< /Type /Font /Subtype /Type1 /BaseFont /Helvetica >>",
	})
	pdf := b.XRefStream("/Root 1 0 R")
	i := bytes.LastIndex(pdf, []byte("startxref\n"))
	pdf = append(pdf[:i], []byte("startxref\n12\n%%EOF\n")...)
	if got := contentOf(t, readBytes(t, pdf)); !strings.Contains(got, "Object stream text") {
		t.Fatalf("content = %q", got)
	}
}

// Regression: xref entries were read as fixed 20-byte records, so tables with 19-byte
// (LF only) entries, which many producers write, pointed every object at the wrong offset.
func TestXRefTableWithNineteenByteEntries(t *testing.T) {
	b := onePageWithContentObj("<< /Length 44 >>\nstream\nBT /F1 12 Tf 20 100 Td (Short lines) Tj ET\nendstream")
	pdf := b.XRefTable("/Root 1 0 R")
	pdf = bytes.ReplaceAll(pdf, []byte(" n \n"), []byte(" n\n"))
	pdf = bytes.ReplaceAll(pdf, []byte(" f \n"), []byte(" f\n"))
	if got := contentOf(t, readBytes(t, pdf)); !strings.Contains(got, "Short lines") {
		t.Fatalf("content = %q", got)
	}
}
