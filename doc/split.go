package doc

import (
	"fmt"

	"github.com/gsoultan/gpdf/model"
)

// Split splits a document into multiple documents based on the provided page ranges.
// Each range is a slice of 0-based page indices.
func Split(src Document, ranges ...[]int) ([]Document, error) {
	if len(ranges) == 0 {
		return nil, fmt.Errorf("split: no ranges provided")
	}

	allPages, err := src.Pages()
	if err != nil {
		return nil, fmt.Errorf("split: failed to get pages: %w", err)
	}

	var docs []Document
	for i, r := range ranges {
		doc, err := splitOne(src, allPages, r)
		if err != nil {
			return nil, fmt.Errorf("split: failed at range %d: %w", i, err)
		}
		docs = append(docs, doc)
	}

	return docs, nil
}

// SplitEvery splits a document into multiple documents, each containing at most n pages.
func SplitEvery(src Document, n int) ([]Document, error) {
	if n <= 0 {
		return nil, fmt.Errorf("split: n must be positive")
	}

	allPages, err := src.Pages()
	if err != nil {
		return nil, fmt.Errorf("split: failed to get pages: %w", err)
	}

	numPages := len(allPages)
	var ranges [][]int
	for i := 0; i < numPages; i += n {
		end := i + n
		if end > numPages {
			end = numPages
		}
		r := make([]int, end-i)
		for j := range r {
			r[j] = i + j
		}
		ranges = append(ranges, r)
	}

	return Split(src, ranges...)
}

// Extract extracts a contiguous range of pages from the document.
// The range is inclusive of start and exclusive of end (start <= page < end).
func Extract(src Document, start, end int) (Document, error) {
	if start < 0 || end < start {
		return nil, fmt.Errorf("split: invalid range [%d, %d)", start, end)
	}

	allPages, err := src.Pages()
	if err != nil {
		return nil, fmt.Errorf("split: failed to get pages: %w", err)
	}

	if start >= len(allPages) {
		return nil, fmt.Errorf("split: start index %d out of range", start)
	}
	if end > len(allPages) {
		end = len(allPages)
	}

	r := make([]int, end-start)
	for i := range r {
		r[i] = start + i
	}

	return splitOne(src, allPages, r)
}

func splitOne(src Document, allPages []model.Page, indices []int) (Document, error) {
	if len(indices) == 0 {
		return nil, fmt.Errorf("split: no indices provided")
	}
	selected := make([]model.Page, 0, len(indices))
	for _, idx := range indices {
		if idx < 0 || idx >= len(allPages) {
			return nil, fmt.Errorf("split: page index %d out of range (0-%d)", idx, len(allPages)-1)
		}
		selected = append(selected, allPages[idx])
	}

	dst := &builtDocument{
		objects: make(map[int]model.Object),
		size:    1,
	}
	copyInfo(src, dst)
	refs, err := copyPages(src, dst, selected)
	if err != nil {
		return nil, fmt.Errorf("split: %w", err)
	}
	newPageTree(dst, refs)
	return dst, nil
}
