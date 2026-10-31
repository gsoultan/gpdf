package writer

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"math/bits"
	"sort"

	"github.com/gsoultan/gpdf/model"
)

// WriteLinearized writes a linearized (fast web view) PDF to ws, laid out as ISO 32000-1
// Annex F describes: the linearization parameter dictionary, the first-page
// cross-reference table and trailer, the primary hint stream, the first-page section
// (catalog, first page and everything it uses), then one section per remaining page,
// the objects shared between pages, the remaining objects, and the main
// cross-reference table. Objects are renumbered so that each section is consecutive:
// the remaining sections count up from 1 and the first-page section follows them.
//
// The hint stream holds the page offset and shared object hint tables, so a viewer
// can fetch any page with a single byte-range request. Offsets inside it are counted
// as if the hint stream itself were not in the file, as poppler and Acrobat expect.
func (pw *PDFWriter) WriteLinearized(ws WriteSeeker, doc Document) error {
	objs, err := collectObjects(doc)
	if err != nil {
		return err
	}
	plan, err := planLinearization(doc, objs)
	if err != nil {
		return err
	}
	lw := &linearizedWriter{pw: pw, doc: doc, plan: plan}
	data, err := lw.build()
	if err != nil {
		return err
	}
	_, err = ws.Write(data)
	return err
}

// linearizationPlan groups the document's objects (by their original numbers) into the
// sections of a linearized file.
type linearizationPlan struct {
	objects   map[int]model.Object
	catalog   int
	pages     []int   // page object numbers, in document order
	firstPage []int   // first-page section (part 6): catalog, page 1, what page 1 uses
	pageSecs  [][]int // one per remaining page: the page object and its exclusive objects
	shared    []int   // objects used by two or more pages, not in the first-page section
	other     []int   // everything else
	pageUses  []map[int]bool
}

func planLinearization(doc Document, objs []indirectObject) (*linearizationPlan, error) {
	p := &linearizationPlan{objects: make(map[int]model.Object, len(objs))}
	for _, o := range objs {
		p.objects[o.ref.ObjectNumber] = o.obj
	}
	root := doc.Trailer().Root()
	if root == nil {
		return nil, fmt.Errorf("%w: no /Root", ErrInvalidDocumentGraph)
	}
	p.catalog = root.ObjectNumber
	cat, _ := p.objects[p.catalog].(model.Dict)
	if cat == nil {
		return nil, fmt.Errorf("%w: catalog %d is not a dictionary", ErrInvalidDocumentGraph, p.catalog)
	}
	if pagesRef, ok := cat[model.Name("Pages")].(model.Ref); ok {
		p.collectPages(pagesRef.ObjectNumber, map[int]bool{})
	}
	if len(p.pages) == 0 {
		return nil, fmt.Errorf("%w: document has no pages", ErrInvalidDocumentGraph)
	}
	pageSet := make(map[int]bool, len(p.pages))
	for _, n := range p.pages {
		pageSet[n] = true
	}
	useCount := make(map[int]int)
	p.pageUses = make([]map[int]bool, len(p.pages))
	for i, pg := range p.pages {
		used := map[int]bool{}
		p.reach(pg, pageSet, used)
		p.pageUses[i] = used
		for n := range used {
			useCount[n]++
		}
	}
	placed := map[int]bool{}
	place := func(n int) bool {
		if placed[n] {
			return false
		}
		placed[n] = true
		return true
	}
	place(p.catalog)
	place(p.pages[0])
	p.firstPage = []int{p.catalog, p.pages[0]}
	p.firstPage = append(p.firstPage, sortedPlace(p.pageUses[0], place)...)
	for i := 1; i < len(p.pages); i++ {
		sec := []int{p.pages[i]}
		placed[p.pages[i]] = true
		own := map[int]bool{}
		for n := range p.pageUses[i] {
			if useCount[n] == 1 {
				own[n] = true
			}
		}
		sec = append(sec, sortedPlace(own, place)...)
		p.pageSecs = append(p.pageSecs, sec)
	}
	sharedSet := map[int]bool{}
	for n, c := range useCount {
		if c > 1 && !placed[n] {
			sharedSet[n] = true
		}
	}
	p.shared = sortedPlace(sharedSet, place)
	for _, o := range objs {
		if !placed[o.ref.ObjectNumber] {
			p.other = append(p.other, o.ref.ObjectNumber)
		}
	}
	return p, nil
}

func sortedPlace(set map[int]bool, place func(int) bool) []int {
	nums := make([]int, 0, len(set))
	for n := range set {
		nums = append(nums, n)
	}
	sort.Ints(nums)
	out := nums[:0]
	for _, n := range nums {
		if place(n) {
			out = append(out, n)
		}
	}
	return out
}

// collectPages walks the page tree in order.
func (p *linearizationPlan) collectPages(num int, visited map[int]bool) {
	if visited[num] || len(visited) > 1<<16 {
		return
	}
	visited[num] = true
	d, ok := p.objects[num].(model.Dict)
	if !ok {
		return
	}
	kids, hasKids := d[model.Name("Kids")].(model.Array)
	if t, _ := d[model.Name("Type")].(model.Name); t == "Page" || (!hasKids && t != "Pages") {
		p.pages = append(p.pages, num)
		return
	}
	for _, k := range kids {
		if r, ok := k.(model.Ref); ok {
			p.collectPages(r.ObjectNumber, visited)
		}
	}
}

// reach adds to used every object reachable from num, not entering page objects (other
// than num itself) or page tree nodes, so a page's resources are found but not the
// whole document through /Parent.
func (p *linearizationPlan) reach(num int, pageSet map[int]bool, used map[int]bool) {
	var walk func(o model.Object)
	walk = func(o model.Object) {
		switch v := o.(type) {
		case model.Ref:
			n := v.ObjectNumber
			if used[n] || n == num {
				return
			}
			obj, ok := p.objects[n]
			if !ok || pageSet[n] || isPagesNode(obj) {
				return
			}
			used[n] = true
			walk(obj)
		case model.Dict:
			for _, c := range v {
				walk(c)
			}
		case model.Array:
			for _, c := range v {
				walk(c)
			}
		case *model.Stream:
			walk(v.Dict)
		}
	}
	walk(p.objects[num])
}

func isPagesNode(o model.Object) bool {
	d, ok := o.(model.Dict)
	if !ok {
		return false
	}
	t, _ := d[model.Name("Type")].(model.Name)
	return t == "Pages"
}

// remapRefs returns a copy of o with every reference renumbered through m. A reference
// to an object that is not written becomes null (ISO 32000-1 §7.3.10).
func remapRefs(o model.Object, m map[int]int) model.Object {
	switch v := o.(type) {
	case model.Ref:
		if n, ok := m[v.ObjectNumber]; ok {
			return model.Ref{ObjectNumber: n}
		}
		return model.Null{}
	case model.Dict:
		out := make(model.Dict, len(v))
		for k, c := range v {
			out[k] = remapRefs(c, m)
		}
		return out
	case model.Array:
		out := make(model.Array, len(v))
		for i, c := range v {
			out[i] = remapRefs(c, m)
		}
		return out
	case *model.Stream:
		return &model.Stream{Dict: remapRefs(v.Dict, m).(model.Dict), Content: v.Content}
	}
	return o
}

type linearizedWriter struct {
	pw   *PDFWriter
	doc  Document
	plan *linearizationPlan

	renum   map[int]int    // old number -> new number
	body    map[int][]byte // new number -> serialised object
	offsets map[int]int64  // new number -> file offset (set once the layout is known)
}

const linHeader = "%PDF-1.4\n%\xE2\xE3\xCF\xD3\n"

func (lw *linearizedWriter) build() ([]byte, error) {
	p := lw.plan
	// Numbering: parts 7-9 count up from 1, then the linearization dictionary, the
	// first-page section and finally the hint stream.
	lw.renum = make(map[int]int)
	next := 1
	var rest []int
	for _, sec := range p.pageSecs {
		rest = append(rest, sec...)
	}
	rest = append(rest, p.shared...)
	rest = append(rest, p.other...)
	for _, n := range rest {
		lw.renum[n] = next
		next++
	}
	restCount := next - 1
	linNum := next
	next++
	for _, n := range p.firstPage {
		lw.renum[n] = next
		next++
	}
	hintNum := next
	total := hintNum // highest object number

	lw.body = make(map[int][]byte, len(p.objects))
	for old, new := range lw.renum {
		var buf bytes.Buffer
		if err := lw.pw.writeIndirectObject(&buf, model.Ref{ObjectNumber: new}, remapRefs(p.objects[old], lw.renum), nil, nil); err != nil {
			return nil, err
		}
		lw.body[new] = buf.Bytes()
	}
	firstNums := make([]int, len(p.firstPage))
	for i, n := range p.firstPage {
		firstNums[i] = lw.renum[n]
	}
	firstPageNew := lw.renum[p.pages[0]]
	catalogNew := lw.renum[p.catalog]
	var infoNew int
	if r, ok := lw.doc.Trailer().Dict[model.Name("Info")].(model.Ref); ok {
		infoNew = lw.renum[r.ObjectNumber]
	}
	id, _ := lw.doc.Trailer().Dict[model.Name("ID")].(model.Array)

	// Fixed-width pieces, so every offset can be computed before anything is written.
	linDict := func(L, hOff, hLen, E, T int64) []byte {
		return fmt.Appendf(nil, "%d 0 obj\n<< /Linearized 1 /L %010d /H [%010d %010d] /O %d /E %010d /N %d /T %010d >>\nendobj\n",
			linNum, L, hOff, hLen, firstPageNew, E, len(p.pages), T)
	}
	firstTrailer := func(prev int64) []byte {
		var b bytes.Buffer
		fmt.Fprintf(&b, "trailer\n<< /Size %d /Root %d 0 R", total+1, catalogNew)
		if infoNew != 0 {
			fmt.Fprintf(&b, " /Info %d 0 R", infoNew)
		}
		if len(id) == 2 {
			b.WriteString(" /ID [")
			for _, v := range id {
				switch s := v.(type) {
				case model.HexString:
					fmt.Fprintf(&b, "<%s>", hex.EncodeToString([]byte(s)))
				case model.String:
					fmt.Fprintf(&b, "<%s>", hex.EncodeToString([]byte(s)))
				}
			}
			b.WriteString("]")
		}
		fmt.Fprintf(&b, " /Prev %010d >>\nstartxref\n0\n%%%%EOF\n", prev)
		return b.Bytes()
	}
	hintObject := func(content []byte, sharedOff int) []byte {
		var b bytes.Buffer
		fmt.Fprintf(&b, "%d 0 obj\n<< /Length %d /S %d >>\nstream\n", hintNum, len(content), sharedOff)
		b.Write(content)
		b.WriteString("\nendstream\nendobj\n")
		return b.Bytes()
	}

	// Layout. The hint stream's length shifts everything after it, and its tables hold
	// relative values whose width does not depend on that shift, so the layout is
	// computed twice: once with an empty hint stream to learn its size, then for real.
	firstSection := append(append([]int{linNum}, firstNums...), hintNum)
	linOff := int64(len(linHeader))
	firstXrefOff := linOff + int64(len(linDict(0, 0, 0, 0, 0)))
	firstXrefLen := int64(len(fmt.Sprintf("xref\n%d %d\n", linNum, len(firstSection)))) + 20*int64(len(firstSection))
	hintOff := firstXrefOff + firstXrefLen + int64(len(firstTrailer(0)))
	mainXrefHead := fmt.Sprintf("xref\n0 %d\n", restCount+1)
	var offsets map[int]int64
	var firstPageEnd, mainXrefOff int64
	layout := func(hintLen int64) {
		offsets = make(map[int]int64, total)
		off := hintOff + hintLen
		for _, n := range firstNums {
			offsets[n] = off
			off += int64(len(lw.body[n]))
		}
		firstPageEnd = off
		for n := 1; n <= restCount; n++ {
			offsets[n] = off
			off += int64(len(lw.body[n]))
		}
		mainXrefOff = off
		offsets[linNum] = linOff
		offsets[hintNum] = hintOff
		lw.offsets = offsets
	}
	layout(0)
	ht := lw.hintTables(hintOff, 0)
	hintLen := int64(len(hintObject(append(ht.pageOffsets, ht.shared...), len(ht.pageOffsets))))
	layout(hintLen)
	ht = lw.hintTables(hintOff, hintLen)
	hint := hintObject(append(ht.pageOffsets, ht.shared...), len(ht.pageOffsets))
	if int64(len(hint)) != hintLen {
		return nil, fmt.Errorf("linearize: hint stream length changed (%d != %d)", len(hint), hintLen)
	}
	mainEntriesOff := mainXrefOff + int64(len(mainXrefHead))
	mainXrefLen := int64(len(mainXrefHead)) + 20*int64(restCount+1)
	mainTrailer := fmt.Sprintf("trailer\n<< /Size %d /Root %d 0 R >>\nstartxref\n%d\n%%%%EOF\n", total+1, catalogNew, firstXrefOff)
	fileLen := mainXrefOff + mainXrefLen + int64(len(mainTrailer))

	// Emit.
	out := bytes.NewBuffer(make([]byte, 0, fileLen))
	out.WriteString(linHeader)
	out.Write(linDict(fileLen, hintOff, hintLen, firstPageEnd, mainEntriesOff))
	fmt.Fprintf(out, "xref\n%d %d\n", linNum, len(firstSection))
	for _, n := range firstSection {
		fmt.Fprintf(out, "%010d 00000 n \n", offsets[n])
	}
	out.Write(firstTrailer(mainXrefOff))
	out.Write(hint)
	for _, n := range firstNums {
		out.Write(lw.body[n])
	}
	for n := 1; n <= restCount; n++ {
		out.Write(lw.body[n])
	}
	out.WriteString(mainXrefHead)
	out.WriteString("0000000000 65535 f \n")
	for n := 1; n <= restCount; n++ {
		fmt.Fprintf(out, "%010d 00000 n \n", offsets[n])
	}
	out.WriteString(mainTrailer)
	if int64(out.Len()) != fileLen {
		return nil, fmt.Errorf("linearize: file length changed (%d != %d)", out.Len(), fileLen)
	}
	return out.Bytes(), nil
}

// hintTables builds the page offset and shared object hint tables (ISO 32000-1 F.4).
// Absolute offsets are written as if the hint stream were absent (hintLen is subtracted
// from offsets past hintOff). Value arrays are byte-aligned, as qpdf and poppler do.
func (lw *linearizedWriter) hintTables(hintOff, hintLen int64) hintTableSet {
	p := lw.plan
	nPages := len(p.pages)
	rel := func(off int64) int64 {
		if off >= hintOff {
			return off - hintLen
		}
		return off
	}
	secLen := func(nums []int) int64 {
		var l int64
		for _, n := range nums {
			l += int64(len(lw.body[lw.renum[n]]))
		}
		return l
	}
	// Per page: object count, section length, content stream offset/length and the
	// shared objects it refers to.
	nObj := make([]int64, nPages)
	pageLen := make([]int64, nPages)
	contentOff := make([]int64, nPages)
	contentLen := make([]int64, nPages)
	sharedIDs := make([][]int, nPages)
	// Shared object groups: first the first page's own objects (from the page object
	// on, one group each), then the objects of the shared section. Shared object
	// identifiers index this combined list.
	firstGroups := p.firstPage[1:]
	nFirst := len(firstGroups)
	sharedIndex := make(map[int]int, len(p.shared))
	for i, n := range p.shared {
		sharedIndex[n] = nFirst + i
	}
	contentInfo := func(pageNum int, sectionStart int64) (int64, int64) {
		d, _ := p.objects[pageNum].(model.Dict)
		var ref model.Ref
		switch c := d[model.Name("Contents")].(type) {
		case model.Ref:
			ref = c
		case model.Array:
			if len(c) > 0 {
				ref, _ = c[0].(model.Ref)
			}
		}
		new, ok := lw.renum[ref.ObjectNumber]
		if !ok {
			return 0, 0
		}
		return lw.offsets[new] - sectionStart, int64(len(lw.body[new]))
	}
	// First page: from the page object to the end of the first-page section.
	nObj[0] = int64(nFirst)
	pageLen[0] = secLen(firstGroups)
	for i := range nFirst {
		sharedIDs[0] = append(sharedIDs[0], i)
	}
	contentOff[0], contentLen[0] = contentInfo(p.pages[0], lw.offsets[lw.renum[p.pages[0]]])
	for i, sec := range p.pageSecs {
		nObj[i+1] = int64(len(sec))
		pageLen[i+1] = secLen(sec)
		contentOff[i+1], contentLen[i+1] = contentInfo(p.pages[i+1], lw.offsets[lw.renum[sec[0]]])
		for n := range p.pageUses[i+1] {
			if idx, ok := sharedIndex[n]; ok {
				sharedIDs[i+1] = append(sharedIDs[i+1], idx)
			}
		}
		sort.Ints(sharedIDs[i+1])
	}
	minMax := func(v []int64) (least int64, nbits int) {
		least = v[0]
		var most int64
		for _, x := range v {
			least = min(least, x)
			most = max(most, x)
		}
		return least, bitsFor(most - least)
	}
	leastObj, bObj := minMax(nObj)
	leastLen, bLen := minMax(pageLen)
	leastCOff, bCOff := minMax(contentOff)
	leastCLen, bCLen := minMax(contentLen)
	var maxShared, maxID int64
	for _, ids := range sharedIDs {
		maxShared = max(maxShared, int64(len(ids)))
		for _, id := range ids {
			maxID = max(maxID, int64(id))
		}
	}
	bNumShared, bID := bitsFor(maxShared), bitsFor(maxID)

	var w bitWriter
	w.write(uint64(leastObj), 32)
	w.write(uint64(rel(lw.offsets[lw.renum[p.pages[0]]])), 32)
	w.write(uint64(bObj), 16)
	w.write(uint64(leastLen), 32)
	w.write(uint64(bLen), 16)
	w.write(uint64(leastCOff), 32)
	w.write(uint64(bCOff), 16)
	w.write(uint64(leastCLen), 32)
	w.write(uint64(bCLen), 16)
	w.write(uint64(bNumShared), 16)
	w.write(uint64(bID), 16)
	w.write(1, 16) // bits for the fraction numerator
	w.write(1, 16) // denominator
	for _, v := range nObj {
		w.write(uint64(v-leastObj), bObj)
	}
	w.align()
	for _, v := range pageLen {
		w.write(uint64(v-leastLen), bLen)
	}
	w.align()
	for _, ids := range sharedIDs {
		w.write(uint64(len(ids)), bNumShared)
	}
	w.align()
	for _, ids := range sharedIDs {
		for _, id := range ids {
			w.write(uint64(id), bID)
		}
	}
	w.align()
	for _, ids := range sharedIDs {
		for range ids {
			w.write(0, 1) // numerator: the whole page needs the object
		}
	}
	w.align()
	for _, v := range contentOff {
		w.write(uint64(v-leastCOff), bCOff)
	}
	w.align()
	for _, v := range contentLen {
		w.write(uint64(v-leastCLen), bCLen)
	}
	w.align()
	pageTable := w.bytes()

	// Shared object hint table (F.4.2): one object per group, no signatures.
	var s bitWriter
	var firstNum, firstOff int64
	groups := append(append([]int{}, firstGroups...), p.shared...)
	groupLens := make([]int64, len(groups))
	for i, n := range groups {
		groupLens[i] = int64(len(lw.body[lw.renum[n]]))
	}
	leastGroup, bGroup := minMax(groupLens)
	if len(p.shared) > 0 {
		firstNum = int64(lw.renum[p.shared[0]])
		firstOff = rel(lw.offsets[lw.renum[p.shared[0]]])
	}
	s.write(uint64(firstNum), 32)
	s.write(uint64(firstOff), 32)
	s.write(uint64(nFirst), 32)
	s.write(uint64(len(groups)), 32)
	s.write(1, 16) // bits for the number of objects in a group minus one
	s.write(uint64(leastGroup), 32)
	s.write(uint64(bGroup), 16)
	for _, l := range groupLens {
		s.write(uint64(l-leastGroup), bGroup)
	}
	s.align()
	for range groups {
		s.write(0, 1) // no MD5 signature
	}
	s.align()
	for range groups {
		s.write(0, 1) // one object per group
	}
	s.align()
	return hintTableSet{pageOffsets: pageTable, shared: s.bytes()}
}

type hintTableSet struct {
	pageOffsets, shared []byte
}

// bitsFor returns the number of bits needed for v, at least 1.
func bitsFor(v int64) int {
	return max(1, bits.Len64(uint64(v)))
}

// bitWriter packs big-endian bit fields.
type bitWriter struct {
	buf  []byte
	acc  uint64
	nacc int
}

func (w *bitWriter) write(v uint64, n int) {
	for i := n - 1; i >= 0; i-- {
		w.acc = w.acc<<1 | (v>>uint(i))&1
		w.nacc++
		if w.nacc == 8 {
			w.buf = append(w.buf, byte(w.acc))
			w.acc, w.nacc = 0, 0
		}
	}
}

func (w *bitWriter) align() {
	if w.nacc > 0 {
		w.write(0, 8-w.nacc)
	}
}

func (w *bitWriter) bytes() []byte {
	w.align()
	return w.buf
}
