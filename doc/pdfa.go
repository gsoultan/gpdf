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
)

// ErrPDFA is wrapped by every PDF/A conformance error.
var ErrPDFA = errors.New("pdf/a")

// PDFA makes Build produce a PDF/A conforming file: a PDF 1.7 header, XMP metadata with
// the pdfaid schema mirroring the Info dictionary, an sRGB output intent (unless
// SetDefaultICCProfile gave another profile with data), and no encryption. Build fails
// with an error wrapping ErrPDFA when the document breaks a rule it can check: every font
// must be embedded (register TrueType fonts with RegisterFont; the standard 14 fonts are
// not allowed), and no DeviceCMYK colour may be used with the RGB output intent.
func (b *DocumentBuilder) PDFA(level PDFALevel) *DocumentBuilder {
	switch level {
	case PDFA2B, PDFA3B:
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

func (l PDFALevel) part() string        { return string(l[:1]) }
func (l PDFALevel) conformance() string { return string(l[1:]) }

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
	if len(b.metadata.MetadataXMP) == 0 {
		m := b.metadata
		b.metadata.MetadataXMP = metadata.BuildXMP(metadata.XMPInput{
			Title: m.Title, Author: m.Author, Subject: m.Subject, Keywords: m.Keywords,
			Creator: m.Creator, Producer: m.Producer,
			Created: m.CreationDate, Modified: m.CreationDate,
			PDFAPart: b.pdfa.part(), PDFAConformance: b.pdfa.conformance(),
		})
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
			case *model.Stream:
				if st, _ := v.Dict[model.Name("Subtype")].(model.Name); st == "Image" && rgbIntent {
					if cs, _ := v.Dict[model.Name("ColorSpace")].(model.Name); cs == "DeviceCMYK" {
						err = fmt.Errorf("%w: DeviceCMYK image needs a CMYK output intent", ErrPDFA)
					}
				}
			}
		})
		if err != nil {
			return err
		}
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
