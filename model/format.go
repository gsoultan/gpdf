// Copyright (c) 2026 Gembit Soultan Shirazi <gembit.soultan@gmail.com>

package model

import (
	"math"
	"strconv"
	"strings"
)

// FormatReal formats a PDF real number. PDF has no exponent syntax (1e+06 is not a
// number), so values are written in plain decimal notation with at most 6 fractional
// digits, which is finer than 1/1000 of a point.
func FormatReal(f float64) string {
	if math.IsNaN(f) || math.IsInf(f, 0) {
		return "0"
	}
	s := strconv.FormatFloat(f, 'f', 6, 64)
	if strings.ContainsRune(s, '.') {
		s = strings.TrimRight(s, "0")
		s = strings.TrimSuffix(s, ".")
	}
	if s == "-0" || s == "" {
		return "0"
	}
	return s
}
