// Copyright (c) 2026 Gembit Soultan Shirazi <gembit.soultan@gmail.com>

package model

// SearchResult holds the 0-based page indices and byte offsets where a keyword was found.
// Indices maps each page index to the start-byte positions of the keyword on that page.
type SearchResult struct {
	Keyword string
	Pages   []int
	Indices map[int][]int
}
