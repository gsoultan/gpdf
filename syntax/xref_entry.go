// Copyright (c) 2026 Gembit Soultan Shirazi <gembit.soultan@gmail.com>

package syntax

// XRefEntry describes one entry in the cross-reference table.
// For type-2 (compressed) entries, Compressed is true and StreamObjectNumber/IndexInStream
// identify the object stream and position within it.
type XRefEntry struct {
	Offset             int64
	Generation         int
	InUse              bool
	Compressed         bool
	StreamObjectNumber int
	IndexInStream      int
}
