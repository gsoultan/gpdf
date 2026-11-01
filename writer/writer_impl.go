// Copyright (c) 2026 Gembit Soultan Shirazi <gembit.soultan@gmail.com>

package writer

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"maps"
	"slices"
	"sort"
	"strconv"

	"github.com/gsoultan/gpdf/model"
	"github.com/gsoultan/gpdf/security"
	"github.com/gsoultan/gpdf/stream"
	"github.com/gsoultan/gpdf/stream/defaults"
	"github.com/gsoultan/gpdf/stream/predictor"
)

const (
	headerPDF20 = "%PDF-2.0\n%\xE2\xE3\xCF\xD3\n"
	// RC4 encryption is deprecated in PDF 2.0, so RC4-encrypted output is labelled 1.7.
	headerPDF17 = "%PDF-1.7\n%\xE2\xE3\xCF\xD3\n"
	headerPDF14 = "%PDF-1.4\n%\xE2\xE3\xCF\xD3\n"
)

// PDFWriter implements Writer for PDF 2.0 output.
type PDFWriter struct {
	filters       stream.FilterRegistry
	randomSource  io.Reader
	validateGraph bool
	objectStreams bool
}

// NewPDFWriter returns a PDF writer with default stream filters registered.
func NewPDFWriter() *PDFWriter {
	return NewPDFWriterWithOptions(WriterOptions{Filters: defaults.NewStandardRegistry()})
}

// NewPDFWriterWithFilters returns a PDF writer using the given filter registry.
func NewPDFWriterWithFilters(filters stream.FilterRegistry) *PDFWriter {
	return NewPDFWriterWithOptions(WriterOptions{Filters: filters})
}

// NewPDFWriterWithOptions returns a PDF writer using custom options.
func NewPDFWriterWithOptions(options WriterOptions) *PDFWriter {
	opts := normalizeWriterOptions(options)
	filters := opts.Filters
	if filters == nil {
		filters = defaults.NewStandardRegistry()
	}
	return &PDFWriter{
		filters:       filters,
		randomSource:  opts.RandomSource,
		validateGraph: *opts.ValidateGraph,
		objectStreams: *opts.ObjectStreams,
	}
}

// compactFor reports whether the output for header uses object streams and a
// cross-reference stream: when enabled and the version is at least 1.5.
func (pw *PDFWriter) compactFor(header string) bool {
	return pw.objectStreams && header != headerPDF14
}

// generationProvider is implemented by documents read from a file, whose objects can have
// a generation number other than 0.
type generationProvider interface {
	Generation(objectNumber int) int
}

// objectRef returns the reference of object num in doc, with its real generation.
func objectRef(doc Document, num int) model.Ref {
	if gp, ok := doc.(generationProvider); ok {
		return model.Ref{ObjectNumber: num, Generation: gp.Generation(num)}
	}
	return model.Ref{ObjectNumber: num}
}

// indirectObject is one object to be written.
type indirectObject struct {
	ref model.Ref
	obj model.Object
}

// collectObjects resolves the objects to write, sorted by number. It leaves out objects
// that only describe the source file's own layout: the /Encrypt dictionary (the output is
// re-encrypted or written in the clear), cross-reference streams and object streams (their
// contents are written as ordinary objects).
func collectObjects(doc Document) ([]indirectObject, error) {
	objNums := doc.ObjectNumbers()
	if len(objNums) == 0 {
		return nil, fmt.Errorf("%w: document has no objects", ErrInvalidDocumentGraph)
	}
	sort.Ints(objNums)
	skip := excludedObjectNumbers(doc)
	out := make([]indirectObject, 0, len(objNums))
	for _, num := range objNums {
		if skip[num] {
			continue
		}
		ref := objectRef(doc, num)
		obj, err := doc.Resolve(ref)
		if err != nil {
			return nil, err
		}
		if isFileStructureObject(obj) {
			continue
		}
		out = append(out, indirectObject{ref: ref, obj: obj})
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("%w: document has no objects", ErrInvalidDocumentGraph)
	}
	return out, nil
}

func excludedObjectNumbers(doc Document) map[int]bool {
	skip := make(map[int]bool)
	if enc, ok := doc.Trailer().Dict[model.Name("Encrypt")].(model.Ref); ok {
		skip[enc.ObjectNumber] = true
	}
	return skip
}

func isFileStructureObject(obj model.Object) bool {
	s, ok := obj.(*model.Stream)
	if !ok {
		return false
	}
	t, _ := s.Dict[model.Name("Type")].(model.Name)
	return t == "XRef" || t == "ObjStm"
}

// outputTrailer builds a clean trailer for a full rewrite: only /Root, /Info and /ID are
// carried over. Keys that describe the source file (/Prev, /XRefStm, /Encrypt, and the
// /W, /Index, /Filter... of a cross-reference stream dictionary) must not be copied.
func outputTrailer(d model.Dict, size int) model.Dict {
	out := model.Dict{model.Name("Size"): model.Integer(int64(size))}
	for _, k := range []model.Name{"Root", "Info", "ID"} {
		if v, ok := d[k]; ok {
			out[k] = v
		}
	}
	return out
}

// fileID returns a two-element /ID array with the same value twice.
func fileID(id []byte) model.Array {
	return model.Array{model.HexString(id), model.HexString(id)}
}

// Write writes the document to w. Non-stream objects are packed into object streams
// and the cross-reference is a stream unless WriterOptions.ObjectStreams is false or
// the document's version is below 1.5, in which case a classic table is written.
func (pw *PDFWriter) Write(w io.Writer, doc Document) error {
	if pw.validateGraph {
		if err := validateDocumentGraph(doc); err != nil {
			return err
		}
	}
	objs, err := collectObjects(doc)
	if err != nil {
		return err
	}
	known := knownNumbers(objs)
	maxNum := objs[len(objs)-1].ref.ObjectNumber
	trailer := outputTrailer(doc.Trailer().Dict, maxNum+1)
	header := headerFor(doc)
	var body bytes.Buffer
	out := &offsetWriter{w: &body}
	if _, ok := trailer[model.Name("ID")].(model.Array); !ok {
		out.h = md5.New()
	}
	io.WriteString(out, header)
	pk := newPacker(pw, nil, known, pw.compactFor(header), maxNum+1)
	for _, o := range objs {
		if err := pk.write(out, o.ref, o.obj); err != nil {
			return err
		}
	}
	if err := pk.flush(out); err != nil {
		return err
	}
	if out.h != nil {
		trailer[model.Name("ID")] = fileID(out.h.Sum(nil))
	}
	if err := pk.finish(out, trailer); err != nil {
		return err
	}
	_, err = w.Write(body.Bytes())
	return err
}

// WriteWithPassword writes the document encrypted with the Standard security handler
// using 128-bit RC4 (V 2, R 3), which every reader supports. The output is labelled
// PDF 1.7 because PDF 2.0 deprecates RC4; use WriteWithAES256Password for PDF 2.0.
func (pw *PDFWriter) WriteWithPassword(w io.Writer, doc Document, userPassword, ownerPassword string) error {
	return pw.writeEncrypted(w, doc, headerPDF17, func(id []byte) (model.Dict, security.Encryptor, error) {
		return security.BuildEncryptDictForWrite(userPassword, ownerPassword, id, -4)
	})
}

// WriteWithAES256Password writes the document encrypted with AES-256 using the
// ISO 32000-2 Standard security handler (V 5, R 6).
func (pw *PDFWriter) WriteWithAES256Password(w io.Writer, doc Document, userPassword, ownerPassword string) error {
	return pw.writeEncrypted(w, doc, headerPDF20, func(id []byte) (model.Dict, security.Encryptor, error) {
		return security.BuildAES256EncryptDictForWriteWithRandom(userPassword, ownerPassword, id, -4, pw.randomSource)
	})
}

func (pw *PDFWriter) writeEncrypted(w io.Writer, doc Document, header string, buildEncrypt func(id []byte) (model.Dict, security.Encryptor, error)) error {
	if pw.validateGraph {
		if err := validateDocumentGraph(doc); err != nil {
			return err
		}
	}
	objs, err := collectObjects(doc)
	if err != nil {
		return err
	}
	known := knownNumbers(objs)
	maxNum := objs[len(objs)-1].ref.ObjectNumber
	encryptNum := maxNum + 1
	known[encryptNum] = true
	id := make([]byte, 16)
	if _, err := io.ReadFull(pw.randomSource, id); err != nil {
		return err
	}
	encDict, enc, err := buildEncrypt(id)
	if err != nil {
		return err
	}
	trailer := outputTrailer(doc.Trailer().Dict, encryptNum+1)
	trailer[model.Name("Encrypt")] = model.Ref{ObjectNumber: encryptNum}
	trailer[model.Name("ID")] = fileID(id)

	var body bytes.Buffer
	out := &offsetWriter{w: &body}
	io.WriteString(out, header)
	pk := newPacker(pw, enc, known, pw.compactFor(header), encryptNum+1)
	for _, o := range objs {
		if err := pk.write(out, o.ref, o.obj); err != nil {
			return err
		}
	}
	// The encryption dictionary itself is never encrypted, nor stored in an object stream.
	if err := pk.writeDirect(out, model.Ref{ObjectNumber: encryptNum}, encDict, nil); err != nil {
		return err
	}
	if err := pk.finish(out, trailer); err != nil {
		return err
	}
	_, err = w.Write(body.Bytes())
	return err
}

func knownNumbers(objs []indirectObject) map[int]bool {
	known := make(map[int]bool, len(objs))
	for _, o := range objs {
		known[o.ref.ObjectNumber] = true
	}
	return known
}

func (pw *PDFWriter) writeTrailer(w io.Writer, trailer model.Dict, xrefStart int64, known map[int]bool) error {
	ew := &errWriter{w: w}
	io.WriteString(ew, "trailer\n")
	if err := pw.writeObject(ew, trailer, nil, nil, known); err != nil {
		return err
	}
	fmt.Fprintf(ew, "\nstartxref\n%d\n%%%%EOF\n", xrefStart)
	return ew.err
}

// errWriter remembers the first write error so long write sequences can be checked once.
type errWriter struct {
	w   io.Writer
	err error
}

func (e *errWriter) Write(p []byte) (int, error) {
	if e.err != nil {
		return 0, e.err
	}
	n, err := e.w.Write(p)
	e.err = err
	return n, err
}

func asErrWriter(w io.Writer) *errWriter {
	if ew, ok := w.(*errWriter); ok {
		return ew
	}
	return &errWriter{w: w}
}

// writeXRefTable writes a classic cross-reference table for objects 0..maxNum.
func (pw *PDFWriter) writeXRefTable(w io.Writer, offsets map[int]int64, gens map[int]int, maxNum int) error {
	ew := asErrWriter(w)
	fmt.Fprintf(ew, "xref\n0 %d\n", maxNum+1)
	for i := 0; i <= maxNum; i++ {
		if off, ok := offsets[i]; ok && i > 0 {
			fmt.Fprintf(ew, "%010d %05d n \n", off, gens[i])
		} else {
			fmt.Fprintf(ew, "%010d %05d f \n", 0, 65535)
		}
	}
	return ew.err
}

// WriteIncremental appends an incremental update (new objects + xref + trailer with /Prev + startxref + %%EOF) to w.
// The cross-reference section has the form of the base file's: a stream when the base
// trailer is a cross-reference stream dictionary (/Type /XRef), a table otherwise
// (ISO 32000-2 §7.5.8.1). Objects are written directly, not in object streams, so
// that a signature's /ByteRange and /Contents are found at their offsets.
func (pw *PDFWriter) WriteIncremental(w io.Writer, appendOffset int64, baseStartXRef int64, doc Document) error {
	if pw.validateGraph {
		if err := validateIncrementalDocumentGraph(doc); err != nil {
			return err
		}
	}
	objNums := doc.ObjectNumbers()
	if len(objNums) == 0 {
		return fmt.Errorf("%w: incremental update requires at least one object", ErrInvalidDocumentGraph)
	}
	sort.Ints(objNums)
	offsets := make(map[int]int64)
	gens := make(map[int]int)
	var body bytes.Buffer
	for _, num := range objNums {
		ref := objectRef(doc, num)
		obj, err := doc.Resolve(ref)
		if err != nil {
			return err
		}
		offsets[num] = appendOffset + int64(body.Len())
		gens[num] = ref.Generation
		// References to objects of the base file are valid here, so no dangling-ref filter.
		if err := pw.writeIndirectObject(&body, ref, obj, nil, nil); err != nil {
			return err
		}
	}
	if _, err := w.Write(body.Bytes()); err != nil {
		return err
	}
	xrefStart := appendOffset + int64(body.Len())
	trailer := copyTrailerWithPrev(doc.Trailer().Dict, baseStartXRef, objNums)
	if t, _ := doc.Trailer().Dict[model.Name("Type")].(model.Name); t == "XRef" {
		size, _ := trailer[model.Name("Size")].(model.Integer)
		pk := newPacker(pw, nil, nil, false, int(size))
		for _, num := range objNums {
			pk.entries[num] = xrefEntry{typ: 1, field2: offsets[num], field3: gens[num]}
			pk.note(num)
		}
		out := &offsetWriter{w: w, off: xrefStart}
		return pk.writeXRefStream(out, trailer, objNums)
	}
	if err := pw.writeXRefSubsection(w, objNums, offsets, gens); err != nil {
		return err
	}
	return pw.writeTrailer(w, trailer, xrefStart, nil)
}

// copyTrailerWithPrev builds the trailer of an incremental update.
func copyTrailerWithPrev(d model.Dict, prev int64, objNums []int) model.Dict {
	out := make(model.Dict, len(d)+2)
	for k, v := range d {
		switch k {
		case "Prev", "XRefStm", "Type", "W", "Index", "Filter", "DecodeParms", "Length":
			continue // belong to the previous section (or to an xref stream dictionary)
		}
		out[k] = v
	}
	out[model.Name("Prev")] = model.Integer(prev)
	maxNum := slices.Max(objNums)
	size := maxNum + 1
	if s, ok := d[model.Name("Size")].(model.Integer); ok && int(s) > size {
		size = int(s)
	}
	out[model.Name("Size")] = model.Integer(int64(size))
	return out
}

// writeXRefSubsection writes one subsection per run of consecutive object numbers, so
// objects that are not part of the update keep their entries from the base file.
func (pw *PDFWriter) writeXRefSubsection(w io.Writer, objNums []int, offsets map[int]int64, gens map[int]int) error {
	if len(objNums) == 0 {
		return nil
	}
	ew := asErrWriter(w)
	io.WriteString(ew, "xref\n")
	for i := 0; i < len(objNums); {
		j := i
		for j+1 < len(objNums) && objNums[j+1] == objNums[j]+1 {
			j++
		}
		fmt.Fprintf(ew, "%d %d\n", objNums[i], j-i+1)
		for k := i; k <= j; k++ {
			fmt.Fprintf(ew, "%010d %05d n \n", offsets[objNums[k]], gens[objNums[k]])
		}
		i = j + 1
	}
	return ew.err
}

func (pw *PDFWriter) writeIndirectObject(w io.Writer, ref model.Ref, obj model.Object, enc security.Encryptor, known map[int]bool) error {
	ew := asErrWriter(w)
	fmt.Fprintf(ew, "%d %d obj\n", ref.ObjectNumber, ref.Generation)
	if err := pw.writeObject(ew, obj, &ref, enc, known); err != nil {
		return err
	}
	io.WriteString(ew, "\nendobj\n")
	return ew.err
}

// writeObject serialises obj. When enc is non-nil, strings and streams are encrypted with
// the key of the enclosing indirect object ref. When known is non-nil, references to
// objects that are not written are replaced by null (ISO 32000-1 §7.3.10).
func (pw *PDFWriter) writeObject(w io.Writer, obj model.Object, ref *model.Ref, enc security.Encryptor, known map[int]bool) error {
	ew := asErrWriter(w)
	switch v := obj.(type) {
	case nil, model.Null:
		io.WriteString(ew, "null")
	case model.Boolean:
		if v {
			io.WriteString(ew, "true")
		} else {
			io.WriteString(ew, "false")
		}
	case model.Integer:
		io.WriteString(ew, strconv.FormatInt(int64(v), 10))
	case model.Real:
		io.WriteString(ew, FormatReal(float64(v)))
	case model.String:
		if enc != nil && ref != nil {
			cipher, err := enc.EncryptString(*ref, []byte(v))
			if err != nil {
				return err
			}
			pw.writeHexString(ew, cipher)
		} else {
			pw.writeString(ew, string(v))
		}
	case model.HexString:
		data := []byte(v)
		if enc != nil && ref != nil {
			cipher, err := enc.EncryptString(*ref, data)
			if err != nil {
				return err
			}
			data = cipher
		}
		pw.writeHexString(ew, data)
	case model.Name:
		io.WriteString(ew, "/"+escapeName(string(v)))
	case model.Ref:
		if known != nil && !known[v.ObjectNumber] {
			io.WriteString(ew, "null")
		} else {
			fmt.Fprintf(ew, "%d %d R", v.ObjectNumber, v.Generation)
		}
	case model.Array:
		io.WriteString(ew, "[")
		for i, e := range v {
			if i > 0 {
				io.WriteString(ew, " ")
			}
			if err := pw.writeObject(ew, e, ref, enc, known); err != nil {
				return err
			}
		}
		io.WriteString(ew, "]")
	case model.Dict:
		return pw.writeDictObj(ew, v, ref, enc, known)
	case *model.Stream:
		return pw.writeStream(ew, v, ref, enc, known)
	default:
		return fmt.Errorf("unknown object type %T", obj)
	}
	return ew.err
}

func (pw *PDFWriter) writeDictObj(ew *errWriter, d model.Dict, ref *model.Ref, enc security.Encryptor, known map[int]bool) error {
	io.WriteString(ew, "<<")
	// Sorted keys make the output reproducible.
	for _, k := range slices.Sorted(maps.Keys(d)) {
		io.WriteString(ew, "\n/")
		io.WriteString(ew, escapeName(string(k)))
		io.WriteString(ew, " ")
		if err := pw.writeObject(ew, d[k], ref, enc, known); err != nil {
			return err
		}
	}
	io.WriteString(ew, "\n>>")
	return ew.err
}

func (pw *PDFWriter) writeStream(ew *errWriter, v *model.Stream, ref *model.Ref, enc security.Encryptor, known map[int]bool) error {
	dict, content := pw.encodeStream(v)
	if enc != nil && ref != nil {
		encrypted, err := enc.EncryptStream(*ref, content)
		if err != nil {
			return err
		}
		content = encrypted
	}
	// /Length is the length of the bytes actually written (after encryption).
	dict[model.Name("Length")] = model.Integer(int64(len(content)))
	if err := pw.writeDictObj(ew, dict, ref, enc, known); err != nil {
		return err
	}
	io.WriteString(ew, "\nstream\n")
	ew.Write(content)
	io.WriteString(ew, "\nendstream")
	return ew.err
}

// writeDict writes d without encryption or reference filtering.
func (pw *PDFWriter) writeDict(w io.Writer, d model.Dict) error {
	return pw.writeDictObj(asErrWriter(w), d, nil, nil, nil)
}

// filterNames returns the filter names of a stream's /Filter (a name or an array of names).
func filterNames(obj model.Object) ([]string, bool) {
	switch v := obj.(type) {
	case nil:
		return nil, true
	case model.Name:
		return []string{string(v)}, true
	case model.Array:
		names := make([]string, len(v))
		for i, o := range v {
			n, ok := o.(model.Name)
			if !ok {
				return nil, false
			}
			names[i] = string(n)
		}
		return names, true
	}
	return nil, false
}

// encodeStream re-encodes the (decoded) stream content with the stream's filter chain and
// returns the dictionary to write and the encoded bytes. Filters are applied last to first,
// since a reader decodes them first to last. Flate and LZW apply the PNG or TIFF predictor
// their /DecodeParms ask for (image samples and cross-reference rows compress far better
// with one); other filters keep their parameters untouched. If any filter cannot
// encode, the content is written unfiltered and /Filter and /DecodeParms are removed, so the
// dictionary never claims an encoding the bytes do not have. A filter chain with a name the
// registry does not know is passed through untouched (see reader: such streams stay encoded).
func (pw *PDFWriter) encodeStream(v *model.Stream) (model.Dict, []byte) {
	dict := maps.Clone(v.Dict)
	if dict == nil {
		dict = make(model.Dict)
	}
	raw := func() (model.Dict, []byte) {
		delete(dict, model.Name("Filter"))
		delete(dict, model.Name("DecodeParms"))
		return dict, v.Content
	}
	names, ok := filterNames(dict[model.Name("Filter")])
	if !ok {
		return raw()
	}
	if len(names) == 0 {
		delete(dict, model.Name("DecodeParms"))
		return dict, v.Content
	}
	if pw.filters == nil {
		return raw()
	}
	parms := make([]model.Object, len(names))
	switch p := dict[model.Name("DecodeParms")].(type) {
	case model.Dict:
		parms[0] = p
	case model.Array:
		for i := range min(len(p), len(parms)) {
			parms[i] = p[i]
		}
	}
	data := v.Content
	for i := len(names) - 1; i >= 0; i-- {
		f := pw.filters.Get(names[i])
		if f == nil {
			// No codec for this filter: the reader left such a stream encoded, so the
			// bytes and the dictionary are copied as they are.
			return maps.Clone(v.Dict), v.Content
		}
		if names[i] == "FlateDecode" || names[i] == "LZWDecode" {
			// The predictor the parameters ask for is applied before compression;
			// if it cannot be, the parameters are dropped so they never describe
			// an encoding the bytes do not have.
			if pd, ok := parms[i].(model.Dict); ok {
				predicted, err := predictor.Encode(data, predictorParams(pd))
				if err != nil {
					parms[i] = nil
				} else {
					data = predicted
				}
			}
		}
		var buf bytes.Buffer
		if err := f.Encode(&buf, bytes.NewReader(data), names[i]); err != nil {
			return raw()
		}
		data = buf.Bytes()
	}
	setDecodeParms(dict, parms)
	return dict, data
}

// predictorParams reads the predictor entries of a /DecodeParms dictionary.
func predictorParams(d model.Dict) predictor.Params {
	intOf := func(k string) int {
		switch v := d[model.Name(k)].(type) {
		case model.Integer:
			return int(v)
		case model.Real:
			return int(v)
		}
		return 0
	}
	return predictor.Params{
		Predictor:        intOf("Predictor"),
		Colors:           intOf("Colors"),
		BitsPerComponent: intOf("BitsPerComponent"),
		Columns:          intOf("Columns"),
	}
}

func setDecodeParms(dict model.Dict, parms []model.Object) {
	any := false
	for i, p := range parms {
		if _, ok := p.(model.Dict); ok {
			any = true
		} else {
			parms[i] = model.Null{}
		}
	}
	switch {
	case !any:
		delete(dict, model.Name("DecodeParms"))
	case len(parms) == 1:
		dict[model.Name("DecodeParms")] = parms[0]
	default:
		dict[model.Name("DecodeParms")] = model.Array(parms)
	}
}

// encodeStreamContent applies stream filters to produce encoded content.
func (pw *PDFWriter) encodeStreamContent(v *model.Stream) []byte {
	_, data := pw.encodeStream(v)
	return data
}

// FormatReal formats a PDF real number in plain decimal notation (see model.FormatReal).
func FormatReal(f float64) string { return model.FormatReal(f) }

func (pw *PDFWriter) writeHexString(w io.Writer, b []byte) {
	io.WriteString(w, "<")
	io.WriteString(w, hex.EncodeToString(b))
	io.WriteString(w, ">")
}

func (pw *PDFWriter) writeString(w io.Writer, s string) {
	var b bytes.Buffer
	b.WriteByte('(')
	for i := range len(s) {
		c := s[i]
		switch c {
		case '\\', '(', ')':
			b.WriteByte('\\')
			b.WriteByte(c)
		case '\n':
			b.WriteString("\\n")
		case '\r':
			b.WriteString("\\r")
		case '\t':
			b.WriteString("\\t")
		case '\b':
			b.WriteString("\\b")
		case '\f':
			b.WriteString("\\f")
		default:
			if c < 0x20 || c > 0x7e {
				fmt.Fprintf(&b, "\\%03o", c)
				continue
			}
			b.WriteByte(c)
		}
	}
	b.WriteByte(')')
	w.Write(b.Bytes())
}

func escapeName(s string) string {
	var b bytes.Buffer
	for i := range len(s) {
		c := s[i]
		if c <= ' ' || c >= 127 || c == '#' || c == '/' || c == '(' || c == ')' || c == '<' || c == '>' || c == '[' || c == ']' || c == '{' || c == '}' || c == '%' {
			fmt.Fprintf(&b, "#%02x", c)
		} else {
			b.WriteByte(c)
		}
	}
	return b.String()
}
