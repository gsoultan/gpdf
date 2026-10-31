# gPDF — deep review findings (2026-10-31, commit 036d17f; fixes 2026-10-31 on branch claude/project-thread-tv91nd)

Status: open unless marked FIXED. "Verified" = reproduced with a hand-built PDF; "by reading" = from code inspection only.
Every FIXED item has a regression test (see the test named in brackets). Interop checked with poppler 24.02
(pdfinfo/pdftotext) and pypdf 6.17 (RC4-40/128, AES-128, AES-256 R5/R6 files open with user and owner password).

## Critical / high
1. FIXED — /DecodeParms predictors ignored. New `stream/predictor` (PNG 10-15, TIFF 2, all bit depths) applied by
   `reader.applyFilterChain` after Flate/LZW; writer drops predictor parms when it re-encodes. [TestReadXRefStreamWithPNGPredictor]
2. FIXED — huge /Length OOM: parseStream clamps to the file. [TestStreamLengthHardening]
3. FIXED — /Length past EOF hang: endstream must follow, else scan for endstream, else error. [TestStreamWithoutEndstreamFails]
4. FIXED — indirect /Length via `impl.NewParserWithLengthResolver` (documentCore.resolveLength, non-recursive);
   short/missing/negative /Length recover by endstream scan. [TestStreamLengthHardening]
5. FIXED — wrong password accepted: R2-R4 check /U (Alg. 4/5), owner password via Alg. 7, `security.ErrWrongPassword`.
   [TestRC4WrongPasswordRejectedAndOwnerAccepted, TestEncryptedOutputInteroperatesAndResaves]
6. FIXED — AES-256 is now ISO 32000-2 R6 (Alg. 2.A/2.B, U/UE/O/OE/Perms) for read and write; R5 read. x/crypto dropped.
   [TestAES256R6UserOwnerAndWrongPassword]
7. FIXED — RC4 output: V2 R3 128-bit, /Length, 2-element /ID, header %PDF-1.7. Poppler opens it.
8. FIXED — Save with gen > 0: writer uses `Generation(num)` (documentCore) via `writer.objectRef`. [TestSaveKeepsNonZeroGenerations]
9. FIXED — saving opened encrypted docs: Encrypt dict never decrypted, skipped on write; clean trailer
   (Root/Info/ID/Size only); XRef/ObjStm objects skipped. [TestSaveFromXRefStreamWritesCleanTrailer]
10. FIXED — Extract/Split/Merge copied the page tree: `doc/page_copier.go` pre-assigns selected pages, drops
    /Parent /StructParents /B, cloner nulls refs to unselected Page/Pages. `model.Page.Ref` added. [TestExtractCopiesOnlySelectedPage]
11. FIXED — R4 /EncryptMetadata false key derivation; metadata stream left unencrypted. [TestR4EncryptMetadataFalseChangesKey]

## Medium
- FIXED — tokenizer: stray `>`/`)`/`{`/`}` (no more EOF or infinite loop), comment offsets, 1-digit octal, odd hex,
  form feed, `\`+CRLF, `#xx` first char, lenient numbers. Parser: "n gen X" keeps gen, missing endobj accepted,
  nesting depth limit 512. [syntax/impl/regression_test.go, FuzzParseObject]
- FIXED — xref repair keeps LAST definition, reads `trailer` dicts, else /Root from XRef stream dict or last Catalog,
  registers ObjStm objects; tolerant mode also repairs when the trailer has no /Root. [tests/reader/repair_regression_test.go]
- FIXED — hybrid /XRefStm merged. [TestHybridFileReadsXRefStm]
- FIXED — LZW /EarlyChange: own codec in `stream/lzw` (spec example passes). [stream/lzw tests]
- FIXED — named /Crypt stream filters (CryptFilterDecryptor); HexStrings are decrypted/encrypted too.
- FIXED — writer: reals via `model.FormatReal` (also in content encoder), filter arrays encoded, failed encode writes
  raw without /Filter, sorted keys, errors propagated, dangling refs written as null (reader/file uses ValidateGraph=false).
- FIXED — text: simple-font /Encoding + /Differences (`reader/text_encoding.go`, Standard/WinAnsi/MacRoman, glyph names,
  uniXXXX); token-based ToUnicode parser with codespace ranges and correct multi-char bfrange. [reader/text_encoding_test.go]
- FIXED — Merge/split nil trailer panic; page-tree /Kids cycles bounded.
- FIXED — `sign.Sign` now returns `sign.ErrNotImplemented` and writes nothing. [TestSignReportsNotImplemented]
- FIXED — ObjStm cache entries go through cacheOrder; header object number checked. [TestExtractFromObjectStreamChecksObjectNumber]
- FIXED — letter spacing leak: builder tracks Tc/Tw/Tz/Tr per stream (`textStateTracker`). [TestLetterSpacingDoesNotLeakIntoNextRun]
  proteus-converter's `minSpacing` workaround in render/gpdf.go can be removed once it uses this gPDF version.
- FIXED — link annotations get /Border [0 0 0]. [TestLinkAnnotationsHaveNoBorder]
- FIXED (found 2026-10-31) — SaveLinearized: fixed-width /L /H /E /T, real object number for the parameter dict,
  correct first-section xref and /Prev. [TestSaveLinearizedIsWellFormed]
- FIXED (found 2026-10-31) — reader/file.Document finalizer could close the file while a method was still reading
  (e.g. SaveLinearized on a large file: "file already closed"). Finalizer removed.
- FIXED (proteus-converter PR #3) — paint order was fixed to graphics, text, images. Runs now carry a per-page
  `Seq` (stamped by `builder.Page.AddText/AddImage/AddGraphic`; always use these, never append to the slices) and
  `DocumentBuilder.SetPaintOrder(doc.PaintInCallOrder)` interleaves them; `PaintLayered` stays the default.
  Images are still named /Im<n> by their index in ImageRuns. [TestPaintInCallOrderInterleavesRuns]
- FIXED (proteus-converter PR #3) — PNG alpha was dropped (premultiplied, so transparent pixels went black).
  `image.DecodePNG` returns non-premultiplied samples + Alpha; `builder.ImageRun.SetPNG` writes Alpha as /SMask; every
  PNG path (DrawPNG*, imgdraw, Page image objects, table cell images, DrawImageWith with PNG/JPEG file bytes, so
  FlowBuilder.Image works) uses it. `DecodePNGToRaw` is deprecated. [TestPNGAlphaBecomesSoftMaskAndImagesAreCompressed, poppler render]
- FIXED (proteus-converter PR #3) — raw image XObjects and soft masks are written with /FlateDecode.

## Still open
- JPEG CMYK (Adobe inverted) gets DeviceCMYK with no /Decode array, so such files can render inverted.
- SaveLinearized writes no hint stream (poppler warns "Hints table length ... invalid"; files still open).
- AES-256 passwords are not SASLprep-normalised (ASCII and NFC input fine).
- Reader decodes every stream on Resolve; a stream with an unsupported filter makes Resolve (and Save) fail.
- Writer always re-encodes Flate content (no pass-through of the original compressed bytes).
- Xref table parser assumes 20-byte entries.
- Committed binaries despite `.gitignore`: `CV CONTOH.pdf` (779 KB), `test_out.pdf`, `.idea/` (left for the owner to decide).

## Repo / process
- FIXED 2026-10-31 — `go.sum` was gitignored.
- FIXED 2026-10-31 — CI: `.github/workflows/go.yml` runs gofmt, vet, `go test -race` (with poppler installed).
