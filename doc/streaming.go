// Copyright (c) 2026 Gembit Soultan Shirazi <gembit.soultan@gmail.com>

package doc

import (
	"errors"
	"fmt"
	"io"
	"sort"

	"github.com/gsoultan/gpdf/model"
	"github.com/gsoultan/gpdf/writer"
)

// ErrStreaming reports a misuse of the streaming output mode.
var ErrStreaming = errors.New("streaming")

// docStream is the state of a builder that writes its output as it goes.
type docStream struct {
	w       io.Writer
	sw      *writer.StreamWriter
	version string
}

// StreamTo makes the builder write the document to w as it is built instead of
// holding every object in memory: each image is written when it is drawn and its
// samples released, each page's content stream is written as the page is built,
// and Finish writes what remains (pages, fonts, structure, metadata) with the
// cross-reference. Memory then grows with the text on the pages, not with the
// images. Call it before drawing; set PDFA or PDFUA before the first image.
// Encryption and linearization are not available in this mode.
func (b *DocumentBuilder) StreamTo(w io.Writer) *DocumentBuilder {
	if b.stream != nil {
		b.setErr(fmt.Errorf("%w: StreamTo called twice", ErrStreaming))
		return b
	}
	if w == nil {
		b.setErr(fmt.Errorf("%w: nil writer", ErrStreaming))
		return b
	}
	b.stream = &docStream{w: w}
	return b
}

// outputVersion is the header version the document's conformance level needs.
func (b *DocumentBuilder) outputVersion() string {
	if b.pdfa != "" {
		return b.pdfa.PDFVersion()
	}
	if v := b.pdfuaVersion(); v != "" {
		return v
	}
	return "2.0"
}

// writer returns the stream writer, started with the current output version.
func (ds *docStream) writer(b *DocumentBuilder) (*writer.StreamWriter, error) {
	if ds.sw == nil {
		opts := writer.WriterOptions{ValidateGraph: new(false)}
		if b.noObjectStreams {
			opts.ObjectStreams = new(false)
		}
		ds.sw = writer.NewStreamWriter(ds.w, opts)
		ds.version = b.outputVersion()
		if err := ds.sw.SetVersion(ds.version); err != nil {
			return nil, err
		}
	} else if v := b.outputVersion(); v != ds.version {
		return nil, fmt.Errorf("%w: PDF version changed to %s after output started as %s (set PDFA or PDFUA before drawing)", ErrStreaming, v, ds.version)
	}
	return ds.sw, nil
}

// writeImage writes the image objects of run and releases its samples.
func (ds *docStream) writeImage(b *DocumentBuilder, run *imageRun) error {
	sw, err := ds.writer(b)
	if err != nil {
		return err
	}
	img, mask := b.imageStreams(*run)
	if mask != nil {
		ref := sw.Alloc()
		if err := sw.WriteObject(ref, mask); err != nil {
			return err
		}
		run.MaskObjNum = ref.ObjectNumber
		img.Dict[model.Name("SMask")] = ref
	}
	ref := sw.Alloc()
	if err := sw.WriteObject(ref, img); err != nil {
		return err
	}
	run.ObjNum = ref.ObjectNumber
	run.Raw, run.Mask = nil, nil
	return nil
}

// writeObject writes one object straight to the output.
func (ds *docStream) writeObject(b *DocumentBuilder, num int, obj model.Object) error {
	sw, err := ds.writer(b)
	if err != nil {
		return err
	}
	return sw.WriteObject(model.Ref{ObjectNumber: num}, obj)
}

// Finish completes a document started with StreamTo: it builds the remaining
// objects, writes them and the cross-reference, and returns the first error of
// the build or the output. The builder cannot be used afterwards.
func (b *DocumentBuilder) Finish() error {
	if b.stream == nil {
		return fmt.Errorf("%w: Finish needs StreamTo", ErrStreaming)
	}
	d, err := b.build()
	if err != nil {
		return err
	}
	sw, err := b.stream.writer(b)
	if err != nil {
		return err
	}
	nums := make([]int, 0, len(d.objects))
	for num := range d.objects {
		nums = append(nums, num)
	}
	sort.Ints(nums)
	for _, num := range nums {
		if err := sw.WriteObject(model.Ref{ObjectNumber: num}, d.objects[num]); err != nil {
			return err
		}
	}
	b.stream.sw = nil
	return sw.Close(d.trailer.Dict)
}

// BuildTo builds the document and writes it to w in one go, streaming the page
// content as it is built (see StreamTo). It suits documents that are saved once.
func (b *DocumentBuilder) BuildTo(w io.Writer) error {
	return b.StreamTo(w).Finish()
}
