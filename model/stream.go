// Copyright (c) 2026 Gembit Soultan Shirazi <gembit.soultan@gmail.com>

package model

// Stream is a PDF stream: dictionary plus raw (decoded) byte content.
type Stream struct {
	Dict    Dict
	Content []byte
}

func (Stream) IsIndirect() bool { return false }
