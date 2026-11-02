// Copyright (c) 2026 Gembit Soultan Shirazi <gembit.soultan@gmail.com>

package doc_test

import (
	"strings"
	"testing"

	"github.com/gsoultan/gpdf/doc"
	"github.com/gsoultan/gpdf/model"
)

// A translucent text run gets a constant-alpha ExtGState in the page
// resources and is wrapped in q/Q so its state never leaks into the next run.
func TestDrawTextWithStyleOpacity(t *testing.T) {
	b := doc.New().PageSize(300, 300).AddPage()
	b.DrawTextWithStyle(0, "DRAFT", 20, 150, doc.TextStyle{FontName: "Helvetica-Bold", FontSize: 48,
		Color: doc.Color{R: 1}, Rotation: -30, Opacity: 0.25})
	b.DrawTextWithSpacing("Body", 10, 100, "Helvetica", 12, doc.ColorBlack, 0, 0)
	d, err := b.Build()
	if err != nil {
		t.Fatal(err)
	}
	ops := firstPageContent(t, d)
	i := strings.Index(ops, "/TXGS1 gs")
	if i < 0 {
		t.Fatalf("no gs operator for the translucent run:\n%s", ops)
	}
	run := ops[i:strings.Index(ops, "(Body)")]
	if !strings.Contains(run, "(DRAFT)") || !strings.Contains(run, "Q") {
		t.Fatalf("translucent run not wrapped in q/Q before the next run:\n%s", run)
	}
	if !strings.Contains(run, " Tm") {
		t.Fatalf("rotation not applied:\n%s", run)
	}
	pages, _ := d.Pages()
	res, _ := pages[0].Dict[model.Name("Resources")].(model.Dict)
	if res == nil {
		r := d.(resolver)
		obj, err := r.Resolve(pages[0].Dict[model.Name("Resources")].(model.Ref))
		if err != nil {
			t.Fatal(err)
		}
		res = obj.(model.Dict)
	}
	gs, _ := res[model.Name("ExtGState")].(model.Dict)
	st, _ := gs[model.Name("TXGS1")].(model.Dict)
	if st == nil {
		t.Fatalf("ExtGState TXGS1 missing from page resources: %v", res)
	}
	if ca, _ := st[model.Name("ca")].(model.Real); ca != 0.25 {
		t.Fatalf("ca = %v, want 0.25", st[model.Name("ca")])
	}
}

// An opaque run is unchanged: no gs, no q/Q wrapper.
func TestDrawTextWithStyleOpaque(t *testing.T) {
	b := doc.New().PageSize(300, 300).AddPage()
	b.DrawTextWithStyle(0, "Plain", 20, 150, doc.TextStyle{FontName: "Helvetica", FontSize: 12})
	d, err := b.Build()
	if err != nil {
		t.Fatal(err)
	}
	if ops := firstPageContent(t, d); strings.Contains(ops, " gs") {
		t.Fatalf("opaque run got a graphics state:\n%s", ops)
	}
}
