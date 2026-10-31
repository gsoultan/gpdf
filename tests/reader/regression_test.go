package reader_test

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	"github.com/gsoultan/gpdf/internal/pdftest"
	"github.com/gsoultan/gpdf/reader"
)

func readBytes(t *testing.T, pdf []byte) reader.Document {
	t.Helper()
	d, err := reader.NewPDFReader().ReadDocument(bytes.NewReader(pdf), int64(len(pdf)))
	if err != nil {
		t.Fatalf("ReadDocument: %v", err)
	}
	return d
}

func contentOf(t *testing.T, d reader.Document) string {
	t.Helper()
	text, err := d.Content()
	if err != nil {
		t.Fatalf("Content: %v", err)
	}
	return text
}

// Regression: /DecodeParms /Predictor was ignored, so PDF 1.5+ files with a predicted
// xref stream (almost all of them) failed with "object 1 0 R not in xref".
func TestReadXRefStreamWithPNGPredictor(t *testing.T) {
	content := []byte("BT /F1 12 Tf 20 100 Td (Predicted hello) Tj ET")
	b := pdftest.New("1.7").OnePage()
	b.Stream(4, 0, "/Filter /FlateDecode /DecodeParms << /Predictor 12 /Columns 8 >>", pdftest.Flate(pdftest.PNGUp(content, 8)))
	d := readBytes(t, b.XRefStream("/Root 1 0 R"))
	if got := contentOf(t, d); !strings.Contains(got, "Predicted hello") {
		t.Fatalf("content = %q", got)
	}
}

func onePageWithContentObj(content string) *pdftest.Builder {
	b := pdftest.New("1.7").OnePage()
	b.Obj(4, 0, content)
	return b
}

const helloOps = "BT /F1 12 Tf 20 100 Td (Hello length) Tj ET"

// Regression: a huge /Length made the parser allocate it (fatal OOM); a /Length past
// EOF hung forever; an indirect /Length failed; a short /Length silently lost text.
func TestStreamLengthHardening(t *testing.T) {
	cases := map[string]string{
		"huge":     "<< /Length 99999999999 >>\nstream\n" + helloOps + "\nendstream",
		"pastEOF":  "<< /Length 1000000 >>\nstream\n" + helloOps + "\nendstream",
		"tooShort": "<< /Length 7 >>\nstream\n" + helloOps + "\nendstream",
		"negative": "<< /Length -5 >>\nstream\n" + helloOps + "\nendstream",
		"missing":  "<< >>\nstream\n" + helloOps + "\nendstream",
		"indirect": "<< /Length 6 0 R >>\nstream\n" + helloOps + "\nendstream",
	}
	for name, body := range cases {
		t.Run(name, func(t *testing.T) {
			b := onePageWithContentObj(body)
			b.Obj(6, 0, fmt.Sprint(len(helloOps)))
			d := readBytes(t, b.XRefTable("/Root 1 0 R"))
			if got := contentOf(t, d); !strings.Contains(got, "Hello length") {
				t.Fatalf("content = %q", got)
			}
		})
	}
}

// A stream with no endstream at all must fail cleanly (it used to loop forever).
func TestStreamWithoutEndstreamFails(t *testing.T) {
	b := onePageWithContentObj("<< /Length 1000000 >>\nstream\n" + helloOps)
	pdf := b.XRefTable("/Root 1 0 R")
	pdf = bytes.ReplaceAll(pdf, []byte("endobj\n6"), []byte("xxxxxx\n6"))
	d, err := reader.NewPDFReader().ReadDocument(bytes.NewReader(pdf), int64(len(pdf)))
	if err != nil {
		return
	}
	_, _ = d.Content() // must return, error or not
}
