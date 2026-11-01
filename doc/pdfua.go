package doc

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/gsoultan/gpdf/doc/tagged"
	"github.com/gsoultan/gpdf/model"
)

// PDFUALevel is a PDF/UA (ISO 14289) part.
type PDFUALevel string

const (
	// PDFUA1 is ISO 14289-1 (PDF/UA-1, based on PDF 1.7).
	PDFUA1 PDFUALevel = "1"
	// PDFUA2 is ISO 14289-2 (PDF/UA-2, based on PDF 2.0 with namespaced structure).
	PDFUA2 PDFUALevel = "2"
)

// ErrPDFUA is wrapped by every PDF/UA conformance error.
var ErrPDFUA = errors.New("pdf/ua")

// PDFUA makes Build produce a PDF/UA conforming, tagged file: a Document structure
// element in reading order, pdfuaid identification in the XMP packet, /Lang and a
// displayed title, Link structure elements for link annotations, decorative vector
// graphics marked as artifacts, Lbl/LBody list items, and (for PDF/UA-2) the PDF 2.0
// structure namespace. Substitutes for the standard 14 fonts are embedded when the
// system has them (see EmbedStandardFonts), since PDF/UA requires embedded fonts.
//
// Build fails with an error wrapping ErrPDFUA when a rule it can check is broken: a
// missing Title or SetLanguage, text or images drawn without a structure element (use
// DrawParagraph, DrawHeading, DrawTaggedParagraphBox, DrawList, BeginTable, the tagged
// image methods, or SetHeader/SetFooter for artifacts), figures without alternative
// text, heading levels that skip (H1 before H2, and so on), or a font that is not
// embedded. PDF/UA-2 cannot be combined with PDF/A-2 or PDF/A-3, which are PDF 1.7.
func (b *DocumentBuilder) PDFUA(level PDFUALevel) *DocumentBuilder {
	switch level {
	case PDFUA1, PDFUA2:
		b.pdfua = level
	default:
		b.setErr(fmt.Errorf("%w: unsupported part %q", ErrPDFUA, level))
		return b
	}
	b.useTagged = true
	b.tagging.TagAnnotations = true
	b.tagging.UseNamespaces = level == PDFUA2
	if b.fc.embeddedFonts["Helvetica"] == nil {
		b.EmbedStandardFonts()
	}
	return b
}

// pdfuaPDF20 reports whether the PDF/UA part is based on PDF 2.0.
func (b *DocumentBuilder) pdfuaPDF20() bool { return b.pdfua == PDFUA2 }

// pdfuaVersion returns the header version PDF/UA needs, or "" when the
// document is not PDF/UA.
func (b *DocumentBuilder) pdfuaVersion() string {
	switch b.pdfua {
	case PDFUA1:
		return "1.7"
	case PDFUA2:
		return "2.0"
	}
	return ""
}

// checkPDFUA inspects the builder state and the built objects for the PDF/UA rules
// that can be checked here.
func (b *DocumentBuilder) checkPDFUA(objs map[int]model.Object) error {
	if b.pdfua == "" {
		return nil
	}
	if b.pdfua == PDFUA2 && b.pdfa != "" && !b.pdfa.pdf20() {
		return fmt.Errorf("%w: PDF/UA-2 needs PDF 2.0; combine it with PDF/A-4, not PDF/A-%s", ErrPDFUA, b.pdfa.part())
	}
	if b.metadata.Title == "" {
		return fmt.Errorf("%w: a document title is required (Title)", ErrPDFUA)
	}
	if b.metadata.Lang == "" {
		return fmt.Errorf("%w: a document language is required (SetLanguage)", ErrPDFUA)
	}
	for i := range b.pc.pages {
		for _, r := range b.pc.pages[i].TextRuns {
			if !r.HasMCID && !r.IsArtifact {
				return fmt.Errorf("%w: page %d has text %q outside any structure element; draw it with a tagged method or as a header/footer artifact", ErrPDFUA, i+1, clip(r.Text))
			}
		}
		for _, im := range b.pc.pages[i].ImageRuns {
			if !im.HasMCID && !im.IsArtifact {
				return fmt.Errorf("%w: page %d has an image outside any structure element; use DrawTaggedPNG, DrawTaggedJPEG or DrawTaggedImageWith with alternative text", ErrPDFUA, i+1)
			}
		}
	}
	for i, fig := range b.tagging.Figures {
		if fig.Alt == "" {
			return fmt.Errorf("%w: figure %d on page %d has no alternative text", ErrPDFUA, i+1, fig.PageIndex+1)
		}
	}
	if err := checkHeadingLevels(b.tagging.Blocks); err != nil {
		return err
	}
	var fontErr error
	for _, obj := range objs {
		walkBuilt(obj, 0, func(o model.Object) {
			if fontErr != nil {
				return
			}
			if v, ok := o.(model.Dict); ok && v[model.Name("Type")] == model.Name("Font") {
				if err := checkPDFAFont(v, objs); err != nil {
					fontErr = fmt.Errorf("%w: %s", ErrPDFUA, trimPrefix(err.Error()))
				}
			}
		})
		if fontErr != nil {
			return fontErr
		}
	}
	return nil
}

// checkHeadingLevels requires headings to start at H1 and to descend one level at a
// time (ISO 14289-1 §7.4.2 and 14289-2 §8.2.5.7).
func checkHeadingLevels(blocks []tagged.Block) error {
	level := 0
	for _, blk := range blocks {
		role := string(blk.Role)
		if len(role) < 2 || role[0] != 'H' {
			continue
		}
		n, err := strconv.Atoi(role[1:])
		if err != nil {
			continue
		}
		if n > level+1 {
			return fmt.Errorf("%w: heading %s on page %d follows H%d; heading levels must not skip", ErrPDFUA, role, blk.PageIndex+1, level)
		}
		level = n
	}
	return nil
}

func clip(s string) string {
	r := []rune(s)
	if len(r) > 30 {
		return string(r[:30]) + "…"
	}
	return s
}

// trimPrefix strips the "pdf/a: " prefix of a font error so it can be re-wrapped.
func trimPrefix(msg string) string {
	const p = "pdf/a: "
	if len(msg) > len(p) && msg[:len(p)] == p {
		return msg[len(p):]
	}
	return msg
}
