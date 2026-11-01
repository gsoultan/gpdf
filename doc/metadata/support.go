package metadata

import (
	"time"

	"github.com/gsoultan/gpdf/model"
)

// Support owns document metadata fields for a document builder.
type Support struct {
	Title       string
	Author      string
	Subject     string
	Keywords    string
	Creator     string
	Producer    string
	MetadataXMP []byte
	Lang        string
	// CreationDate, when set, is written as /CreationDate and /ModDate.
	CreationDate time.Time
	// PDF20 selects the ISO 32000-2 date form and UTF-8 text strings.
	PDF20 bool
	// DatesOnly marks PDF/A-4 output, where the Info dictionary is deprecated
	// and the XMP packet carries every entry: the builder writes no Info at all
	// and BuildInfoDict keeps only the dates.
	DatesOnly bool
}

// HasFields reports whether any descriptive metadata is set.
func (s *Support) HasFields() bool {
	return s.Title != "" || s.Author != "" || s.Subject != "" || s.Keywords != "" || s.Creator != "" || s.Producer != ""
}

// BuildInfoDict builds the PDF /Info dictionary from the stored metadata fields.
func (s *Support) BuildInfoDict() model.Dict {
	info := model.Dict{}
	if !s.CreationDate.IsZero() {
		d := model.String(model.FormatDate(s.CreationDate, s.PDF20))
		info[model.Name("CreationDate")] = d
		info[model.Name("ModDate")] = d
	}
	if s.DatesOnly {
		return info
	}
	if s.Title != "" {
		info[model.Name("Title")] = model.String(s.Title)
	}
	if s.Author != "" {
		info[model.Name("Author")] = model.String(s.Author)
	}
	if s.Subject != "" {
		info[model.Name("Subject")] = model.String(s.Subject)
	}
	if s.Keywords != "" {
		info[model.Name("Keywords")] = model.String(s.Keywords)
	}
	if s.Creator != "" {
		info[model.Name("Creator")] = model.String(s.Creator)
	}
	if s.Producer != "" {
		info[model.Name("Producer")] = model.String(s.Producer)
	}
	return info
}

// BuildMetadataStream writes the XMP metadata stream object into objs and
// wires it into the catalog via /Metadata. Returns the updated next object number.
func (s *Support) BuildMetadataStream(objs map[int]model.Object, catalogDict model.Dict, nextNum int) int {
	if len(s.MetadataXMP) == 0 {
		return nextNum
	}
	metaNum := nextNum
	nextNum++
	objs[metaNum] = &model.Stream{
		Dict: model.Dict{
			model.Name("Type"):    model.Name("Metadata"),
			model.Name("Subtype"): model.Name("XML"),
			model.Name("Length"):  model.Integer(int64(len(s.MetadataXMP))),
		},
		Content: s.MetadataXMP,
	}
	catalogDict[model.Name("Metadata")] = model.Ref{ObjectNumber: metaNum, Generation: 0}
	return nextNum
}
