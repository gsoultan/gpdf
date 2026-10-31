// Package pdftest builds small hand-made PDF files for regression tests.
// It writes exactly the bytes it is given, so tests can produce malformed input on purpose.
package pdftest

import (
	"bytes"
	"compress/zlib"
	"fmt"
	"sort"
)

// Builder accumulates raw indirect objects.
type Builder struct {
	Version string
	objs    map[int]obj
}

type obj struct {
	gen  int
	body []byte
}

// New returns a builder for a file with the given header version (e.g. "1.7").
func New(version string) *Builder {
	return &Builder{Version: version, objs: make(map[int]obj)}
}

// Obj adds "num gen obj body endobj".
func (b *Builder) Obj(num, gen int, body string) *Builder {
	b.objs[num] = obj{gen: gen, body: []byte(body)}
	return b
}

// Stream adds a stream object. dict is the inside of the dictionary without /Length;
// a correct /Length is appended unless dict already contains "/Length".
func (b *Builder) Stream(num, gen int, dict string, data []byte) *Builder {
	var buf bytes.Buffer
	buf.WriteString("<< ")
	buf.WriteString(dict)
	if !bytes.Contains([]byte(dict), []byte("/Length")) {
		fmt.Fprintf(&buf, " /Length %d", len(data))
	}
	buf.WriteString(" >>\nstream\n")
	buf.Write(data)
	buf.WriteString("\nendstream")
	b.objs[num] = obj{gen: gen, body: buf.Bytes()}
	return b
}

// Flate zlib-compresses data.
func Flate(data []byte) []byte {
	var buf bytes.Buffer
	w := zlib.NewWriter(&buf)
	w.Write(data)
	w.Close()
	return buf.Bytes()
}

// PNGUp applies the PNG "Up" predictor (filter type 2) to rows of the given width.
func PNGUp(data []byte, columns int) []byte {
	var out []byte
	prev := make([]byte, columns)
	for off := 0; off < len(data); off += columns {
		row := data[off:min(off+columns, len(data))]
		out = append(out, 2)
		for i, c := range row {
			out = append(out, c-prev[i])
		}
		copy(prev, row)
	}
	return out
}

func (b *Builder) body() ([]byte, map[int]int, []int) {
	var buf bytes.Buffer
	fmt.Fprintf(&buf, "%%PDF-%s\n%%\xE2\xE3\xCF\xD3\n", b.Version)
	nums := make([]int, 0, len(b.objs))
	for n := range b.objs {
		nums = append(nums, n)
	}
	sort.Ints(nums)
	offsets := make(map[int]int)
	for _, n := range nums {
		o := b.objs[n]
		offsets[n] = buf.Len()
		fmt.Fprintf(&buf, "%d %d obj\n", n, o.gen)
		buf.Write(o.body)
		buf.WriteString("\nendobj\n")
	}
	return buf.Bytes(), offsets, nums
}

// XRefTable serialises the file with a classic xref table. trailer is the inside of the
// trailer dictionary without /Size.
func (b *Builder) XRefTable(trailer string) []byte {
	body, offsets, nums := b.body()
	buf := bytes.NewBuffer(body)
	size := nums[len(nums)-1] + 1
	xref := buf.Len()
	fmt.Fprintf(buf, "xref\n0 %d\n", size)
	for i := range size {
		if off, ok := offsets[i]; ok {
			fmt.Fprintf(buf, "%010d %05d n \n", off, b.objs[i].gen)
		} else {
			buf.WriteString("0000000000 65535 f \n")
		}
	}
	fmt.Fprintf(buf, "trailer\n<< /Size %d %s >>\nstartxref\n%d\n%%%%EOF\n", size, trailer, xref)
	return buf.Bytes()
}

// XRefStream serialises the file with a cross-reference stream (PDF 1.5) compressed with
// FlateDecode and the PNG Up predictor (/Predictor 12), as most producers write it.
// trailer is extra dictionary content (e.g. "/Root 1 0 R").
func (b *Builder) XRefStream(trailer string) []byte {
	body, offsets, nums := b.body()
	xrefNum := nums[len(nums)-1] + 1
	size := xrefNum + 1
	xrefOff := len(body)
	offsets[xrefNum] = xrefOff
	const cols = 1 + 4 + 2
	var rows []byte
	for i := range size {
		off, ok := offsets[i]
		if !ok {
			rows = append(rows, 0, 0, 0, 0, 0, 0xFF, 0xFF)
			continue
		}
		gen := 0
		if o, ok := b.objs[i]; ok {
			gen = o.gen
		}
		rows = append(rows, 1, byte(off>>24), byte(off>>16), byte(off>>8), byte(off), byte(gen>>8), byte(gen))
	}
	data := Flate(PNGUp(rows, cols))
	buf := bytes.NewBuffer(body)
	fmt.Fprintf(buf, "%d 0 obj\n<< /Type /XRef /Size %d /W [1 4 2] /Filter /FlateDecode /DecodeParms << /Predictor 12 /Columns %d >> /Length %d %s >>\nstream\n",
		xrefNum, size, cols, len(data), trailer)
	buf.Write(data)
	fmt.Fprintf(buf, "\nendstream\nendobj\nstartxref\n%d\n%%%%EOF\n", xrefOff)
	return buf.Bytes()
}

// OnePage adds a catalog (1), page tree (2), page (3) and Helvetica font (5) whose content
// stream is object 4. Callers add object 4.
func (b *Builder) OnePage() *Builder {
	b.Obj(1, 0, "<< /Type /Catalog /Pages 2 0 R >>")
	b.Obj(2, 0, "<< /Type /Pages /Kids [3 0 R] /Count 1 >>")
	b.Obj(3, 0, "<< /Type /Page /Parent 2 0 R /MediaBox [0 0 200 200] /Resources << /Font << /F1 5 0 R >> >> /Contents 4 0 R >>")
	b.Obj(5, 0, "<< /Type /Font /Subtype /Type1 /BaseFont /Helvetica /Encoding /WinAnsiEncoding >>")
	return b
}

// ObjStm adds object stream num holding the given objects (number -> body), FlateDecode'd.
func (b *Builder) ObjStm(num int, objs map[int]string) *Builder {
	nums := make([]int, 0, len(objs))
	for n := range objs {
		nums = append(nums, n)
	}
	sort.Ints(nums)
	var header, data bytes.Buffer
	for _, n := range nums {
		fmt.Fprintf(&header, "%d %d ", n, data.Len())
		data.WriteString(objs[n])
		data.WriteString("\n")
	}
	payload := append(header.Bytes(), data.Bytes()...)
	return b.Stream(num, 0, fmt.Sprintf("/Type /ObjStm /N %d /First %d /Filter /FlateDecode", len(nums), header.Len()), Flate(payload))
}

// Hybrid serialises a hybrid-reference file (ISO 32000-1 §7.5.8.4): a classic xref table
// that omits the objects stored in object streams, plus a cross-reference stream, named
// by the trailer's /XRefStm, with type-2 entries for them. compressed maps an object
// number to its object stream number and index.
func (b *Builder) Hybrid(trailer string, compressed map[int][2]int) []byte {
	body, offsets, nums := b.body()
	maxNum := nums[len(nums)-1]
	for n := range compressed {
		maxNum = max(maxNum, n)
	}
	stmNum := maxNum + 1
	size := stmNum + 1
	var rows []byte
	for i := range size {
		if c, ok := compressed[i]; ok {
			rows = append(rows, 2, byte(c[0]>>8), byte(c[0]), byte(c[1]))
		} else {
			rows = append(rows, 0, 0, 0, 0)
		}
	}
	buf := bytes.NewBuffer(body)
	stmOff := buf.Len()
	fmt.Fprintf(buf, "%d 0 obj\n<< /Type /XRef /Size %d /W [1 2 1] /Length %d >>\nstream\n", stmNum, size, len(rows))
	buf.Write(rows)
	buf.WriteString("\nendstream\nendobj\n")
	xref := buf.Len()
	fmt.Fprintf(buf, "xref\n0 %d\n", size)
	for i := range size {
		if off, ok := offsets[i]; ok {
			fmt.Fprintf(buf, "%010d %05d n \n", off, b.objs[i].gen)
		} else {
			buf.WriteString("0000000000 65535 f \n")
		}
	}
	fmt.Fprintf(buf, "trailer\n<< /Size %d /XRefStm %d %s >>\nstartxref\n%d\n%%%%EOF\n", size, stmOff, trailer, xref)
	return buf.Bytes()
}
