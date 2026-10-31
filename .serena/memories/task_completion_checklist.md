# gPDF — task workflow and completion checklist

Standing rules from the owner (2026-10-31):
1. At the start of every task, read the Serena memories (`.serena/memories/*.md`) and the graphify report
   (`graphify-out/GRAPH_REPORT.md`, query `graphify-out/graph.json` as needed).
2. Read Serena/graphify output through `rtk` (Rust Token Killer) when it is installed, to keep token use low.
3. When a task finishes, update the relevant Serena memories and rerun `graphify update .`, and commit both.

Before calling a code task done:
- `gofmt -l .` is empty, `go vet ./...` and `go test ./...` pass (add `-race` for concurrency changes).
- Add a regression test for every bug fix; for parser/reader fixes prefer a tiny hand-built PDF in the test (`internal/pdftest`).
- For writer/encryption changes, check the output with an external reader (`pdfinfo`/`pdftotext`) when available.
- Update `review_findings` memory: mark fixed items, add new ones.
- Attribution: commit as `gsoultan <gembit.shirazi@gmail.com>`, no Claude credit lines anywhere (commits, PR text, comments).
