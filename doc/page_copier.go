// Copyright (c) 2026 Gembit Soultan Shirazi <gembit.soultan@gmail.com>

package doc

import (
	"fmt"

	"github.com/gsoultan/gpdf/model"
)

// pageCopyDroppedKeys are page entries that point back into the source document's
// structure and must not be followed when a page is copied on its own: /Parent leads to
// the whole source page tree, /StructParents and /B to the structure tree and article beads.
var pageCopyDroppedKeys = []model.Name{"Parent", "StructParents", "B"}

// copyPages clones pages from src into dst and returns the new page references.
// Every selected page is given its destination number before anything is cloned, so
// references between selected pages (link /Dest, annotation /P) are remapped, while
// references to pages that are not copied become null instead of pulling them in.
func copyPages(src Document, dst *builtDocument, pages []model.Page) ([]model.Ref, error) {
	c := newObjectCloner(src, dst)
	newRefs := make([]model.Ref, len(pages))
	placed := make(map[int]bool)
	for i, p := range pages {
		num := p.Ref.ObjectNumber
		if num > 0 && !placed[num] {
			if _, reserved := c.remapped[num]; !reserved {
				c.remapped[num] = dst.size
				dst.size++
			}
			placed[num] = true
			newRefs[i] = model.Ref{ObjectNumber: c.remapped[num]}
			continue
		}
		// Unknown or repeated page: it gets a number of its own.
		newRefs[i] = model.Ref{ObjectNumber: dst.size}
		dst.size++
	}
	for i, p := range pages {
		pageDict := make(model.Dict, len(p.Dict))
		for k, v := range p.Dict {
			pageDict[k] = v
		}
		for _, k := range pageCopyDroppedKeys {
			delete(pageDict, k)
		}
		cloned, err := c.clone(pageDict)
		if err != nil {
			return nil, fmt.Errorf("clone page %d: %w", i, err)
		}
		dst.objects[newRefs[i].ObjectNumber] = cloned
	}
	return newRefs, nil
}

// newPageTree adds a flat /Pages node over refs and a catalog to dst, and sets the trailer.
func newPageTree(dst *builtDocument, refs []model.Ref) {
	kids := make(model.Array, len(refs))
	for i, ref := range refs {
		kids[i] = ref
	}
	pagesRef := dst.addObject(model.Dict{
		model.Name("Type"):  model.Name("Pages"),
		model.Name("Count"): model.Integer(int64(len(refs))),
		model.Name("Kids"):  kids,
	})
	for _, ref := range refs {
		if pageDict, ok := dst.objects[ref.ObjectNumber].(model.Dict); ok {
			pageDict[model.Name("Parent")] = pagesRef
		}
	}
	catalogRef := dst.addObject(model.Dict{
		model.Name("Type"):  model.Name("Catalog"),
		model.Name("Pages"): pagesRef,
	})
	if dst.trailer.Dict == nil {
		dst.trailer.Dict = make(model.Dict)
	}
	dst.trailer.Dict[model.Name("Root")] = catalogRef
	dst.trailer.Dict[model.Name("Size")] = model.Integer(int64(dst.size))
}

// copyInfo clones the source /Info dictionary into dst. A failure only drops the metadata.
func copyInfo(src Document, dst *builtDocument) {
	if dst.trailer.Dict == nil {
		dst.trailer.Dict = make(model.Dict)
	}
	info, err := src.Info()
	if err != nil || len(info) == 0 {
		return
	}
	newInfo, err := newObjectCloner(src, dst).clone(info)
	if err != nil {
		return
	}
	dst.trailer.Dict[model.Name("Info")] = dst.addObject(newInfo)
}
