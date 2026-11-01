// Copyright (c) 2026 Gembit Soultan Shirazi <gembit.soultan@gmail.com>

package doc

import (
	"cmp"
	"slices"

	"github.com/gsoultan/gpdf/content"
	"github.com/gsoultan/gpdf/model"
)

// PaintOrder selects how a page's text, image and graphic runs are ordered in its content
// stream. Later content paints over earlier content.
type PaintOrder int

const (
	// PaintLayered paints all graphics first, then all text, then all images. It is the
	// default, so backgrounds and borders added after text still end up behind it.
	PaintLayered PaintOrder = iota
	// PaintInCallOrder paints runs in the order they were added to the page, interleaving
	// text, images and graphics. Use it to reproduce a stacking order, such as an HTML
	// layout where a background image sits under later text and an icon over it.
	PaintInCallOrder
)

// SetPaintOrder sets the paint order used for every page when the document is built.
func (b *DocumentBuilder) SetPaintOrder(order PaintOrder) *DocumentBuilder {
	b.paintOrder = order
	return b
}

type runKind int

const (
	graphicKind runKind = iota
	textKind
	imageKind
)

type orderedRun struct {
	seq   int
	kind  runKind
	index int
}

// buildOpsInCallOrder builds the page's ops with all runs sorted by the sequence number
// they got when added. Runs without one (seq 0) keep the layered order ahead of the rest.
func (b *DocumentBuilder) buildOpsInCallOrder(page *pageSpec, imageStreamNums []int) ([]content.Op, map[string]model.Name, model.Dict, model.Dict) {
	runs := make([]orderedRun, 0, len(page.GraphicRuns)+len(page.TextRuns)+len(page.ImageRuns))
	for i, r := range page.GraphicRuns {
		runs = append(runs, orderedRun{r.Seq, graphicKind, i})
	}
	for i, r := range page.TextRuns {
		runs = append(runs, orderedRun{r.Seq, textKind, i})
	}
	for i, r := range page.ImageRuns {
		runs = append(runs, orderedRun{r.Seq, imageKind, i})
	}
	slices.SortStableFunc(runs, func(x, y orderedRun) int { return cmp.Compare(x.seq, y.seq) })

	tw := b.newTextOpsWriter(nil)
	iw := newImageOpsWriter(imageStreamNums)
	var ops []content.Op
	for _, r := range runs {
		switch r.kind {
		case graphicKind:
			ops = b.appendGraphicOps(ops, page.GraphicRuns[r.index].Ops)
			// Graphics may set the fill color or text state (e.g. inside an unbalanced
			// marked-content or layer run), so the next text run sets its own again.
			tw.forget()
		case textKind:
			ops = tw.add(ops, page.TextRuns[r.index])
		case imageKind:
			ops = iw.add(ops, r.index, page.ImageRuns[r.index])
		}
	}
	return ops, tw.fontRes, iw.xobj, iw.gs
}
