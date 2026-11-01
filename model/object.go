// Copyright (c) 2026 Gembit Soultan Shirazi <gembit.soultan@gmail.com>

package model

// Object represents a PDF value type (ISO 32000-2). Implementations are immutable
// value types: Boolean, Integer, Real, String, Name, Array, Dictionary, Stream, Null, Ref.
type Object interface {
	// IsIndirect returns true if this object is an indirect reference.
	IsIndirect() bool
}
