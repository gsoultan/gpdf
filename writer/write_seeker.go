// Copyright (c) 2026 Gembit Soultan Shirazi <gembit.soultan@gmail.com>

package writer

import "io"

// WriteSeeker is the subset of io.Writer required for linearized write (must support Seek).
type WriteSeeker interface {
	io.Writer
	io.Seeker
}
