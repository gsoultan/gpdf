package file

import (
	"errors"
	"fmt"
	"io"

	"github.com/gsoultan/gpdf/model"
	"github.com/gsoultan/gpdf/reader"
	"github.com/gsoultan/gpdf/writer"
)

// Document wraps a reader.Document and an optional closer, implementing doc.Document (Save, Close).
type Document struct {
	closer io.Closer
	doc    reader.Document
	w      writer.Writer
}

// NewDocument returns a Document that delegates read operations to doc and calls closer.Close on Close.
func NewDocument(closer io.Closer, doc reader.Document) *Document {
	d := &Document{
		closer: closer,
		doc:    doc,
		// Real-world files often reference deleted objects. The strict graph check is meant
		// for documents built in memory; for files read from disk the writer writes such
		// dangling references as null, as ISO 32000-1 §7.3.10 defines them.
		w: writer.NewPDFWriterWithOptions(writer.WriterOptions{ValidateGraph: new(false)}),
	}
	// No finalizer: one on this wrapper could run while a method is still using the
	// underlying reader (the wrapper is unreachable once its fields are loaded) and close
	// the file mid-read. An *os.File closes itself when it becomes unreachable anyway.
	return d
}

// Catalog returns the document catalog.
func (d *Document) Catalog() (*model.Catalog, error) {
	return d.doc.Catalog()
}

// Pages returns the list of pages.
func (d *Document) Pages() ([]model.Page, error) {
	return d.doc.Pages()
}

// Info returns the document Info dictionary.
func (d *Document) Info() (model.Dict, error) {
	return d.doc.Info()
}

// Attachments lists the embedded files (EmbeddedFiles name tree and /AF).
func (d *Document) Attachments() ([]reader.EmbeddedFile, error) {
	cat, err := d.doc.Catalog()
	if err != nil {
		return nil, err
	}
	return reader.ExtractEmbeddedFiles(d.doc, cat)
}

// MetadataStream returns the XMP metadata stream bytes.
func (d *Document) MetadataStream() ([]byte, error) {
	return d.doc.MetadataStream()
}

// StartXRefOffset returns the file offset of the xref used to read this document (0 if not from file).
func (d *Document) StartXRefOffset() int64 {
	return d.doc.StartXRefOffset()
}

// Trailer returns the document trailer.
func (d *Document) Trailer() model.Trailer {
	return d.doc.Trailer()
}

// Save writes the document to w.
func (d *Document) Save(w io.Writer) error {
	return d.w.Write(w, d.doc)
}

// SaveWithPassword writes the document encrypted with the given passwords.
func (d *Document) SaveWithPassword(w io.Writer, userPassword, ownerPassword string) error {
	return d.w.WriteWithPassword(w, d.doc, userPassword, ownerPassword)
}

// SaveWithAES256Password writes the document encrypted with AES-256.
func (d *Document) SaveWithAES256Password(w io.Writer, userPassword, ownerPassword string) error {
	return d.w.WriteWithAES256Password(w, d.doc, userPassword, ownerPassword)
}

// SaveLinearized writes a linearized (fast web view) PDF to ws.
func (d *Document) SaveLinearized(ws writer.WriteSeeker) error {
	return d.w.WriteLinearized(ws, d.doc)
}

// ReadContent returns all extracted text from the document.
func (d *Document) ReadContent() (string, error) {
	return d.doc.Content()
}

// ReadContentPerPage returns extracted text for each page in order.
func (d *Document) ReadContentPerPage() ([]string, error) {
	return d.doc.ContentPerPage()
}

// ReadContentInReadingOrder returns the text in reading order (structure tree or
// layout analysis).
func (d *Document) ReadContentInReadingOrder() (string, error) {
	return reader.ExtractTextInReadingOrder(d.doc)
}

// ReadStructuredText returns the reading-order text as elements with their role and page.
func (d *Document) ReadStructuredText() ([]reader.TextElement, error) {
	return reader.ExtractStructuredText(d.doc)
}

// changeTracker is what the reader document offers for incremental updates.
type changeTracker interface {
	Source() (io.ReaderAt, int64)
	MarkModified(ref model.Ref, obj model.Object)
	NewObject(obj model.Object) model.Ref
	ModifiedObjects() map[model.Ref]model.Object
	Redacted() bool
	Generation(num int) int
}

// MarkModified records obj as the new value of ref: Save writes it, and
// SaveIncremental appends it.
func (d *Document) MarkModified(ref model.Ref, obj model.Object) {
	if t, ok := d.doc.(changeTracker); ok {
		t.MarkModified(ref, obj)
	}
}

// SetObject replaces or adds the indirect object ref (see MarkModified).
func (d *Document) SetObject(ref model.Ref, obj model.Object) { d.MarkModified(ref, obj) }

// MarkRedacted records that content was removed, after which SaveIncremental is refused.
func (d *Document) MarkRedacted() {
	if t, ok := d.doc.(interface{ MarkRedacted() }); ok {
		t.MarkRedacted()
	}
}

// NewObject adds obj as a new indirect object and returns its reference.
func (d *Document) NewObject(obj model.Object) model.Ref {
	if t, ok := d.doc.(changeTracker); ok {
		return t.NewObject(obj)
	}
	return model.Ref{}
}

// SetInfo sets one entry of the document information dictionary (creating the
// dictionary when the file has none), for example "Title" or "Author".
func (d *Document) SetInfo(key string, value model.Object) error {
	t, ok := d.doc.(changeTracker)
	if !ok {
		return errors.New("document does not track changes")
	}
	trailer := d.doc.Trailer()
	if ref := trailer.Info(); ref != nil {
		obj, err := d.doc.Resolve(*ref)
		if err != nil {
			return err
		}
		info, _ := obj.(model.Dict)
		updated := make(model.Dict, len(info)+1)
		for k, v := range info {
			updated[k] = v
		}
		updated[model.Name(key)] = value
		t.MarkModified(*ref, updated)
		return nil
	}
	ref := t.NewObject(model.Dict{model.Name(key): value})
	trailer.Dict[model.Name("Info")] = ref
	return nil
}

// incrementalUpdate is the object set appended by SaveIncremental.
type incrementalUpdate struct {
	objects map[int]model.Object
	gens    map[int]int
	trailer model.Trailer
}

func (u *incrementalUpdate) Trailer() model.Trailer { return u.trailer }
func (u *incrementalUpdate) Generation(num int) int { return u.gens[num] }
func (u *incrementalUpdate) Resolve(ref model.Ref) (model.Object, error) {
	o, ok := u.objects[ref.ObjectNumber]
	if !ok {
		return nil, fmt.Errorf("object %d not in update", ref.ObjectNumber)
	}
	return o, nil
}
func (u *incrementalUpdate) ObjectNumbers() []int {
	out := make([]int, 0, len(u.objects))
	for n := range u.objects {
		out = append(out, n)
	}
	return out
}

// SaveIncremental writes the original file bytes followed by an incremental
// update holding the changed and added objects. With nothing changed, the file
// is copied as it is.
func (d *Document) SaveIncremental(w io.Writer) error {
	t, ok := d.doc.(changeTracker)
	if !ok {
		return errors.New("document does not track changes")
	}
	if t.Redacted() {
		return ErrRedacted
	}
	r, size := t.Source()
	if r == nil {
		return errors.New("document has no source file")
	}
	if _, err := io.Copy(w, io.NewSectionReader(r, 0, size)); err != nil {
		return err
	}
	changed := t.ModifiedObjects()
	if len(changed) == 0 {
		return nil
	}
	u := &incrementalUpdate{objects: make(map[int]model.Object, len(changed)), gens: make(map[int]int, len(changed))}
	for ref, obj := range changed {
		u.objects[ref.ObjectNumber] = obj
		u.gens[ref.ObjectNumber] = ref.Generation
	}
	base := d.doc.Trailer().Dict
	u.trailer = model.Trailer{Dict: make(model.Dict, len(base))}
	for k, v := range base {
		u.trailer.Dict[k] = v
	}
	pw := writer.NewPDFWriterWithOptions(writer.WriterOptions{ValidateGraph: new(false)})
	return pw.WriteIncremental(w, size, d.doc.StartXRefOffset(), u)
}

// ErrRedacted is returned by SaveIncremental after content was redacted: an
// incremental update would keep the removed bytes in the file.
var ErrRedacted = errors.New("redacted document must be saved with Save, not incrementally")

// ReadImages returns metadata and raw bytes for all images across all pages.
func (d *Document) ReadImages() ([]reader.ImageInfo, error) {
	return d.doc.Images()
}

// ReadImagesPerPage returns image info grouped by page (one slice per page).
func (d *Document) ReadImagesPerPage() ([][]reader.ImageInfo, error) {
	return d.doc.ImagesPerPage()
}

// ReadLayout returns one PageLayout per page with positioned, styled TextBlocks.
func (d *Document) ReadLayout() ([]reader.PageLayout, error) {
	return d.doc.Layout()
}

// ReadTables detects table-like grids from the page layout and returns them grouped by page.
func (d *Document) ReadTables() ([][]reader.Table, error) {
	return d.doc.Tables()
}

// Search returns, for each keyword, the 0-based page indices where it was found.
func (d *Document) Search(keywords ...string) ([]model.SearchResult, error) {
	return d.doc.Search(keywords...)
}

// Replace replaces all occurrences of old with new in content streams.
func (d *Document) Replace(old, new string) error {
	return d.doc.Replace(old, new)
}

// Replaces applies multiple replacements to content streams.
func (d *Document) Replaces(replacements map[string]string) error {
	return d.doc.Replaces(replacements)
}

// Resolve returns the indirect object at the given reference.
func (d *Document) Resolve(ref model.Ref) (model.Object, error) {
	return d.doc.Resolve(ref)
}

// ObjectNumbers returns all indirect object numbers (for writing).
func (d *Document) ObjectNumbers() []int {
	return d.doc.ObjectNumbers()
}

// Close closes the underlying resources. Idempotent.
func (d *Document) Close() error {
	if d.closer == nil {
		return nil
	}
	err := d.closer.Close()
	d.closer = nil
	return err
}
