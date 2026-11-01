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

- FIXED (round 2, 2026-10-31) — Adobe CMYK JPEGs get /Decode [1 0 1 0 1 0 1 0] (`jpegHasAdobeMarker`). [TestAdobeCMYKJPEGGetsInvertedDecode]
- FIXED (round 2) — SaveLinearized writes a real primary hint stream (page offset + shared object tables) with
  spec layout and renumbering; poppler parses it with no warnings. [TestSaveLinearizedWritesUsableHintTables]
- FIXED (round 2) — unsupported filter: Resolve keeps the stream encoded (with /Filter), writer copies it unchanged.
  [TestUnknownFilterIsPassedThrough]
- FIXED (round 2) — xref table entries are read as tokens (19/21-byte entries work). [TestXRefTableWithNineteenByteEntries]
- FIXED (round 2, found by FuzzReadDocument) — a stray `)` in a content stream hung text extraction forever
  (`content/impl/tokenizer.go` readOperator). [TestParseStrayDelimiterTerminates]

- FIXED (round 2, reported by proteus-converter) — embedded TrueType subsets carried the full name/cmap/post tables
  (post alone was ~450 KB for CJK fonts) and full-length loca/hmtx. `font/truetype/subset.go` now keeps only
  head, OS/2, cvt, fpgm, prep, hhea, maxp, hmtx, post (format 3.0), loca, glyf and cuts maxp/hhea/hmtx/loca to the
  highest used GID (+1), so /CIDToGIDMap /Identity stays valid. DejaVuSans: 760 KB -> 6.7 KB for a short string.
  Note: `truetype.Parse` cannot re-read a subset (it requires cmap); test it through poppler instead.
  [tests/doc/font_subset_test.go]

- FIXED (round 3, PDF 2.0 audit, 2026-11-01) — text strings were written as raw UTF-8 and read as raw bytes: now
  `model.EncodeTextString`/`DecodeTextString` (UTF-8 BOM for 2.0 output, UTF-16BE BOM for PDF/A-2/3, PDFDocEncoding
  on read), applied by `doc/text_strings.go` to every text-string key after the objects are built.
- FIXED (round 3) — PDF dates: `model.ParseDate` (Z, optional apostrophes, partial dates) and `model.FormatDate`
  (2.0 form `...Z` / `+HH'mm` vs 1.x `+HH'mm'`). Info and XMP for every build with metadata, not only PDF/A.
- FIXED (round 3) — AES-256 R6 passwords go through SASLprep (`security/saslprep.go`: mapping, prohibited
  characters, bidi rule; **no NFKC**, the standard library has no normaliser). Prohibited characters -> ErrPasswordPrep.
- FIXED (round 3) — forms: every widget has an appearance stream (`doc/form/appearance.go`: text, check mark and
  radio dot as vector paths, push button), AcroForm has /DR and /DA, NeedAppearances (deprecated in 2.0, forbidden
  in PDF/A) is gone.
- FIXED (round 3) — CIDSet (deprecated in PDF 2.0) is only written for PDF/A-2/3 output.
- ADDED (round 3) — attachments/associated files (`doc/attach.go`, `reader/embedded_files.go`, `Document.Attachments()`),
  PDF/A-4/4e/4f (PDF 2.0 header, pdfaid:rev 2020, Info limited to dates, 4f needs an attachment, PDF/A-2 rejects
  attachments), page-level output intents, document parts (`DocumentPart`), ExtGState /UseBlackPtComp,
  reader `EffectiveVersion()` (catalog /Version), catalog `Extensions()`, `DPartRootRef()`, page `OutputIntents()`.

## Still open
- SASLprep is applied without NFKC normalisation (no normaliser without golang.org/x/text).
- DONE (roadmap phase 1, 2026-11-01): PAdES signatures (`sign` package: CAdES CMS, sign/verify/DSS, pdfsig says
  "Signature is Valid"), portable collections (`doc/collection.go`, Attachment.CollectionItem -> /CI), 3D annotations
  (`doc/annot3d.go`, U3D/PRC stream, poster appearance) and geospatial viewports (`doc/geo.go` /VP /Measure GEO,
  `reader.ExtractGeoViewports` + `ToGeo`). Annotations from builder code go through `doc/annotations.go` annotBuilder.
- PDF 2.0 features still not implemented: unencrypted wrapper documents, RichMedia, tagged PDF 2.0 namespaces in the
  builder (reader exposes /Namespaces). Standard 14 fonts are still written unembedded (phase 2).
- Signature limits: time-stamp tokens are embedded but their own signature is not checked; no OCSP/CRL fetching
  (caller passes DER data to AddValidationData); incremental update writes an xref table even after an xref stream;
  signing an encrypted file is refused.
- Writer always re-encodes Flate content (no pass-through of the original compressed bytes).
- PDF/A: only level B of parts 2 and 3; checks cover fonts, CMYK and encryption, not everything veraPDF checks
  (no veraPDF available here). Not done: PDF/A-1, level A/U, embedded-file (AF) support for PDF/A-3.
- Committed binaries despite `.gitignore`: `CV CONTOH.pdf` (779 KB), `test_out.pdf`, `.idea/` (left for the owner to decide).

## Repo / process
- FIXED 2026-10-31 — `go.sum` was gitignored.
- FIXED 2026-10-31 — CI: `.github/workflows/go.yml` runs gofmt, vet, `go test -race` (with poppler installed).
