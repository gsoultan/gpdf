package doc

import (
	"errors"
	"io"

	"github.com/gsoultan/gpdf/reader/file"
	"github.com/gsoultan/gpdf/writer"
)

// ErrNoBaseFile is returned by SaveIncremental on a document that was built in
// memory rather than opened from a file.
var ErrNoBaseFile = errors.New("incremental save needs a document opened from a file")

// ErrRedacted is returned by SaveIncremental after Redact: an incremental update
// would keep the removed bytes in the file.
var ErrRedacted = file.ErrRedacted

// Saver writes a PDF document in various formats and encryption modes.
type Saver interface {
	Save(w io.Writer) error
	SaveWithPassword(w io.Writer, userPassword, ownerPassword string) error
	SaveWithAES256Password(w io.Writer, userPassword, ownerPassword string) error
	SaveLinearized(ws writer.WriteSeeker) error
	// SaveIncremental writes the file as it was opened followed by an update
	// section holding only the objects changed since (ISO 32000-2 §7.5.6), so
	// earlier bytes, and any signatures over them, stay intact. Documents built
	// in memory return ErrNoBaseFile; redacted ones ErrRedacted.
	SaveIncremental(w io.Writer) error
}
