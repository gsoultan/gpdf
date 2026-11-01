// Copyright (c) 2026 Gembit Soultan Shirazi <gembit.soultan@gmail.com>

package model

// Null represents the PDF null object.
type Null struct{}

func (Null) IsIndirect() bool { return false }
