# Graph Report - gpdf  (2026-11-14)

## Corpus Check
- cluster-only mode — file stats not available

## Summary
- 3121 nodes · 17284 edges · 154 communities (83 shown, 71 thin omitted)
- Extraction: 95% EXTRACTED · 5% INFERRED · 0% AMBIGUOUS · INFERRED: 843 edges (avg confidence: 0.87)
- Token cost: 0 input · 0 output

## Graph Freshness
- Built from commit: `f666bac6`
- Run `git rev-parse HEAD` and compare to check if the graph is stale.
- Run `graphify update .` after code changes (no API cost).

## Community Hubs (Navigation)
- Page
- scanner.go
- syntax/impl/tokenizer.go
- Name
- DocumentBuilder
- Op
- Drawer
- Color
- New
- detectLinearization
- testing.T
- LineStyle
- PageAccess
- documentCore
- io.Writer
- pdfDocument
- EmbeddableFont
- Open
- layout.go
- TextBoxObject
- pdftest.go
- FlowBuilder
- PageText
- Font
- PDFVersion
- stream/fuzz_test.go
- steps
- ITableBuilder
- ExtractLayout
- builtDocument
- with
- cms.go
- writer_impl.go
- io.ReaderAt
- contentTokenizer
- parser.go
- security/standard.go
- LayoutExtractor
- ParserImpl
- validation.go
- testing.B
- StreamWriter
- NewPDFReader
- crypto/x509.Certificate
- DefaultStandardSubstitutes
- .setErr
- steps
- memoryRegistry
- MergeTo
- ListObject
- bidi.go
- Decode
- Font
- Still open
- TextLayoutOptions
- shape_test.go
- decode.go
- steps
- reader_impl.go
- RedactOptions
- corpus
- ExtractGeoViewports
- gpos.go
- Support
- packer
- TaggingAccess
- .addRun
- arabic.go
- Object
- SearchPages
- planLinearization
- runlength.go
- NewPDFWriterWithOptions
- Writer
- mockFont
- Usage
- Verify
- Handle
- String
- Page
- pagesize.go
- TextStyle
- .Subset
- LineObject
- ContentSearcher
- saslprep.go
- project.yml
- StructureProvider
- bitWriter
- Tokenizer
- CFFFont
- TableImpl
- RectObject
- Document
- Table
- ObjectResolver
- structWalker
- pageConfig
- objHeaderBefore
- OutlineProvider
- TestSaveLinearizedWritesUsableHintTables
- .DecodeWithParams
- ImageExtractor
- multiFlag
- srgb_icc.go
- CLAUDE.md
- suggested_commands.md
- task_completion_checklist.md
- github.com/gsoultan/gpdf
- orderedRun
- Boolean
- OpenReaderWithPassword
- TableCell
- Document
- .BuildFields

## God Nodes (most connected - your core abstractions)
1. `Name` - 333 edges
2. `Object` - 281 edges
3. `Ref` - 255 edges
4. `Stream` - 164 edges
5. `Page` - 158 edges
6. `Trailer` - 145 edges
7. `DocumentBuilder` - 140 edges
8. `Catalog` - 138 edges
9. `PageAccess` - 129 edges
10. `Outlines` - 128 edges

## Surprising Connections (you probably didn't know these)
- `gPDF — architecture map` --references--> `textOpsWriter`  [INFERRED]
  .serena/memories/architecture.md → doc/builder_content.go
- `gPDF — architecture map` --references--> `shapedTextArg()`  [INFERRED]
  .serena/memories/architecture.md → doc/builder_content.go
- `gPDF — architecture map` --references--> `textStateTracker`  [INFERRED]
  .serena/memories/architecture.md → doc/builder_content.go
- `Medium` --references--> `textStateTracker`  [INFERRED]
  .serena/memories/review_findings.md → doc/builder_content.go
- `gPDF — project overview` --references--> `MergeTo()`  [INFERRED]
  .serena/memories/project_overview.md → doc/merge_stream.go

## Import Cycles
- 2-file cycle: `doc/example_split_test.go -> doc/merge_test.go -> doc/example_split_test.go`
- 2-file cycle: `doc/example_split_test.go -> doc/split_test.go -> doc/example_split_test.go`
- 2-file cycle: `doc/merge_test.go -> doc/split_test.go -> doc/merge_test.go`
- 2-file cycle: `stream/defaults/defaults.go -> stream/fuzz_test.go -> stream/defaults/defaults.go`
- 3-file cycle: `doc/example_split_test.go -> doc/merge_test.go -> doc/split_test.go -> doc/example_split_test.go`
- 3-file cycle: `doc/example_split_test.go -> doc/split_test.go -> doc/merge_test.go -> doc/example_split_test.go`
- 3-file cycle: `stream/dct/dct.go -> stream/fuzz_test.go -> stream/defaults/defaults.go -> stream/dct/dct.go`
- 3-file cycle: `stream/crypt/crypt.go -> stream/fuzz_test.go -> stream/defaults/defaults.go -> stream/crypt/crypt.go`
- 3-file cycle: `stream/ascii85/ascii85.go -> stream/fuzz_test.go -> stream/defaults/defaults.go -> stream/ascii85/ascii85.go`
- 3-file cycle: `stream/defaults/defaults.go -> stream/jpx/jpx.go -> stream/fuzz_test.go -> stream/defaults/defaults.go`
- 3-file cycle: `stream/ccitt/ccitt.go -> stream/fuzz_test.go -> stream/defaults/defaults.go -> stream/ccitt/ccitt.go`
- 3-file cycle: `stream/defaults/defaults.go -> stream/jbig2/jbig2.go -> stream/fuzz_test.go -> stream/defaults/defaults.go`
- 3-file cycle: `stream/defaults/defaults.go -> stream/flate/flate.go -> stream/fuzz_test.go -> stream/defaults/defaults.go`
- 3-file cycle: `stream/defaults/defaults.go -> stream/lzw/lzw.go -> stream/fuzz_test.go -> stream/defaults/defaults.go`
- 3-file cycle: `stream/asciihex/asciihex.go -> stream/fuzz_test.go -> stream/defaults/defaults.go -> stream/asciihex/asciihex.go`
- 3-file cycle: `stream/defaults/defaults.go -> stream/runlength/runlength.go -> stream/fuzz_test.go -> stream/defaults/defaults.go`

## Communities (154 total, 71 thin omitted)

### Community 0 - "Page"
Cohesion: 0.19
Nodes (27): Document, embeddedFontObjects, graphicRun, Page, Saver, Options, AcroForm, Action (+19 more)

### Community 1 - "scanner.go"
Cohesion: 0.12
Nodes (10): GlyphEncoder, Metrics, Registry, registerPaths(), scanDir(), writeUTF16BE(), surrogates(), LinuxScanner (+2 more)

### Community 2 - "syntax/impl/tokenizer.go"
Cohesion: 0.22
Nodes (9): tokenizer, hexByte(), hexVal(), isDelimiter(), isHex(), isWhitespace(), numberToken(), Token (+1 more)

### Community 3 - "Name"
Cohesion: 0.03
Nodes (12): OutlineItem, Ref, StructElem, StructTreeRoot, Array, Name, makePageWithMediaBox(), TestExtractLayout_CapturesPositionAndStyle() (+4 more)

### Community 4 - "DocumentBuilder"
Cohesion: 0.03
Nodes (11): annotBuilder, DocumentBuilder, documentPart, FormBuilder, ImageBuilder, MetadataBuilder, OutlineBuilder, PageBuilder (+3 more)

### Community 5 - "Op"
Cohesion: 0.06
Nodes (86): TestParseStrayDelimiterTerminates(), Parser, ctokenKind, Dict, cmapToken, codeRange, contentSource, abs64() (+78 more)

### Community 6 - "Drawer"
Cohesion: 0.10
Nodes (6): defaultBPC(), defaultColorSpace(), lastPage(), NewDrawer(), Figure, Drawer

### Community 7 - "Color"
Cohesion: 0.09
Nodes (13): Op, balancedMarkedContent(), circlePathOps(), ellipsePathOps(), polygonPathOps(), roundedRectPathOps(), fillColorOps(), strokeStateOps() (+5 more)

### Community 8 - "New"
Cohesion: 0.06
Nodes (71): main(), main(), main(), createLogo(), main(), New(), R(), TestDrawImage() (+63 more)

### Community 9 - "detectLinearization"
Cohesion: 0.33
Nodes (5): buildLinearizationInfo(), detectLinearization(), intFromDict(), isLinearized(), LinearizationInfo

### Community 10 - "testing.T"
Cohesion: 0.09
Nodes (58): TestStreamParser_ImplementsInterface(), TestTable_CellLookup(), baseEncoding(), TestEncodingTablesComplete(), TestDecodeSpecExample(), parseInline(), TestMalformedNumbers(), TestTokenizerNameAndWhitespace() (+50 more)

### Community 11 - "LineStyle"
Cohesion: 0.06
Nodes (11): addGraphicRunWithState(), circlePathOps(), ellipsePathOps(), fillColorOps(), NewDrawer(), polygonPathOps(), roundedRectPathOps(), strokeStateOps() (+3 more)

### Community 12 - "PageAccess"
Cohesion: 0.09
Nodes (11): PageAccess, alignLine(), countWordSpaces(), defaultFont(), fontWidthFunc(), measureText(), NewDrawer(), wrapLines() (+3 more)

### Community 13 - "documentCore"
Cohesion: 0.10
Nodes (6): incrementalUpdate, cryptFilterName(), documentCore, TestExtractFromObjectStreamChecksObjectNumber(), generationProvider, Entry

### Community 14 - "io.Writer"
Cohesion: 0.09
Nodes (13): Filter, Filter, Filter, Filter, Filter, Filter, Filter, Filter (+5 more)

### Community 15 - "pdfDocument"
Cohesion: 0.04
Nodes (31): NewStreamParser(), TestStreamParser_InlineImage(), TestStreamParser_LineContinuation(), TestStreamParser_OctalEscape(), TestStreamParser_OperandTypes(), TestStreamParser_Parse(), mergeInherited(), outlineItemFirstRef() (+23 more)

### Community 16 - "EmbeddableFont"
Cohesion: 0.06
Nodes (24): aliasedFont, shapedTextArg(), buildCIDSet(), buildCIDWidthArray(), buildEmbeddedFontObjects(), embeddedProgramName(), fontBBoxArray(), newEmbeddedFontUsage() (+16 more)

### Community 17 - "Open"
Cohesion: 0.07
Nodes (39): cleanup(), main(), createDummyPDF(), main(), openFromURL(), main(), cleanup(), main() (+31 more)

### Community 18 - "layout.go"
Cohesion: 0.35
Nodes (6): FontWidthFunc, LineRectFunc, LineWidthFunc, WrapLines(), WrapLinesDynamic(), WrapLinesRect()

### Community 19 - "TextBoxObject"
Cohesion: 0.05
Nodes (7): At(), HeadingObject, ImageObject, Pt, TextAlignment, TextBoxObject, TextObject

### Community 20 - "pdftest.go"
Cohesion: 0.09
Nodes (31): Flate(), Builder, New(), PNGUp(), obj, roundTripSeeds(), fuzzSeeds(), contentOf() (+23 more)

### Community 21 - "FlowBuilder"
Cohesion: 0.15
Nodes (4): FlowBuilder, FlowOptions, DefaultImageStyle(), ImageStyle

### Community 22 - "PageText"
Cohesion: 0.33
Nodes (3): PageDecorator, PageText, TextAlign

### Community 24 - "PDFVersion"
Cohesion: 0.13
Nodes (7): CatalogProvider, PDFVersion, defaultVersion(), detectPDFVersion(), digitAt(), parseVersionHeader(), parseVersionName()

### Community 25 - "stream/fuzz_test.go"
Cohesion: 0.22
Nodes (13): NewFilter(), NewFilter(), NewFilter(), NewFilter(), NewFilter(), RegisterStandardFilters(), Filter, NewFilter() (+5 more)

### Community 26 - "steps"
Cohesion: 0.06
Nodes (33): name, permissions, runs-on, steps, strategy, build-mode, language, jobs (+25 more)

### Community 27 - "ITableBuilder"
Cohesion: 0.06
Nodes (10): cellLayout, ITableBuilder, ImageSide, computeImageDimensions(), CellImageSpec, TableBuilder, TableBuilderAPI, TableCellImageSpec (+2 more)

### Community 28 - "ExtractLayout"
Cohesion: 0.04
Nodes (47): LayoutReader, candidateRow, ColorRGB, ExtractLayout(), clusterAnchors(), DetectTables(), detectTablesOnPage(), extractTableFromGroup() (+39 more)

### Community 29 - "builtDocument"
Cohesion: 0.13
Nodes (6): builtDocument, newObjectCloner(), objectCloner, copyInfo(), copyPages(), newPageTree()

### Community 30 - "with"
Cohesion: 0.06
Nodes (30): runs-on, steps, strategy, default, description, fuzztime, jobs, fuzz (+22 more)

### Community 31 - "cms.go"
Cohesion: 0.14
Nodes (20): algorithmIdentifier, attribute, digestOID(), explicit0(), hashFromOID(), parseCertificateSet(), parseCMS(), pick() (+12 more)

### Community 32 - "writer_impl.go"
Cohesion: 0.11
Nodes (20): headerFor(), validateDocumentGraph(), validateDocumentGraphWithOptions(), validateIncrementalDocumentGraph(), validateObjectRefs(), newPacker(), VersionedDocument, collectObjects() (+12 more)

### Community 33 - "io.ReaderAt"
Cohesion: 0.12
Nodes (14): decodeXRefStreamEntry(), findLastXRefOffset(), findStartXRef(), intFromModelObj(), parseStartXRefOffset(), readBEField(), repairResult, readScanChunk() (+6 more)

### Community 34 - "contentTokenizer"
Cohesion: 0.22
Nodes (8): hexByte(), hexVal(), isHexChar(), isSpace(), newContentTokenizer(), contentTokenizer, ctoken, StreamParser

### Community 36 - "security/standard.go"
Cohesion: 0.05
Nodes (43): HexString, CryptFilterDecryptor, cryptMethod, Decryptor, MetadataEncryption, aes256Password(), aesNoPadIV(), authenticateAES256() (+35 more)

### Community 39 - "validation.go"
Cohesion: 0.26
Nodes (11): appendValidationError(), appendValidationWarn(), isValidationLevel(), validationError(), validationErrorOrNil(), validationWarn(), ValidationDiagnostic, ValidationLevel (+3 more)

### Community 40 - "testing.B"
Cohesion: 0.24
Nodes (19): BenchmarkParseObject(), benchBuilder(), benchDocument(), BenchmarkBuild200Pages(), BenchmarkBuildAndSave(), BenchmarkBuildAndSaveClassic(), BenchmarkCompressImages(), BenchmarkLayout() (+11 more)

### Community 42 - "NewPDFReader"
Cohesion: 0.10
Nodes (23): TestStrictModeRejectsMissingStartXRef(), TestTolerantModeRecoversMissingStartXRef(), TestValidateDocumentReportsStructuralStatus(), TestValidateRejectsInvalidLevel(), NewPDFReader(), minimalPDFBytes(), safeByte(), TestMinimalPDFOffsets() (+15 more)

### Community 43 - "crypto/x509.Certificate"
Cohesion: 0.28
Nodes (10): NewCMSSigner(), CMSSigner, popplerSig(), samplePDF(), testCert(), TestSecondSignatureKeepsFirstValid(), TestSignAndVerify(), TestTimestampAndValidationData() (+2 more)

### Community 44 - "DefaultStandardSubstitutes"
Cohesion: 0.13
Nodes (18): DefaultFallback(), fileScore(), NewFallback(), parseFaces(), GetDefaultScanner(), DefaultStandardSubstitutes(), loadFont(), StandardSubstitutes() (+10 more)

### Community 45 - ".setErr"
Cohesion: 0.05
Nodes (22): Attachment, CollectionField, CollectionFieldType, CollectionOptions, CollectionView, FileRelationship, components(), decodeJPEGToRaw() (+14 more)

### Community 46 - "steps"
Cohesion: 0.09
Nodes (21): env, GPDF_REQUIRE_TOOLS, VERAPDF_SHA256, VERAPDF_VERSION, jobs, test, name, on (+13 more)

### Community 49 - "MergeTo"
Cohesion: 0.20
Nodes (18): CopyOptions, MergeTo(), SplitTo(), maxSelectedPages(), ParsePageRanges(), 2026-11-14 (merge/split for proteus-server), checkWithPoppler(), openPDF() (+10 more)

### Community 50 - "ListObject"
Cohesion: 0.11
Nodes (5): toRoman(), ListObject, DefaultListStyle(), ListMarker, ListStyle

### Community 51 - "bidi.go"
Cohesion: 0.17
Nodes (18): TextDirection, BaseDirection(), classify(), Direction, Run, isArabicLetterScript(), IsRTL(), Levels() (+10 more)

### Community 52 - "Decode"
Cohesion: 0.19
Nodes (15): Encode(), encodePNG(), encodeTIFF(), pngFilter(), sign(), Decode(), decodePNG(), decodeTIFF() (+7 more)

### Community 53 - "Font"
Cohesion: 0.13
Nodes (3): Font, GSUB, tableRecord

### Community 54 - "Still open"
Cohesion: 0.17
Nodes (11): allScriptFeatures(), featureLookups(), featureTables(), langSysFeatures(), parseCoverageTable(), parseGSUBType1(), scriptFeatures(), Critical / high (+3 more)

### Community 55 - "TextLayoutOptions"
Cohesion: 0.12
Nodes (3): countWordSpaces(), TextBuilder, TextLayoutOptions

### Community 56 - "shape_test.go"
Cohesion: 0.18
Nodes (14): cffIndex(), cffIndexCount(), cffString(), parseCFFCharset(), parseCFFDict(), TestGPOSPairKerning(), loadFont(), TestCFFDictAndCharset() (+6 more)

### Community 57 - "decode.go"
Cohesion: 0.24
Nodes (11): DecodePNG(), DecodePNGToRaw(), DetectDimensions(), HasAlpha(), isGrayImage(), nrgbaAt(), ProcessImage(), Decoded (+3 more)

### Community 58 - "steps"
Cohesion: 0.11
Nodes (17): runs-on, steps, jobs, bench, name, on, pull_request, workflow_dispatch (+9 more)

### Community 59 - "reader_impl.go"
Cohesion: 0.16
Nodes (20): boundedWriter, DefaultReaderOptions(), normalizeReaderOptions(), ParseLimits, ParserMode, PDFReader, applyFilterChain(), applyFilters() (+12 more)

### Community 61 - "corpus"
Cohesion: 0.13
Nodes (17): sample, PDFUALevel, Redact(), RedactionReport, basic(), corpus(), corpusDir(), reopen() (+9 more)

### Community 62 - "ExtractGeoViewports"
Cohesion: 0.12
Nodes (11): CatalogReader, embeddedFileFromSpec(), ExtractEmbeddedFiles(), EmbeddedFile, textOf(), walkNameTree(), ExtractGeoViewports(), numberOf() (+3 more)

### Community 63 - "gpos.go"
Cohesion: 0.23
Nodes (11): coverageGlyphs(), forEachLookupSubtable(), parsePairPos(), sortedInts(), TestParsePairPosClassBased(), valueRecordSize(), xAdvance(), classDef (+3 more)

### Community 64 - "Support"
Cohesion: 0.08
Nodes (21): EncodeBytes(), escapeContentName(), NewEncoder(), sortedKeys(), TestEncodeBytes(), writeHexString(), writeLiteralString(), Encoder (+13 more)

### Community 65 - "packer"
Cohesion: 0.32
Nodes (5): objectGroup, packable(), offsetWriter, packer, xrefEntry

### Community 67 - ".addRun"
Cohesion: 0.17
Nodes (8): DocumentBuilder, symbolEncode(), syntheticBoldStroke(), winAnsiEncode(), zapfDingbatsEncode(), pageSpec, textOpsWriter, textStateTracker

### Community 68 - "arabic.go"
Cohesion: 0.32
Nodes (9): ArabicForms(), IsArabicLetter(), JoiningTypeOf(), LamAlefLigature(), PresentationForm(), TestPresentationForms(), nextBase(), Form (+1 more)

### Community 69 - "Object"
Cohesion: 0.03
Nodes (56): asciiFileName(), md5sum(), buildToUnicodeCMap(), buildToUnicodeSymbolCMap(), buildToUnicodeWinAnsiCMap(), buildToUnicodeZapfDingbatsCMap(), newImageOpsWriter(), copyPageDict() (+48 more)

### Community 70 - "SearchPages"
Cohesion: 0.15
Nodes (7): SearchResult, ContentExtractor, SearchPages(), TestSearchPages_EmptyKeyword(), TestSearchPages_FindsKeywordOnCorrectPage(), TestSearchPages_MultipleKeywords(), TestSearchPages_RecordsIndices()

### Community 71 - "planLinearization"
Cohesion: 0.19
Nodes (9): hintTableSet, indirectObject, linearizationPlan, bitsFor(), isPagesNode(), planLinearization(), remapRefs(), sortedPlace() (+1 more)

### Community 72 - "runlength.go"
Cohesion: 0.23
Nodes (9): Filter, canEncodeRepeatRun(), copyLiteralRun(), copyRepeatedRun(), decode(), decodeLiteral(), encode(), writeLiteralRun() (+1 more)

### Community 73 - "NewPDFWriterWithOptions"
Cohesion: 0.22
Nodes (11): TestApplyFiltersWithLimitsRejectsOversizedDecode(), NewStandardRegistry(), FuzzFilters(), NewRegistry(), DefaultWriterOptions(), normalizeWriterOptions(), NewStreamWriter(), writeEncryptedWithSeed() (+3 more)

### Community 76 - "Usage"
Cohesion: 0.15
Nodes (12): Checking conformance, Extracting Pages, gPDF, Merging from Remote Sources (URL/S3), Merging PDFs, Reading a PDF, Reading order, updating in place, redacting, Searching and Replacing (+4 more)

### Community 77 - "Verify"
Cohesion: 0.06
Nodes (35): rectArray(), jpegHasAdobeMarker(), GeoViewport, Support, PDFDate(), Model3D, FormatDate(), ParseDate() (+27 more)

### Community 78 - "Handle"
Cohesion: 0.22
Nodes (4): Handle, Support, LayerBuilder, Spec

### Community 80 - "Page"
Cohesion: 0.39
Nodes (4): GraphicRun, ImageRun, TextRun, Page

### Community 81 - "pagesize.go"
Cohesion: 0.40
Nodes (3): Custom(), Size, TestPageSizeAPI()

### Community 82 - "TextStyle"
Cohesion: 0.12
Nodes (5): NewRichText(), RichText, RichTextSegment, DefaultTextStyle(), TextStyle

### Community 86 - "saslprep.go"
Cohesion: 0.33
Nodes (8): checkBidi(), isLCat(), isMappedToNothing(), isNonASCIISpace(), isProhibited(), isRandAL(), saslprep(), TestSASLprep()

### Community 87 - "project.yml"
Cohesion: 0.22
Nodes (8): encoding, excluded_tools, ignore_all_files_in_gitignore, ignored_paths, initial_prompt, languages, project_name, read_only

### Community 90 - "Tokenizer"
Cohesion: 0.25
Nodes (6): BenchmarkTokenizer(), TestTokenizerCommentAdvancesOffset(), TestTokenizerStrayDelimitersMakeProgress(), NewTokenizer(), TestTokenizer_CurrentOffset_AfterTokens(), Tokenizer

### Community 94 - "Document"
Cohesion: 0.07
Nodes (4): ImageReader, Document, NewDocument(), ImageInfo

### Community 97 - "structWalker"
Cohesion: 0.13
Nodes (11): ContentReader, ExtractStructuredText(), ExtractTextInReadingOrder(), TextElement, isCellRole(), JoinTextElements(), layoutOrderText(), nameOf() (+3 more)

### Community 99 - "objHeaderBefore"
Cohesion: 0.60
Nodes (5): isPDFWhitespace(), isDigit(), isRegular(), objHeaderBefore(), scanChunk()

### Community 101 - "TestSaveLinearizedWritesUsableHintTables"
Cohesion: 0.50
Nodes (3): linValue(), TestSaveLinearizedWritesUsableHintTables(), bitReader

### Community 102 - ".DecodeWithParams"
Cohesion: 0.20
Nodes (5): bitReader, bitWriter, Filter, decode(), Encode()

### Community 106 - "srgb_icc.go"
Cohesion: 0.10
Nodes (24): FloatingImage, pageOutputIntent, ColorFromHex(), Gray(), NewICCProfile(), SRGBProfile(), DefaultImageLayout(), ImageLayout (+16 more)

### Community 143 - "OpenReaderWithPassword"
Cohesion: 0.20
Nodes (9): OpenReaderWithPassword(), FuzzTextString(), DecodeTextString(), IsTextStringBOM(), TestDecodePDFDocEncoding(), TestAES256PasswordIsSASLprepped(), TestPDFAOutput(), TestEncryptedObjectStreams() (+1 more)

### Community 145 - "Document"
Cohesion: 0.22
Nodes (11): dictInt(), floatFromObj(), main(), openPDF(), printDocumentInfo(), printMetadata(), printPages(), printSearch() (+3 more)

### Community 146 - ".BuildFields"
Cohesion: 0.24
Nodes (13): buttonAppearance(), emptyAppearance(), fontResources(), formXObject(), frameOps(), num(), pdfLiteral(), symbolAppearance() (+5 more)

## Knowledge Gaps
- **1 isolated node(s):** `github.com/gsoultan/gpdf`
  These have ≤1 connection - possible missing edges or undocumented components. (Counts symbols only; 462 node(s) total have ≤1 connection when file, concept and rationale nodes are included.)
- **71 thin communities (<3 nodes) omitted from report** — run `graphify query` to explore isolated nodes.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **Why does `DocumentBuilder` connect `DocumentBuilder` to `Page`, `Drawer`, `Color`, `New`, `testing.T`, `LineStyle`, `PageAccess`, `EmbeddableFont`, `.BuildFields`, `TextBoxObject`, `pdftest.go`, `FlowBuilder`, `PageText`, `Font`, `ITableBuilder`, `testing.B`, `StreamWriter`, `DefaultStandardSubstitutes`, `.setErr`, `.ApplyPageSize`, `bidi.go`, `TextLayoutOptions`, `corpus`, `Support`, `Object`, `Verify`, `Handle`, `pageConfig`, `srgb_icc.go`?**
  _High betweenness centrality (0.115) - this node is a cross-community bridge._
- **Are the 44 inferred relationships involving `Name` (e.g. with `.Fields()` and `.SigFlags()`) actually correct?**
  _`Name` has 44 INFERRED edges - model-reasoned connections that need verification._
- **What connects `github.com/gsoultan/gpdf` to the rest of the system?**
  _1 weakly-connected nodes found - possible documentation gaps or missing edges._
- **Should `scanner.go` be split into smaller, more focused modules?**
  _Cohesion score 0.12462462462462462 - nodes in this community are weakly interconnected._
- **Why does `Name` connect `Name` to `Page`, `Op`, `Color`, `New`, `detectLinearization`, `testing.T`, `LineStyle`, `documentCore`, `io.Writer`, `pdfDocument`, `EmbeddableFont`, `Document`, `.BuildFields`, `TextBoxObject`, `Open`, `FlowBuilder`, `OpenReaderWithPassword`, `pdftest.go`, `ITableBuilder`, `ExtractLayout`, `builtDocument`, `writer_impl.go`, `io.ReaderAt`, `contentTokenizer`, `security/standard.go`, `ParserImpl`, `validation.go`, `StreamWriter`, `NewPDFReader`, `crypto/x509.Certificate`, `.setErr`, `MergeTo`, `TextLayoutOptions`, `reader_impl.go`, `ExtractGeoViewports`, `Support`, `packer`, `TaggingAccess`, `.addRun`, `Object`, `planLinearization`, `NewPDFWriterWithOptions`, `Verify`, `Handle`, `String`, `Page`, `Document`, `structWalker`, `pageConfig`, `TestSaveLinearizedWritesUsableHintTables`, `.DecodeWithParams`?**
  _High betweenness centrality (0.106) - this node is a cross-community bridge._
- **Should `Name` be split into smaller, more focused modules?**
  _Cohesion score 0.03471444568868981 - nodes in this community are weakly interconnected._
- **Why does `Object` connect `Object` to `Page`, `Name`, `Op`, `Color`, `New`, `testing.T`, `LineStyle`, `documentCore`, `io.Writer`, `pdfDocument`, `EmbeddableFont`, `Document`, `.BuildFields`, `stream/fuzz_test.go`, `ExtractLayout`, `builtDocument`, `writer_impl.go`, `io.ReaderAt`, `contentTokenizer`, `security/standard.go`, `ParserImpl`, `StreamWriter`, `.setErr`, `reader_impl.go`, `ExtractGeoViewports`, `Support`, `packer`, `planLinearization`, `Verify`, `Handle`, `Document`, `ObjectResolver`, `structWalker`, `pageConfig`?**
  _High betweenness centrality (0.070) - this node is a cross-community bridge._