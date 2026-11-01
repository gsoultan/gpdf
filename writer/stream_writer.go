package writer

import (
	"crypto/md5"
	"errors"
	"fmt"
	"io"

	"github.com/gsoultan/gpdf/model"
)

// StreamWriter writes a PDF file object by object, straight to an io.Writer, so
// that a document of any size can be produced without holding its objects in
// memory. Streams (page content, images, fonts) go out as soon as they are
// written; dictionaries and other small objects are packed into object streams
// (unless WriterOptions.ObjectStreams is false) and written at Close together with
// the cross-reference stream, so only they and the cross-reference entries stay in
// memory.
//
// Object numbers come from Alloc, or the caller picks its own; each number may be
// written once. The output is not encrypted and not validated for dangling
// references: a reference to a number never written reads as null.
type StreamWriter struct {
	pw      *PDFWriter
	out     *offsetWriter
	pk      *packer
	version string
	started bool
	closed  bool
	used    map[int]bool
	next    int
}

// NewStreamWriter returns a writer that produces a PDF 2.0 file on w.
func NewStreamWriter(w io.Writer, options WriterOptions) *StreamWriter {
	pw := NewPDFWriterWithOptions(options)
	return &StreamWriter{
		pw:      pw,
		out:     &offsetWriter{w: w, h: md5.New()},
		version: "2.0",
		used:    make(map[int]bool),
		next:    1,
	}
}

// SetVersion sets the header version ("1.4" to "1.7", or "2.0") before the first
// object is written. Object streams need 1.5 or later; a 1.4 file gets a classic
// cross-reference table.
func (sw *StreamWriter) SetVersion(v string) error {
	if sw.started {
		return errors.New("writer: version must be set before the first object is written")
	}
	switch v {
	case "1.4", "1.5", "1.6", "1.7", "2.0":
		sw.version = v
		return nil
	}
	return fmt.Errorf("writer: unsupported PDF version %q", v)
}

func (sw *StreamWriter) header() string {
	if sw.version == "2.0" {
		return headerPDF20
	}
	return "%PDF-" + sw.version + "\n%\xE2\xE3\xCF\xD3\n"
}

func (sw *StreamWriter) start() error {
	if sw.started {
		return nil
	}
	sw.started = true
	header := sw.header()
	sw.pk = newPacker(sw.pw, nil, nil, sw.pw.compactFor(header), 0)
	sw.pk.deferred = true
	_, err := io.WriteString(sw.out, header)
	return err
}

// Alloc reserves the next unused object number.
func (sw *StreamWriter) Alloc() model.Ref {
	for sw.used[sw.next] {
		sw.next++
	}
	ref := model.Ref{ObjectNumber: sw.next}
	sw.used[sw.next] = true
	sw.next++
	return ref
}

// NextObjectNumber returns the lowest object number that is neither allocated nor
// written, so a caller numbering objects itself can continue from it.
func (sw *StreamWriter) NextObjectNumber() int {
	for sw.used[sw.next] {
		sw.next++
	}
	return sw.next
}

// WriteObject writes obj as indirect object ref. Streams go to the output at once;
// other objects may be held for an object stream until Close.
func (sw *StreamWriter) WriteObject(ref model.Ref, obj model.Object) error {
	if sw.closed {
		return errors.New("writer: write after Close")
	}
	if ref.ObjectNumber < 1 {
		return fmt.Errorf("writer: invalid object number %d", ref.ObjectNumber)
	}
	if sw.pk != nil {
		if _, written := sw.pk.entries[ref.ObjectNumber]; written || sw.pendingHas(ref.ObjectNumber) {
			return fmt.Errorf("writer: object %d written twice", ref.ObjectNumber)
		}
	}
	if err := sw.start(); err != nil {
		return err
	}
	sw.used[ref.ObjectNumber] = true
	return sw.pk.write(sw.out, ref, obj)
}

func (sw *StreamWriter) pendingHas(num int) bool {
	for _, n := range sw.pk.pendingNums {
		if n == num {
			return true
		}
	}
	for _, g := range sw.pk.groups {
		for _, n := range g.nums {
			if n == num {
				return true
			}
		}
	}
	return false
}

// Offset returns the number of bytes written so far.
func (sw *StreamWriter) Offset() int64 { return sw.out.off }

// Close writes the object streams held back, the cross-reference section and the
// trailer, which must carry /Root and may carry /Info and /ID (an /ID is derived
// from the bytes written when absent). Nothing may be written afterwards.
func (sw *StreamWriter) Close(trailer model.Dict) error {
	if sw.closed {
		return errors.New("writer: closed twice")
	}
	if _, ok := trailer[model.Name("Root")].(model.Ref); !ok {
		return fmt.Errorf("%w: trailer has no /Root", ErrInvalidDocumentGraph)
	}
	if err := sw.start(); err != nil {
		return err
	}
	sw.closed = true
	out := outputTrailer(trailer, 0)
	delete(out, model.Name("Size"))
	// Object and cross-reference streams are numbered after every number used.
	sw.pk.next = max(sw.pk.next, sw.pk.maxNum+1, sw.NextObjectNumber())
	if err := sw.pk.flush(sw.out); err != nil {
		return err
	}
	if _, ok := out[model.Name("ID")].(model.Array); !ok {
		out[model.Name("ID")] = fileID(sw.out.h.Sum(nil))
	}
	return sw.pk.finish(sw.out, out)
}
