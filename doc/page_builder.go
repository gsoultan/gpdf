// Copyright (c) 2026 Gembit Soultan Shirazi <gembit.soultan@gmail.com>

package doc

// PageBuilder configures page size and manages pages.
type PageBuilder interface {
	PageSize(width, height float64) *DocumentBuilder
	AddPage() *DocumentBuilder
}
