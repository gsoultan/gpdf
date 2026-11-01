package doc

import (
	"strconv"
	"strings"
)

// PageDecorator draws on one page once the final page count is known. It runs at
// Build time for every page, after all other content was added, so it can print
// "Page 3 of 7" or draw on pages added later by text flow. Decorators must only add
// content; what they draw is marked as a pagination artifact for tagged output.
type PageDecorator func(p *Page, pageIndex, pageCount int)

// PageText describes a running header or footer line.
type PageText struct {
	// Text may contain {page} and {pages}, replaced by the 1-based page number and
	// the page count.
	Text     string
	FontName string  // default Helvetica
	FontSize float64 // default 9
	Color    Color   // default black
	Align    TextAlign
	// Margin is the distance from the top (header) or bottom (footer) page edge to the
	// text baseline. The default is 36pt.
	Margin float64
	// Left and Right are the side margins. The default is 72pt.
	Left, Right float64
	// SkipFirst leaves the first page without this text (for title pages).
	SkipFirst bool
}

// OnEveryPage registers a decorator that runs for each page when the document is
// built. Several decorators run in the order they were added.
func (b *DocumentBuilder) OnEveryPage(fn PageDecorator) *DocumentBuilder {
	if fn != nil {
		b.decorators = append(b.decorators, fn)
	}
	return b
}

// SetHeader draws t at the top of every page.
func (b *DocumentBuilder) SetHeader(t PageText) *DocumentBuilder {
	return b.OnEveryPage(func(p *Page, i, n int) {
		if t.SkipFirst && i == 0 {
			return
		}
		y := p.Height() - t.margin()
		b.drawPageText(p, t, y, i, n)
	})
}

// SetFooter draws t at the bottom of every page.
func (b *DocumentBuilder) SetFooter(t PageText) *DocumentBuilder {
	return b.OnEveryPage(func(p *Page, i, n int) {
		if t.SkipFirst && i == 0 {
			return
		}
		b.drawPageText(p, t, t.margin(), i, n)
	})
}

func (t PageText) margin() float64 {
	if t.Margin > 0 {
		return t.Margin
	}
	return 36
}

// Expand replaces the {page} and {pages} placeholders.
func (t PageText) Expand(pageIndex, pageCount int) string {
	r := strings.NewReplacer("{page}", strconv.Itoa(pageIndex+1), "{pages}", strconv.Itoa(pageCount))
	return r.Replace(t.Text)
}

func (b *DocumentBuilder) drawPageText(p *Page, t PageText, y float64, pageIndex, pageCount int) {
	s := t.Expand(pageIndex, pageCount)
	if s == "" {
		return
	}
	font, size := t.FontName, t.FontSize
	if font == "" {
		font = "Helvetica"
	}
	if size <= 0 {
		size = 9
	}
	left, right := t.Left, t.Right
	if left <= 0 {
		left = 72
	}
	if right <= 0 {
		right = 72
	}
	x := left
	switch t.Align {
	case TextAlignCenter:
		x = (left + p.Width() - right - b.textWidth(s, size, font)) / 2
	case TextAlignRight:
		x = p.Width() - right - b.textWidth(s, size, font)
	}
	ps := &b.pc.pages[p.pageIndex]
	for _, seg := range b.fc.resolveFont(s, font) {
		ps.AddText(textRun{
			Text: seg.drawText(), Glyphs: seg.glyphs, X: x, Y: y, FontName: seg.fontName, FontSize: size,
			TextColorRGB: [3]float64{t.Color.R, t.Color.G, t.Color.B},
			IsArtifact:   true,
		})
		x += b.segmentWidth(seg, size, 0, 0)
	}
}

// applyDecorators runs the page decorators and returns a function that removes what
// they added, so that Build can be called again without duplicating headers.
func (b *DocumentBuilder) applyDecorators() (restore func()) {
	if len(b.decorators) == 0 {
		return func() {}
	}
	type mark struct{ text, image, graphic, seq int }
	marks := make([]mark, len(b.pc.pages))
	for i := range b.pc.pages {
		ps := &b.pc.pages[i]
		marks[i] = mark{len(ps.TextRuns), len(ps.ImageRuns), len(ps.GraphicRuns), ps.NextSeq}
	}
	n := len(b.pc.pages)
	for i := range n {
		p := &Page{builder: b, pageIndex: i}
		for _, fn := range b.decorators {
			fn(p, i, n)
		}
	}
	return func() {
		for i := range marks {
			if i >= len(b.pc.pages) {
				break
			}
			ps := &b.pc.pages[i]
			m := marks[i]
			ps.TextRuns = ps.TextRuns[:m.text]
			ps.ImageRuns = ps.ImageRuns[:m.image]
			ps.GraphicRuns = ps.GraphicRuns[:m.graphic]
			ps.NextSeq = m.seq
		}
	}
}
