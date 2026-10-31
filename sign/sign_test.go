package sign

import (
	"bytes"
	"errors"
	"testing"
)

// Regression: Sign copied the input and returned nil, so callers thought it was signed.
func TestSignReportsNotImplemented(t *testing.T) {
	var out bytes.Buffer
	err := Sign(bytes.NewReader([]byte("%PDF-1.7")), 8, &out, nil, SignOptions{})
	if !errors.Is(err, ErrNotImplemented) {
		t.Fatalf("err = %v, want ErrNotImplemented", err)
	}
	if out.Len() != 0 {
		t.Fatal("Sign wrote output although it did not sign")
	}
}
