package writer

import (
	"bytes"
	"fmt"
	"hash"
	"io"
	"maps"
	"slices"
	"sort"

	"github.com/gsoultan/gpdf/model"
	"github.com/gsoultan/gpdf/security"
)

// maxObjectsPerStream bounds the number of objects packed into one object stream.
// ISO 32000-1 Annex C recommended at most 100; PDF 2.0 has no limit, but older
// readers still apply the old one.
const maxObjectsPerStream = 100

// xrefEntry is one cross-reference stream entry (ISO 32000-2 §7.5.8.3).
type xrefEntry struct {
	typ    byte  // 0 free, 1 in file, 2 in an object stream
	field2 int64 // offset (type 1) or object stream number (type 2)
	field3 int   // generation (type 1) or index in the object stream (type 2)
}

// offsetWriter counts the bytes written, remembers the first error, and feeds an
// optional hash (for the file identifier).
type offsetWriter struct {
	w   io.Writer
	off int64
	h   hash.Hash
	err error
}

func (o *offsetWriter) Write(p []byte) (int, error) {
	if o.err != nil {
		return 0, o.err
	}
	n, err := o.w.Write(p)
	o.off += int64(n)
	if o.h != nil {
		o.h.Write(p[:n])
	}
	o.err = err
	return n, err
}

// packer writes indirect objects, packing the ones that may live in an object stream
// (ISO 32000-2 §7.5.7: no streams, generation 0, not the encryption dictionary) into
// object streams of at most maxObjectsPerStream objects, and records the
// cross-reference entry of everything written.
type packer struct {
	pw      *PDFWriter
	enc     security.Encryptor
	known   map[int]bool
	compact bool // pack objects into object streams

	entries map[int]xrefEntry
	maxNum  int
	// next is the next free object number, for object and cross-reference streams.
	// When deferred is set it is unknown until finish, and full object streams are
	// held in groups until then, so that callers may keep writing objects under any
	// number of their own.
	next     int
	deferred bool
	groups   []objectGroup

	pendingNums []int
	pendingBody bytes.Buffer
	pendingHead bytes.Buffer
}

// objectGroup is the content of one object stream waiting to be written.
type objectGroup struct {
	nums    []int
	content []byte
	first   int
}

func newPacker(pw *PDFWriter, enc security.Encryptor, known map[int]bool, compact bool, next int) *packer {
	return &packer{pw: pw, enc: enc, known: known, compact: compact, entries: make(map[int]xrefEntry), next: next}
}

func (p *packer) note(num int) {
	if num > p.maxNum {
		p.maxNum = num
	}
}

// write writes obj under ref: into the current object stream when it may be packed,
// otherwise directly to out.
func (p *packer) write(out *offsetWriter, ref model.Ref, obj model.Object) error {
	if p.compact && ref.Generation == 0 && packable(obj) {
		return p.pack(out, ref, obj)
	}
	return p.writeDirect(out, ref, obj, p.enc)
}

// writeDirect writes obj as a top-level indirect object, encrypted with enc.
func (p *packer) writeDirect(out *offsetWriter, ref model.Ref, obj model.Object, enc security.Encryptor) error {
	p.entries[ref.ObjectNumber] = xrefEntry{typ: 1, field2: out.off, field3: ref.Generation}
	p.note(ref.ObjectNumber)
	return p.pw.writeIndirectObject(out, ref, obj, enc, p.known)
}

func packable(obj model.Object) bool {
	_, isStream := obj.(*model.Stream)
	return !isStream
}

func (p *packer) pack(out *offsetWriter, ref model.Ref, obj model.Object) error {
	if len(p.pendingNums) >= maxObjectsPerStream {
		if p.deferred {
			p.groups = append(p.groups, p.takePending())
		} else if err := p.flush(out); err != nil {
			return err
		}
	}
	fmt.Fprintf(&p.pendingHead, "%d %d\n", ref.ObjectNumber, p.pendingBody.Len())
	// Strings inside an object stream are not encrypted on their own: the whole
	// stream is (ISO 32000-2 §7.6.2).
	if err := p.pw.writeObject(&p.pendingBody, obj, nil, nil, p.known); err != nil {
		return err
	}
	p.pendingBody.WriteString("\n")
	p.pendingNums = append(p.pendingNums, ref.ObjectNumber)
	p.note(ref.ObjectNumber)
	return nil
}

// takePending returns the pending objects as a group and resets the buffers.
func (p *packer) takePending() objectGroup {
	content := make([]byte, 0, p.pendingHead.Len()+p.pendingBody.Len())
	content = append(content, p.pendingHead.Bytes()...)
	content = append(content, p.pendingBody.Bytes()...)
	g := objectGroup{nums: slices.Clone(p.pendingNums), content: content, first: p.pendingHead.Len()}
	p.pendingNums = p.pendingNums[:0]
	p.pendingHead.Reset()
	p.pendingBody.Reset()
	return g
}

// flush writes the object streams held so far, if any.
func (p *packer) flush(out *offsetWriter) error {
	if len(p.pendingNums) > 0 {
		p.groups = append(p.groups, p.takePending())
	}
	if p.deferred {
		p.next = max(p.next, p.maxNum+1)
	}
	for _, g := range p.groups {
		num := p.next
		p.next++
		for i, n := range g.nums {
			p.entries[n] = xrefEntry{typ: 2, field2: int64(num), field3: i}
		}
		stm := &model.Stream{
			Dict: model.Dict{
				model.Name("Type"):   model.Name("ObjStm"),
				model.Name("N"):      model.Integer(int64(len(g.nums))),
				model.Name("First"):  model.Integer(int64(g.first)),
				model.Name("Filter"): model.Name("FlateDecode"),
			},
			Content: g.content,
		}
		if err := p.writeDirect(out, model.Ref{ObjectNumber: num}, stm, p.enc); err != nil {
			return err
		}
	}
	p.groups = nil
	return nil
}

// writeXRefStream writes the cross-reference stream for every object written so far
// (and for itself), followed by startxref and %%EOF. trailer supplies /Root, /Info,
// /ID, /Encrypt and /Prev; /Size, /Type, /W, /Index and the filter are set here.
// When nums is nil the section covers 0..Size-1 with free entries for the gaps (a
// full rewrite); otherwise it covers the given object numbers (an incremental
// update). The stream is never encrypted (ISO 32000-2 §7.5.8.2).
func (p *packer) writeXRefStream(out *offsetWriter, trailer model.Dict, nums []int) error {
	if err := p.flush(out); err != nil {
		return err
	}
	num := p.next
	p.next++
	p.entries[num] = xrefEntry{typ: 1, field2: out.off, field3: 0}
	p.note(num)
	size := p.maxNum + 1
	if s, ok := trailer[model.Name("Size")].(model.Integer); ok && int(s) > size {
		size = int(s)
	}
	dict := maps.Clone(trailer)
	dict[model.Name("Type")] = model.Name("XRef")
	dict[model.Name("Size")] = model.Integer(int64(size))
	dict[model.Name("Filter")] = model.Name("FlateDecode")

	if nums == nil {
		nums = make([]int, size)
		for i := range nums {
			nums[i] = i
		}
	} else {
		nums = append(slices.Clone(nums), num)
		sort.Ints(nums)
		var index model.Array
		for i := 0; i < len(nums); {
			j := i
			for j+1 < len(nums) && nums[j+1] == nums[j]+1 {
				j++
			}
			index = append(index, model.Integer(int64(nums[i])), model.Integer(int64(j-i+1)))
			i = j + 1
		}
		dict[model.Name("Index")] = index
	}

	var maxField2 int64
	for _, e := range p.entries {
		maxField2 = max(maxField2, e.field2)
	}
	w2 := 4
	if maxField2 > 0xFFFFFFFF {
		w2 = 8
	}
	dict[model.Name("W")] = model.Array{model.Integer(1), model.Integer(int64(w2)), model.Integer(2)}
	columns := 1 + w2 + 2
	dict[model.Name("DecodeParms")] = model.Dict{
		model.Name("Predictor"): model.Integer(12),
		model.Name("Columns"):   model.Integer(int64(columns)),
	}
	rows := make([]byte, 0, len(nums)*columns)
	for _, n := range nums {
		e, ok := p.entries[n]
		if !ok {
			e = xrefEntry{typ: 0}
			if n == 0 {
				e.field3 = 65535
			}
		}
		rows = append(rows, e.typ)
		for i := w2 - 1; i >= 0; i-- {
			rows = append(rows, byte(e.field2>>(8*i)))
		}
		rows = append(rows, byte(e.field3>>8), byte(e.field3))
	}
	xrefStart := out.off
	stm := &model.Stream{Dict: dict, Content: rows}
	if err := p.pw.writeIndirectObject(out, model.Ref{ObjectNumber: num}, stm, nil, nil); err != nil {
		return err
	}
	fmt.Fprintf(out, "startxref\n%d\n%%%%EOF\n", xrefStart)
	return out.err
}

// finish writes the cross-reference section and trailer in the form the packer was
// configured for.
func (p *packer) finish(out *offsetWriter, trailer model.Dict) error {
	if p.compact {
		return p.writeXRefStream(out, trailer, nil)
	}
	return p.writeXRefTableAndTrailer(out, trailer)
}

// writeXRefTableAndTrailer writes a classic cross-reference table for the objects
// written so far, the trailer, startxref and %%EOF.
func (p *packer) writeXRefTableAndTrailer(out *offsetWriter, trailer model.Dict) error {
	if err := p.flush(out); err != nil {
		return err
	}
	offsets := make(map[int]int64, len(p.entries))
	gens := make(map[int]int, len(p.entries))
	for n, e := range p.entries {
		offsets[n] = e.field2
		gens[n] = e.field3
	}
	size := p.maxNum + 1
	if s, ok := trailer[model.Name("Size")].(model.Integer); ok && int(s) > size {
		size = int(s)
	}
	trailer = maps.Clone(trailer)
	trailer[model.Name("Size")] = model.Integer(int64(size))
	xrefStart := out.off
	if err := p.pw.writeXRefTable(out, offsets, gens, size-1); err != nil {
		return err
	}
	return p.pw.writeTrailer(out, trailer, xrefStart, nil)
}
