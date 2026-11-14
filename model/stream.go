// Copyright (c) 2026 Gembit Soultan Shirazi <gembit.soultan@gmail.com>

package model

// Stream is a PDF stream: dictionary plus byte content. Content is decoded
// unless Encoded is set.
type Stream struct {
	Dict    Dict
	Content []byte
	// Encoded marks Content as already encoded with the dictionary's /Filter
	// chain (and its /DecodeParms), as it was stored in the source file. Writers
	// copy such content unchanged instead of re-encoding it, which keeps the
	// original compression and skips a decode/encode round trip when pages are
	// copied between files.
	Encoded bool
}

func (Stream) IsIndirect() bool { return false }
