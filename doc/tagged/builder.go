package tagged

import "github.com/gsoultan/gpdf/model"

// inSection returns whether the given content index belongs to any section.
func inSection(sections []Section, blockIdx, figureIdx, tableIdx, listIdx int, kind byte) bool {
	for _, s := range sections {
		switch kind {
		case 'b':
			for _, i := range s.BlockIndices {
				if i == blockIdx {
					return true
				}
			}
		case 'f':
			for _, i := range s.FigureIndices {
				if i == figureIdx {
					return true
				}
			}
		case 't':
			for _, i := range s.TableIndices {
				if i == tableIdx {
					return true
				}
			}
		case 'l':
			for _, i := range s.ListIndices {
				if i == listIdx {
					return true
				}
			}
		}
	}
	return false
}

// BuildStructure builds MarkInfo, StructTreeRoot, and ParentTree for tagged PDF.
// It writes objects into objs, updates catalogDict, and returns the updated next object number.
func (s *Support) BuildStructure(objs map[int]model.Object, catalogDict model.Dict, pageNums []int, nextNum int) int {
	markInfoNum := nextNum
	nextNum++
	objs[markInfoNum] = model.Dict{
		model.Name("Marked"): model.Boolean(true),
	}
	catalogDict[model.Name("MarkInfo")] = model.Ref{ObjectNumber: markInfoNum, Generation: 0}
	structTreeNum := nextNum
	nextNum++

	type cellElemRef struct {
		pageIndex int
		mcid      int
		objNum    int
	}
	tableObjNums := make([]int, len(s.Tables))
	rowObjNums := make([][]int, len(s.Tables))
	var cellElems []cellElemRef

	for ti, tbl := range s.Tables {
		if len(tbl.Rows) == 0 {
			continue
		}
		tableObjNums[ti] = nextNum
		nextNum++
		rowObjNums[ti] = make([]int, len(tbl.Rows))
		for ri, row := range tbl.Rows {
			if len(row.Cells) == 0 {
				continue
			}
			rowObjNums[ti][ri] = nextNum
			nextNum++
			for _, c := range row.Cells {
				for _, mcid := range c.MCIDs {
					cellNum := nextNum
					nextNum++
					cellElems = append(cellElems, cellElemRef{
						pageIndex: c.PageIndex,
						mcid:      mcid,
						objNum:    cellNum,
					})
				}
			}
		}
	}

	cellIdx := 0
	for ti, tbl := range s.Tables {
		for ri, row := range tbl.Rows {
			for _, c := range row.Cells {
				if len(c.MCIDs) == 0 {
					continue
				}
				pageRef := model.Ref{ObjectNumber: pageNums[c.PageIndex], Generation: 0}
				rowRef := model.Ref{ObjectNumber: rowObjNums[ti][ri], Generation: 0}

				var kids model.Array
				var cellRefNums []int
				for _, mcid := range c.MCIDs {
					if cellIdx >= len(cellElems) {
						break
					}
					cellRef := cellElems[cellIdx]
					cellIdx++
					if cellRef.pageIndex != c.PageIndex || cellRef.mcid != mcid {
						continue
					}
					kids = append(kids, model.Dict{
						model.Name("Type"): model.Name("MCR"),
						model.Name("Pg"):   pageRef,
						model.Name("MCID"): model.Integer(int64(mcid)),
					})
					cellRefNums = append(cellRefNums, cellRef.objNum)
				}
				if len(kids) == 0 || len(cellRefNums) == 0 {
					continue
				}

				elemDict := model.Dict{
					model.Name("Type"): model.Name("StructElem"),
					model.Name("S"):    c.Role,
					model.Name("P"):    rowRef,
					model.Name("Pg"):   pageRef,
				}
				if len(kids) == 1 {
					elemDict[model.Name("K")] = kids[0]
				} else {
					elemDict[model.Name("K")] = kids
				}
				// Scope is a table attribute (owner /Table, ISO 32000-2 §14.8.5.7);
				// Alt and Lang are entries of the element itself (§14.9.3, §14.9.2).
				attr := model.Dict{}
				if c.Role == model.Name("TH") && c.Scope != "" {
					attr[model.Name("O")] = model.Name("Table")
					attr[model.Name("Scope")] = model.Name(c.Scope)
				}
				if c.Alt != "" {
					elemDict[model.Name("Alt")] = model.String(c.Alt)
				}
				if c.Lang != "" {
					elemDict[model.Name("Lang")] = model.String(c.Lang)
				}
				if len(attr) > 0 {
					elemDict[model.Name("A")] = attr
				}
				for _, num := range cellRefNums {
					objs[num] = elemDict
				}
			}
		}
	}

	for ti, tbl := range s.Tables {
		if len(tbl.Rows) == 0 || tableObjNums[ti] == 0 {
			continue
		}
		tableNum := tableObjNums[ti]
		var rowRefs model.Array
		for ri, row := range tbl.Rows {
			if len(row.Cells) == 0 || rowObjNums[ti][ri] == 0 {
				continue
			}
			rowNum := rowObjNums[ti][ri]
			var cellRefs model.Array
			for _, ce := range cellElems {
				for _, c := range row.Cells {
					for _, mcid := range c.MCIDs {
						if ce.pageIndex == c.PageIndex && ce.mcid == mcid {
							cellRefs = append(cellRefs, model.Ref{ObjectNumber: ce.objNum, Generation: 0})
							break
						}
					}
				}
			}
			rowDict := model.Dict{
				model.Name("Type"): model.Name("StructElem"),
				model.Name("S"):    model.Name("TR"),
				model.Name("P"):    model.Ref{ObjectNumber: tableNum, Generation: 0},
				model.Name("K"):    cellRefs,
			}
			objs[rowNum] = rowDict
			rowRefs = append(rowRefs, model.Ref{ObjectNumber: rowNum, Generation: 0})
		}
		tableDict := model.Dict{
			model.Name("Type"): model.Name("StructElem"),
			model.Name("S"):    model.Name("Table"),
			model.Name("K"):    rowRefs,
		}
		objs[tableNum] = tableDict
	}

	blockObjNums := make([]int, len(s.Blocks))
	for i := range s.Blocks {
		blockObjNums[i] = nextNum
		nextNum++
	}
	figureObjNums := make([]int, len(s.Figures))
	for i, fig := range s.Figures {
		if fig.PageIndex >= 0 && fig.PageIndex < len(pageNums) {
			figureObjNums[i] = nextNum
			nextNum++
		}
	}
	for i, blk := range s.Blocks {
		if blk.PageIndex < 0 || blk.PageIndex >= len(pageNums) {
			continue
		}
		pageRef := model.Ref{ObjectNumber: pageNums[blk.PageIndex], Generation: 0}
		k := model.Dict{
			model.Name("Type"): model.Name("MCR"),
			model.Name("Pg"):   pageRef,
			model.Name("MCID"): model.Integer(int64(blk.MCID)),
		}
		elemDict := model.Dict{
			model.Name("Type"): model.Name("StructElem"),
			model.Name("S"):    blk.Role,
			model.Name("Pg"):   pageRef,
			model.Name("K"):    k,
		}
		if blk.Lang != "" {
			elemDict[model.Name("Lang")] = model.String(blk.Lang)
		}
		if blk.Alt != "" {
			elemDict[model.Name("Alt")] = model.String(blk.Alt)
		}
		objs[blockObjNums[i]] = elemDict
	}

	for i, fig := range s.Figures {
		if figureObjNums[i] == 0 {
			continue
		}
		pageRef := model.Ref{ObjectNumber: pageNums[fig.PageIndex], Generation: 0}
		mcr := model.Dict{
			model.Name("Type"): model.Name("MCR"),
			model.Name("Pg"):   pageRef,
			model.Name("MCID"): model.Integer(int64(fig.MCID)),
		}
		elemDict := model.Dict{
			model.Name("Type"): model.Name("StructElem"),
			model.Name("S"):    model.Name("Figure"),
			model.Name("Pg"):   pageRef,
			model.Name("K"):    mcr,
		}
		if fig.Alt != "" {
			elemDict[model.Name("Alt")] = model.String(fig.Alt)
		}
		objs[figureObjNums[i]] = elemDict
	}

	listObjNums := make([]int, len(s.Lists))
	itemObjNums := make([][]int, len(s.Lists))
	for i, lst := range s.Lists {
		if lst.PageIndex < 0 || lst.PageIndex >= len(pageNums) || len(lst.Items) == 0 {
			continue
		}
		listObjNums[i] = nextNum
		nextNum++
		itemObjNums[i] = make([]int, len(lst.Items))
		for j := range lst.Items {
			itemObjNums[i][j] = nextNum
			nextNum++
		}
	}
	for i, lst := range s.Lists {
		if listObjNums[i] == 0 || len(lst.Items) == 0 || lst.PageIndex < 0 || lst.PageIndex >= len(pageNums) {
			continue
		}
		pageRef := model.Ref{ObjectNumber: pageNums[lst.PageIndex], Generation: 0}
		var liRefs model.Array
		for j, it := range lst.Items {
			liRef := model.Ref{ObjectNumber: itemObjNums[i][j], Generation: 0}
			// LI holds an optional Lbl (the bullet or number) and an LBody (the text).
			var liKids model.Array
			if it.HasLabel {
				lblNum := nextNum
				nextNum++
				objs[lblNum] = model.Dict{
					model.Name("Type"): model.Name("StructElem"),
					model.Name("S"):    model.Name("Lbl"),
					model.Name("P"):    liRef,
					model.Name("Pg"):   pageRef,
					model.Name("K"):    mcr(pageRef, it.LabelMCID),
				}
				liKids = append(liKids, model.Ref{ObjectNumber: lblNum})
			}
			bodyNum := nextNum
			nextNum++
			objs[bodyNum] = model.Dict{
				model.Name("Type"): model.Name("StructElem"),
				model.Name("S"):    model.Name("LBody"),
				model.Name("P"):    liRef,
				model.Name("Pg"):   pageRef,
				model.Name("K"):    mcr(pageRef, it.MCID),
			}
			liKids = append(liKids, model.Ref{ObjectNumber: bodyNum})
			objs[itemObjNums[i][j]] = model.Dict{
				model.Name("Type"): model.Name("StructElem"),
				model.Name("S"):    model.Name("LI"),
				model.Name("P"):    model.Ref{ObjectNumber: listObjNums[i], Generation: 0},
				model.Name("Pg"):   pageRef,
				model.Name("K"):    liKids,
			}
			liRefs = append(liRefs, liRef)
		}
		listDict := model.Dict{
			model.Name("Type"): model.Name("StructElem"),
			model.Name("S"):    model.Name("L"),
			model.Name("K"):    liRefs,
		}
		if lst.Ordered {
			listDict[model.Name("A")] = model.Dict{model.Name("O"): model.Name("List"), model.Name("ListNumbering"): model.Name("Decimal")}
		}
		objs[listObjNums[i]] = listDict
	}

	// Elements are emitted in the order they were recorded (the reading order).
	objNumOf := func(ref ElemRef) int {
		switch ref.Kind {
		case KindBlock:
			if ref.Index >= 0 && ref.Index < len(blockObjNums) {
				return blockObjNums[ref.Index]
			}
		case KindFigure:
			if ref.Index >= 0 && ref.Index < len(figureObjNums) {
				return figureObjNums[ref.Index]
			}
		case KindTable:
			if ref.Index >= 0 && ref.Index < len(tableObjNums) {
				return tableObjNums[ref.Index]
			}
		case KindList:
			if ref.Index >= 0 && ref.Index < len(listObjNums) {
				return listObjNums[ref.Index]
			}
		}
		return 0
	}
	sectObjNums := make([]int, len(s.Sections))
	for i := range s.Sections {
		sectObjNums[i] = nextNum
		nextNum++
	}
	for i, sec := range s.Sections {
		var kids model.Array
		for _, ref := range sec.Kids {
			if num := objNumOf(ref); num != 0 {
				kids = append(kids, model.Ref{ObjectNumber: num, Generation: 0})
			}
		}
		objs[sectObjNums[i]] = model.Dict{
			model.Name("Type"): model.Name("StructElem"),
			model.Name("S"):    model.Name("Sect"),
			model.Name("K"):    kids,
		}
	}

	var docKids model.Array
	seen := map[int]bool{}
	addKid := func(num int) {
		if num != 0 && !seen[num] {
			seen[num] = true
			docKids = append(docKids, model.Ref{ObjectNumber: num, Generation: 0})
		}
	}
	for _, ref := range s.Order {
		if ref.Kind == KindSection {
			if ref.Index >= 0 && ref.Index < len(sectObjNums) {
				addKid(sectObjNums[ref.Index])
			}
			continue
		}
		addKid(objNumOf(ref))
	}
	// Elements recorded without going through the reading order (older callers
	// that filled the slices directly) follow, grouped by kind.
	for i := range s.Sections {
		addKid(sectObjNums[i])
	}
	inSomeSection := func(ref ElemRef) bool {
		for _, sec := range s.Sections {
			for _, k := range sec.Kids {
				if k == ref {
					return true
				}
			}
		}
		return false
	}
	for ti := range s.Tables {
		if !inSomeSection(ElemRef{KindTable, ti}) && !inSection(s.Sections, 0, 0, ti, 0, 't') {
			addKid(tableObjNums[ti])
		}
	}
	for li := range s.Lists {
		if !inSomeSection(ElemRef{KindList, li}) && !inSection(s.Sections, 0, 0, 0, li, 'l') {
			addKid(listObjNums[li])
		}
	}
	for bi := range s.Blocks {
		if !inSomeSection(ElemRef{KindBlock, bi}) && !inSection(s.Sections, bi, 0, 0, 0, 'b') {
			addKid(blockObjNums[bi])
		}
	}
	for fi := range s.Figures {
		if !inSomeSection(ElemRef{KindFigure, fi}) && !inSection(s.Sections, 0, fi, 0, 0, 'f') {
			addKid(figureObjNums[fi])
		}
	}

	// Link annotations become Link elements with an object reference, and the
	// annotation points back through /StructParent.
	type annotElem struct {
		objNum int
	}
	var annotElems []annotElem
	nextParentKey := 0
	if s.TagAnnotations {
		for _, pageNum := range pageNums {
			pd, ok := objs[pageNum].(model.Dict)
			if !ok {
				continue
			}
			annots, _ := pd[model.Name("Annots")].(model.Array)
			for _, a := range annots {
				ref, ok := a.(model.Ref)
				if !ok {
					continue
				}
				ad, ok := objs[ref.ObjectNumber].(model.Dict)
				if !ok || ad[model.Name("Subtype")] != model.Name("Link") {
					continue
				}
				elemNum := nextNum
				nextNum++
				pageRef := model.Ref{ObjectNumber: pageNum, Generation: 0}
				elem := model.Dict{
					model.Name("Type"): model.Name("StructElem"),
					model.Name("S"):    model.Name("Link"),
					model.Name("Pg"):   pageRef,
					model.Name("K"): model.Array{model.Dict{
						model.Name("Type"): model.Name("OBJR"),
						model.Name("Obj"):  ref,
						model.Name("Pg"):   pageRef,
					}},
				}
				if c, ok := ad[model.Name("Contents")]; ok {
					elem[model.Name("Alt")] = c
				}
				objs[elemNum] = elem
				annotElems = append(annotElems, annotElem{objNum: elemNum})
				addKid(elemNum)
			}
		}
	}

	documentNum := nextNum
	nextNum++
	documentDict := model.Dict{
		model.Name("Type"): model.Name("StructElem"),
		model.Name("S"):    model.Name("Document"),
		model.Name("K"):    docKids,
	}
	objs[documentNum] = documentDict

	structRoot := model.Dict{
		model.Name("Type"): model.Name("StructTreeRoot"),
		model.Name("K"):    model.Array{model.Ref{ObjectNumber: documentNum, Generation: 0}},
	}

	// Parent tree (ISO 32000-2 §14.7.5.4): one array per page with marked
	// content, indexed by MCID and naming the element that holds that marked
	// content, then one entry per tagged annotation.
	pageMCIDs := mcidParents(objs, structTreeNum, nextNum)
	if len(pageMCIDs) > 0 || len(annotElems) > 0 {
		parentTreeNum := nextNum
		nextNum++
		var nums model.Array
		for _, pageNum := range pageNums {
			parents, ok := pageMCIDs[pageNum]
			if !ok {
				continue
			}
			maxMCID := -1
			for mcid := range parents {
				maxMCID = max(maxMCID, mcid)
			}
			arr := make(model.Array, maxMCID+1)
			for i := range arr {
				if elem, ok := parents[i]; ok {
					arr[i] = model.Ref{ObjectNumber: elem, Generation: 0}
				} else {
					arr[i] = model.Null{}
				}
			}
			nums = append(nums, model.Integer(int64(nextParentKey)), arr)
			if pd, ok := objs[pageNum].(model.Dict); ok {
				pd[model.Name("StructParents")] = model.Integer(int64(nextParentKey))
				objs[pageNum] = pd
			}
			nextParentKey++
		}
		for _, ae := range annotElems {
			nums = append(nums, model.Integer(int64(nextParentKey)), model.Ref{ObjectNumber: ae.objNum, Generation: 0})
			elem := objs[ae.objNum].(model.Dict)
			kids := elem[model.Name("K")].(model.Array)
			objr := kids[0].(model.Dict)
			annotRef := objr[model.Name("Obj")].(model.Ref)
			if ad, ok := objs[annotRef.ObjectNumber].(model.Dict); ok {
				ad[model.Name("StructParent")] = model.Integer(int64(nextParentKey))
			}
			nextParentKey++
		}
		objs[parentTreeNum] = model.Dict{model.Name("Nums"): nums}
		structRoot[model.Name("ParentTree")] = model.Ref{ObjectNumber: parentTreeNum, Generation: 0}
		structRoot[model.Name("ParentTreeNextKey")] = model.Integer(int64(nextParentKey))
	}

	if s.UseNamespaces {
		nextNum = s.applyNamespaces(objs, structRoot, structTreeNum, nextNum)
	}

	objs[structTreeNum] = structRoot
	catalogDict[model.Name("StructTreeRoot")] = model.Ref{ObjectNumber: structTreeNum, Generation: 0}
	setStructParents(objs, structTreeNum)

	return nextNum
}

// setStructParents gives every structure element reachable from the structure
// tree root its required /P entry (ISO 32000-2 §14.7.2), the root itself for
// the top-level elements.
func setStructParents(objs map[int]model.Object, structTreeNum int) {
	seen := map[int]bool{}
	var walk func(num int)
	walk = func(num int) {
		if seen[num] {
			return
		}
		seen[num] = true
		d, ok := objs[num].(model.Dict)
		if !ok {
			return
		}
		var kids model.Array
		switch k := d[model.Name("K")].(type) {
		case model.Array:
			kids = k
		case model.Ref:
			kids = model.Array{k}
		}
		for _, k := range kids {
			ref, ok := k.(model.Ref)
			if !ok {
				continue
			}
			if kd, ok := objs[ref.ObjectNumber].(model.Dict); ok && kd[model.Name("Type")] == model.Name("StructElem") {
				kd[model.Name("P")] = model.Ref{ObjectNumber: num, Generation: 0}
				walk(ref.ObjectNumber)
			}
		}
	}
	walk(structTreeNum)
}

// mcr returns a marked-content reference for a sequence on a page.
func mcr(pageRef model.Ref, mcid int) model.Dict {
	return model.Dict{
		model.Name("Type"): model.Name("MCR"),
		model.Name("Pg"):   pageRef,
		model.Name("MCID"): model.Integer(int64(mcid)),
	}
}

// pdf17OnlyTypes are standard structure types of PDF 1.7 that the PDF 2.0
// namespace no longer defines; elements with these roles are placed in the
// PDF 1.7 namespace, whose types ISO 32000-2 Annex L maps onto PDF 2.0 ones.
var pdf17OnlyTypes = map[model.Name]bool{
	"Quote": true, "Code": true, "BlockQuote": true, "Note": true, "Reference": true,
	"BibEntry": true, "Index": true, "TOC": true, "TOCI": true, "Private": true,
}

// applyNamespaces writes the namespace dictionaries and tags every structure
// element created after structTreeNum with /NS.
func (s *Support) applyNamespaces(objs map[int]model.Object, structRoot model.Dict, structTreeNum, nextNum int) int {
	ns2Num := nextNum
	nextNum++
	objs[ns2Num] = model.Dict{model.Name("Type"): model.Name("Namespace"), model.Name("NS"): model.String(NamespacePDF2)}
	ns2 := model.Ref{ObjectNumber: ns2Num, Generation: 0}
	namespaces := model.Array{ns2}
	var ns17 *model.Ref
	for num, obj := range objs {
		d, ok := obj.(model.Dict)
		if !ok || num <= structTreeNum || d[model.Name("Type")] != model.Name("StructElem") {
			continue
		}
		role, _ := d[model.Name("S")].(model.Name)
		if pdf17OnlyTypes[role] {
			if ns17 == nil {
				n := nextNum
				nextNum++
				objs[n] = model.Dict{model.Name("Type"): model.Name("Namespace"), model.Name("NS"): model.String(NamespacePDF17)}
				r := model.Ref{ObjectNumber: n, Generation: 0}
				ns17 = &r
				namespaces = append(namespaces, r)
			}
			d[model.Name("NS")] = *ns17
			continue
		}
		d[model.Name("NS")] = ns2
	}
	structRoot[model.Name("Namespaces")] = namespaces
	return nextNum
}

// mcidParents maps, per page object number, each MCID to the structure element
// (object number in [from, to)) whose /K holds its marked-content reference.
func mcidParents(objs map[int]model.Object, from, to int) map[int]map[int]int {
	out := map[int]map[int]int{}
	record := func(pg model.Object, mcid model.Object, elem int) {
		ref, ok := pg.(model.Ref)
		n, isInt := mcid.(model.Integer)
		if !ok || !isInt {
			return
		}
		if out[ref.ObjectNumber] == nil {
			out[ref.ObjectNumber] = map[int]int{}
		}
		out[ref.ObjectNumber][int(n)] = elem
	}
	for num := from; num < to; num++ {
		d, ok := objs[num].(model.Dict)
		if !ok || d[model.Name("Type")] != model.Name("StructElem") {
			continue
		}
		var kids model.Array
		switch k := d[model.Name("K")].(type) {
		case model.Array:
			kids = k
		case nil:
		default:
			kids = model.Array{k}
		}
		for _, k := range kids {
			switch k := k.(type) {
			case model.Integer:
				record(d[model.Name("Pg")], k, num)
			case model.Dict:
				if k[model.Name("Type")] == model.Name("MCR") {
					pg := k[model.Name("Pg")]
					if pg == nil {
						pg = d[model.Name("Pg")]
					}
					record(pg, k[model.Name("MCID")], num)
				}
			}
		}
	}
	return out
}
