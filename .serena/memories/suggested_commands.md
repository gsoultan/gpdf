# gPDF — commands

- Dependencies: `go mod download` (go.sum is committed since 2026-10-31; before that it was gitignored and a fresh clone failed to build).
- Build/vet: `go build ./... && go vet ./...`
- Tests: `go test ./...`; race: `go test -race ./...`; coverage: `go test -cover ./...`
- Format: `gofmt -l .` (should print nothing).
- Example CLIs (each takes an output path): `go run ./cmd/example out.pdf`, `./cmd/cv`, `./cmd/quotation`, `./cmd/pdf20demo`;
  readers: `go run ./cmd/readpdf file.pdf`, `./cmd/readcontent`, `./cmd/mergepdf`, `./cmd/splitpdf`.
- External validation (if poppler is installed): `pdfinfo out.pdf`, `pdftotext out.pdf -` — use it to check interoperability; gPDF reading its own output is not enough.
- Knowledge graph: `graphify update .` (AST-only, no LLM, refreshes `graphify-out/`); query with `graphify query "<question>"` or `graphify explain "<symbol>"`.
