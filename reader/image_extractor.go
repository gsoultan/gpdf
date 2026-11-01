// Copyright (c) 2026 Gembit Soultan Shirazi <gembit.soultan@gmail.com>

package reader

// ImageExtractor extracts image metadata and raw bytes from a PDF document.
type ImageExtractor interface {
	ImagesPerPage() ([][]ImageInfo, error)
	Images() ([]ImageInfo, error)
}
