package writer_test

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gsoultan/gpdf/doc"
	"github.com/gsoultan/gpdf/internal/pdftest"
	"github.com/gsoultan/gpdf/model"
	"github.com/gsoultan/gpdf/security"
	"github.com/gsoultan/gpdf/writer"
)

func open(t *testing.T, pdf []byte, password string) doc.Document {
	t.Helper()
	d, err := doc.OpenReaderWithPassword(bytes.NewReader(pdf), int64(len(pdf)), password)
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	return d
}

func save(t *testing.T, d doc.Document) []byte {
	t.Helper()
	var buf bytes.Buffer
	if err := d.Save(&buf); err != nil {
		t.Fatalf("Save: %v", err)
	}
	return buf.Bytes()
}

func text(t *testing.T, d doc.Document) string {
	t.Helper()
	s, err := d.ReadContent()
	if err != nil {
		t.Fatalf("ReadContent: %v", err)
	}
	return s
}

// popplerText runs pdftotext on pdf when poppler is installed (interoperability check:
// gPDF reading its own output is not enough).
func popplerText(t *testing.T, pdf []byte, args ...string) (string, bool) {
	t.Helper()
	bin, err := exec.LookPath("pdftotext")
	if err != nil {
		return "", false
	}
	path := filepath.Join(t.TempDir(), "out.pdf")
	if err := os.WriteFile(path, pdf, 0o644); err != nil {
		t.Fatal(err)
	}
	out, err := exec.Command(bin, append(args, path, "-")...).CombinedOutput()
	if err != nil {
		t.Fatalf("pdftotext %v: %v\n%s", args, err, out)
	}
	return string(out), true
}

// Regression: Save resolved every object as generation 0 and failed with
// "generation mismatch" for files containing objects with generation > 0.
func TestSaveKeepsNonZeroGenerations(t *testing.T) {
	b := pdftest.New("1.7").OnePage()
	b.Obj(3, 2, "<< /Type /Page /Parent 2 0 R /MediaBox [0 0 200 200] /Resources << /Font << /F1 5 0 R >> >> /Contents 4 1 R >>")
	b.Obj(2, 0, "<< /Type /Pages /Kids [3 2 R] /Count 1 >>")
	b.Stream(4, 1, "", []byte("BT /F1 12 Tf 20 100 Td (Generation two) Tj ET"))
	out := save(t, open(t, b.XRefTable("/Root 1 0 R"), ""))
	if got := text(t, open(t, out, "")); !strings.Contains(got, "Generation two") {
		t.Fatalf("text = %q", got)
	}
	if s, ok := popplerText(t, out); ok && !strings.Contains(s, "Generation two") {
		t.Fatalf("poppler text = %q", s)
	}
}

// Regression: saving a PDF 1.5 file copied the xref-stream dictionary (/W, /Index,
// /Filter, /DecodeParms...) and /Prev into the new trailer. The output has its own
// cross-reference stream; the source one is not copied, and the predictor the
// content stream asks for is applied again on the way out.
func TestSaveFromXRefStreamWritesCleanTrailer(t *testing.T) {
	content := []byte("BT /F1 12 Tf 20 100 Td (Clean trailer) Tj ET")
	b := pdftest.New("1.7").OnePage()
	b.Stream(4, 0, "/Filter /FlateDecode /DecodeParms << /Predictor 12 /Columns 6 >>", pdftest.Flate(pdftest.PNGUp(content, 6)))
	out := save(t, open(t, b.XRefStream("/Root 1 0 R /Prev 999999"), ""))
	if bytes.Contains(out, []byte("/Prev")) {
		t.Fatalf("output carries the source /Prev:\n%s", out)
	}
	if n := bytes.Count(out, []byte("/XRef")); n != 1 {
		t.Fatalf("found %d cross-reference streams, want the output's own", n)
	}
	if bytes.Contains(out, []byte("trailer")) {
		t.Fatal("a cross-reference stream file has no trailer keyword")
	}
	if got := text(t, open(t, out, "")); !strings.Contains(got, "Clean trailer") {
		t.Fatalf("text = %q", got)
	}
	if s, ok := popplerText(t, out); ok && !strings.Contains(s, "Clean trailer") {
		t.Fatalf("poppler text = %q", s)
	}
	// Without object streams the classic layout comes back, still without /Prev.
	var classic bytes.Buffer
	pw := writer.NewPDFWriterWithOptions(writer.WriterOptions{ObjectStreams: new(false)})
	if err := pw.Write(&classic, open(t, out, "").(writer.Document)); err != nil {
		t.Fatal(err)
	}
	tail := classic.Bytes()[bytes.LastIndex(classic.Bytes(), []byte("trailer")):]
	for _, bad := range []string{"/W", "/Index", "/Prev", "/DecodeParms", "/Type"} {
		if bytes.Contains(tail, []byte(bad)) {
			t.Fatalf("trailer contains %s:\n%s", bad, tail)
		}
	}
	if bytes.Contains(classic.Bytes(), []byte("/XRef")) || bytes.Contains(classic.Bytes(), []byte("/ObjStm")) {
		t.Fatal("classic output must not use cross-reference or object streams")
	}
}

func builtDoc(t *testing.T) doc.Document {
	t.Helper()
	d, err := doc.New().Title("Secret title").PageSize(300, 300).AddPage().
		Flow(doc.FlowOptions{Margin: 36}).Paragraph("Encrypted body text").End().Build()
	if err != nil {
		t.Fatal(err)
	}
	return d
}

// Regression: RC4 and AES-256 output was rejected by poppler, wrong passwords were
// accepted, and re-saving an opened encrypted file failed or kept a broken /Encrypt.
func TestEncryptedOutputInteroperatesAndResaves(t *testing.T) {
	for name, encrypt := range map[string]func(doc.Document, *bytes.Buffer) error{
		"rc4":    func(d doc.Document, w *bytes.Buffer) error { return d.SaveWithPassword(w, "user", "owner") },
		"aes256": func(d doc.Document, w *bytes.Buffer) error { return d.SaveWithAES256Password(w, "user", "owner") },
	} {
		t.Run(name, func(t *testing.T) {
			var enc bytes.Buffer
			if err := encrypt(builtDoc(t), &enc); err != nil {
				t.Fatal(err)
			}
			if _, err := doc.OpenReaderWithPassword(bytes.NewReader(enc.Bytes()), int64(enc.Len()), "WRONG"); !errors.Is(err, security.ErrWrongPassword) {
				t.Fatalf("wrong password: err = %v", err)
			}
			for _, pw := range []string{"user", "owner"} {
				d := open(t, enc.Bytes(), pw)
				if got := text(t, d); !strings.Contains(got, "Encrypted body text") {
					t.Fatalf("%s: text = %q", pw, got)
				}
				info, _ := d.Info()
				if title, _ := info["Title"].(model.String); title != "Secret title" {
					t.Fatalf("%s: Title = %q", pw, title)
				}
				resaved := save(t, d)
				if bytes.Contains(resaved, []byte("/Encrypt")) {
					t.Fatal("re-saved document still references /Encrypt")
				}
				if got := text(t, open(t, resaved, "")); !strings.Contains(got, "Encrypted body text") {
					t.Fatalf("re-saved text = %q", got)
				}
			}
			if s, ok := popplerText(t, enc.Bytes(), "-upw", "user"); ok && !strings.Contains(s, "Encrypted body text") {
				t.Fatalf("poppler (user) text = %q", s)
			}
			if s, ok := popplerText(t, enc.Bytes(), "-opw", "owner"); ok && !strings.Contains(s, "Encrypted body text") {
				t.Fatalf("poppler (owner) text = %q", s)
			}
		})
	}
}

// A reference to an object that does not exist is null (ISO 32000-1 §7.3.10); saving a
// file that contains one must work and write null.
func TestSaveFileWithDanglingReference(t *testing.T) {
	b := pdftest.New("1.7").OnePage()
	b.Obj(1, 0, "<< /Type /Catalog /Pages 2 0 R /Outlines 77 0 R >>")
	b.Stream(4, 0, "", []byte("BT /F1 12 Tf 20 100 Td (Dangling) Tj ET"))
	out := save(t, open(t, b.XRefTable("/Root 1 0 R"), ""))
	if bytes.Contains(out, []byte("77 0 R")) {
		t.Fatal("dangling reference was written as a reference")
	}
	if got := text(t, open(t, out, "")); !strings.Contains(got, "Dangling") {
		t.Fatalf("text = %q", got)
	}
}

type seekBuffer struct {
	data []byte
	pos  int64
}

func (s *seekBuffer) Write(p []byte) (int, error) {
	end := s.pos + int64(len(p))
	if end > int64(len(s.data)) {
		s.data = append(s.data, make([]byte, end-int64(len(s.data)))...)
	}
	copy(s.data[s.pos:], p)
	s.pos = end
	return len(p), nil
}

func (s *seekBuffer) Seek(offset int64, whence int) (int64, error) {
	switch whence {
	case 1:
		offset += s.pos
	case 2:
		offset += int64(len(s.data))
	}
	s.pos = offset
	return offset, nil
}

// Regression: SaveLinearized patched "/L 0000000000" with a shorter number (leaving
// garbage in the dictionary), used object number 0 and wrote a wrong /Prev offset.
func TestSaveLinearizedIsWellFormed(t *testing.T) {
	d, err := doc.New().PageSize(300, 300).AddPage().Flow(doc.FlowOptions{Margin: 36}).Paragraph("Page one").End().
		AddPage().Flow(doc.FlowOptions{Margin: 36}).Paragraph("Page two").End().Build()
	if err != nil {
		t.Fatal(err)
	}
	var ws seekBuffer
	if err := d.SaveLinearized(&ws); err != nil {
		t.Fatal(err)
	}
	head := string(ws.data[:min(300, len(ws.data))])
	if strings.Contains(head, "\n0 0 obj") {
		t.Fatal("linearization dictionary uses object number 0")
	}
	if !strings.Contains(head, "/L "+fmt.Sprintf("%010d", len(ws.data))) {
		t.Fatalf("/L does not hold the file length:\n%s", head)
	}
	got := text(t, open(t, ws.data, ""))
	if !strings.Contains(got, "Page one") || !strings.Contains(got, "Page two") {
		t.Fatalf("text = %q", got)
	}
	if s, ok := popplerText(t, ws.data); ok && !strings.Contains(s, "Page two") {
		t.Fatalf("poppler text = %q", s)
	}
}

// Regression: a stream with a filter gPDF has no codec for made Resolve, and so Save,
// fail. The reader now keeps such a stream encoded and the writer copies it unchanged.
func TestUnknownFilterIsPassedThrough(t *testing.T) {
	b := pdftest.New("1.4").OnePage()
	b.Stream(10, 0, "/Filter /MadeUpDecode /DecodeParms << /K 1 >>", []byte("opaque bytes"))
	pdf := b.XRefTable("/Root 1 0 R")
	d := open(t, pdf, "")
	obj, err := d.Resolve(model.Ref{ObjectNumber: 10})
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	s := obj.(*model.Stream)
	if string(s.Content) != "opaque bytes" || s.Dict[model.Name("Filter")] != model.Name("MadeUpDecode") {
		t.Fatalf("stream = %v %q", s.Dict, s.Content)
	}
	out := save(t, d)
	if !bytes.Contains(out, []byte("/Filter /MadeUpDecode")) || !bytes.Contains(out, []byte("/DecodeParms <<\n/K 1\n>>")) ||
		!bytes.Contains(out, []byte("stream\nopaque bytes\nendstream")) {
		t.Fatalf("saved file does not carry the stream through:\n%s", out)
	}
}
