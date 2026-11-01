// Copyright (c) 2026 Gembit Soultan Shirazi <gembit.soultan@gmail.com>

package writer

import (
	"crypto/rand"
	"io"

	"github.com/gsoultan/gpdf/stream"
)

// WriterOptions configures a PDFWriter. A nil pointer field takes its default.
type WriterOptions struct {
	Filters      stream.FilterRegistry
	RandomSource io.Reader
	// ValidateGraph checks the object graph (dangling references, cycles in the
	// page tree) before writing. Default true.
	ValidateGraph *bool
	// ObjectStreams packs non-stream objects into object streams and writes a
	// cross-reference stream instead of a table (ISO 32000-2 §7.5.7, §7.5.8), which
	// makes files markedly smaller. Default true. It is ignored, and a classic table
	// written, for documents whose header version is below 1.5.
	ObjectStreams *bool
}

func DefaultWriterOptions() WriterOptions {
	return WriterOptions{
		RandomSource:  rand.Reader,
		ValidateGraph: new(true),
		ObjectStreams: new(true),
	}
}

func normalizeWriterOptions(options WriterOptions) WriterOptions {
	defaults := DefaultWriterOptions()
	if options.RandomSource == nil {
		options.RandomSource = defaults.RandomSource
	}
	if options.ValidateGraph == nil {
		options.ValidateGraph = defaults.ValidateGraph
	}
	if options.ObjectStreams == nil {
		options.ObjectStreams = defaults.ObjectStreams
	}
	return options
}

func boolRef(v bool) *bool {
	return &v
}
