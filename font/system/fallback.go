// Copyright (c) 2026 Gembit Soultan Shirazi <gembit.soultan@gmail.com>

package system

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	"github.com/gsoultan/gpdf/font"
	"github.com/gsoultan/gpdf/font/truetype"
)

// Fallback finds installed fonts that cover characters a document's own
// fonts lack. Font files are parsed lazily, once, on the first lookup that
// reaches them, and results are cached per rune.
type Fallback struct {
	scanner FontScanner

	mu      sync.Mutex
	paths   []string
	scanned bool
	parsed  map[string][]*truetype.Font
	byRune  map[rune]font.EmbeddableFont
}

// NewFallback returns a Fallback over the fonts the scanner finds.
func NewFallback(scanner FontScanner) *Fallback {
	return &Fallback{scanner: scanner, parsed: map[string][]*truetype.Font{}, byRune: map[rune]font.EmbeddableFont{}}
}

var (
	defaultFallbackOnce sync.Once
	defaultFallback     *Fallback
)

// DefaultFallback returns a process-wide Fallback over the default scanner.
func DefaultFallback() *Fallback {
	defaultFallbackOnce.Do(func() { defaultFallback = NewFallback(GetDefaultScanner()) })
	return defaultFallback
}

// FontFor returns an installed font that has a glyph for r, or nil. Upright
// regular-weight text faces are preferred over bold, italic, condensed,
// monospace and colour-emoji faces.
func (fb *Fallback) FontFor(r rune) font.EmbeddableFont {
	fb.mu.Lock()
	defer fb.mu.Unlock()
	if f, ok := fb.byRune[r]; ok {
		return f
	}
	if !fb.scanned {
		fb.scanned = true
		paths, _ := fb.scanner.Scan()
		sort.SliceStable(paths, func(i, j int) bool { return fileScore(paths[i]) < fileScore(paths[j]) })
		fb.paths = paths
	}
	var best font.EmbeddableFont
	for _, p := range fb.paths {
		faces, ok := fb.parsed[p]
		if !ok {
			faces = parseFaces(p)
			fb.parsed[p] = faces
		}
		for _, f := range faces {
			if f.Contains(r) {
				best = f
				break
			}
		}
		if best != nil {
			break
		}
	}
	fb.byRune[r] = best
	return best
}

// fileScore ranks font files: lower is tried first.
func fileScore(p string) int {
	name := strings.ToLower(filepath.Base(p))
	score := 0
	for _, bad := range []string{"bold", "italic", "oblique", "condensed", "light", "thin", "black", "mono", "emoji", "display", "ui-", "math", "symbols"} {
		if strings.Contains(name, bad) {
			score += 10
		}
	}
	if strings.Contains(name, "noto") || strings.Contains(name, "dejavu") {
		score -= 5
	}
	if strings.HasSuffix(name, ".ttc") {
		score += 3
	}
	return score
}

func parseFaces(p string) []*truetype.Font {
	data, err := os.ReadFile(p)
	if err != nil {
		return nil
	}
	if fonts, err := truetype.ParseCollection(data); err == nil {
		return fonts
	}
	f, err := truetype.Parse(data)
	if err != nil {
		return nil
	}
	return []*truetype.Font{f}
}
