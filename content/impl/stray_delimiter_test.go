package impl

import (
	"testing"
	"time"
)

// Regression (found by FuzzReadDocument): a stray ')' made the tokenizer return an empty
// operator without consuming the byte, so Parse looped forever on content such as
// "BT (x) Tj ) ET".
func TestParseStrayDelimiterTerminates(t *testing.T) {
	done := make(chan struct{})
	go func() {
		defer close(done)
		for _, src := range []string{"BT /F1 12 Tf 20 100 Td Seed) Tj ET", ")", "} { ) ( ", "BT ) ) ) ET"} {
			ops, err := NewStreamParser().Parse([]byte(src))
			if err != nil {
				t.Errorf("%q: %v", src, err)
			}
			if len(ops) == 0 {
				t.Errorf("%q: no ops", src)
			}
		}
	}()
	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("parser hung on a stray delimiter")
	}
}
