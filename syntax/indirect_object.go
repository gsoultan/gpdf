// Copyright (c) 2026 Gembit Soultan Shirazi <gembit.soultan@gmail.com>

package syntax

import "github.com/gsoultan/gpdf/model"

// IndirectObject is an indirect object: object number, generation, and value.
type IndirectObject struct {
	ObjectNumber int
	Generation   int
	Value        model.Object
}
