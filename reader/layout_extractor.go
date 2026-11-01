// Copyright (c) 2026 Gembit Soultan Shirazi <gembit.soultan@gmail.com>

package reader

// LayoutExtractor extracts positioned text blocks and detects tables from page content.
type LayoutExtractor interface {
	Layout() ([]PageLayout, error)
	Tables() ([][]Table, error)
}
