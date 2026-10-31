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
| `writer` | `PDFWriter.Write` (full rewrite, xref table, real generations via optional `Generation(num)`, clean trailer, skips source Encrypt/XRef/ObjStm objects, filter chains re-encoded, sorted keys, `model.FormatReal`), `WriteIncremental`, `WriteLinearized` (no hint stream), encrypted variants, `graph_validation.go` (refs must resolve; disabled for files read from disk, where dangling refs are written as null). |
| `content` + `content/impl` | Content-stream tokenizer/parser (`Op`) and encoder. |
| `font`, `font/truetype`, `font/system` | Standard-14 metrics, TrueType parsing/subsetting/GSUB/ToUnicode, system font lookup. |
| `xobject` | Image/form XObject helpers. |
| `doc` | Public facade. `Document` = CatalogReader + ContentReader + ContentSearcher + ImageReader + LayoutReader + Saver. `DocumentBuilder` (+ `builder/graphics|imgdraw|text`, `table`, `form`, `layer`, `tagged`, `outline`, `metadata`, `style`, `pagesize`), `builtDocument`, `merge.go`, `split.go`, `cloner.go` (ref-remapping deep copy). |
| `sign` | `Sign`/`SignFile` — not implemented, return `sign.ErrNotImplemented`. |
| `tests/` | Black-box tests (`tests/doc`, `tests/reader`, `tests/writer` incl. poppler interop when installed, `tests/merge_test.go`). |
| `internal/pdftest` | Builder for hand-made test PDFs: raw objects, streams, xref table, predicted xref stream, object streams, hybrid files. |

Key flows
- Open: `doc.Open` → `reader/file` → `reader.PDFReader` → `documentCore.Resolve` (parse at xref offset, decrypt strings/streams per object, decode filters, cache with FIFO limit).
- Save: `file.Document.Save` → `writer.Write` resolves each object with its real generation and rewrites streams, re-encoding the full `/Filter` chain (Flate/LZW without predictor).
- Merge/Split/Extract: `copyPages` (`doc/page_copier.go`) reserves numbers for selected pages, drops /Parent /StructParents /B, and `objectCloner` deep-copies the rest (refs to unselected Page/Pages become null) into a fresh `builtDocument`.
- Builder text: `buildTextOps` tracks Tc/Tw/Tz/Tr (`textStateTracker`) so text state never leaks between runs; links get /Border [0 0 0].

Graph: `graphify-out/GRAPH_REPORT.md` and `graphify-out/graph.json` (regenerate with `graphify update .`). As of 2026-10-31 the graph is built from 036d17f and is stale for the fix branch: graphify was not installable in that session, so rerun `graphify update .`.
