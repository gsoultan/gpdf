package doc

import (
	"crypto/md5"
	"fmt"
	"sort"
	"time"

	"github.com/gsoultan/gpdf/model"
)

// FileRelationship is the /AFRelationship of an associated file (ISO 32000-2 §14.13.5).
type FileRelationship string

const (
	RelationshipSource           FileRelationship = "Source"
	RelationshipData             FileRelationship = "Data"
	RelationshipAlternative      FileRelationship = "Alternative"
	RelationshipSupplement       FileRelationship = "Supplement"
	RelationshipEncryptedPayload FileRelationship = "EncryptedPayload"
	RelationshipFormData         FileRelationship = "FormData"
	RelationshipSchema           FileRelationship = "Schema"
	RelationshipUnspecified      FileRelationship = "Unspecified"
)

// Attachment is a file embedded in the document (ISO 32000-2 §7.11.4) and listed as an
// associated file (§14.13). It appears in the reader's attachment panel and, through
// /AF, is tied to the document or to one page.
type Attachment struct {
	// Name is the file name shown to the user (Unicode is fine).
	Name string
	// MIME is the media type, such as "application/xml". Default application/octet-stream.
	MIME        string
	Description string
	// Relationship says what the file is to the document. Default Unspecified.
	Relationship FileRelationship
	// Modified, when set, is written as the file's modification date.
	Modified time.Time
	Data     []byte
	// CollectionItem holds the values for the document's collection schema
	// (see Collection): string, int, float64 or time-formatted string per field key.
	CollectionItem map[string]any

	pageIndex int // -1 for a document-level attachment
}

// AttachFile embeds a file and associates it with the document (catalog /AF and the
// EmbeddedFiles name tree). PDF/A-2 rejects it (only PDF/A-3 and -4f allow arbitrary
// embedded files).
func (b *DocumentBuilder) AttachFile(a Attachment) *DocumentBuilder {
	a.pageIndex = -1
	b.addAttachment(a)
	return b
}

// AttachFile embeds a file and associates it with this page (page /AF) as well as
// listing it in the document's EmbeddedFiles name tree.
func (p *Page) AttachFile(a Attachment) *Page {
	a.pageIndex = p.pageIndex
	p.builder.addAttachment(a)
	return p
}

func (b *DocumentBuilder) addAttachment(a Attachment) {
	if a.Name == "" {
		b.setErr(fmt.Errorf("attachment needs a name"))
		return
	}
	if a.MIME == "" {
		a.MIME = "application/octet-stream"
	}
	if a.Relationship == "" {
		a.Relationship = RelationshipUnspecified
	}
	b.attachments = append(b.attachments, a)
}

// buildAttachments writes the embedded file streams, file specifications, the
// EmbeddedFiles name tree and the /AF arrays.
func (b *DocumentBuilder) buildAttachments(objs map[int]model.Object, catalogDict model.Dict, pageNums []int, nextNum int) int {
	if len(b.attachments) == 0 {
		return nextNum
	}
	pdf20 := b.metadata.PDF20
	type named struct {
		name string
		ref  model.Ref
	}
	var entries []named
	var docAF model.Array
	for i, a := range b.attachments {
		streamNum := nextNum
		nextNum++
		specNum := nextNum
		nextNum++
		params := model.Dict{
			model.Name("Size"):     model.Integer(int64(len(a.Data))),
			model.Name("CheckSum"): model.HexString(md5sum(a.Data)),
		}
		if !a.Modified.IsZero() {
			params[model.Name("ModDate")] = model.String(model.FormatDate(a.Modified, pdf20))
		}
		objs[streamNum] = &model.Stream{
			Dict: model.Dict{
				model.Name("Type"):    model.Name("EmbeddedFile"),
				model.Name("Subtype"): model.Name(a.MIME),
				model.Name("Params"):  params,
				model.Name("Length"):  model.Integer(int64(len(a.Data))),
				model.Name("Filter"):  model.Name("FlateDecode"),
			},
			Content: a.Data,
		}
		spec := model.Dict{
			model.Name("Type"):           model.Name("Filespec"),
			model.Name("F"):              model.String(asciiFileName(a.Name)),
			model.Name("UF"):             model.String(a.Name),
			model.Name("AFRelationship"): model.Name(a.Relationship),
			model.Name("EF"): model.Dict{
				model.Name("F"):  model.Ref{ObjectNumber: streamNum},
				model.Name("UF"): model.Ref{ObjectNumber: streamNum},
			},
		}
		if a.Description != "" {
			spec[model.Name("Desc")] = model.String(a.Description)
		}
		if ci := collectionItem(a.CollectionItem, pdf20); ci != nil {
			spec[model.Name("CI")] = ci
		}
		objs[specNum] = spec
		ref := model.Ref{ObjectNumber: specNum}
		// Name tree keys must be unique; number duplicates.
		key := a.Name
		for _, e := range entries {
			if e.name == key {
				key = fmt.Sprintf("%s#%d", a.Name, i)
				break
			}
		}
		entries = append(entries, named{key, ref})
		if a.pageIndex >= 0 && a.pageIndex < len(pageNums) {
			if pd, ok := objs[pageNums[a.pageIndex]].(model.Dict); ok {
				af, _ := pd[model.Name("AF")].(model.Array)
				pd[model.Name("AF")] = append(af, ref)
			}
		} else {
			docAF = append(docAF, ref)
		}
	}
	// A name tree's /Names array is sorted by key (byte order).
	sort.Slice(entries, func(i, j int) bool { return entries[i].name < entries[j].name })
	names := make(model.Array, 0, 2*len(entries))
	for _, e := range entries {
		names = append(names, model.String(e.name), e.ref)
	}
	treeNum := nextNum
	nextNum++
	objs[treeNum] = model.Dict{model.Name("Names"): names}
	namesDict, _ := catalogDict[model.Name("Names")].(model.Dict)
	if namesDict == nil {
		namesDict = model.Dict{}
	}
	namesDict[model.Name("EmbeddedFiles")] = model.Ref{ObjectNumber: treeNum}
	catalogDict[model.Name("Names")] = namesDict
	if len(docAF) > 0 {
		catalogDict[model.Name("AF")] = docAF
	}
	return nextNum
}

func md5sum(b []byte) []byte {
	s := md5.Sum(b)
	return s[:]
}

// asciiFileName gives the /F form of a file name: PDF 1.x readers that ignore /UF
// still get something usable.
func asciiFileName(s string) string {
	out := make([]byte, 0, len(s))
	for _, r := range s {
		if r < 0x80 && r >= 0x20 {
			out = append(out, byte(r))
		} else {
			out = append(out, '_')
		}
	}
	return string(out)
}
