# gPDF — project overview

- Module: `github.com/gsoultan/gpdf`, Go 1.26 (uses range-over-int, `strings.SplitSeq`, `new(v)` on values).
- No external dependencies (golang.org/x/crypto was dropped on 2026-10-31 with the ISO 32000-2 AES-256 rewrite).
- Purpose: pure-Go library to read PDFs (text, layout, tables, images, vectors), search/replace text,
  build new PDFs with a fluent builder (text, tables, images, forms, layers, tagged/PDF-UA, outlines),
  merge/split/extract pages, and write encrypted or linearized output. Writer emits `%PDF-2.0`.
- Size: ~250 Go files, ~27.6k lines. Largest packages: `doc` (7.0k), `reader` (6.5k), `tests/doc` (1.7k),
  `font/truetype` (1.2k), `writer` (1.1k).
- Entry points for users: `doc.Open`, `doc.OpenWithPassword`, `doc.OpenReader`, `doc.New()` builder,
  `doc.Merge`, `doc.Split`, `doc.SplitEvery`, `doc.Extract`. CLI examples in `cmd/*` (each takes an output path).
- Sample input committed at repo root: `CV CONTOH.pdf` (Canva, xref table, 1 page). `cmd/cv` recreates it.
- PDF 2.0 audit 2026-11-01 (round 3): text strings, dates, SASLprep, XMP, attachments/AF, PDF/A-4, form appearances,
  page output intents, document parts, black point compensation. See `review_findings` for what is still open.
- Last full review: 2026-10-31 (see memory `review_findings`); round 2 the same day added headers/footers, PDF/A,
  hint streams, fuzz targets (`FuzzReadDocument`, `FuzzTokenizer`, `FuzzParseObject`) and benchmarks
  (`go test -bench . ./syntax/impl ./tests/reader ./tests/doc`).
- Roadmap 2026-11-01 (branch claude/project-thread-tv91nd, PR #4), six phases all done: PAdES/collections/3D/geo;
  fonts (standard-14 substitutes, CFF, fallback, shaping); tagged PDF + PDF/UA; object streams, image
  recompression, streaming writer; reading order, incremental updates, redaction; qpdf + veraPDF in CI, nightly
  fuzzing, benchmarks. Output validated with veraPDF 1.28.2 and qpdf (via pikepdf 12.4) in that session.
