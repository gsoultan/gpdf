package sign

import (
	"errors"
	"io"
	"os"
)

// ErrNotImplemented is returned by Sign and SignFile: digital signatures are not
// implemented yet. (They used to copy the input unchanged and return nil, so callers
// believed the file had been signed.)
var ErrNotImplemented = errors.New("sign: digital signatures are not implemented")

// Sign will read a PDF from r (size bytes) and write a signed PDF to out. It currently
// writes nothing and returns ErrNotImplemented.
func Sign(r io.ReaderAt, size int64, out io.Writer, s Signer, opts SignOptions) error {
	return ErrNotImplemented
}

// SignFile opens an existing PDF from path, applies signing via Sign, and writes
// the result to out.
func SignFile(path string, out io.Writer, s Signer, opts SignOptions) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	info, err := f.Stat()
	if err != nil {
		return err
	}
	return Sign(f, info.Size(), out, s, opts)
}
