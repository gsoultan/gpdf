package reader

import (
	"errors"
	"fmt"
	"math"

	"github.com/gsoultan/gpdf/content"
	contentimpl "github.com/gsoultan/gpdf/content/impl"
	"github.com/gsoultan/gpdf/model"
)

// RedactionReport says what RedactContent did.
type RedactionReport struct {
	// Removed counts the text-showing operators whose text was taken out of the
	// page content streams.
	Removed int
	// Boxes counts the filled rectangles drawn over the removed text.
	Boxes int
	// SkippedInForms counts matches inside form XObjects, which are shared
	// resources and are left untouched (and not covered by a box).
	SkippedInForms int
	// Pages lists the 0-based pages that were changed.
	Pages []int
}

// ErrNothingToRedact is returned when no text matched.
var ErrNothingToRedact = errors.New("redact: nothing matched")

// RedactMatcher decides whether a text block on a page is to be removed.
type RedactMatcher func(page int, block TextBlock) bool

// RedactContent removes the text that match selects from the page content
// streams and, when fill is non-nil, draws a rectangle of that RGB colour over
// each removed run. A text-showing operator is removed whole when any of its
// text matches, and replaced by an equivalent displacement so the text after
// it keeps its place. The document must then be saved with a full rewrite
// (Save), never as an incremental update, so that the removed bytes are gone;
// documents that track changes are marked accordingly. Fonts keep their glyph
// programs, and annotations, metadata and bookmarks are not searched.
func RedactContent(src contentSource, match RedactMatcher, fill *[3]float64) (RedactionReport, error) {
	var report RedactionReport
	pages, err := src.Pages()
	if err != nil {
		return report, err
	}
	parser := contentimpl.NewStreamParser()
	for pageIdx, page := range pages {
		refs, err := pageContentRefs(page)
		if err != nil || len(refs) == 0 {
			continue
		}
		raw, err := pageContentBytes(src, page)
		if err != nil || len(raw) == 0 {
			continue
		}
		ops, err := parser.Parse(raw)
		if err != nil {
			return report, fmt.Errorf("redact: page %d: %w", pageIdx+1, err)
		}
		resources, _ := page.Resources()
		var boxes [][4]float64
		removed := map[int]*layoutState{}
		walkTextOps(ops, src, parser, resources, map[model.Ref]struct{}{}, map[model.Ref]*toUnicodeDecoder{}, nil,
			func(i int, st *layoutState, blocks []TextBlock, nested bool) {
				var hit bool
				var box [4]float64
				for _, b := range blocks {
					if !match(pageIdx, b) {
						continue
					}
					if nested {
						report.SkippedInForms++
						continue
					}
					r := blockBox(b)
					if !hit {
						box = r
					} else {
						box = unionBox(box, r)
					}
					hit = true
				}
				if hit {
					removed[i] = st
					boxes = append(boxes, box)
				}
			})
		if len(removed) == 0 {
			continue
		}
		out := make([]content.Op, 0, len(ops)+2*len(boxes)+2)
		out = append(out, content.Op{Name: "q"})
		for i, op := range ops {
			if st, ok := removed[i]; ok {
				out = append(out, displacementOps(op, st)...)
				continue
			}
			out = append(out, op)
		}
		out = append(out, content.Op{Name: "Q"})
		if fill != nil {
			out = append(out, content.Op{Name: "q"}, content.Op{Name: "rg", Args: []model.Object{
				model.Real(fill[0]), model.Real(fill[1]), model.Real(fill[2]),
			}})
			for _, b := range boxes {
				out = append(out,
					content.Op{Name: "re", Args: []model.Object{model.Real(b[0]), model.Real(b[1]), model.Real(b[2]), model.Real(b[3])}},
					content.Op{Name: "f"})
			}
			out = append(out, content.Op{Name: "Q"})
			report.Boxes += len(boxes)
		}
		encoded, err := content.EncodeBytes(out)
		if err != nil {
			return report, err
		}
		// The rewritten content replaces the first stream; the others are emptied,
		// so every reference stays valid.
		for i, ref := range refs {
			obj, err := src.Resolve(ref)
			if err != nil {
				continue
			}
			s, ok := obj.(*model.Stream)
			if !ok {
				continue
			}
			if i == 0 {
				s.Content = encoded
			} else {
				s.Content = nil
			}
			s.Dict[model.Name("Length")] = model.Integer(int64(len(s.Content)))
			markModified(src, ref, s)
		}
		report.Removed += len(removed)
		report.Pages = append(report.Pages, pageIdx)
	}
	if report.Removed == 0 && report.SkippedInForms == 0 {
		return report, ErrNothingToRedact
	}
	if m, ok := src.(interface{ MarkRedacted() }); ok && report.Removed > 0 {
		m.MarkRedacted()
	}
	return report, nil
}

// blockBox is the rectangle a block covers, with room for descenders.
func blockBox(b TextBlock) [4]float64 {
	return [4]float64{b.X, b.Y - 0.25*b.Height, b.Width, b.Height * 1.25}
}

func unionBox(a, b [4]float64) [4]float64 {
	x0 := math.Min(a[0], b[0])
	y0 := math.Min(a[1], b[1])
	x1 := math.Max(a[0]+a[2], b[0]+b[2])
	y1 := math.Max(a[1]+a[3], b[1]+b[3])
	return [4]float64{x0, y0, x1 - x0, y1 - y0}
}

// displacementOps replaces a text-showing operator by operators that move the
// text position as far as the removed text did, without showing anything.
func displacementOps(op content.Op, st *layoutState) []content.Op {
	fs := st.fontSize
	if fs <= 0 {
		fs = 12
	}
	// A TJ number n moves by -n/1000*Tfs*Th; advanceForBytes already includes Th,
	// so the number for an advance a is -a*1000/(Tfs*Th).
	th := st.horizScale / 100
	if th == 0 {
		th = 1
	}
	number := func(adv float64) model.Object {
		return model.Real(-adv * 1000 / (fs * th))
	}
	var pre []content.Op
	var arr model.Array
	switch op.Name {
	case "Tj", "'", "\"":
		arg := op.Args[len(op.Args)-1]
		if op.Name == "\"" && len(op.Args) >= 3 {
			pre = append(pre,
				content.Op{Name: "Tw", Args: []model.Object{op.Args[0]}},
				content.Op{Name: "Tc", Args: []model.Object{op.Args[1]}})
		}
		if op.Name != "Tj" {
			pre = append(pre, content.Op{Name: "T*"})
		}
		if s, ok := arg.(model.String); ok {
			arr = model.Array{number(advanceForBytes([]byte(s), st, nil))}
		}
	case "TJ":
		if a, ok := op.Args[0].(model.Array); ok {
			for _, e := range a {
				if s, ok := e.(model.String); ok {
					arr = append(arr, number(advanceForBytes([]byte(s), st, nil)))
				} else {
					arr = append(arr, e)
				}
			}
		}
	}
	if len(arr) == 0 {
		return pre
	}
	return append(pre, content.Op{Name: "TJ", Args: []model.Object{arr}})
}
