package doc

import "github.com/gsoultan/gpdf/reader"

// ContentReader extracts text content from a PDF document.
type ContentReader interface {
	ReadContent() (string, error)
	ReadContentPerPage() ([]string, error)
	// ReadContentInReadingOrder returns the text in reading order: through the
	// structure tree of a tagged document, or through layout analysis (columns,
	// paragraphs) of an untagged one. See reader.ExtractTextInReadingOrder.
	ReadContentInReadingOrder() (string, error)
	// ReadStructuredText returns the reading-order text as elements with their
	// structure role and page.
	ReadStructuredText() ([]reader.TextElement, error)
}
