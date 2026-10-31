# gPDF — agent instructions

Standing workflow for every task in this repo:

1. Start: read the Serena memories in `.serena/memories/` (start with `project_overview`, `architecture`,
   `review_findings`, `task_completion_checklist`) and the graphify report `graphify-out/GRAPH_REPORT.md`
   (query `graphify-out/graph.json` with `graphify query` / `graphify explain` as needed).
2. Read Serena and graphify output through `rtk` (Rust Token Killer) when it is available.
3. Finish: update the affected Serena memories, run `graphify update .` to refresh `graphify-out/`, and commit both
   with the code change.

Checks before done: `gofmt -l .` empty, `go vet ./...`, `go test ./...`.
