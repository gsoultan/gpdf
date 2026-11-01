package model

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// ErrDate is wrapped by ParseDate errors.
var ErrDate = errors.New("invalid PDF date")

// ParseDate reads a PDF date string (ISO 32000-2 §7.9.4): D:YYYYMMDDHHmmSSOHH'mm, where
// everything after the year is optional, O is +, - or Z, and the apostrophes that PDF 1.x
// wrote around the minutes are accepted whether present or not. A date without an
// offset is returned in UTC with no location information lost (time.Time has none).
func ParseDate(s string) (time.Time, error) {
	orig := s
	s = strings.TrimSpace(s)
	s = strings.TrimPrefix(s, "D:")
	num := func(n int, lo, hi, def int) (int, error) {
		if len(s) < n {
			if len(s) == 0 {
				return def, nil
			}
			return 0, fmt.Errorf("%w: %q", ErrDate, orig)
		}
		v, err := strconv.Atoi(s[:n])
		if err != nil || v < lo || v > hi {
			return 0, fmt.Errorf("%w: %q", ErrDate, orig)
		}
		s = s[n:]
		return v, nil
	}
	if len(s) < 4 {
		return time.Time{}, fmt.Errorf("%w: %q", ErrDate, orig)
	}
	year, err := num(4, 0, 9999, 0)
	if err != nil {
		return time.Time{}, err
	}
	month, err := num(2, 1, 12, 1)
	if err != nil {
		return time.Time{}, err
	}
	day, err := num(2, 1, 31, 1)
	if err != nil {
		return time.Time{}, err
	}
	hour, err := num(2, 0, 23, 0)
	if err != nil {
		return time.Time{}, err
	}
	minute, err := num(2, 0, 59, 0)
	if err != nil {
		return time.Time{}, err
	}
	sec, err := num(2, 0, 59, 0)
	if err != nil {
		return time.Time{}, err
	}
	loc := time.UTC
	if len(s) > 0 {
		sign := s[0]
		s = s[1:]
		switch sign {
		case 'Z':
			if strings.Trim(s, "0'") != "" {
				return time.Time{}, fmt.Errorf("%w: %q", ErrDate, orig)
			}
		case '+', '-':
			s = strings.ReplaceAll(s, "'", "")
			oh, err := num(2, 0, 23, 0)
			if err != nil {
				return time.Time{}, err
			}
			om, err := num(2, 0, 59, 0)
			if err != nil {
				return time.Time{}, err
			}
			off := oh*3600 + om*60
			if sign == '-' {
				off = -off
			}
			if off != 0 {
				loc = time.FixedZone("", off)
			}
		default:
			return time.Time{}, fmt.Errorf("%w: %q", ErrDate, orig)
		}
	}
	return time.Date(year, time.Month(month), day, hour, minute, sec, 0, loc), nil
}

// FormatDate writes t as a PDF date string. With pdf20 set it uses the ISO 32000-2
// form (Z for UTC, no apostrophe after the minutes); otherwise the PDF 1.x form
// (+HH'mm'), which readers of every version accept.
func FormatDate(t time.Time, pdf20 bool) string {
	base := "D:" + t.Format("20060102150405")
	_, off := t.Zone()
	if pdf20 && off == 0 {
		return base + "Z"
	}
	sign := '+'
	if off < 0 {
		sign = '-'
		off = -off
	}
	if pdf20 {
		return fmt.Sprintf("%s%c%02d'%02d", base, sign, off/3600, off%3600/60)
	}
	return fmt.Sprintf("%s%c%02d'%02d'", base, sign, off/3600, off%3600/60)
}
