package builder

import "github.com/gsoultan/gpdf/model"

// Page holds a page dict and collected content runs.
type Page struct {
	Dict        model.Dict
	TextRuns    []TextRun
	ImageRuns   []ImageRun
	GraphicRuns []GraphicRun
	NextMCID    int
	NextGSIndex int
	CurrX       float64
	CurrY       float64
	// NextSeq is the sequence number the next added run gets. Runs record it in Seq so
	// that pages built in call order can interleave text, images and graphics.
	NextSeq int
}

// AddText appends a text run and stamps it with the page's next sequence number.
func (p *Page) AddText(r TextRun) {
	r.Seq = p.nextSeq()
	p.TextRuns = append(p.TextRuns, r)
}

// AddImage appends an image run and stamps it with the page's next sequence number.
func (p *Page) AddImage(r ImageRun) {
	r.Seq = p.nextSeq()
	p.ImageRuns = append(p.ImageRuns, r)
}

// AddGraphic appends a graphic run and stamps it with the page's next sequence number.
func (p *Page) AddGraphic(r GraphicRun) {
	r.Seq = p.nextSeq()
	p.GraphicRuns = append(p.GraphicRuns, r)
}

func (p *Page) nextSeq() int {
	p.NextSeq++
	return p.NextSeq
}
