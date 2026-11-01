package doc

import (
	"errors"
	"fmt"
	"time"

	"github.com/gsoultan/gpdf/doc/metadata"
	"github.com/gsoultan/gpdf/doc/style"
	"github.com/gsoultan/gpdf/model"
)

// PDFALevel is a PDF/A conformance level.
type PDFALevel string

const (
	// PDFA2B is ISO 19005-2 level B (visual reproducibility, based on PDF 1.7).
	PDFA2B PDFALevel = "2B"
	// PDFA3B is ISO 19005-3 level B: PDF/A-2 plus arbitrary embedded files.
	PDFA3B PDFALevel = "3B"
	// PDFA4 is ISO 19005-4 (based on PDF 2.0; no conformance levels).
	PDFA4 PDFALevel = "4"
	// PDFA4E is PDF/A-4e (engineering: PDF/A-4 plus 3D and rich media).
	PDFA4E PDFALevel = "4E"
	// PDFA4F is PDF/A-4f: PDF/A-4 with at least one embedded file.
	PDFA4F PDFALevel = "4F"
)

// ErrPDFA is wrapped by every PDF/A conformance error.
var ErrPDFA = errors.New("pdf/a")

// PDFA makes Build produce a PDF/A conforming file: a PDF 1.7 header (2.0 for part 4), XMP metadata with
// the pdfaid schema mirroring the Info dictionary, an sRGB output intent (unless
// SetDefaultICCProfile gave another profile with data), and no encryption. Build fails
// with an error wrapping ErrPDFA when the document breaks a rule it can check: every font
// must be embedded (register TrueType fonts with RegisterFont; the standard 14 fonts are
// not allowed), and no DeviceCMYK colour may be used with the RGB output intent.
func (b *DocumentBuilder) PDFA(level PDFALevel) *DocumentBuilder {
	switch level {
	case PDFA2B, PDFA3B, PDFA4, PDFA4E, PDFA4F:
		b.pdfa = level
	default:
		b.setErr(fmt.Errorf("%w: unsupported level %q", ErrPDFA, level))
	}
	return b
}

// CreationDate sets the Info /CreationDate (and XMP CreateDate). Without it, PDF/A
// output uses the build time.
func (b *DocumentBuilder) CreationDate(t time.Time) *DocumentBuilder {
	b.metadata.CreationDate = t
	return b
}

func (l PDFALevel) part() string { return string(l[:1]) }

// conformance is the pdfaid:conformance value: B for parts 2 and 3; parts 4 have none
// (4e and 4f are expressed through pdfaid:rev plus the file's content).
func (l PDFALevel) conformance() string {
	if l.part() == "4" {
		if len(l) > 1 {
			return string(l[1:])
		}
		return ""
	}
	return string(l[1:])
}

// pdf20 reports whether the level is based on PDF 2.0 (part 4).
func (l PDFALevel) pdf20() bool { return l.part() == "4" }

// PDFVersion is the header version the level requires.
func (l PDFALevel) PDFVersion() string {
	if l.pdf20() {
		return "2.0"
	}
	return "1.7"
}

// preparePDFA fills in what PDF/A needs before the objects are built.
func (b *DocumentBuilder) preparePDFA() {
	if b.pdfa == "" {
		return
	}
	if b.iccProfile == nil || len(b.iccProfile.Data) == 0 {
		p := style.SRGBICCProfile()
		b.iccProfile = &p
	}
	if b.metadata.CreationDate.IsZero() {
		b.metadata.CreationDate = time.Now()
	}
	if b.pdfa.pdf20() {
		// PDF/A-4 forbids the Info dictionary (ISO 19005-4 §6.1.3): XMP carries the metadata.
		b.metadata.DatesOnly = true
	}
	if len(b.metadata.MetadataXMP) == 0 {
		b.metadata.MetadataXMP = metadata.BuildXMP(b.xmpInput())
	}
}

// xmpInput mirrors the Info dictionary into an XMP packet.
func (b *DocumentBuilder) xmpInput() metadata.XMPInput {
	m := b.metadata
	in := metadata.XMPInput{
		Title: m.Title, Author: m.Author, Subject: m.Subject, Keywords: m.Keywords,
		Creator: m.Creator, Producer: m.Producer,
		Created: m.CreationDate, Modified: m.CreationDate,
	}
	if b.pdfa != "" {
		in.PDFAPart = b.pdfa.part()
		in.PDFAConformance = b.pdfa.conformance()
		if b.pdfa.pdf20() {
			in.PDFARev = "2020"
		}
	}
	if b.pdfua != "" {
		in.PDFUAPart = string(b.pdfua)
		if b.pdfua == PDFUA2 {
			in.PDFUARev = "2024"
		}
	}
	return in
}

// prepareMetadata writes an XMP packet for documents that have descriptive metadata.
// PDF 2.0 deprecates the Info dictionary entries other than the dates (ISO 32000-2
// §14.3.3), so the builder always mirrors them into /Metadata.
func (b *DocumentBuilder) prepareMetadata() {
	b.metadata.PDF20 = b.pdfa == "" || b.pdfa.pdf20()
	if b.pdfua == PDFUA1 && !(b.pdfa != "" && b.pdfa.pdf20()) {
		b.metadata.PDF20 = false // PDF/UA-1 files are PDF 1.7
	}
	if b.pdfa != "" {
		b.preparePDFA()
		return
	}
	if len(b.metadata.MetadataXMP) == 0 && (b.metadata.HasFields() || !b.metadata.CreationDate.IsZero() || b.pdfua != "") {
		b.metadata.MetadataXMP = metadata.BuildXMP(b.xmpInput())
	}
}

// checkPDFA inspects the built objects for rules PDF/A level B needs.
func (b *DocumentBuilder) checkPDFA(objs map[int]model.Object) error {
	if b.pdfa == "" {
		return nil
	}
	rgbIntent := b.iccProfile.N == 3
	var err error
	for _, obj := range objs {
		// Font dictionaries may be inline in a page's /Resources, so walk every object.
		walkBuilt(obj, 0, func(o model.Object) {
			if err != nil {
				return
			}
			switch v := o.(type) {
			case model.Dict:
				if t, _ := v[model.Name("Type")].(model.Name); t == "Font" {
					err = checkPDFAFont(v, objs)
				}
			}
		})
		if err != nil {
			return err
		}
	}
	if rgbIntent {
		for i := range b.pc.pages {
			for _, im := range b.pc.pages[i].ImageRuns {
				if im.ColorSpace == "DeviceCMYK" {
					return fmt.Errorf("%w: DeviceCMYK image needs a CMYK output intent", ErrPDFA)
				}
			}
		}
	}
	if b.pdfa == PDFA4F && len(b.attachments) == 0 {
		return fmt.Errorf("%w: PDF/A-4f needs at least one embedded file (AttachFile)", ErrPDFA)
	}
	if b.pdfa == PDFA2B && len(b.attachments) > 0 {
		return fmt.Errorf("%w: PDF/A-2 does not allow embedded files; use PDFA3B or PDFA4F", ErrPDFA)
	}
	if rgbIntent {
		for i := range b.pc.pages {
			for _, gr := range b.pc.pages[i].GraphicRuns {
				for _, op := range gr.Ops {
					if op.Name == "k" || op.Name == "K" {
						return fmt.Errorf("%w: page %d uses a DeviceCMYK colour with an RGB output intent", ErrPDFA, i+1)
					}
				}
			}
		}
	}
	return nil
}

func checkPDFAFont(f model.Dict, objs map[int]model.Object) error {
	base, _ := f[model.Name("BaseFont")].(model.Name)
	subtype, _ := f[model.Name("Subtype")].(model.Name)
	if subtype == "Type0" {
		return nil // checked through its descendant CIDFont
	}
	desc, ok := resolveBuilt(f[model.Name("FontDescriptor")], objs).(model.Dict)
	if !ok {
		return fmt.Errorf("%w: font %s is not embedded; register an embeddable font with RegisterFont", ErrPDFA, base)
	}
	for _, k := range []model.Name{"FontFile", "FontFile2", "FontFile3"} {
		if _, ok := desc[k]; ok {
			return nil
		}
	}
	return fmt.Errorf("%w: font %s has no embedded font program", ErrPDFA, base)
}

func resolveBuilt(o model.Object, objs map[int]model.Object) model.Object {
	if r, ok := o.(model.Ref); ok {
		return objs[r.ObjectNumber]
	}
	return o
}

// walkBuilt calls fn for o and every dictionary, array and stream nested in it.
func walkBuilt(o model.Object, depth int, fn func(model.Object)) {
	if depth > 32 {
		return
	}
	fn(o)
	switch v := o.(type) {
	case model.Dict:
		for _, c := range v {
			walkBuilt(c, depth+1, fn)
		}
	case model.Array:
		for _, c := range v {
			walkBuilt(c, depth+1, fn)
		}
	case *model.Stream:
		walkBuilt(v.Dict, depth+1, fn)
	}
}
