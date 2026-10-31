# gPDF — project overview

- Module: `github.com/gsoultan/gpdf`, Go 1.26 (uses range-over-int, `strings.SplitSeq`, `new(v)` on values).
- Only external dependency: `golang.org/x/crypto` (pbkdf2, used by the AES-256 handler).
- Purpose: pure-Go library to read PDFs (text, layout, tables, images, vectors), search/replace text,
  build new PDFs with a fluent builder (text, tables, images, forms, layers, tagged/PDF-UA, outlines),
  merge/split/extract pages, and write encrypted or linearized output. Writer emits `%PDF-2.0`.
- Size: ~250 Go files, ~27.6k lines. Largest packages: `doc` (7.0k), `reader` (6.5k), `tests/doc` (1.7k),
  `font/truetype` (1.2k), `writer` (1.1k).
- Entry points for users: `doc.Open`, `doc.OpenWithPassword`, `doc.OpenReader`, `doc.New()` builder,
  `doc.Merge`, `doc.Split`, `doc.SplitEvery`, `doc.Extract`. CLI examples in `cmd/*` (each takes an output path).
- Sample input committed at repo root: `CV CONTOH.pdf` (Canva, xref table, 1 page). `cmd/cv` recreates it.
- Last full review: 2026-10-31 (see memory `review_findings`).
