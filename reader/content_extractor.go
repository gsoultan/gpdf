// Copyright (c) 2026 Gembit Soultan Shirazi <gembit.soultan@gmail.com>

package reader

import "github.com/gsoultan/gpdf/model"

// ContentExtractor extracts text, searches keywords, and replaces text in content streams.
type ContentExtractor interface {
	Content() (string, error)
	ContentPerPage() ([]string, error)
	Search(keywords ...string) ([]model.SearchResult, error)
	Replace(old, new string) error
	Replaces(replacements map[string]string) error
}
