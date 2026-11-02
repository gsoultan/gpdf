// Copyright (c) 2026 Gembit Soultan Shirazi <gembit.soultan@gmail.com>

package reader

import (
	"bytes"
	"fmt"
	"io"
	"strconv"

	"github.com/gsoultan/gpdf/model"
	"github.com/gsoultan/gpdf/syntax"
	"github.com/gsoultan/gpdf/syntax/impl"
	"github.com/gsoultan/gpdf/xref"
)

const (
	xrefScanChunk   = 1 << 20
	xrefScanOverlap = 64
)

var (
	objKeyword     = []byte("obj")
	trailerKeyword = []byte("trailer")
)

// repairResult is what a full-file scan recovers from a file whose cross-reference
// data is unusable.
type repairResult struct {
	entries map[int]syntax.XRefEntry
	trailer model.Dict
}

// rebuildXRefByScan reconstructs the cross-reference table by scanning the whole file for
// "N G obj" headers. Later definitions win, as they do with incremental updates. The trailer
// is taken from the last "trailer" dictionary that has a /Root; repairDocument fills in /Root
// from a catalog object when there is none.
func rebuildXRefByScan(r io.ReaderAt, size int64) (map[int]syntax.XRefEntry, model.Dict, error) {
	res := repairResult{entries: make(map[int]syntax.XRefEntry)}
	var trailerOffsets []int64
	for offset := int64(0); offset < size; {
		buf, err := readScanChunk(r, size, offset)
		if err != nil {
			return nil, nil, err
		}
		limit := len(buf)
		if offset+int64(len(buf)) < size {
			limit -= xrefScanOverlap // the overlap is scanned again with the next chunk
		}
		scanChunk(buf, limit, offset, res.entries, &trailerOffsets)
		if offset+int64(len(buf)) >= size {
			break
		}
		offset += int64(limit)
	}
	if len(res.entries) == 0 {
		return nil, nil, fmt.Errorf("xref repair: no objects found during scan")
	}
	for i := len(trailerOffsets) - 1; i >= 0; i-- {
		p := impl.NewParser(r, size)
		if err := p.SetPosition(trailerOffsets[i]); err != nil {
			continue
		}
		if d, err := p.ParseTrailer(); err == nil {
			if res.trailer == nil {
				res.trailer = d
			}
			if _, ok := d[model.Name("Root")].(model.Ref); ok {
				res.trailer = d
				break
			}
		}
	}
	if res.trailer == nil {
		res.trailer = model.Dict{}
	}
	return res.entries, res.trailer, nil
}

func readScanChunk(r io.ReaderAt, size, offset int64) ([]byte, error) {
	end := min(offset+xrefScanChunk, size)
	buf := make([]byte, end-offset)
	n, err := r.ReadAt(buf, offset)
	if err != nil && err != io.EOF {
		return nil, fmt.Errorf("xref repair: read at %d: %w", offset, err)
	}
	return buf[:n], nil
}

func isRegular(b byte) bool {
	return !isPDFWhitespace(b) && !bytes.ContainsRune([]byte("()<>[]{}/%"), rune(b))
}

func isDigit(b byte) bool { return b >= '0' && b <= '9' }

// scanChunk records "N G obj" headers and "trailer" keywords that start before limit.
func scanChunk(buf []byte, limit int, base int64, entries map[int]syntax.XRefEntry, trailers *[]int64) {
	for pos := 0; pos < limit; {
		i := bytes.Index(buf[pos:], objKeyword)
		t := bytes.Index(buf[pos:min(limit, len(buf))], trailerKeyword)
		if t >= 0 && (i < 0 || t < i) {
			at := pos + t
			if (at == 0 || !isRegular(buf[at-1])) && (at+len(trailerKeyword) >= len(buf) || !isRegular(buf[at+len(trailerKeyword)])) {
				*trailers = append(*trailers, base+int64(at))
			}
			pos = at + len(trailerKeyword)
			continue
		}
		if i < 0 {
			return
		}
		at := pos + i
		pos = at + len(objKeyword)
		if at >= limit {
			return
		}
		if at+len(objKeyword) < len(buf) && isRegular(buf[at+len(objKeyword)]) {
			continue // "objstm" etc.
		}
		start, num, gen, ok := objHeaderBefore(buf, at)
		if !ok {
			continue
		}
		entries[num] = syntax.XRefEntry{Offset: base + int64(start), Generation: gen, InUse: true}
	}
}

// objHeaderBefore parses "N G" backwards from the "obj" keyword at kw.
func objHeaderBefore(buf []byte, kw int) (start, num, gen int, ok bool) {
	j := kw
	skipWS := func() bool {
		n := 0
		for j > 0 && isPDFWhitespace(buf[j-1]) {
			j--
			n++
		}
		return n > 0
	}
	digits := func() (int, bool) {
		end := j
		for j > 0 && isDigit(buf[j-1]) && end-j < 10 {
			j--
		}
		if j == end {
			return 0, false
		}
		v, err := strconv.Atoi(string(buf[j:end]))
		return v, err == nil
	}
	if !skipWS() {
		return 0, 0, 0, false
	}
	gen, ok = digits()
	if !ok || !skipWS() {
		return 0, 0, 0, false
	}
	num, ok = digits()
	if !ok || num <= 0 || gen > 65535 {
		return 0, 0, 0, false
	}
	if j > 0 && isRegular(buf[j-1]) {
		return 0, 0, 0, false
	}
	return j, num, gen, true
}

// repairTrailer completes the trailer of a scan-rebuilt document: a missing /Root (and
// /Info, /ID, /Encrypt) is recovered from cross-reference stream dictionaries or from the
// last catalog object. It returns the object streams found, which registerObjectStream
// handles once decryption is set up.
func (d *pdfDocument) repairTrailer() []int {
	nums := d.xref.ObjectNumbers()
	var xrefDict model.Dict
	var catalogRef *model.Ref
	var objStms []int
	for _, num := range nums {
		ref := model.Ref{ObjectNumber: num, Generation: d.Generation(num)}
		obj, err := d.resolveRaw(ref)
		if err != nil {
			continue
		}
		switch v := obj.(type) {
		case *model.Stream:
			switch t, _ := v.Dict[model.Name("Type")].(model.Name); t {
			case "ObjStm":
				objStms = append(objStms, num)
			case "XRef":
				xrefDict = v.Dict
			}
		case model.Dict:
			if t, _ := v[model.Name("Type")].(model.Name); t == "Catalog" {
				catalogRef = &model.Ref{ObjectNumber: num, Generation: ref.Generation}
			}
		}
	}
	if _, ok := d.trailer.Dict[model.Name("Root")].(model.Ref); ok {
		return objStms
	}
	if d.trailer.Dict == nil {
		d.trailer.Dict = model.Dict{}
	}
	if xrefDict != nil {
		for _, k := range []model.Name{"Root", "Info", "ID", "Encrypt"} {
			if v, ok := xrefDict[k]; ok {
				d.trailer.Dict[k] = v
			}
		}
	}
	if _, ok := d.trailer.Dict[model.Name("Root")].(model.Ref); !ok && catalogRef != nil {
		d.trailer.Dict[model.Name("Root")] = *catalogRef
	}
	return objStms
}

// registerObjectStream adds xref entries for the objects stored in object stream num.
func (d *pdfDocument) registerObjectStream(num int) {
	obj, err := d.Resolve(model.Ref{ObjectNumber: num, Generation: d.Generation(num)})
	if err != nil {
		return
	}
	s, ok := obj.(*model.Stream)
	if !ok {
		return
	}
	n := intFromModelObj(s.Dict[model.Name("N")])
	first := intFromModelObj(s.Dict[model.Name("First")])
	if first <= 0 || first > len(s.Content) {
		return
	}
	fields := bytes.Fields(s.Content[:first])
	for i := 0; i < n && 2*i+1 < len(fields); i++ {
		objNum, err := strconv.Atoi(string(fields[2*i]))
		if err != nil || objNum <= 0 {
			continue
		}
		if _, exists := d.xref.Get(objNum); exists {
			continue
		}
		d.xref.Set(objNum, xref.Entry{InUse: true, Compressed: true, StreamObjectNumber: num, IndexInStream: i})
	}
}
