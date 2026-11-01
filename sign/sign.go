package sign

import (
	"bytes"
	"crypto"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/gsoultan/gpdf/model"
	"github.com/gsoultan/gpdf/reader"
	"github.com/gsoultan/gpdf/writer"
)

// ErrNotImplemented is kept for callers that checked for it; Sign now signs.
//
// Deprecated: Sign no longer returns it.
var ErrNotImplemented = errors.New("sign: digital signatures are not implemented")

// ErrEncrypted is returned when the input is encrypted: an incremental update would
// have to encrypt the new objects with the file key, which is not supported.
var ErrEncrypted = errors.New("sign: encrypted documents cannot be signed")

// ErrCertified is returned when a document carries a certification signature that
// forbids further changes (DocMDP /P 1).
var ErrCertified = errors.New("sign: document is certified and allows no changes")

// Sign reads a PDF from r (size bytes) and writes it to out with a signature appended
// as an incremental update (ISO 32000-2 §12.8): the original bytes are untouched,
// so earlier signatures stay valid. The signature covers every byte except the
// /Contents placeholder; s produces the CMS container (see CMSSigner).
func Sign(r io.ReaderAt, size int64, out io.Writer, s Signer, opts SignOptions) error {
	if s == nil {
		return errors.New("sign: no Signer")
	}
	doc, err := reader.NewPDFReader().ReadDocument(r, size)
	if err != nil {
		return fmt.Errorf("sign: read: %w", err)
	}
	if doc.Trailer().Dict[model.Name("Encrypt")] != nil {
		return ErrEncrypted
	}
	if level, _ := docMDPLevel(doc); level == 1 {
		return ErrCertified
	}
	if opts.Hash == 0 {
		opts.Hash = crypto.SHA256
	}
	if opts.SubFilter == "" {
		opts.SubFilter = SubFilterCAdES
	}
	if opts.SigningTime.IsZero() {
		opts.SigningTime = time.Now()
	}
	if opts.ContentsSize <= 0 {
		opts.ContentsSize = 16384
	}
	if cs, ok := s.(*CMSSigner); ok && cs.Cert != nil && opts.signerName == "" {
		opts.signerName = subjectName(cs.Cert.Subject)
	}
	update, err := buildSignatureUpdate(doc, opts)
	if err != nil {
		return err
	}
	// Serialize the update with placeholders, then patch /ByteRange and /Contents in place.
	var buf bytes.Buffer
	pw := writer.NewPDFWriterWithOptions(writer.WriterOptions{ValidateGraph: new(false)})
	if err := pw.WriteIncremental(&buf, size, doc.StartXRefOffset(), update); err != nil {
		return fmt.Errorf("sign: write update: %w", err)
	}
	tail := buf.Bytes()
	contentsMark := []byte("/Contents <" + hex.EncodeToString(make([]byte, opts.ContentsSize)) + ">")
	ci := bytes.Index(tail, contentsMark)
	bi := bytes.Index(tail, byteRangePlaceholder)
	if ci < 0 || bi < 0 {
		return errors.New("sign: placeholders not found in the serialized update")
	}
	// The hole is the hex string itself, "<...>".
	holeStart := size + int64(ci) + int64(len("/Contents "))
	holeLen := int64(len(contentsMark) - len("/Contents "))
	total := size + int64(len(tail))
	br := fmt.Sprintf("[0 %d %d %d", holeStart, holeStart+holeLen, total-(holeStart+holeLen))
	if len(br)+1 > len(byteRangePlaceholder) {
		return errors.New("sign: ByteRange does not fit its placeholder")
	}
	for len(br) < len(byteRangePlaceholder)-1 {
		br += " "
	}
	copy(tail[bi:], br+"]")

	h := opts.Hash.New()
	if _, err := io.Copy(h, io.NewSectionReader(r, 0, size)); err != nil {
		return err
	}
	h.Write(tail[:ci+len("/Contents ")])
	h.Write(tail[ci+len(contentsMark):])
	cms, err := s.Sign(h.Sum(nil), opts.Hash, opts.SigningTime)
	if err != nil {
		return err
	}
	if len(cms) > opts.ContentsSize {
		return fmt.Errorf("sign: CMS container is %d bytes, more than the %d reserved (raise ContentsSize)", len(cms), opts.ContentsSize)
	}
	hexCMS := make([]byte, 2*opts.ContentsSize)
	for i := range hexCMS {
		hexCMS[i] = '0'
	}
	hex.Encode(hexCMS, cms)
	copy(tail[ci+len("/Contents <"):], hexCMS)

	if _, err := io.Copy(out, io.NewSectionReader(r, 0, size)); err != nil {
		return err
	}
	_, err = out.Write(tail)
	return err
}

// SignFile opens an existing PDF from path, applies signing via Sign, and writes
// the result to out.
func SignFile(path string, out io.Writer, s Signer, opts SignOptions) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	info, err := f.Stat()
	if err != nil {
		return err
	}
	return Sign(f, info.Size(), out, s, opts)
}

// byteRangePlaceholder is how the writer prints the placeholder /ByteRange: wide
// enough for three 12-digit offsets.
var byteRangePlaceholder = []byte("[0 100000000000 100000000000 100000000000]")

// update is the incremental section: the new and changed objects plus a trailer.
type update struct {
	objects map[int]model.Object
	gens    map[int]int
	trailer model.Trailer
}

func (u *update) Trailer() model.Trailer { return u.trailer }
func (u *update) Resolve(ref model.Ref) (model.Object, error) {
	o, ok := u.objects[ref.ObjectNumber]
	if !ok {
		return nil, fmt.Errorf("object %d not in update", ref.ObjectNumber)
	}
	return o, nil
}
func (u *update) ObjectNumbers() []int {
	out := make([]int, 0, len(u.objects))
	for n := range u.objects {
		out = append(out, n)
	}
	return out
}
func (u *update) Generation(num int) int { return u.gens[num] }

func buildSignatureUpdate(doc reader.Document, opts SignOptions) (*update, error) {
	cat, err := doc.Catalog()
	if err != nil {
		return nil, fmt.Errorf("sign: %w", err)
	}
	pages, err := doc.Pages()
	if err != nil {
		return nil, fmt.Errorf("sign: %w", err)
	}
	if opts.PageIndex < 0 || opts.PageIndex >= len(pages) {
		return nil, fmt.Errorf("sign: page %d out of range (%d pages)", opts.PageIndex, len(pages))
	}
	trailer := doc.Trailer().Dict
	size, _ := trailer[model.Name("Size")].(model.Integer)
	next := int(size)
	if next < 1 {
		next = 1
	}
	u := &update{objects: map[int]model.Object{}, gens: map[int]int{}}
	root := doc.Trailer().Root()
	if root == nil {
		return nil, errors.New("sign: trailer has no /Root")
	}
	alloc := func() int { n := next; next++; return n }
	sigNum := alloc()
	fieldNum := alloc()
	sigRef := model.Ref{ObjectNumber: sigNum}
	fieldRef := model.Ref{ObjectNumber: fieldNum}

	// AcroForm: reuse the existing dictionary or add one.
	var acro model.Dict
	acroNum := 0
	acroInline := false
	switch v := cat.Dict[model.Name("AcroForm")].(type) {
	case model.Ref:
		o, err := doc.Resolve(v)
		if err != nil {
			return nil, fmt.Errorf("sign: AcroForm: %w", err)
		}
		if d, ok := o.(model.Dict); ok {
			acro = copyDict(d)
			acroNum = v.ObjectNumber
			u.gens[acroNum] = v.Generation
		}
	case model.Dict:
		acro = copyDict(v)
		acroInline = true
	}
	if acro == nil {
		acro = model.Dict{}
	}
	fields, _ := resolveArray(doc, acro[model.Name("Fields")])
	existing := 0
	for _, f := range fields {
		if d, ok := resolveDict(doc, f); ok {
			if ft, _ := d[model.Name("FT")].(model.Name); ft == "Sig" {
				existing++
			}
		}
	}
	if opts.Certify > 0 && existing > 0 {
		return nil, errors.New("sign: only the first signature can certify a document")
	}
	name := opts.FieldName
	if name == "" {
		name = fmt.Sprintf("Signature%d", existing+1)
	}
	acro[model.Name("Fields")] = append(append(model.Array{}, fields...), fieldRef)
	flags, _ := acro[model.Name("SigFlags")].(model.Integer)
	acro[model.Name("SigFlags")] = flags | 3
	delete(acro, model.Name("NeedAppearances"))

	// Catalog (changed when the AcroForm is new or inline, or when certifying).
	catalog := copyDict(cat.Dict)
	catalogChanged := false
	if acroInline || acroNum == 0 {
		acroNum = alloc()
		catalog[model.Name("AcroForm")] = model.Ref{ObjectNumber: acroNum}
		catalogChanged = true
	}
	u.objects[acroNum] = acro
	if opts.Certify > 0 {
		perms, _ := resolveDict(doc, catalog[model.Name("Perms")])
		perms = copyDict(perms)
		perms[model.Name("DocMDP")] = sigRef
		catalog[model.Name("Perms")] = perms
		catalogChanged = true
	}
	if catalogChanged {
		u.objects[root.ObjectNumber] = catalog
		u.gens[root.ObjectNumber] = root.Generation
	}

	// Signature dictionary with placeholders.
	sig := model.Dict{
		model.Name("Type"):      model.Name("Sig"),
		model.Name("Filter"):    model.Name("Adobe.PPKLite"),
		model.Name("SubFilter"): model.Name(opts.SubFilter),
		model.Name("ByteRange"): model.Array{model.Integer(0), model.Integer(100000000000), model.Integer(100000000000), model.Integer(100000000000)},
		model.Name("Contents"):  model.HexString(make([]byte, opts.ContentsSize)),
		model.Name("M"):         model.String(model.FormatDate(opts.SigningTime, false)),
	}
	if opts.Reason != "" {
		sig[model.Name("Reason")] = model.EncodeTextString(opts.Reason, false)
	}
	if opts.Location != "" {
		sig[model.Name("Location")] = model.EncodeTextString(opts.Location, false)
	}
	if opts.Contact != "" {
		sig[model.Name("ContactInfo")] = model.EncodeTextString(opts.Contact, false)
	}
	if opts.Certify > 0 {
		p := min(opts.Certify, 3)
		sig[model.Name("Reference")] = model.Array{model.Dict{
			model.Name("Type"):            model.Name("SigRef"),
			model.Name("TransformMethod"): model.Name("DocMDP"),
			model.Name("TransformParams"): model.Dict{
				model.Name("Type"): model.Name("TransformParams"),
				model.Name("P"):    model.Integer(p),
				model.Name("V"):    model.Name("1.2"),
			},
		}}
	}
	u.objects[sigNum] = sig

	// Signature field merged with its widget annotation.
	page := pages[opts.PageIndex]
	rect := model.Array{model.Real(opts.Rect[0]), model.Real(opts.Rect[1]), model.Real(opts.Rect[2]), model.Real(opts.Rect[3])}
	field := model.Dict{
		model.Name("Type"):    model.Name("Annot"),
		model.Name("Subtype"): model.Name("Widget"),
		model.Name("FT"):      model.Name("Sig"),
		model.Name("T"):       model.EncodeTextString(name, false),
		model.Name("V"):       sigRef,
		model.Name("F"):       model.Integer(4 | 128), // Print, Locked
		model.Name("Rect"):    rect,
		model.Name("P"):       page.Ref,
	}
	w, h := opts.Rect[2]-opts.Rect[0], opts.Rect[3]-opts.Rect[1]
	if w > 0 && h > 0 {
		apNum := alloc()
		u.objects[apNum] = appearanceStream(opts, w, h)
		field[model.Name("AP")] = model.Dict{model.Name("N"): model.Ref{ObjectNumber: apNum}}
	}
	u.objects[fieldNum] = field

	// Page: append the widget to /Annots.
	pageDict := copyDict(page.Dict)
	annots, _ := resolveArray(doc, pageDict[model.Name("Annots")])
	pageDict[model.Name("Annots")] = append(append(model.Array{}, annots...), fieldRef)
	u.objects[page.Ref.ObjectNumber] = pageDict
	u.gens[page.Ref.ObjectNumber] = page.Ref.Generation

	td := model.Dict{model.Name("Root"): *root, model.Name("Size"): model.Integer(next)}
	if info := doc.Trailer().Info(); info != nil {
		td[model.Name("Info")] = *info
	}
	if id, ok := trailer[model.Name("ID")]; ok {
		td[model.Name("ID")] = id
	}
	u.trailer = model.Trailer{Dict: td}
	return u, nil
}

// appearanceStream draws a plain text block for a visible signature.
func appearanceStream(opts SignOptions, w, h float64) *model.Stream {
	size := 9.0
	lines := []string{"Digitally signed", "Date: " + opts.SigningTime.Format("2006-01-02 15:04:05 MST")}
	if opts.signerName != "" {
		lines[0] = "Digitally signed by " + opts.signerName
	}
	if opts.Reason != "" {
		lines = append(lines, "Reason: "+opts.Reason)
	}
	if opts.Location != "" {
		lines = append(lines, "Location: "+opts.Location)
	}
	var b bytes.Buffer
	fmt.Fprintf(&b, "q 0.5 G 0.5 w 0.25 0.25 %s %s re S Q\nBT /Helv %s Tf 0 g %s TL 4 %s Td\n",
		model.FormatReal(w-0.5), model.FormatReal(h-0.5), model.FormatReal(size), model.FormatReal(size*1.2), model.FormatReal(h-size-3))
	for _, l := range lines {
		b.WriteString("(" + escapeLiteral(l) + ") Tj T*\n")
	}
	b.WriteString("ET\n")
	return &model.Stream{
		Dict: model.Dict{
			model.Name("Type"):    model.Name("XObject"),
			model.Name("Subtype"): model.Name("Form"),
			model.Name("BBox"):    model.Array{model.Integer(0), model.Integer(0), model.Real(w), model.Real(h)},
			model.Name("Resources"): model.Dict{model.Name("Font"): model.Dict{model.Name("Helv"): model.Dict{
				model.Name("Type"): model.Name("Font"), model.Name("Subtype"): model.Name("Type1"),
				model.Name("BaseFont"): model.Name("Helvetica"), model.Name("Encoding"): model.Name("WinAnsiEncoding"),
			}}},
			model.Name("Length"): model.Integer(b.Len()),
		},
		Content: b.Bytes(),
	}
}

func escapeLiteral(s string) string {
	var b bytes.Buffer
	for _, r := range s {
		switch {
		case r == '(' || r == ')' || r == '\\':
			b.WriteByte('\\')
			b.WriteByte(byte(r))
		case r < 0x20 || r > 0xFF:
			b.WriteByte('?')
		case r > 0x7E:
			fmt.Fprintf(&b, "\\%03o", r)
		default:
			b.WriteByte(byte(r))
		}
	}
	return b.String()
}

func copyDict(d model.Dict) model.Dict {
	out := make(model.Dict, len(d)+2)
	for k, v := range d {
		out[k] = v
	}
	return out
}

func resolveArray(doc reader.Document, o model.Object) (model.Array, bool) {
	if ref, ok := o.(model.Ref); ok {
		o, _ = doc.Resolve(ref)
	}
	a, ok := o.(model.Array)
	return a, ok
}

func resolveDict(doc reader.Document, o model.Object) (model.Dict, bool) {
	if ref, ok := o.(model.Ref); ok {
		o, _ = doc.Resolve(ref)
	}
	d, ok := o.(model.Dict)
	return d, ok
}

// docMDPLevel returns the DocMDP /P of a certification signature, if any.
func docMDPLevel(doc reader.Document) (int, bool) {
	cat, err := doc.Catalog()
	if err != nil || cat == nil {
		return 0, false
	}
	perms, ok := resolveDict(doc, cat.Dict[model.Name("Perms")])
	if !ok {
		return 0, false
	}
	sig, ok := resolveDict(doc, perms[model.Name("DocMDP")])
	if !ok {
		return 0, false
	}
	refs, _ := resolveArray(doc, sig[model.Name("Reference")])
	for _, r := range refs {
		if d, ok := resolveDict(doc, r); ok {
			if tp, ok := resolveDict(doc, d[model.Name("TransformParams")]); ok {
				if p, ok := tp[model.Name("P")].(model.Integer); ok {
					return int(p), true
				}
				return 2, true
			}
		}
	}
	return 2, true
}
