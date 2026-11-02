// Copyright (c) 2026 Gembit Soultan Shirazi <gembit.soultan@gmail.com>

package syntax

// Tokenizer produces a sequence of PDF tokens from input.
type Tokenizer interface {
	Next() (Token, error)
	CurrentOffset() int64
}
