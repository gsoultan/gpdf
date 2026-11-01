package doc

import (
	"github.com/gsoultan/gpdf/model"
	"github.com/gsoultan/gpdf/reader"
)

// CatalogReader provides access to the document catalog, page tree, info dict, and metadata.
type CatalogReader interface {
	Catalog() (*model.Catalog, error)
	Pages() ([]model.Page, error)
	Info() (model.Dict, error)
	MetadataStream() ([]byte, error)
	StartXRefOffset() int64
	Trailer() model.Trailer
	// Attachments lists the embedded files (EmbeddedFiles name tree and /AF).
	Attachments() ([]reader.EmbeddedFile, error)
}
