package writer_test

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	"github.com/gsoultan/gpdf/doc"
	"github.com/gsoultan/gpdf/internal/pdftest"
	"github.com/gsoultan/gpdf/model"
	"github.com/gsoultan/gpdf/writer"
)

// updatable is what WriteIncremental needs from an opened document.
type updatable interface {
	writer.Document
	StartXRefOffset() int64
}

// update is the object set of an incremental update.
type update struct {
	trailer model.Dict
	objects map[int]model.Object
}

func (u *update) Trailer() model.Trailer { return model.Trailer{Dict: u.trailer} }
func (u *update) Resolve(ref model.Ref) (model.Object, error) {
	o, ok := u.objects[ref.ObjectNumber]
	if !ok {
		return nil, fmt.Errorf("no object %d", ref.ObjectNumber)
	}
	return o, nil
}
func (u *update) ObjectNumbers() []int {
	nums := make([]int, 0, len(u.objects))
	for n := range u.objects {
		nums = append(nums, n)
	}
	return nums
}

// An incremental update takes the form of the base file's cross-reference: a
// stream after a stream, a table after a table.
func TestIncrementalUpdateFollowsBaseXRefForm(t *testing.T) {
	for _, tc := range []struct {
		name    string
		classic bool
	}{{"xref stream", false}, {"xref table", true}} {
		t.Run(tc.name, func(t *testing.T) {
			b := doc.New().Title("Base").PageSize(200, 200).AddPage()
			if tc.classic {
				b.NoObjectStreams()
			}
			b.DrawText("Base text", 20, 100, "Helvetica", 12)
			base := save(t, builtOf(t, b))
			d := open(t, base, "").(updatable)
			size := int(d.Trailer().Dict[model.Name("Size")].(model.Integer))
			// Replace the Info dictionary with a new title.
			infoRef := *d.Trailer().Info()
			u := &update{
				trailer: map[model.Name]model.Object{
					model.Name("Root"): *d.Trailer().Root(),
					model.Name("Info"): infoRef,
					model.Name("Size"): model.Integer(int64(size)),
				},
				objects: map[int]model.Object{infoRef.ObjectNumber: model.Dict{model.Name("Title"): model.String("Updated")}},
			}
			for k, v := range d.Trailer().Dict {
				if _, set := u.trailer[k]; !set {
					u.trailer[k] = v
				}
			}
			var out bytes.Buffer
			out.Write(base)
			pw := writer.NewPDFWriterWithOptions(writer.WriterOptions{ValidateGraph: new(false)})
			if err := pw.WriteIncremental(&out, int64(len(base)), d.StartXRefOffset(), u); err != nil {
				t.Fatal(err)
			}
			tail := out.Bytes()[len(base):]
			if tc.classic {
				if !bytes.Contains(tail, []byte("trailer")) || bytes.Contains(tail, []byte("/XRef")) {
					t.Fatalf("update after a table should use a table:\n%s", tail)
				}
			} else {
				if bytes.Contains(tail, []byte("trailer")) || !bytes.Contains(tail, []byte("/Type /XRef")) || !bytes.Contains(tail, []byte("/Index")) {
					t.Fatalf("update after a stream should use a stream:\n%s", tail)
				}
			}
			if !bytes.Contains(tail, []byte("/Prev")) {
				t.Fatal("update has no /Prev")
			}
			updated := open(t, out.Bytes(), "")
			info, _ := updated.Info()
			if title, _ := info[model.Name("Title")].(model.String); title.Text() != "Updated" {
				t.Fatalf("title after update = %q", title)
			}
			if txt := text(t, updated); !strings.Contains(txt, "Base text") {
				t.Fatalf("text = %q", txt)
			}
			if s, ok := popplerText(t, out.Bytes()); ok && !strings.Contains(s, "Base text") {
				t.Fatalf("poppler text = %q", s)
			}
		})
	}
}

func builtOf(t *testing.T, b *doc.DocumentBuilder) doc.Document {
	t.Helper()
	d, err := b.Build()
	if err != nil {
		t.Fatal(err)
	}
	return d
}

// The stream writer produces a complete file object by object; small objects are
// packed into object streams at Close, streams go out at once.
func TestStreamWriter(t *testing.T) {
	var out bytes.Buffer
	sw := writer.NewStreamWriter(&out, writer.WriterOptions{})
	catalog, pages := sw.Alloc(), sw.Alloc()
	var kids model.Array
	const nPages = 250
	for i := range nPages {
		content := fmt.Sprintf("BT /F1 12 Tf 20 100 Td (Page %d) Tj ET", i+1)
		cs := sw.Alloc()
		if err := sw.WriteObject(cs, &model.Stream{
			Dict:    model.Dict{model.Name("Filter"): model.Name("FlateDecode")},
			Content: []byte(content),
		}); err != nil {
			t.Fatal(err)
		}
		if i == 0 && !bytes.Contains(out.Bytes(), []byte("stream")) {
			t.Fatal("the first stream was not written at once")
		}
		page := sw.Alloc()
		kids = append(kids, page)
		if err := sw.WriteObject(page, model.Dict{
			model.Name("Type"):     model.Name("Page"),
			model.Name("Parent"):   pages,
			model.Name("MediaBox"): model.Array{model.Integer(0), model.Integer(0), model.Integer(200), model.Integer(200)},
			model.Name("Contents"): cs,
			model.Name("Resources"): model.Dict{model.Name("Font"): model.Dict{model.Name("F1"): model.Dict{
				model.Name("Type"): model.Name("Font"), model.Name("Subtype"): model.Name("Type1"), model.Name("BaseFont"): model.Name("Helvetica"),
			}}},
		}); err != nil {
			t.Fatal(err)
		}
	}
	if err := sw.WriteObject(pages, model.Dict{model.Name("Type"): model.Name("Pages"), model.Name("Kids"): kids, model.Name("Count"): model.Integer(nPages)}); err != nil {
		t.Fatal(err)
	}
	if err := sw.WriteObject(catalog, model.Dict{model.Name("Type"): model.Name("Catalog"), model.Name("Pages"): pages}); err != nil {
		t.Fatal(err)
	}
	if err := sw.WriteObject(catalog, model.Dict{}); err == nil {
		t.Fatal("writing an object twice was accepted")
	}
	if err := sw.SetVersion("1.7"); err == nil {
		t.Fatal("SetVersion after the first object was accepted")
	}
	if err := sw.Close(model.Dict{}); err == nil {
		t.Fatal("Close without /Root was accepted")
	}
	if err := sw.Close(model.Dict{model.Name("Root"): catalog}); err != nil {
		t.Fatal(err)
	}
	if err := sw.WriteObject(sw.Alloc(), model.Dict{}); err == nil {
		t.Fatal("write after Close was accepted")
	}
	pdf := out.Bytes()
	if n := bytes.Count(pdf, []byte("/Type /ObjStm")); n != 3 {
		t.Errorf("%d object streams, want 3 for 252 packed objects", n)
	}
	if !bytes.HasSuffix(pdf, []byte("%%EOF\n")) || !bytes.Contains(pdf, []byte("/Type /XRef")) {
		t.Error("file is not finished with a cross-reference stream")
	}
	d := open(t, pdf, "")
	pages2, err := d.Pages()
	if err != nil || len(pages2) != nPages {
		t.Fatalf("pages = %d, %v", len(pages2), err)
	}
	per, err := d.ReadContentPerPage()
	if err != nil || len(per) != nPages || !strings.Contains(per[249], "Page 250") {
		t.Fatalf("text per page: %v", err)
	}
	if s, ok := popplerText(t, pdf, "-f", "250", "-l", "250"); ok && !strings.Contains(s, "Page 250") {
		t.Fatalf("poppler text = %q", s)
	}

	// A 1.4 file gets a classic table and no object streams.
	out.Reset()
	sw = writer.NewStreamWriter(&out, writer.WriterOptions{})
	if err := sw.SetVersion("1.4"); err != nil {
		t.Fatal(err)
	}
	c := sw.Alloc()
	p := sw.Alloc()
	sw.WriteObject(c, model.Dict{model.Name("Type"): model.Name("Catalog"), model.Name("Pages"): p})
	sw.WriteObject(p, model.Dict{model.Name("Type"): model.Name("Pages"), model.Name("Kids"): model.Array{}, model.Name("Count"): model.Integer(0)})
	if err := sw.Close(model.Dict{model.Name("Root"): c}); err != nil {
		t.Fatal(err)
	}
	if !bytes.HasPrefix(out.Bytes(), []byte("%PDF-1.4")) || !bytes.Contains(out.Bytes(), []byte("trailer")) || bytes.Contains(out.Bytes(), []byte("ObjStm")) {
		t.Errorf("1.4 output:\n%s", out.Bytes())
	}
	if _, err := open(t, out.Bytes(), "").Pages(); err != nil {
		t.Fatal(err)
	}
}

// A stream that asks for a predictor gets it on the way out, and reads back.
func TestWriterAppliesRequestedPredictor(t *testing.T) {
	content := []byte("BT /F1 12 Tf 20 100 Td (Predicted) Tj ET")
	b := pdftest.New("1.7").OnePage()
	b.Stream(4, 0, "/Filter /FlateDecode /DecodeParms << /Predictor 12 /Columns 6 >>", pdftest.Flate(pdftest.PNGUp(content, 6)))
	out := save(t, open(t, b.XRefTable("/Root 1 0 R"), ""))
	if !bytes.Contains(out, []byte("/Predictor 12")) {
		t.Fatal("predictor parameters were dropped")
	}
	if got := text(t, open(t, out, "")); !strings.Contains(got, "Predicted") {
		t.Fatalf("text = %q", got)
	}
	if s, ok := popplerText(t, out); ok && !strings.Contains(s, "Predicted") {
		t.Fatalf("poppler text = %q", s)
	}
}
