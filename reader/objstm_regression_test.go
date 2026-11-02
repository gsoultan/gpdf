// Copyright (c) 2026 Gembit Soultan Shirazi <gembit.soultan@gmail.com>

package reader

import (
	"testing"

	"github.com/gsoultan/gpdf/model"
)

// Regression: the object number in an object stream header was never checked, so a wrong
// index silently returned another object.
func TestExtractFromObjectStreamChecksObjectNumber(t *testing.T) {
	content := []byte("10 0 11 4 (a) (b)")
	stm := &model.Stream{Dict: model.Dict{"N": model.Integer(2), "First": model.Integer(10)}, Content: content}
	c := &documentCore{}
	// Index 0 holds object 10, but object 11 is requested: the header is searched instead.
	obj, err := c.extractFromObjectStream(stm, 5, 0, 11)
	if err != nil || obj != model.String("b") {
		t.Fatalf("got %v, %v; want (b)", obj, err)
	}
	if _, err := c.extractFromObjectStream(stm, 5, 0, 12); err == nil {
		t.Fatal("expected an error for an object that is not in the stream")
	}
}
