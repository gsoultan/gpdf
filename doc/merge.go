package doc

import (
	"fmt"

	"github.com/gsoultan/gpdf/model"
)

// Merge combines multiple documents into a single new document.
// Pages from the documents are appended in the order they are provided.
func Merge(docs ...Document) (Document, error) {
	if len(docs) == 0 {
		return nil, fmt.Errorf("merge: no documents provided")
	}

	merged := &builtDocument{
		objects: make(map[int]model.Object),
		size:    1, // Start object numbers at 1
	}
	// The merged document gets a new catalog and page tree; Info comes from the first document.
	copyInfo(docs[0], merged)

	var allPages []model.Ref
	for i, srcDoc := range docs {
		pages, err := srcDoc.Pages()
		if err != nil {
			return nil, fmt.Errorf("merge: failed to get pages of document %d: %w", i, err)
		}
		refs, err := copyPages(srcDoc, merged, pages)
		if err != nil {
			return nil, fmt.Errorf("merge: document %d: %w", i, err)
		}
		allPages = append(allPages, refs...)
	}
	newPageTree(merged, allPages)
	return merged, nil
}
