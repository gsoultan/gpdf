package doc

import (
	"errors"
	"regexp"
	"strings"

	"github.com/gsoultan/gpdf/reader"
)

// RedactArea is a rectangle on a page, in PDF user space (origin bottom left).
type RedactArea struct {
	Page                int
	X, Y, Width, Height float64
}

// RedactOptions selects the text Redact removes.
type RedactOptions struct {
	// Texts are substrings to remove (case-sensitive); every text run that
	// contains one is removed whole.
	Texts []string
	// Pattern, when set, removes every run it matches.
	Pattern *regexp.Regexp
	// Areas remove every run whose box intersects them.
	Areas []RedactArea
	// Fill is the colour of the box drawn over removed text; black by default.
	Fill *Color
	// NoFill leaves the removed text's place empty instead of drawing a box.
	NoFill bool
}

// Redact removes matching text from the page content of d, leaving a filled box
// where it was, and reports what changed. The removal is real: the text bytes
// are no longer in the content stream, and the text after a removed run keeps
// its place. Save the result with Save (a full rewrite); SaveIncremental is
// refused afterwards because it would keep the old bytes. Text inside form
// XObjects (shared drawings) is reported in SkippedInForms and left alone;
// annotations, bookmarks, metadata and font programs are not searched.
func Redact(d Document, opts RedactOptions) (reader.RedactionReport, error) {
	if len(opts.Texts) == 0 && opts.Pattern == nil && len(opts.Areas) == 0 {
		return reader.RedactionReport{}, errors.New("redact: nothing to look for")
	}
	match := func(page int, b reader.TextBlock) bool {
		for _, t := range opts.Texts {
			if t != "" && strings.Contains(b.Text, t) {
				return true
			}
		}
		if opts.Pattern != nil && opts.Pattern.MatchString(b.Text) {
			return true
		}
		for _, a := range opts.Areas {
			if a.Page != page {
				continue
			}
			if b.X < a.X+a.Width && b.X+b.Width > a.X && b.Y < a.Y+a.Height && b.Y+b.Height > a.Y {
				return true
			}
		}
		return false
	}
	var fill *[3]float64
	if !opts.NoFill {
		c := ColorBlack
		if opts.Fill != nil {
			c = *opts.Fill
		}
		fill = &[3]float64{c.R, c.G, c.B}
	}
	return reader.RedactContent(d, match, fill)
}
