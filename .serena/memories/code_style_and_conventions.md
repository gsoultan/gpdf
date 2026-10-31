# gPDF — code style and conventions

- Interfaces at package root, implementations in `impl/` subpackages or `*_impl.go` files; constructors `NewX` return the interface.
- Small focused files per concern (e.g. `builder_text.go`, `builder_image.go`, `extract_tables.go`).
- Errors: `fmt.Errorf` with `%w`; sentinel errors in `reader/errors.go` and `writer/errors.go` (`ErrParseLimitExceeded`, `ErrStreamDecodeLimit`, `ErrMalformedXRef`, `ErrInvalidDocumentGraph`, ...).
- Modern Go idioms used freely: `for i := range n`, `min`/`max`, `maps.Clone`, `slices`, `strings.SplitSeq`.
- Doc comments on exported identifiers; comment density is moderate.
- PDF names are typed `model.Name("Key")`; dict lookups use type assertions with `ok`.
- Reader limits are configurable through `reader.ReaderOptions` (strict vs tolerant parser mode).
- Tests: white-box `_test.go` next to code for internals, black-box under `tests/`; use `t.TempDir()` for files.
- Don't commit generated PDFs (`/*.pdf` is gitignored; older ones were committed before the rule).
