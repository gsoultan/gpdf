// Copyright (c) 2026 Gembit Soultan Shirazi <gembit.soultan@gmail.com>

package doc

import "github.com/gsoultan/gpdf/reader"

// LayoutReader extracts positioned text blocks and detects tables from page content.
type LayoutReader interface {
	ReadLayout() ([]reader.PageLayout, error)
	ReadTables() ([][]reader.Table, error)
}
