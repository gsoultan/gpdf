# gPDF — architecture map

Layering (low → high). Interfaces sit at the package root; concrete code in `impl/` subpackages or `*_impl.go`.

| Package | Role |
| --- | --- |
| `model` | PDF object model: `Object`, `Dict`, `Array`, `Name`, `String`, `HexString`, `Integer`, `Real`, `Ref`, `*Stream`, plus typed views (`Catalog`, `Page`, `Trailer`, `AcroForm`, `StructTree`, ...). |
| `syntax` + `syntax/impl` | Tokenizer (`tokenizer.go`, bufio over `io.SectionReader`) and parser (`parser_impl.go`): objects, streams (needs direct `/Length`), xref tables, trailer. |
| `xref` | In-memory xref `Table` (`Entry`: offset/gen/inUse or compressed stream+index). |
| `stream/*` | Filters behind `stream.Filter` (Decode/Encode by name) + `stream/defaults` registry; optional `stream.ParamFilter` gets /DecodeParms. Flate, LZW (own codec with /EarlyChange), ASCII85, ASCIIHex, RunLength are real; DCT/JPX/JBIG2/CCITT are pass-through; Crypt passes through (decryption picks the named crypt filter). `stream/predictor` = PNG/TIFF predictors, applied by `reader.applyFilterChain`. |
| `security` | Standard handler `StandardDecryptor` (also Encryptor): `standard.go` (dispatch, crypt filters /CF /StmF /StrF, RC4/AES helpers, `ErrWrongPassword`), `standard_legacy.go` (R2-R4 Algorithms 2-7, user+owner auth, EncryptMetadata), `standard_aes256.go` (R5/R6 Alg. 2.A/2.B, `ValidatePerms`). Writers in `standard_write.go`: `BuildEncryptDictForWrite` (V2 R3 RC4-128), `BuildAES256EncryptDictForWrite[WithRandom]` (V5 R6). No external deps. |
| `reader` | `PDFReader.ReadDocument` → version + linearization detect → `startxref` → `/Prev` chain (table or xref stream, hybrid /XRefStm) → repair scan when that fails or has no /Root (`xref_repair.go`: last definition wins, trailer/Catalog recovery, ObjStm registration) → `pdfDocument` (`document_core.go` resolve/cache/decrypt/objstm, `Generation(num)`, indirect /Length resolver). Text decoding: `extract.go` (ToUnicode tokenizer + codespace) and `text_encoding.go` (simple-font /Encoding, /Differences, glyph names). Extraction: `extract.go` (text, ToUnicode), `extract_layout.go`, `extract_tables.go`, `extract_images.go`, `extract_vectors.go`, `replace.go`. Options/limits in `options.go` (max decoded stream size, filter chain length, object count, cache size). |
| `reader/file` | `file.Document`: wraps a reader doc + `writer.PDFWriter` for Save / SaveWithPassword / SaveWithAES256Password / SaveLinearized. |
| `writer` | `PDFWriter.Write` (full rewrite; by default `packer` (`objstm.go`) puts non-stream gen-0 objects into /ObjStm of ≤100 and writes a /XRef stream (W [1 4 2], Predictor 12), `ObjectStreams=false` or a 1.4 header gives the classic table; real generations via optional `Generation(num)`, clean trailer, skips source Encrypt/XRef/ObjStm objects, filter chains re-encoded with the requested predictor (`predictor.Encode`), sorted keys, `model.FormatReal`), `WriteIncremental` (xref stream when the base trailer is /Type /XRef, else table), `WriteLinearized` (hint stream), encrypted variants (object streams encrypted whole, strings inside not), `StreamWriter` (`stream_writer.go`: Alloc/WriteObject/Close straight to an io.Writer; packed objects held until Close, deferred packer numbering), `graph_validation.go` (refs must resolve; disabled for files read from disk, where dangling refs are written as null). |
| `content` + `content/impl` | Content-stream tokenizer/parser (`Op`) and encoder. |
| `font`, `font/truetype`, `font/shaping`, `font/system` | Standard-14 metrics and the `Font`/`EmbeddableFont`/`Shaper`/`GlyphEncoder`/`CFFFont` interfaces; `truetype`: TrueType and OpenType-CFF parsing (`cff.go`), subsetting (CFF embedded whole), GSUB (liga/clig/rlig ligatures, Arabic init/medi/fina), GPOS pair kerning of the `kern` feature (`gpos.go`, PairPos 1/2, used instead of the legacy `kern` table when present, like HarfBuzz; Coverage/ClassDef looked up in place), `Shape` (forms, ligatures, kerning → `font.Glyph`), ToUnicode by rune or by glyph; `shaping`: UAX #9 bidi `Reorder`/`Levels`, `Mirror`, Arabic `JoiningTypeOf`/`ArabicForms`/presentation forms; `system`: font directory scanners, `StandardSubstitutes` (standard-14 → Liberation/Nimbus/...), `Fallback` (installed font covering a rune). |
| `xobject` | Image/form XObject helpers. |
| `doc` | Public facade. `Document` = CatalogReader + ContentReader + ContentSearcher + ImageReader + LayoutReader + Saver. `DocumentBuilder` (+ `builder/graphics|imgdraw|text`, `table`, `form`, `layer`, `tagged`, `outline`, `metadata`, `style`, `pagesize`), `builtDocument`, `merge.go`, `split.go`, `cloner.go` (ref-remapping deep copy). |
| `sign` | PAdES/CMS signatures: `CMSSigner` (`cms.go`, encoding/asn1 SignedData with content-type, message-digest, ESS signing-certificate-v2, optional RFC 3161 token), `Sign` (`sign.go`: incremental update with Sig dict + merged Sig field/widget, ByteRange and Contents patched in place, DocMDP certification, optional appearance), `Verify` (`verify.go`: byte-range digest, CMS signature, chain via x509, CoversWholeDocument), `AddValidationData` (`dss.go`: /DSS Certs/CRLs/OCSPs, ESIC extension). Encrypted inputs refused. |
| `tests/` | Black-box tests (`tests/doc`, `tests/reader`, `tests/writer` incl. poppler interop when installed, `tests/merge_test.go`). |
| `internal/pdftest` | Builder for hand-made test PDFs: raw objects, streams, xref table, predicted xref stream, object streams, hybrid files. |

Key flows
- Open: `doc.Open` → `reader/file` → `reader.PDFReader` → `documentCore.Resolve` (parse at xref offset, decrypt strings/streams per object, decode filters, cache with FIFO limit).
- Save: `file.Document.Save` → `writer.Write` resolves each object with its real generation and rewrites streams, re-encoding the full `/Filter` chain (Flate/LZW with the predictor their /DecodeParms ask for). `builtDocument.newWriter` passes `NoObjectStreams()` through as `WriterOptions.ObjectStreams=false`.
- Merge/Split/Extract: `copyPages` (`doc/page_copier.go`) reserves numbers for selected pages, drops /Parent /StructParents /B, and `objectCloner` deep-copies the rest (refs to unselected Page/Pages become null) into a fresh `builtDocument`.
- Builder paint: `buildPageContent` uses `buildGraphicsOps`/`buildTextOps`/`buildImageOps` (layered, default) or
  `buildOpsInCallOrder` (`doc/paint_order.go`, sorted by run `Seq`); both go through `textOpsWriter` and
  `imageOpsWriter`. In call order, the text writer forgets its color/text state after each graphic run.
- Builder images: every image goes through `DocumentBuilder.addImage` (`builder_image.go`): `prepareImage`
  (`image_compression.go`: `CompressImages` → MaxDPI box downsampling of raw/JPEG samples and masks, JPEGQuality
  re-encoding of 8-bit gray/RGB without mask) then, in streaming mode, `docStream.writeImage`. PNG →
  `image.DecodePNG` → `ImageRun.SetPNG` (alpha as /SMask); raw images and masks are FlateDecode with
  `predictorParms` (Predictor 15); `imageStreams` builds image + mask streams for Build and streaming alike.
- Streaming output (`doc/streaming.go`): `StreamTo(w)` sets `b.stream`; images are written when drawn
  (`ImageRun.ObjNum/MaskObjNum`, samples released), `build()` starts numbering at `StreamWriter.NextObjectNumber()`
  and writes content streams at once, `Finish()` writes the remaining objects and closes; `BuildTo(w)` = both.
  `Build()` is refused after `StreamTo`; changing PDFA/PDFUA after the first write is an `ErrStreaming` error.
- Builder text: `fontConfig.resolveFont` (`doc/font_config.go`) reorders each line with `shaping.Reorder`, splits
  runs by font (primary → block fallback → fallback chain → system fallback, registering fonts on the fly), and
  shapes segments whose font is a `font.Shaper` (RTL runs get reversed glyphs). `fontSegment.glyphs` →
  `TextRun.Glyphs` → `shapedTextArg` in the content writer (TJ with kerning); `embeddedFontUsage.markGlyphs`
  tracks glyph → text for ToUnicode. Standard names map to embedded programs through `aliasedFont`
  (`registerFontAs`), whose `EmbeddedFontName` becomes the subset BaseFont. Inline simple-font dicts are turned
  into shared indirect objects at the end of Build (`indirectStandardFonts`).
- Tagged PDF (`doc/tagged`): the builder records Blocks/Figures/Tables/Lists plus `Order`/`Section.Kids` (reading
  order); `BuildStructure` writes MarkInfo, StructTreeRoot (K = one Document element), Sect/P/Hn/L(LI>Lbl,LBody)/
  Table(TR>TH,TD)/Figure/Link elements, the ParentTree (per-page arrays, then annotation keys, `ParentTreeNextKey`),
  and with `UseNamespaces` the PDF 2.0 namespace objects. PDF/UA lives in `doc/pdfua.go` (`PDFUA`, `checkPDFUA`);
  `appendGraphicOps` wraps decorative graphics as artifacts in PDF/UA output.
- Builder page decorators (`doc/page_decoration.go`): `OnEveryPage`/`SetHeader`/`SetFooter` run inside Build with the
  final page count (runs added are removed again after Build, so Build is repeatable); output is a pagination artifact.
- PDF/A (`doc/pdfa.go`): `PDFA(PDFA2B|PDFA3B)` → `preparePDFA` (sRGB ICC from `style.SRGBICCProfile`, XMP from
  `metadata.BuildXMP`, CreationDate) and `checkPDFA` after object build (fonts embedded, no DeviceCMYK with RGB
  intent); `builtDocument.version="1.7"` → writer `VersionedDocument` picks the header; encrypted saves refused.
- Linearized save (`writer/linearized.go`): `planLinearization` splits objects into first-page section, per-page
  sections, shared, other; objects are renumbered (rest from 1, first-page section after) and a hint stream with
  page-offset + shared-object tables is written (offsets as if the hint stream were absent, arrays byte-aligned,
  as poppler/qpdf read them).
- PDF 2.0 output (`doc/pdf20.go`, `doc/attach.go`, `doc/text_strings.go`): Build runs `prepareMetadata` (XMP for any
  metadata, `metadata.Support.PDF20` picks UTF-8 text strings and the 2.0 date form), `buildAttachments` (EmbeddedFile
  streams, Filespec with /AFRelationship, EmbeddedFiles name tree, catalog/page /AF), `buildPageOutputIntents`,
  `buildDocumentParts` (DPartRoot/DPart, gaps filled), then `encodeTextStrings` over every object. PDF/A-4 levels set
  `version="2.0"` and `DatesOnly` Info. Forms: `doc/form/appearance.go` builds /AP streams and /DR /DA.
- Reading order (`reader/reading_order.go`): `ExtractStructuredText` walks StructTreeRoot /K (RoleMap applied,
  MCR/OBJR/inline kids, ActualText, Alt also read from /A for foreign files), gathering per-page MCID text with
  `collectMarkedContent` (BDC/BMC/EMC stack, form XObjects); inline roles (Span, Link, ...) join their parent;
  `TextElement{Role, Page, Depth, Parent, Text, Alt}`; `JoinTextElements` (Lbl+LBody with a space, cells of a
  row with tabs, pages with a blank line). Untagged fallback `OrderBlocks`: XY cut (wider band wins: y gap ≥1.5
  line, x gap ≥1 line) then rows → paragraphs (`regionParagraphs`).
- Change tracking (`documentCore.modified`): `MarkModified`/`NewObject`/`ModifiedObjects`/`Source`/`MarkRedacted`;
  modified objects are pinned in the cache and included in `ObjectNumbers`; `reader.ReplaceContent` and
  `RedactContent` call `markModified`. `file.Document.SaveIncremental` copies the source bytes and
  `WriteIncremental`s the modified set (trailer copied from the base); `SetInfo`, `SetObject`, `NewObject`
  (`doc.ObjectEditor`, also on `builtDocument`). `doc.ErrNoBaseFile`, `doc.ErrRedacted` (= `file.ErrRedacted`).
- Redaction (`reader/redact.go`, `doc/redact.go`): `walkTextOps` (shared with `extractBlocksFromOps`) reports each
  text-showing op with its pre-state and blocks; matched ops become `[n] TJ` displacements (`displacementOps`,
  ' and " keep their T*/Tw/Tc), content is wrapped in q/Q and boxes appended (`re f`), page content goes into the
  first stream (others emptied); matches in form XObjects are only counted (`SkippedInForms`).
- Reading PDF 2.0: `model.DecodeTextString` (`String.Text()`), `model.ParseDate`, `reader.ExtractEmbeddedFiles`
  (name tree + /AF), `pdfDocument.EffectiveVersion` (header vs catalog /Version), R6 passwords via `security.saslprep`.
- Builder text: `buildTextOps` tracks Tc/Tw/Tz/Tr (`textStateTracker`) so text state never leaks between runs; links get /Border [0 0 0].

- Validation (`tests/conformance/conformance_test.go`): `corpus(t)` builds every output form (`sample{name,
  flavour, password, pdf}`, flavour = space-separated veraPDF profiles), `corpusDir` writes them (to
  `GPDF_CONFORMANCE_DIR` if set), `tool(t, name)` finds `qpdf`/`verapdf` (`<NAME>_BIN`, skip or fail under
  `GPDF_REQUIRE_TOOLS`). Fuzz targets: `syntax/impl` (objects, tokenizer), `content/impl` FuzzParseContent
  (parse → Encoder → parse), `font/truetype` FuzzParse, `stream` FuzzFilters (filters + predictor round trip),
  `model` FuzzTextString/FuzzParseDate, `tests/reader` FuzzReadDocument/FuzzRoundTrip (open, extract, save,
  SaveIncremental, Redact, reopen). Crash inputs are committed under `testdata/fuzz/<Target>/`. Benchmarks in
  `tests/doc/bench_test.go` (classic vs object streams, StreamTo vs Build, images, linearize, reading order,
  layout, redact, incremental save, merge) and `tests/reader` (BenchmarkReadDocument).
- Structure tree parent tree (`doc/tagged/builder.go`): `mcidParents` derives the per-page `/ParentTree` arrays
  from the MCR kids (index = MCID), `setStructParents` adds `/P` to every element.

Graph: `graphify-out/GRAPH_REPORT.md` and `graphify-out/graph.json`. Rebuilt 2026-11-01 from dff41a7 plus the
phase 6 tree with `graphify update . --no-cluster` then `graphify cluster-only . --no-label --no-viz` (community
names are placeholders: no LLM backend here). graphify installs with `pip install graphifyy` (PyPI name).
