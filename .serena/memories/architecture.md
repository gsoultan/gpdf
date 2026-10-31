# gPDF — architecture map

Layering (low → high). Interfaces sit at the package root; concrete code in `impl/` subpackages or `*_impl.go`.

| Package | Role |
| --- | --- |
| `model` | PDF object model: `Object`, `Dict`, `Array`, `Name`, `String`, `HexString`, `Integer`, `Real`, `Ref`, `*Stream`, plus typed views (`Catalog`, `Page`, `Trailer`, `AcroForm`, `StructTree`, ...). |
| `syntax` + `syntax/impl` | Tokenizer (`tokenizer.go`, bufio over `io.SectionReader`) and parser (`parser_impl.go`): objects, streams (needs direct `/Length`), xref tables, trailer. |
| `xref` | In-memory xref `Table` (`Entry`: offset/gen/inUse or compressed stream+index). |
| `stream/*` | Filters behind `stream.Filter` (Decode/Encode by name) + `stream/defaults` registry. Flate, LZW (Go `compress/lzw`), ASCII85, ASCIIHex, RunLength are real; DCT/JPX/JBIG2/CCITT are pass-through; Crypt only handles Identity. No `/DecodeParms` support in the generic pipeline. |
| `security` | Standard handler: R2/R3 RC4 (`standard.go`), R4 RC4/AESV2 (`standard_r4.go`), "R6" AES-256 (`aes256_*.go`, custom PBKDF2 scheme, NOT ISO 32000-2). Writers: `BuildEncryptDictForWrite` (R2 40-bit), `BuildAES256EncryptDictForWrite`. |
| `reader` | `PDFReader.ReadDocument` → version + linearization detect → `startxref` → `/Prev` chain (table or xref stream) → optional repair scan (`xref_repair.go`) → `pdfDocument` (`document_core.go` resolve/cache/decrypt/objstm). Extraction: `extract.go` (text, ToUnicode), `extract_layout.go`, `extract_tables.go`, `extract_images.go`, `extract_vectors.go`, `replace.go`. Options/limits in `options.go` (max decoded stream size, filter chain length, object count, cache size). |
| `reader/file` | `file.Document`: wraps a reader doc + `writer.PDFWriter` for Save / SaveWithPassword / SaveWithAES256Password / SaveLinearized. |
| `writer` | `PDFWriter.Write` (full rewrite, xref table, gen always 0), `WriteIncremental`, `WriteLinearized`, encrypted variants, `graph_validation.go` (refs must resolve). |
| `content` + `content/impl` | Content-stream tokenizer/parser (`Op`) and encoder. |
| `font`, `font/truetype`, `font/system` | Standard-14 metrics, TrueType parsing/subsetting/GSUB/ToUnicode, system font lookup. |
| `xobject` | Image/form XObject helpers. |
| `doc` | Public facade. `Document` = CatalogReader + ContentReader + ContentSearcher + ImageReader + LayoutReader + Saver. `DocumentBuilder` (+ `builder/graphics|imgdraw|text`, `table`, `form`, `layer`, `tagged`, `outline`, `metadata`, `style`, `pagesize`), `builtDocument`, `merge.go`, `split.go`, `cloner.go` (ref-remapping deep copy). |
| `sign` | `Sign`/`SignFile` — currently a byte-for-byte pass-through stub. |
| `tests/` | Black-box tests (`tests/doc`, `tests/reader`, `tests/merge_test.go`). |

Key flows
- Open: `doc.Open` → `reader/file` → `reader.PDFReader` → `documentCore.Resolve` (parse at xref offset, decrypt strings/streams per object, decode filters, cache with FIFO limit).
- Save: `file.Document.Save` → `writer.Write` resolves every object number as `Ref{n,0}` and rewrites streams, re-encoding only a single-name `/Filter`.
- Merge/Split: `objectCloner` deep-copies page dicts (with inherited MediaBox/CropBox/Resources/Rotate) into a fresh `builtDocument`.

Graph: `graphify-out/GRAPH_REPORT.md` and `graphify-out/graph.json` (regenerate with `graphify update .`).
