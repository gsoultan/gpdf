// Copyright (c) 2026 Gembit Soultan Shirazi <gembit.soultan@gmail.com>

package writer

import "github.com/gsoultan/gpdf/model"

// Document is the input to the writer: object graph and trailer (root, size).
type Document interface {
	Trailer() model.Trailer
	Resolve(ref model.Ref) (model.Object, error)
	ObjectNumbers() []int
}

// VersionedDocument is implemented by documents that need a particular header version,
// such as PDF/A-2 output ("1.7"). Write uses 2.0 otherwise.
type VersionedDocument interface {
	PDFVersion() string
}

func headerFor(doc Document) string {
	if v, ok := doc.(VersionedDocument); ok {
		switch v.PDFVersion() {
		case "1.4", "1.5", "1.6", "1.7":
			return "%PDF-" + v.PDFVersion() + "\n%\xE2\xE3\xCF\xD3\n"
		}
	}
	return headerPDF20
}
