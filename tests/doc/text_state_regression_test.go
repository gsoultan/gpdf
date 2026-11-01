// Copyright (c) 2026 Gembit Soultan Shirazi <gembit.soultan@gmail.com>

package doc_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/gsoultan/gpdf/doc"
	"github.com/gsoultan/gpdf/model"
)

type resolver interface {
	Resolve(model.Ref) (model.Object, error)
}

func firstPageContent(t *testing.T, d doc.Document) string {
	t.Helper()
	pages, err := d.Pages()
	if err != nil || len(pages) == 0 {
		t.Fatalf("pages: %v", err)
	}
	r, ok := d.(resolver)
	if !ok {
		t.Fatal("document cannot resolve objects")
	}
	var refs []model.Ref
	switch c := pages[0].Dict[model.Name("Contents")].(type) {
	case model.Ref:
		refs = []model.Ref{c}
	case model.Array:
		for _, o := range c {
			refs = append(refs, o.(model.Ref))
		}
	}
	var sb strings.Builder
	for _, ref := range refs {
		obj, err := r.Resolve(ref)
		if err != nil {
			t.Fatal(err)
		}
		sb.Write(obj.(*model.Stream).Content)
	}
	return sb.String()
}

// Regression: Tc was only written when non-zero, and Tc survives ET, so a run without
// letter spacing inherited the previous run's spacing (proteus-converter worked around it).
func TestLetterSpacingDoesNotLeakIntoNextRun(t *testing.T) {
	b := doc.New().PageSize(300, 300).AddPage()
	b.DrawTextWithSpacing("Spaced", 10, 200, "Helvetica", 12, doc.ColorBlack, 3, 0)
	b.DrawTextWithSpacing("Plain", 10, 100, "Helvetica", 12, doc.ColorBlack, 0, 0)
	d, err := b.Build()
	if err != nil {
		t.Fatal(err)
	}
	ops := firstPageContent(t, d)
	rest := ops[strings.Index(ops, "(Spaced)"):]
	if !strings.Contains(rest[:strings.Index(rest, "(Plain)")], "0 Tc") {
		t.Fatalf("second run does not reset Tc:\n%s", ops)
	}
}

// Links are drawn without the default 1pt border box.
func TestLinkAnnotationsHaveNoBorder(t *testing.T) {
	b := doc.New().PageSize(300, 300).NoObjectStreams().AddPage().AddPage()
	b.AddLinkToURL(0, 10, 10, 100, 30, "https://example.com")
	b.AddLinkToPage(0, 10, 40, 100, 60, 1)
	d, err := b.Build()
	if err != nil {
		t.Fatal(err)
	}
	var out bytes.Buffer
	if err := d.Save(&out); err != nil {
		t.Fatal(err)
	}
	if n := bytes.Count(out.Bytes(), []byte("/Border [0 0 0]")); n != 2 {
		t.Fatalf("found %d links with /Border [0 0 0], want 2", n)
	}
}
