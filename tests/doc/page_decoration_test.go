package doc_test

import (
	"strings"
	"testing"

	"github.com/gsoultan/gpdf/doc"
)

func pageContents(t *testing.T, d doc.Document) []string {
	t.Helper()
	texts, err := d.ReadContentPerPage()
	if err != nil {
		t.Fatal(err)
	}
	return texts
}

func TestHeaderAndFooterOnEveryPage(t *testing.T) {
	b := doc.New().PageSize(300, 200).
		SetHeader(doc.PageText{Text: "Report", Align: doc.TextAlignCenter}).
		SetFooter(doc.PageText{Text: "Page {page} of {pages}", Align: doc.TextAlignRight, SkipFirst: true})
	b.AddPage().DrawText("one", 72, 100, "Helvetica", 12)
	b.AddPage().DrawText("two", 72, 100, "Helvetica", 12)
	b.AddPage()
	d, err := b.Build()
	if err != nil {
		t.Fatal(err)
	}
	texts := pageContents(t, d)
	if len(texts) != 3 {
		t.Fatalf("got %d pages", len(texts))
	}
	for i, want := range []string{"Report", "Page 2 of 3", "Page 3 of 3"} {
		if !strings.Contains(texts[i], want) {
			t.Errorf("page %d: want %q in %q", i+1, want, texts[i])
		}
	}
	if strings.Contains(texts[0], "Page 1") {
		t.Errorf("SkipFirst ignored: %q", texts[0])
	}
	// The footer is a pagination artifact.
	if ops := firstPageContent(t, d); !strings.Contains(ops, "/Artifact") {
		t.Errorf("header is not marked as an artifact:\n%s", ops)
	}
}

func TestOnEveryPageDecoratorRunsAfterContentAndIsRepeatable(t *testing.T) {
	calls := 0
	b := doc.New().PageSize(300, 200).OnEveryPage(func(p *doc.Page, i, n int) {
		calls++
		p.Text("Decor").At(10, p.Height()-10).Draw()
	})
	b.AddPage()
	b.AddPage()
	for range 2 { // Build twice: the second build must not carry duplicate headers.
		d, err := b.Build()
		if err != nil {
			t.Fatal(err)
		}
		texts := pageContents(t, d)
		if len(texts) != 2 || strings.Count(texts[0], "Decor") != 1 || strings.Count(texts[1], "Decor") != 1 {
			t.Fatalf("texts = %q", texts)
		}
	}
	if calls != 4 {
		t.Fatalf("decorator called %d times, want 4", calls)
	}
}
