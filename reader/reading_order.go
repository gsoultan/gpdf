// Copyright (c) 2026 Gembit Soultan Shirazi <gembit.soultan@gmail.com>

package reader

import (
	"sort"
	"strings"

	"github.com/gsoultan/gpdf/content"
	contentimpl "github.com/gsoultan/gpdf/content/impl"
	"github.com/gsoultan/gpdf/model"
)

// TextElement is one piece of text in reading order: a structure element's own
// text for tagged documents, or a paragraph found by layout analysis otherwise.
type TextElement struct {
	// Role is the structure type (P, H1, TD, Lbl, LBody, Figure ...), after the
	// document's RoleMap; "P" for paragraphs found by layout analysis.
	Role string
	// Page is the 0-based page index, or -1 when the element names no page.
	Page int
	// Depth is the nesting depth in the structure tree (0 for the top-level
	// elements), or 0 for layout analysis.
	Depth int
	// Parent identifies the enclosing structure element: elements with the same
	// Parent are siblings (the cells of one row, a list item's label and body).
	Parent int
	Text   string
	// Alt is the element's alternative description (figures), when set.
	Alt string
}

// structuredSource is what reading-order extraction needs: pages, objects and
// the catalog (for the structure tree).
type structuredSource interface {
	contentSource
	Catalog() (*model.Catalog, error)
}

// ExtractStructuredText returns the document's text in reading order. For a
// tagged document (ISO 32000-2 §14.8) it walks the structure tree and gathers
// the marked content of every element, so columns, sidebars, tables and lists
// come out in the order the author tagged them, with ActualText honoured and
// artifacts (headers, footers, decorations) left out. For an untagged document
// it falls back to layout analysis (see ExtractTextInReadingOrder).
func ExtractStructuredText(src structuredSource) ([]TextElement, error) {
	elems, err := structureTreeText(src)
	if err != nil {
		return nil, err
	}
	if len(elems) > 0 {
		return elems, nil
	}
	return layoutOrderText(src)
}

// ExtractTextInReadingOrder returns the document text in reading order as one
// string: tagged documents through their structure tree, untagged ones through
// layout analysis that finds columns and paragraphs (recursive XY cuts on the
// whitespace between text blocks). Block elements are separated by newlines, a
// list label from its body by a space, table cells in a row by tabs, and pages
// by a blank line.
func ExtractTextInReadingOrder(src structuredSource) (string, error) {
	elems, err := ExtractStructuredText(src)
	if err != nil {
		return "", err
	}
	return JoinTextElements(elems), nil
}

// JoinTextElements renders elements as ExtractTextInReadingOrder does.
func JoinTextElements(elems []TextElement) string {
	var sb strings.Builder
	var prev *TextElement
	for i := range elems {
		e := &elems[i]
		text := strings.TrimSpace(e.Text)
		if text == "" && e.Alt == "" {
			continue
		}
		if text == "" {
			text = e.Alt
		}
		if prev != nil {
			switch {
			case prev.Page != e.Page && e.Page >= 0 && prev.Page >= 0:
				sb.WriteString("\n\n")
			case prev.Role == "Lbl" && prev.Parent == e.Parent:
				sb.WriteString(" ")
			case isCellRole(prev.Role) && isCellRole(e.Role) && prev.Parent == e.Parent:
				sb.WriteString("\t")
			default:
				sb.WriteString("\n")
			}
		}
		sb.WriteString(text)
		prev = e
	}
	return sb.String()
}

func isCellRole(role string) bool { return role == "TD" || role == "TH" }

// inlineRoles are the structure types whose text continues their parent's text.
var inlineRoles = map[string]bool{
	"Span": true, "Quote": true, "Note": true, "Reference": true, "BibEntry": true,
	"Code": true, "Link": true, "Annot": true, "Ruby": true, "Warichu": true,
	"RB": true, "RT": true, "RP": true, "WT": true, "WP": true, "Em": true, "Strong": true,
	"Sub": true,
}

// structureTreeText walks the structure tree. It returns nil when the document
// has no structure tree or the tree yields no text.
func structureTreeText(src structuredSource) ([]TextElement, error) {
	cat, err := src.Catalog()
	if err != nil || cat == nil {
		return nil, nil
	}
	rootRef := cat.StructTreeRootRef()
	if rootRef == nil {
		return nil, nil
	}
	root, ok := resolveDictObject(src, *rootRef)
	if !ok {
		return nil, nil
	}
	pages, err := src.Pages()
	if err != nil {
		return nil, err
	}
	w := &structWalker{
		src:      src,
		pageIdx:  make(map[model.Ref]int, len(pages)),
		pages:    pages,
		mcText:   make(map[int]map[int]string),
		roleMap:  map[string]string{},
		visited:  map[model.Ref]bool{},
		maxElems: 1 << 20,
	}
	for i, p := range pages {
		if p.Ref.ObjectNumber != 0 {
			w.pageIdx[p.Ref] = i
		}
	}
	if rm, ok := resolveDictObject(src, root[model.Name("RoleMap")]); ok {
		for k, v := range rm {
			if n, ok := v.(model.Name); ok {
				w.roleMap[string(k)] = string(n)
			}
		}
	}
	w.walkKids(root[model.Name("K")], -1, 0)
	if !w.anyText {
		return nil, nil
	}
	return w.out, nil
}

type structWalker struct {
	src     structuredSource
	pages   []model.Page
	pageIdx map[model.Ref]int
	// mcText caches, per page, the text of each marked-content sequence by MCID.
	mcText   map[int]map[int]string
	roleMap  map[string]string
	visited  map[model.Ref]bool
	out      []TextElement
	anyText  bool
	count    int
	maxElems int
	// parent is the id of the element being walked (ids count visited elements).
	parent int
}

// role resolves an element's structure type through the RoleMap (bounded).
func (w *structWalker) role(s string) string {
	for range 8 {
		mapped, ok := w.roleMap[s]
		if !ok || mapped == s {
			break
		}
		s = mapped
	}
	return s
}

func (w *structWalker) pageOf(obj model.Object, inherited int) int {
	if ref, ok := obj.(model.Ref); ok {
		if i, ok := w.pageIdx[ref]; ok {
			return i
		}
	}
	return inherited
}

// walkKids visits the kids of an element: structure elements are walked, marked
// content references are gathered into the parent's text. The returned string is
// the inline text gathered for the parent.
func (w *structWalker) walkKids(k model.Object, page int, depth int) {
	w.walkElement(model.Dict{model.Name("K"): k, model.Name("S"): model.Name("")}, page, depth-1, true)
}

// walkElement emits the text of elem and its descendants in order. Inline
// elements return their text to the caller instead of emitting it.
func (w *structWalker) walkElement(elem model.Dict, inheritedPage int, depth int, isRoot bool) string {
	w.count++
	if w.count > w.maxElems {
		return ""
	}
	parentID := w.parent
	w.parent = w.count
	defer func() { w.parent = parentID }()
	role := w.role(string(nameOf(elem[model.Name("S")])))
	page := w.pageOf(elem[model.Name("Pg")], inheritedPage)
	inline := inlineRoles[role] && !isRoot
	alt, _ := elem[model.Name("Alt")].(model.String)
	if alt == "" {
		// Some producers put /Alt in the attribute dictionary.
		if a, ok := resolveDictObject(w.src, elem[model.Name("A")]); ok {
			alt, _ = a[model.Name("Alt")].(model.String)
		}
	}

	if actual, ok := elem[model.Name("ActualText")].(model.String); ok {
		text := actual.Text()
		if inline {
			return text
		}
		w.emit(role, page, depth, parentID, text, alt.Text())
		return ""
	}

	var pending strings.Builder
	flush := func() {
		if pending.Len() > 0 || alt != "" {
			w.emit(role, page, depth, parentID, pending.String(), alt.Text())
			alt = ""
		}
		pending.Reset()
	}
	addText := func(t string) {
		if t == "" {
			return
		}
		if pending.Len() > 0 && !strings.HasSuffix(pending.String(), " ") && !strings.HasPrefix(t, " ") {
			pending.WriteByte(' ')
		}
		pending.WriteString(t)
	}

	kids := elem[model.Name("K")]
	var list model.Array
	switch v := kids.(type) {
	case model.Array:
		list = v
	case nil:
	default:
		list = model.Array{v}
	}
	for _, kid := range list {
		switch v := kid.(type) {
		case model.Integer:
			addText(w.markedContentText(page, int(v)))
		case model.Ref:
			if w.visited[v] {
				continue
			}
			obj, err := w.src.Resolve(v)
			if err != nil {
				continue
			}
			d, ok := obj.(model.Dict)
			if !ok {
				if n, ok := obj.(model.Integer); ok {
					addText(w.markedContentText(page, int(n)))
				}
				continue
			}
			w.visited[v] = true
			t := w.kidDict(d, page, depth, inline, &pending, flush, addText)
			delete(w.visited, v)
			addText(t)
		case model.Dict:
			addText(w.kidDict(v, page, depth, inline, &pending, flush, addText))
		}
	}
	if inline {
		return pending.String()
	}
	flush()
	return ""
}

// kidDict handles a kid that is a dictionary: a marked-content reference, an
// object reference, or a child structure element.
func (w *structWalker) kidDict(d model.Dict, page, depth int, parentInline bool, pending *strings.Builder, flush func(), addText func(string)) string {
	switch nameOf(d[model.Name("Type")]) {
	case "MCR":
		if mcid, ok := d[model.Name("MCID")].(model.Integer); ok {
			return w.markedContentText(w.pageOf(d[model.Name("Pg")], page), int(mcid))
		}
		return ""
	case "OBJR":
		return ""
	}
	childRole := w.role(string(nameOf(d[model.Name("S")])))
	if inlineRoles[childRole] {
		return w.walkElement(d, page, depth+1, false)
	}
	if parentInline {
		// A block inside an inline element: emit it in place.
		return w.walkElement(d, page, depth+1, false)
	}
	flush()
	w.walkElement(d, page, depth+1, false)
	return ""
}

func (w *structWalker) emit(role string, page, depth, parent int, text, alt string) {
	text = strings.TrimSpace(text)
	if text == "" && alt == "" {
		return
	}
	if depth < 0 {
		depth = 0
	}
	w.anyText = true
	w.out = append(w.out, TextElement{Role: role, Page: page, Depth: depth, Parent: parent, Text: text, Alt: alt})
}

// markedContentText returns the text of marked-content sequence mcid on page.
func (w *structWalker) markedContentText(page, mcid int) string {
	if page < 0 || page >= len(w.pages) {
		return ""
	}
	m, ok := w.mcText[page]
	if !ok {
		m = w.pageMarkedContent(page)
		w.mcText[page] = m
	}
	return m[mcid]
}

// pageMarkedContent collects the text of every marked-content sequence on a page
// by MCID, including sequences inside form XObjects drawn from the page.
func (w *structWalker) pageMarkedContent(page int) map[int]string {
	out := make(map[int]string)
	p := w.pages[page]
	raw, err := pageContentBytes(w.src, p)
	if err != nil || len(raw) == 0 {
		return out
	}
	parser := contentimpl.NewStreamParser()
	ops, err := parser.Parse(raw)
	if err != nil {
		return out
	}
	resources, _ := p.Resources()
	builders := make(map[int]*strings.Builder)
	collectMarkedContent(ops, w.src, parser, resources, map[model.Ref]struct{}{}, map[model.Ref]*toUnicodeDecoder{}, nil, builders)
	for mcid, sb := range builders {
		out[mcid] = strings.TrimSpace(sb.String())
	}
	return out
}

// collectMarkedContent walks ops with a stack of marked-content ids and appends
// the shown text to the builder of the innermost MCID.
func collectMarkedContent(
	ops []content.Op,
	src contentSource,
	parser content.Parser,
	resources model.Dict,
	visited map[model.Ref]struct{},
	fontDecoders map[model.Ref]*toUnicodeDecoder,
	stack []int,
	builders map[int]*strings.Builder,
) []int {
	inText := false
	var decoder *toUnicodeDecoder
	current := func() *strings.Builder {
		for i := len(stack) - 1; i >= 0; i-- {
			if stack[i] >= 0 {
				sb := builders[stack[i]]
				if sb == nil {
					sb = &strings.Builder{}
					builders[stack[i]] = sb
				}
				return sb
			}
		}
		return nil
	}
	for _, op := range ops {
		switch op.Name {
		case "BDC":
			stack = append(stack, markedContentID(op, resources, src))
		case "BMC":
			stack = append(stack, -1)
		case "EMC":
			if len(stack) > 0 {
				stack = stack[:len(stack)-1]
			}
		case "BT":
			inText = true
		case "ET":
			inText = false
		case "Tf":
			if len(op.Args) > 0 {
				decoder = resolveFontDecoder(src, resources, op.Args[0], fontDecoders)
			}
		case "Tj", "TJ":
			if inText && len(op.Args) > 0 {
				if sb := current(); sb != nil {
					appendTextArg(sb, op.Args[0], decoder)
				}
			}
		case "'", "\"":
			if inText && len(op.Args) > 0 {
				if sb := current(); sb != nil {
					appendBoundary(sb, ' ')
					appendTextArg(sb, op.Args[len(op.Args)-1], decoder)
				}
			}
		case "T*", "Td", "TD", "Tm":
			if inText {
				if sb := current(); sb != nil {
					if op.Name == "Td" || op.Name == "TD" {
						if len(op.Args) >= 2 && abs64(toFloat64(op.Args[1])) <= 0.5 {
							continue
						}
					}
					appendBoundary(sb, ' ')
				}
			}
		case "Do":
			stack = collectMarkedContentXObject(op.Args, src, parser, resources, visited, fontDecoders, stack, builders)
		}
	}
	return stack
}

func collectMarkedContentXObject(
	args model.Array,
	src contentSource,
	parser content.Parser,
	resources model.Dict,
	visited map[model.Ref]struct{},
	fontDecoders map[model.Ref]*toUnicodeDecoder,
	stack []int,
	builders map[int]*strings.Builder,
) []int {
	if len(args) == 0 || resources == nil {
		return stack
	}
	name, ok := args[0].(model.Name)
	if !ok {
		return stack
	}
	xObjects, ok := resolveDictObject(src, resources[model.Name("XObject")])
	if !ok {
		return stack
	}
	xObject, ref, ok := resolveStreamObject(src, xObjects[name])
	if !ok || xObject == nil || len(xObject.Content) == 0 {
		return stack
	}
	if nameOf(xObject.Dict[model.Name("Subtype")]) != "Form" {
		return stack
	}
	if ref != nil {
		if _, seen := visited[*ref]; seen {
			return stack
		}
		visited[*ref] = struct{}{}
		defer delete(visited, *ref)
	}
	ops, err := parser.Parse(xObject.Content)
	if err != nil {
		return stack
	}
	nested := resources
	if r, ok := resolveDictObject(src, xObject.Dict[model.Name("Resources")]); ok {
		nested = mergeResourceDict(resources, r)
	}
	depth := len(stack)
	inner := collectMarkedContent(ops, src, parser, nested, visited, fontDecoders, stack, builders)
	if len(inner) > depth {
		inner = inner[:depth]
	}
	return inner
}

// markedContentID returns the MCID of a BDC operator (from an inline property
// dictionary or a named /Properties resource), or -1.
func markedContentID(op content.Op, resources model.Dict, src contentSource) int {
	if len(op.Args) < 2 {
		return -1
	}
	var props model.Dict
	switch v := op.Args[1].(type) {
	case model.Dict:
		props = v
	case model.Name:
		if all, ok := resolveDictObject(src, resources[model.Name("Properties")]); ok {
			props, _ = resolveDictObject(src, all[v])
		}
	}
	if mcid, ok := props[model.Name("MCID")].(model.Integer); ok {
		return int(mcid)
	}
	return -1
}

func nameOf(o model.Object) model.Name {
	n, _ := o.(model.Name)
	return n
}

// layoutOrderText orders the text blocks of every page with recursive XY cuts
// and groups them into paragraphs.
func layoutOrderText(src contentSource) ([]TextElement, error) {
	layouts, err := ExtractLayout(src)
	if err != nil {
		return nil, err
	}
	var out []TextElement
	for _, pl := range layouts {
		for _, para := range OrderBlocks(pl.Blocks) {
			out = append(out, TextElement{Role: "P", Page: pl.Page, Text: para})
		}
	}
	return out, nil
}

// OrderBlocks arranges the text blocks of one page in reading order and groups
// them into paragraphs: the page is cut recursively along the widest horizontal
// or vertical band of whitespace (XY cut), so columns are read one after the
// other; inside a region, lines are read top to bottom and split into paragraphs
// at vertical gaps larger than the line spacing.
func OrderBlocks(blocks []TextBlock) []string {
	var kept []TextBlock
	for _, b := range blocks {
		if strings.TrimSpace(b.Text) != "" && b.Width > 0 && b.Height > 0 {
			kept = append(kept, b)
		}
	}
	if len(kept) == 0 {
		return nil
	}
	lineHeight := medianHeight(kept)
	var paras []string
	xyCut(kept, lineHeight, 0, func(region []TextBlock) {
		paras = append(paras, regionParagraphs(region, lineHeight)...)
	})
	return paras
}

func medianHeight(blocks []TextBlock) float64 {
	hs := make([]float64, len(blocks))
	for i, b := range blocks {
		hs[i] = b.Height
	}
	sort.Float64s(hs)
	h := hs[len(hs)/2]
	if h <= 0 {
		return 12
	}
	return h
}

// xyCut splits blocks at the widest whitespace band that no block crosses and
// recurses; a horizontal band (top/bottom) must be taller than a line gap, a
// vertical band (left/right) wider than a word gap.
func xyCut(blocks []TextBlock, lineHeight float64, depth int, leaf func([]TextBlock)) {
	if len(blocks) <= 1 || depth > 32 {
		leaf(blocks)
		return
	}
	// Candidate cuts: a horizontal band taller than 1.5 lines, a vertical band
	// wider than one line height; the wider band wins (a column gutter over the
	// gap between two paragraphs, a section break over a narrow gutter).
	top, bottom, yGap, yOK := splitAtGap(blocks, lineHeight*1.5, true, func(b TextBlock) (float64, float64) {
		return b.Y - 0.25*b.Height, b.Y + b.Height
	})
	left, right, xGap, xOK := splitAtGap(blocks, lineHeight*1.0, false, func(b TextBlock) (float64, float64) {
		return b.X, b.X + b.Width
	})
	switch {
	case xOK && (!yOK || xGap > yGap):
		xyCut(left, lineHeight, depth+1, leaf)
		xyCut(right, lineHeight, depth+1, leaf)
	case yOK:
		xyCut(top, lineHeight, depth+1, leaf)
		xyCut(bottom, lineHeight, depth+1, leaf)
	default:
		leaf(blocks)
	}
}

// splitAtGap finds the widest gap of at least minGap in the projection of blocks
// onto one axis and returns the blocks on each side and the gap. For the y axis
// the first group is the upper one (reading starts at the top).
func splitAtGap(blocks []TextBlock, minGap float64, yAxis bool, span func(TextBlock) (float64, float64)) (first, second []TextBlock, gap float64, ok bool) {
	type interval struct{ lo, hi float64 }
	ivs := make([]interval, len(blocks))
	for i, b := range blocks {
		lo, hi := span(b)
		ivs[i] = interval{lo, hi}
	}
	order := make([]int, len(ivs))
	for i := range order {
		order[i] = i
	}
	sort.Slice(order, func(i, j int) bool { return ivs[order[i]].lo < ivs[order[j]].lo })
	bestGap, bestAt := minGap, 0.0
	found := false
	reach := ivs[order[0]].hi
	for _, idx := range order[1:] {
		iv := ivs[idx]
		if g := iv.lo - reach; g >= bestGap {
			bestGap, bestAt, found = g, reach+g/2, true
		}
		if iv.hi > reach {
			reach = iv.hi
		}
	}
	if !found {
		return nil, nil, 0, false
	}
	for i, b := range blocks {
		if ivs[i].lo < bestAt {
			first = append(first, b)
		} else {
			second = append(second, b)
		}
	}
	if yAxis { // blocks with a larger Y are higher on the page
		return second, first, bestGap, true
	}
	return first, second, bestGap, true
}

// regionParagraphs reads a region line by line (top to bottom, left to right)
// and joins lines into paragraphs, starting a new one at a vertical gap larger
// than the usual line spacing.
func regionParagraphs(region []TextBlock, lineHeight float64) []string {
	rows := map[int][]TextBlock{}
	var keys []int
	for _, b := range region {
		k := yBucket(b.Y)
		if _, ok := rows[k]; !ok {
			keys = append(keys, k)
		}
		rows[k] = append(rows[k], b)
	}
	sort.Sort(sort.Reverse(sort.IntSlice(keys)))
	var paras []string
	var cur strings.Builder
	prevY := 0.0
	for i, k := range keys {
		row := rows[k]
		sort.Slice(row, func(a, b int) bool { return row[a].X < row[b].X })
		var line strings.Builder
		for j, b := range row {
			t := strings.TrimSpace(b.Text)
			if t == "" {
				continue
			}
			if j > 0 && line.Len() > 0 {
				line.WriteByte(' ')
			}
			line.WriteString(t)
		}
		if line.Len() == 0 {
			continue
		}
		y := row[0].Y
		if i > 0 && cur.Len() > 0 && prevY-y > lineHeight*1.8 {
			paras = append(paras, cur.String())
			cur.Reset()
		}
		if cur.Len() > 0 {
			cur.WriteByte(' ')
		}
		cur.WriteString(line.String())
		prevY = y
	}
	if cur.Len() > 0 {
		paras = append(paras, cur.String())
	}
	return paras
}
