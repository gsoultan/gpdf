// Copyright (c) 2026 Gembit Soultan Shirazi <gembit.soultan@gmail.com>

package reader

// TextStyle describes the visual appearance of a text run.
type TextStyle struct {
	FontName               string
	BaseFont               string
	FontSize               float64
	Bold                   bool
	Italic                 bool
	Monospace              bool
	Serif                  bool
	CharSpacing            float64
	WordSpacing            float64
	HorizontalScale        float64
	Leading                float64
	Rotation               float64
	ColorR, ColorG, ColorB float64
}
