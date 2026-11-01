# gPDF — commands

- Dependencies: `go mod download` (go.sum is committed since 2026-10-31; before that it was gitignored and a fresh clone failed to build).
- Build/vet: `go build ./... && go vet ./...`
- Tests: `go test ./...`; race: `go test -race ./...`; coverage: `go test -cover ./...`
- Format: `gofmt -l .` (should print nothing).
- Example CLIs (each takes an output path): `go run ./cmd/example out.pdf`, `./cmd/cv`, `./cmd/quotation`, `./cmd/pdf20demo`;
  readers: `go run ./cmd/readpdf file.pdf`, `./cmd/readcontent`, `./cmd/mergepdf`, `./cmd/splitpdf`.
- External validation (if poppler is installed): `pdfinfo out.pdf`, `pdftotext out.pdf -` — use it to check interoperability; gPDF reading its own output is not enough.
- Knowledge graph: `pip install graphifyy` (venv is fine), then `graphify update . --no-cluster` and
  `graphify cluster-only . --no-label --no-viz` (AST-only, no LLM, refreshes `graphify-out/`); query with
  `graphify query "<question>"` or `graphify explain "<symbol>"`.
- Fuzzing: `go test ./<pkg> -run '^$' -fuzz '^FuzzName$' -fuzztime 60s` (targets listed in `.github/workflows/fuzz.yml`).
- Benchmarks: `go test ./tests/doc ./tests/reader -run '^$' -bench . -benchmem` (benchstat compares runs).
- Conformance: `GPDF_REQUIRE_TOOLS=1 VERAPDF_BIN=... GPDF_CONFORMANCE_DIR=/tmp/corpus go test ./tests/conformance -v`;
  veraPDF runs from the Maven Central `greenfield-apps` jar (`java -cp greenfield-apps.jar org.verapdf.apps.GreenfieldCliWrapper --format mrr --flavour 2b file.pdf`
  prints the failed rules; `--format text` gives the PASS/FAIL line). qpdf: apt `qpdf`, or pikepdf's `pikepdf.Job(["qpdf","--check",f])`
  (no stdout forwarding) when the binary cannot be installed.
- Encryption interop: pypdf (if installed) can write RC4/AES test files: `PdfWriter(clone_from=...).encrypt(user_password, owner_password, algorithm="AES-256")`.
