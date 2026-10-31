# gPDF — deep review findings (2026-10-31, commit 036d17f)

Status: open unless marked FIXED. "Verified" = reproduced with a hand-built PDF; "by reading" = from code inspection only.
Baseline: build/vet/test/race all pass once go.sum exists. Coverage: doc 7.9%, security 22.6%, syntax/impl 26.3%,
writer 27.5%, reader 42.8%; all `stream/*` filters, `font/*`, `model`, `xref`, `sign` have no tests; no fuzz tests.

## Critical / high
1. Verified — PNG/TIFF predictors (`/DecodeParms /Predictor`) are ignored by the generic filter pipeline
   (`reader/reader_impl.go` applyFiltersWithLimits, `stream/flate`). Xref streams almost always use `/Predictor 12`,
   so most PDF 1.5+ files fail to open (`object 1 0 R not in xref`). Only SMask image decoding handles predictors
   (`reader/extract_images.go`). Writer keeps `/DecodeParms` while re-encoding without a predictor, so Save/Merge/Split
   also corrupt predictor-encoded images.
2. Verified — Untrusted `/Length` is used directly: `/Length 99999999999` → `make([]byte, length)` → fatal OOM
   (unrecoverable, not a panic) in `syntax/impl/parser_impl.go` parseStream. Clamp to file size.
3. Verified — `/Length` past end of file → parseStream loops forever: after `SetPosition` beyond EOF the tokenizer
   returns `TokenEOF` with nil error and the `endstream` search loop never exits. Hang DoS.
4. Verified — Indirect `/Length 5 0 R` (common in pdfTeX/Ghostscript output) → "stream missing Length"; text
   extraction then returns "" with nil error. Too-short `/Length` also yields silent empty text (no endstream recovery).
5. Verified — Wrong user password is accepted: R2/R3/R4 never check `/U`, so `OpenWithPassword(..., "WRONG")`
   succeeds and returns garbage strings/streams. Owner password not supported at all.
6. Verified — AES-256 ("V5/R6") is a custom PBKDF2(ID) scheme, not ISO 32000-2 Algorithm 2.A/2.B (no UE/OE/Perms,
   no SHA-256/384/512 hash loop). It cannot open real AES-256 PDFs, and poppler rejects gPDF's own output
   ("Invalid encryption key length"; `cmd/pdf20demo` output unreadable). R5 is advertised but rejected. `/U` is an
   unsalted SHA-256(password‖ID) stored in the file, which makes offline brute force fast and the PBKDF2 pointless.
7. Verified — RC4 output (`SaveWithPassword`) is rejected by poppler ("Weird encryption info"): Encrypt dict lacks
   `/V`, `/Length`; trailer `/ID` has 1 element (spec: 2); 40-bit RC4 in a `%PDF-2.0` file (deprecated in 2.0).
8. Verified — Save fails for any source with generation > 0 objects: writer resolves every object as `Ref{n,0}`
   (`writer/writer_impl.go` Write) → "generation mismatch".
9. Verified — Saving an opened encrypted document: AES path fails ("invalid padding length", the Encrypt dict is
   run through the decryptor by Resolve); RC4 path writes a garbled Encrypt dict and keeps `/Encrypt` in the trailer.
   Old trailer keys (`/Prev`, `/XRefStm`, `/Encrypt`) are copied verbatim on Save.
10. Verified — Extract/Split/Merge copy the whole source page tree: the cloner follows the page's `/Parent` before it
    is deleted (`doc/cloner.go`, `doc/split.go`, `doc/merge.go`). Extracting 1 of 3 pages kept all 3 content streams.
    Fix: drop `/Parent` (and usually `/StructParents`, `/B`) before cloning.
11. By reading — R4 ignores `/EncryptMetadata false` (key derivation must append 0xFFFFFFFF) → such files decrypt wrong.

## Medium
- Tokenizer (`syntax/impl/tokenizer.go`): stray `>` returns a zero Token (= TokenEOF) → silent truncation;
  comments are skipped without advancing `pos` (CurrentOffset drifts); 1-digit octal escapes (`\5`) dropped;
  odd-length hex strings lose the last nibble (spec: pad with 0); form feed (0x0C) not whitespace; `\` + CRLF
  continuation leaves the LF; `#xx` not decoded as first name char; number parse errors become 0.
- `ParseObject` "n gen X" (X not obj/R) loses `gen` (only `tok3` is pushed back).
- Xref repair (`reader/xref_repair.go`): keeps the FIRST definition of each object (incremental updates need the
  last) and returns an empty trailer (no `/Root`), so repaired files still fail to open. Should scan for
  `trailer`/`/Type /Catalog`.
- Hybrid files: trailer `/XRefStm` is ignored.
- By reading — LZWDecode uses Go `compress/lzw`, which has no PDF `/EarlyChange` handling; likely wrong at code-width
  boundaries (pdfcpu uses hhrutter/lzw for this reason).
- Crypt filter with a non-Identity name passes still-encrypted bytes through.
- Writer: `Real` written with `%g` → exponent form (`1e+06`) which is not PDF syntax; Filter arrays are declared
  but not encoded (content written decoded); encode errors silently write raw bytes under the declared filter;
  dict key order is random (non-reproducible output); many write errors ignored (`writeDict`, `writeDictEnc`).
- Text extraction (`reader/extract.go`): no `/Encoding` / `/Differences` / glyph-name support for simple fonts
  (falls back to UTF-8-or-WinAnsi guess); ToUnicode parser is line-based (one mapping per line, breaks on compact
  or CR-only CMaps, ignores codespace ranges, bfrange with multi-code-point base increments only the first rune).
- `doc.Merge`/`splitOne`: if Info clone fails, `trailer.Dict` stays nil → panic on `Root` assignment.
- `sign.Sign`/`SignFile` are pass-through stubs that return nil — callers think the file was signed.
- Object-stream cache entries bypass `cacheOrder` (never evicted); compressed-object header object numbers not checked.

## Repo / process
- FIXED 2026-10-31 — `go.sum` was gitignored: fresh clone failed `go build`/`go test` ("missing go.sum entry"),
  including CodeQL autobuild.
- No CI runs `go test` (only `.github/workflows/codeql.yml`).
- Committed binaries despite `.gitignore`: `CV CONTOH.pdf` (779 KB), `test_out.pdf`, `.idea/`.
- Builder output (`cmd/example`, `cmd/cv`, `cmd/quotation`) opens cleanly in poppler and text extracts correctly.

## Suggested fix order
1 (predictors) → 2/3/4 (stream length hardening + fuzz test for parser) → 10 (cloner /Parent) → 8/9 (writer gens,
trailer hygiene) → 5/6/7/11 (implement ISO 32000-2 Standard security handler properly, validate U/O) → medium items.
