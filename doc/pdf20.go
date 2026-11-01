// Copyright (c) 2026 Gembit Soultan Shirazi <gembit.soultan@gmail.com>

package doc

import (
	"fmt"
	"sort"

	"github.com/gsoultan/gpdf/model"
)

// PDF 2.0 (ISO 32000-2) document features: page-level output intents (§14.11.5)
// and document parts (§14.12).

// pageOutputIntent is an output intent attached to one page.
type pageOutputIntent struct {
	page      int
	profile   ICCProfile
	condition string
}

// OutputIntent attaches an output intent to this page (PDF 2.0 allows one per page in
// addition to the document's). condition names the printing condition, such as
// "FOGRA39"; the profile's data is embedded as the destination profile.
func (p *Page) OutputIntent(profile ICCProfile, condition string) *Page {
	if len(profile.Data) == 0 {
		p.builder.setErr(fmt.Errorf("page output intent needs ICC profile data"))
		return p
	}
	if condition == "" {
		condition = "Custom"
	}
	p.builder.pageIntents = append(p.builder.pageIntents, pageOutputIntent{p.pageIndex, profile, condition})
	return p
}

func (b *DocumentBuilder) buildPageOutputIntents(objs map[int]model.Object, pageNums []int, nextNum int) int {
	for _, oi := range b.pageIntents {
		if oi.page < 0 || oi.page >= len(pageNums) {
			continue
		}
		pd, ok := objs[pageNums[oi.page]].(model.Dict)
		if !ok {
			continue
		}
		iccNum := nextNum
		nextNum++
		objs[iccNum] = &model.Stream{
			Dict: model.Dict{
				model.Name("N"):         model.Integer(int64(oi.profile.N)),
				model.Name("Alternate"): model.Name(oi.profile.Alternate),
				model.Name("Length"):    model.Integer(int64(len(oi.profile.Data))),
				model.Name("Filter"):    model.Name("FlateDecode"),
			},
			Content: oi.profile.Data,
		}
		intentNum := nextNum
		nextNum++
		objs[intentNum] = model.Dict{
			model.Name("Type"):                      model.Name("OutputIntent"),
			model.Name("S"):                         model.Name("GTS_PDFX"),
			model.Name("OutputConditionIdentifier"): model.String(oi.condition),
			model.Name("Info"):                      model.String(oi.condition),
			model.Name("DestOutputProfile"):         model.Ref{ObjectNumber: iccNum},
		}
		arr, _ := pd[model.Name("OutputIntents")].(model.Array)
		pd[model.Name("OutputIntents")] = append(arr, model.Ref{ObjectNumber: intentNum})
	}
	return nextNum
}

// documentPart is a range of pages with document part metadata.
type documentPart struct {
	first, last int
	metadata    map[string]string
}

// DocumentPart declares pages first..last (0-based, inclusive) as one document part
// (PDF 2.0 §14.12), with the given document part metadata (/DPM). Parts must not
// overlap; pages outside every part get a part of their own without metadata, so the
// whole document is covered as the standard requires. Mail merge and print pipelines
// use parts to find each recipient's pages.
func (b *DocumentBuilder) DocumentPart(first, last int, metadata map[string]string) *DocumentBuilder {
	if first < 0 || last < first {
		b.setErr(fmt.Errorf("document part %d..%d: invalid page range", first, last))
		return b
	}
	for _, p := range b.parts {
		if first <= p.last && last >= p.first {
			b.setErr(fmt.Errorf("document part %d..%d overlaps %d..%d", first, last, p.first, p.last))
			return b
		}
	}
	b.parts = append(b.parts, documentPart{first, last, metadata})
	return b
}

func (b *DocumentBuilder) buildDocumentParts(objs map[int]model.Object, catalogDict model.Dict, pageNums []int, nextNum int) error {
	if len(b.parts) == 0 {
		return nil
	}
	n := len(pageNums)
	parts := append([]documentPart(nil), b.parts...)
	sort.Slice(parts, func(i, j int) bool { return parts[i].first < parts[j].first })
	for _, p := range parts {
		if p.last >= n {
			return fmt.Errorf("document part %d..%d: document has %d pages", p.first, p.last, n)
		}
	}
	// Fill the gaps so every page belongs to a part.
	var all []documentPart
	next := 0
	for _, p := range parts {
		if p.first > next {
			all = append(all, documentPart{first: next, last: p.first - 1})
		}
		all = append(all, p)
		next = p.last + 1
	}
	if next < n {
		all = append(all, documentPart{first: next, last: n - 1})
	}

	rootNum := nextNum
	nextNum++
	rootNodeNum := nextNum
	nextNum++
	rootRef := model.Ref{ObjectNumber: rootNum}
	rootNodeRef := model.Ref{ObjectNumber: rootNodeNum}
	var partRefs model.Array
	for _, p := range all {
		partNum := nextNum
		nextNum++
		partRef := model.Ref{ObjectNumber: partNum}
		d := model.Dict{
			model.Name("Type"):   model.Name("DPart"),
			model.Name("Parent"): rootNodeRef,
			model.Name("Start"):  model.Ref{ObjectNumber: pageNums[p.first]},
		}
		if p.last != p.first {
			d[model.Name("End")] = model.Ref{ObjectNumber: pageNums[p.last]}
		}
		if len(p.metadata) > 0 {
			dpm := model.Dict{}
			for k, v := range p.metadata {
				dpm[model.Name(k)] = model.EncodeTextString(v, b.metadata.PDF20)
			}
			d[model.Name("DPM")] = dpm
		}
		objs[partNum] = d
		partRefs = append(partRefs, partRef)
		for i := p.first; i <= p.last; i++ {
			if pd, ok := objs[pageNums[i]].(model.Dict); ok {
				pd[model.Name("DPart")] = partRef
			}
		}
	}
	objs[rootNodeNum] = model.Dict{
		model.Name("Type"):   model.Name("DPart"),
		model.Name("Parent"): rootRef,
		model.Name("DParts"): model.Array{partRefs},
	}
	objs[rootNum] = model.Dict{
		model.Name("Type"):          model.Name("DPartRoot"),
		model.Name("DPartRootNode"): rootNodeRef,
		model.Name("NodeNameList"):  model.Array{model.Name("Root"), model.Name("Part")},
	}
	catalogDict[model.Name("DPartRoot")] = rootRef
	b.partsNextNum = nextNum
	return nil
}
