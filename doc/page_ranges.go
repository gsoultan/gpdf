// Copyright (c) 2026 Gembit Soultan Shirazi <gembit.soultan@gmail.com>

package doc

import (
	"fmt"
	"strconv"
	"strings"
)

// ParsePageRanges turns a page selection such as "1-3,5,8-" into 0-based page
// indices for a document of pageCount pages, in the order written. Pages are
// numbered from 1. Each comma-separated item is a page ("5"), a range ("1-3"),
// a descending range ("3-1"), a range to the last page ("8-"), or a range from
// the first page ("-3"); "last" names the last page anywhere a number may appear
// ("last-1" reverses the document). An empty spec, or "all", selects every page.
// Pages may repeat.
func ParsePageRanges(spec string, pageCount int) ([]int, error) {
	if pageCount <= 0 {
		return nil, fmt.Errorf("page ranges: document has no pages")
	}
	spec = strings.TrimSpace(spec)
	if spec == "" || strings.EqualFold(spec, "all") {
		out := make([]int, pageCount)
		for i := range out {
			out[i] = i
		}
		return out, nil
	}
	page := func(s string, def int) (int, error) {
		s = strings.TrimSpace(s)
		switch {
		case s == "":
			return def, nil
		case strings.EqualFold(s, "last"):
			return pageCount, nil
		}
		n, err := strconv.Atoi(s)
		if err != nil {
			return 0, fmt.Errorf("page ranges: %q is not a page number", s)
		}
		if n < 1 || n > pageCount {
			return 0, fmt.Errorf("page ranges: page %d is out of range (document has %d pages)", n, pageCount)
		}
		return n, nil
	}
	var out []int
	for item := range strings.SplitSeq(spec, ",") {
		item = strings.TrimSpace(item)
		if item == "" {
			return nil, fmt.Errorf("page ranges: empty item in %q", spec)
		}
		from, to, isRange := strings.Cut(item, "-")
		if !isRange {
			n, err := page(from, 0)
			if err != nil {
				return nil, err
			}
			out = append(out, n-1)
			continue
		}
		if strings.TrimSpace(from) == "" && strings.TrimSpace(to) == "" {
			return nil, fmt.Errorf("page ranges: %q has no page numbers", item)
		}
		a, err := page(from, 1)
		if err != nil {
			return nil, err
		}
		b, err := page(to, pageCount)
		if err != nil {
			return nil, err
		}
		if len(out)+abs(b-a)+1 > maxSelectedPages(pageCount) {
			return nil, fmt.Errorf("page ranges: selection is longer than %d pages", maxSelectedPages(pageCount))
		}
		step := 1
		if b < a {
			step = -1
		}
		for n := a; ; n += step {
			out = append(out, n-1)
			if n == b {
				break
			}
		}
	}
	return out, nil
}

// maxSelectedPages bounds a selection so repeated ranges cannot ask for an
// unbounded number of page copies.
func maxSelectedPages(pageCount int) int { return max(pageCount*10, 10000) }

func abs(n int) int {
	if n < 0 {
		return -n
	}
	return n
}
