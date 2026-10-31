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
}

// BuildInfoDict builds the PDF /Info dictionary from the stored metadata fields.
func (s *Support) BuildInfoDict() model.Dict {
	info := model.Dict{}
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
	if !s.CreationDate.IsZero() {
		d := model.String(PDFDate(s.CreationDate))
		info[model.Name("CreationDate")] = d
		info[model.Name("ModDate")] = d
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
