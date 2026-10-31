package writer

import (
	"fmt"
	"io"
	"maps"
	"sort"

	"github.com/gsoultan/gpdf/model"
)

// resolveFirstPageRef resolves catalog->Pages->Kids to the first page ref and page count.
// Returns (firstPageRef, numPages, error). firstPageRef may be nil if not found.
func resolveFirstPageRef(doc Document) (*model.Ref, int, error) {
	root := doc.Trailer().Root()
	if root == nil {
		return nil, 0, nil
	}
	cat, err := doc.Resolve(*root)
	if err != nil {
		return nil, 0, err
	}
	catDict, ok := cat.(model.Dict)
	if !ok {
		return nil, 0, nil
	}
	pagesRef, ok := catDict[model.Name("Pages")].(model.Ref)
	if !ok {
		return nil, 0, nil
	}
	pagesObj, err := doc.Resolve(pagesRef)
	if err != nil {
		return nil, 0, err
	}
	pagesDict, ok := pagesObj.(model.Dict)
	if !ok {
		return nil, 0, nil
	}
	numPages := 1
	if c, ok := pagesDict[model.Name("Count")].(model.Integer); ok {
		numPages = int(c)
	}
	kids, ok := pagesDict[model.Name("Kids")].(model.Array)
	if !ok || len(kids) == 0 {
		return nil, numPages, nil
	}
	firstRef, ok := kids[0].(model.Ref)
	if !ok {
		return nil, numPages, nil
	}
	return &firstRef, numPages, nil
}

// writeObjectSet writes a set of objects to ws and returns their file offsets.
func (pw *PDFWriter) writeObjectSet(ws WriteSeeker, objNums []int, doc Document) (map[int]int64, error) {
	offsets := make(map[int]int64)
	for _, num := range objNums {
		ref := objectRef(doc, num)
		obj, err := doc.Resolve(ref)
		if err != nil {
			return nil, err
		}
		pos, err := ws.Seek(0, io.SeekCurrent)
		if err != nil {
			return nil, err
		}
		offsets[num] = pos
		if err := pw.writeIndirectObject(ws, ref, obj, nil, nil); err != nil {
			return nil, err
		}
	}
	return offsets, nil
}

// WriteLinearized writes a linearized (fast web view) PDF to ws.
// The writer must be seekable so that the linearization parameter dictionary and the
// first-page trailer can be updated with correct offsets after the rest is written.
// Linearization places first-page objects and a first xref/trailer at the start so that
// a viewer can display page 1 before downloading the full file. No hint stream is
// written, so viewers treat the file as an ordinary one with a first-page-first layout.
func (pw *PDFWriter) WriteLinearized(ws WriteSeeker, doc Document) error {
	allNums := doc.ObjectNumbers()
	if len(allNums) == 0 {
		return fmt.Errorf("document has no objects")
	}
	sort.Ints(allNums)
	firstSet, err := pw.firstPageObjectSet(doc)
	if err != nil {
		return err
	}
	var firstPageNums, restNums []int
	for _, n := range allNums {
		if firstSet[n] {
			firstPageNums = append(firstPageNums, n)
		} else {
			restNums = append(restNums, n)
		}
	}
	firstPageRef, numPages, err := resolveFirstPageRef(doc)
	if err != nil {
		return err
	}
	firstPageObjNum := 0
	if firstPageRef != nil {
		firstPageObjNum = firstPageRef.ObjectNumber
	}
	maxNum := allNums[len(allNums)-1]
	linNum := maxNum + 1 // object 0 is reserved, so the parameter dictionary gets its own number
	size := linNum + 1

	ew := &errWriter{w: ws}
	const header = "%PDF-1.4\n%\xE2\xE3\xCF\xD3\n"
	io.WriteString(ew, header)
	linOffset := int64(len(header))
	// Fixed-width placeholders, patched in place at the end.
	fmt.Fprintf(ew, "%d 0 obj\n<< /Linearized 1 /L %010d /H [%010d %010d] /O %d /E %010d /N %d /T %010d >>\nendobj\n",
		linNum, 0, 0, 0, firstPageObjNum, 0, numPages, 0)
	if ew.err != nil {
		return ew.err
	}
	offsetsFirst, err := pw.writeObjectSet(ws, firstPageNums, doc)
	if err != nil {
		return err
	}
	offsetsFirst[linNum] = linOffset
	firstPageEnd, err := ws.Seek(0, io.SeekCurrent)
	if err != nil {
		return err
	}
	firstXrefStart := firstPageEnd
	gens := generations(doc, allNums)
	firstSection := append(append([]int{}, firstPageNums...), linNum)
	sort.Ints(firstSection)
	if err := pw.writeXRefSubsection(ws, firstSection, offsetsFirst, gens); err != nil {
		return err
	}
	root := model.Ref{ObjectNumber: 1}
	if r := doc.Trailer().Root(); r != nil {
		root = *r
	}
	fmt.Fprintf(ew, "trailer\n<< /Root %d %d R /Size %d /Prev ", root.ObjectNumber, root.Generation, size)
	if ew.err != nil {
		return ew.err
	}
	prevValueOffset, err := ws.Seek(0, io.SeekCurrent)
	if err != nil {
		return err
	}
	fmt.Fprintf(ew, "%010d >>\nstartxref\n%d\n%%%%EOF\n", 0, firstXrefStart)
	if ew.err != nil {
		return ew.err
	}
	offsetsRest, err := pw.writeObjectSet(ws, restNums, doc)
	if err != nil {
		return err
	}
	mainXrefStart, err := ws.Seek(0, io.SeekCurrent)
	if err != nil {
		return err
	}
	allOffsets := make(map[int]int64, len(allNums)+1)
	maps.Copy(allOffsets, offsetsFirst)
	maps.Copy(allOffsets, offsetsRest)
	gens[linNum] = 0
	if err := pw.writeXRefTable(ws, allOffsets, gens, linNum); err != nil {
		return err
	}
	if err := pw.writeTrailer(ws, outputTrailer(doc.Trailer().Dict, size), mainXrefStart, nil); err != nil {
		return err
	}
	totalLen, err := ws.Seek(0, io.SeekCurrent)
	if err != nil {
		return err
	}
	// Patch the parameter dictionary and the first trailer's /Prev.
	if _, err := ws.Seek(linOffset, io.SeekStart); err != nil {
		return err
	}
	fmt.Fprintf(ew, "%d 0 obj\n<< /Linearized 1 /L %010d /H [%010d %010d] /O %d /E %010d /N %d /T %010d >>\nendobj\n",
		linNum, totalLen, firstPageEnd, 0, firstPageObjNum, firstPageEnd, numPages, mainXrefStart)
	if _, err := ws.Seek(prevValueOffset, io.SeekStart); err != nil {
		return err
	}
	fmt.Fprintf(ew, "%010d", mainXrefStart)
	if ew.err != nil {
		return ew.err
	}
	_, err = ws.Seek(totalLen, io.SeekStart)
	return err
}

// firstPageObjectSet returns the set of object numbers required to display the first page
// (catalog, pages tree, first page, and all refs from that page).
func (pw *PDFWriter) firstPageObjectSet(doc Document) (map[int]bool, error) {
	root := doc.Trailer().Root()
	if root == nil {
		return nil, fmt.Errorf("no catalog")
	}
	seen := make(map[int]bool)
	pw.collectRefs(doc, root.ObjectNumber, seen)
	cat, err := doc.Resolve(*root)
	if err != nil {
		return nil, err
	}
	catDict, ok := cat.(model.Dict)
	if !ok {
		return seen, nil
	}
	pagesRef, ok := catDict[model.Name("Pages")].(model.Ref)
	if !ok {
		return seen, nil
	}
	pw.collectRefs(doc, pagesRef.ObjectNumber, seen)
	pagesObj, err := doc.Resolve(pagesRef)
	if err != nil {
		return seen, nil
	}
	pagesDict, ok := pagesObj.(model.Dict)
	if !ok {
		return seen, nil
	}
	kids, ok := pagesDict[model.Name("Kids")].(model.Array)
	if !ok || len(kids) == 0 {
		return seen, nil
	}
	firstRef, ok := kids[0].(model.Ref)
	if !ok {
		return seen, nil
	}
	pw.collectRefs(doc, firstRef.ObjectNumber, seen)
	return seen, nil
}

func (pw *PDFWriter) collectRefs(doc Document, objNum int, seen map[int]bool) {
	if seen[objNum] {
		return
	}
	seen[objNum] = true
	obj, err := doc.Resolve(objectRef(doc, objNum))
	if err != nil || obj == nil {
		return
	}
	switch v := obj.(type) {
	case model.Dict:
		for _, val := range v {
			if ref, ok := val.(model.Ref); ok {
				pw.collectRefs(doc, ref.ObjectNumber, seen)
			}
		}
	case model.Array:
		for _, val := range v {
			if ref, ok := val.(model.Ref); ok {
				pw.collectRefs(doc, ref.ObjectNumber, seen)
			}
		}
	case *model.Stream:
		for _, val := range v.Dict {
			if ref, ok := val.(model.Ref); ok {
				pw.collectRefs(doc, ref.ObjectNumber, seen)
			}
		}
	}
}

func generations(doc Document, nums []int) map[int]int {
	gens := make(map[int]int, len(nums))
	for _, n := range nums {
		gens[n] = objectRef(doc, n).Generation
	}
	return gens
}
