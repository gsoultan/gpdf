// Copyright (c) 2026 Gembit Soultan Shirazi <gembit.soultan@gmail.com>

package reader

// TableCell is a single cell within a detected table, containing the merged
// text of all TextBlocks that fall within its column/row bucket.
type TableCell struct {
	Row   int
	Col   int
	Text  string
	Style TextStyle // style of the first contributing TextBlock
}
