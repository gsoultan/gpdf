// Copyright (c) 2026 Gembit Soultan Shirazi <gembit.soultan@gmail.com>

package reader

// PageLayout groups all positioned text blocks extracted from a single page,
// together with the page dimensions and any vector shapes (used for border detection).
type PageLayout struct {
	Page   int
	Width  float64
	Height float64
	Blocks []TextBlock
	Shapes []VectorShape // optional: vector shapes for table border detection
}
