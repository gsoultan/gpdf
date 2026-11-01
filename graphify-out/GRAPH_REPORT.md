# Graph Report - gpdf  (2026-11-01)

## Corpus Check
- 344 files · ~179,833 words
- Verdict: corpus is large enough that graph structure adds value.
- Unclassified: 8 file(s) not represented in the graph (top: (none) 8)

## Summary
- 3073 nodes · 17178 edges · 148 communities (84 shown, 64 thin omitted)
- Extraction: 95% EXTRACTED · 5% INFERRED · 0% AMBIGUOUS · INFERRED: 776 edges (avg confidence: 0.87)
- Token cost: 0 input · 0 output

## Graph Freshness
- Built from commit: `2437d9eb`
- Run `git rev-parse HEAD` and compare to check if the graph is stale.
- Run `graphify update .` after code changes (no API cost).

## Community Hubs (Navigation)
- Page
- pdf20_test.go
- syntax/impl/tokenizer.go
- Name
- DocumentBuilder
- Op
- Document
- Color
- New
- documentCore
- testing.T
- LineStyle
- PageAccess
- ITableBuilder
- io.Writer
- Integer
- go_pkg_io
- Open
- .layoutTextIntoPages
- TextBoxObject
- pdftest.go
- FlowBuilder
- planLinearization
- Font
- pdfDocument
- OpenReaderWithPassword
- steps
- ReplacesContent
- validation.go
- builtDocument
- with
- cms.go
- PDFWriter
- io.ReaderAt
- contentTokenizer
- parser.go
- Dict
- Document
- gPDF — architecture map
- structWalker
- testing.B
- writer/options.go
- NewPDFReader
- OpenReaderWithPassword
- EmbeddableFont
- .setErr
- steps
- .ApplyPageSize
- registerPaths
- CatalogProvider
- ListObject
- bidi.go
- Decode
- Font
- EmbeddableFont
- Sign
- Verify
- decode.go
- steps
- FilterRegistry
- Still open
- ExtractGeoViewports
- ParserImpl
- TableCellSpec
- Support
- packer
- TaggingAccess
- imageRun
- arabic.go
- EncodeBytes
- SearchPages
- StandardDecryptor
- runlength.go
- time.Time
- WriteSeeker
- mockFont
- Usage
- buildSignatureUpdate
- Handle
- Pt
- ICCProfile
- structWalker
- TextStyle
- .Subset
- ExtractLayout
- ReaderOptions
- saslprep.go
- project.yml
- CatalogReader
- PDFALevel
- MetadataBuilder
- ImageInfo
- TableImpl
- RectObject
- OutlineBuilder
- Table
- CollectionOptions
- memoryRegistry
- .pageHeight
- objHeaderBefore
- CellStyle
- TestSaveLinearizedWritesUsableHintTables
- Encode
- ImageExtractor
- FormBuilder
- multiFlag
- Tokenizer
- CLAUDE.md
- suggested_commands.md
- task_completion_checklist.md
- github.com/gsoultan/gpdf
- .SaveIncremental
- .EmbedStandardFonts
- ImageBuilder
- LayoutExtractor
- Support
- PageBuilder
- RedactOptions

## God Nodes (most connected - your core abstractions)
1. `Name` - 328 edges
2. `Object` - 284 edges
3. `Ref` - 259 edges
4. `Page` - 168 edges
5. `Stream` - 161 edges
6. `Trailer` - 155 edges
7. `Catalog` - 148 edges
8. `DocumentBuilder` - 141 edges
9. `Outlines` - 138 edges
10. `AcroForm` - 137 edges

## Surprising Connections (you probably didn't know these)
- `Still open` --references--> `Encoder`  [INFERRED]
  .serena/memories/review_findings.md → content/encode.go
- `gPDF — architecture map` --references--> `DocumentBuilder`  [INFERRED]
  .serena/memories/architecture.md → doc/builder.go
- `Medium` --references--> `textStateTracker`  [INFERRED]
  .serena/memories/review_findings.md → doc/builder_content.go
- `gPDF — architecture map` --references--> `builtDocument`  [INFERRED]
  .serena/memories/architecture.md → doc/built_document.go
- `gPDF — architecture map` --references--> `objectCloner`  [INFERRED]
  .serena/memories/architecture.md → doc/cloner.go

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
- 3-file cycle: `stream/asciihex/asciihex.go -> stream/fuzz_test.go -> stream/defaults/defaults.go -> stream/asciihex/asciihex.go`
- 3-file cycle: `stream/defaults/defaults.go -> stream/lzw/lzw.go -> stream/fuzz_test.go -> stream/defaults/defaults.go`
- 3-file cycle: `stream/defaults/defaults.go -> stream/runlength/runlength.go -> stream/fuzz_test.go -> stream/defaults/defaults.go`

## Communities (148 total, 64 thin omitted)

### Community 0 - "Page"
Cohesion: 0.12
Nodes (48): floatFromObj(), fontResources(), patchDoc, incrementalUpdate, AcroForm, Action, Annotation, Boolean (+40 more)

### Community 1 - "pdf20_test.go"
Cohesion: 0.21
Nodes (14): cleanup(), cleanup(), pdfFontFlags(), Document, embeddedFontObjects, graphicRun, orderedRun, Page (+6 more)

### Community 2 - "syntax/impl/tokenizer.go"
Cohesion: 0.23
Nodes (9): tokenizer, hexByte(), hexVal(), isDelimiter(), isHex(), isWhitespace(), numberToken(), Token (+1 more)

### Community 3 - "Name"
Cohesion: 0.04
Nodes (14): Support, OutlineItem, StructElem, StructTreeRoot, Array, Name, Entry, LinkAnnotation (+6 more)

### Community 4 - "DocumentBuilder"
Cohesion: 0.04
Nodes (7): annotBuilder, DocumentBuilder, documentPart, PaintOrder, TaggedContentBuilder, TaggedImageBuilder, TextBuilder

### Community 5 - "Op"
Cohesion: 0.05
Nodes (92): Parser, writeUTF16BE(), cmapToken, codeRange, ColorRGB, contentSource, abs64(), addBFRange() (+84 more)

### Community 7 - "Color"
Cohesion: 0.11
Nodes (12): Op, circlePathOps(), ellipsePathOps(), polygonPathOps(), roundedRectPathOps(), fillColorOps(), strokeStateOps(), PathBuilder (+4 more)

### Community 8 - "New"
Cohesion: 0.06
Nodes (73): main(), main(), main(), createLogo(), main(), NewDrawer(), New(), DefaultStandardSubstitutes() (+65 more)

### Community 9 - "documentCore"
Cohesion: 0.12
Nodes (4): cryptFilterName(), documentCore, LinearizationInfo, TestExtractFromObjectStreamChecksObjectNumber()

### Community 10 - "testing.T"
Cohesion: 0.08
Nodes (53): TestStreamParser_ImplementsInterface(), TestGPOSPairKerning(), loadFont(), TestParseCFF(), TestParseCFFCIDKeyed(), TestShapeArabicWithGSUB(), TestShapeLigaturesAndToUnicode(), TestShapePresentationFormsFallback() (+45 more)

### Community 11 - "LineStyle"
Cohesion: 0.05
Nodes (15): addGraphicRunWithState(), circlePathOps(), ellipsePathOps(), fillColorOps(), polygonPathOps(), roundedRectPathOps(), strokeStateOps(), LineObject (+7 more)

### Community 12 - "PageAccess"
Cohesion: 0.06
Nodes (15): defaultBPC(), defaultColorSpace(), lastPage(), NewDrawer(), PageAccess, alignLine(), countWordSpaces(), defaultFont() (+7 more)

### Community 13 - "ITableBuilder"
Cohesion: 0.08
Nodes (3): ITableBuilder, TableBuilder, TableBuilderAPI

### Community 14 - "io.Writer"
Cohesion: 0.08
Nodes (13): Filter, Filter, Filter, Filter, Filter, Filter, Filter, Filter (+5 more)

### Community 15 - "Integer"
Cohesion: 0.08
Nodes (23): asciiFileName(), md5sum(), buildToUnicodeCMap(), buildToUnicodeSymbolCMap(), buildToUnicodeWinAnsiCMap(), buildToUnicodeZapfDingbatsCMap(), copyPageDict(), collectionItem() (+15 more)

### Community 16 - "go_pkg_io"
Cohesion: 0.17
Nodes (16): TestApplyFiltersWithLimitsRejectsOversizedDecode(), NewFilter(), NewFilter(), NewFilter(), NewFilter(), NewFilter(), NewStandardRegistry(), RegisterStandardFilters() (+8 more)

### Community 17 - "Open"
Cohesion: 0.07
Nodes (40): main(), createDummyPDF(), main(), openFromURL(), main(), main(), main(), main() (+32 more)

### Community 18 - ".layoutTextIntoPages"
Cohesion: 0.09
Nodes (14): countWordSpaces(), fontWidthFunc(), extraSpacing(), reverseRunes(), fontSegment, ApproxWidth(), FontWidthFunc, LineRectFunc (+6 more)

### Community 19 - "TextBoxObject"
Cohesion: 0.08
Nodes (3): TextAlignment, TextBoxObject, TextObject

### Community 20 - "pdftest.go"
Cohesion: 0.09
Nodes (31): Flate(), Builder, New(), PNGUp(), obj, BenchmarkReadDocument(), fuzzSeeds(), contentOf() (+23 more)

### Community 22 - "planLinearization"
Cohesion: 0.18
Nodes (8): bitWriter, hintTableSet, linearizationPlan, bitsFor(), isPagesNode(), planLinearization(), sortedPlace(), linearizedWriter

### Community 24 - "pdfDocument"
Cohesion: 0.08
Nodes (10): String, mergeInherited(), MetadataProvider, pdfDocument, PDFVersion, defaultVersion(), detectPDFVersion(), digitAt() (+2 more)

### Community 25 - "OpenReaderWithPassword"
Cohesion: 0.10
Nodes (22): OpenReaderWithPassword(), resolver, before(), diff(), halfTransparentPNG(), imageXObject(), TestAdobeCMYKJPEGGetsInvertedDecode(), TestDefaultPaintOrderIsLayered() (+14 more)

### Community 26 - "steps"
Cohesion: 0.06
Nodes (33): name, permissions, runs-on, steps, strategy, build-mode, language, jobs (+25 more)

### Community 27 - "ReplacesContent"
Cohesion: 0.05
Nodes (41): TestParseStrayDelimiterTerminates(), NewStreamParser(), TestStreamParser_InlineImage(), TestStreamParser_LineContinuation(), TestStreamParser_OctalEscape(), TestStreamParser_OperandTypes(), TestStreamParser_Parse(), ExtractText() (+33 more)

### Community 28 - "validation.go"
Cohesion: 0.26
Nodes (11): appendValidationError(), appendValidationWarn(), isValidationLevel(), validationError(), validationErrorOrNil(), validationWarn(), ValidationDiagnostic, ValidationLevel (+3 more)

### Community 29 - "builtDocument"
Cohesion: 0.10
Nodes (7): builtDocument, isPageTreeNode(), newObjectCloner(), objectCloner, copyInfo(), copyPages(), newPageTree()

### Community 30 - "with"
Cohesion: 0.06
Nodes (30): runs-on, steps, strategy, default, description, fuzztime, jobs, fuzz (+22 more)

### Community 31 - "cms.go"
Cohesion: 0.16
Nodes (20): algorithmIdentifier, attribute, digestOID(), explicit0(), hashFromOID(), parseCertificateSet(), parseCMS(), pick() (+12 more)

### Community 32 - "PDFWriter"
Cohesion: 0.10
Nodes (21): Encryptor, headerFor(), errWriter, generationProvider, validateDocumentGraph(), validateDocumentGraphWithOptions(), validateIncrementalDocumentGraph(), indirectObject (+13 more)

### Community 33 - "io.ReaderAt"
Cohesion: 0.13
Nodes (17): detectLinearization(), isLinearized(), PDFReader, decodeXRefStreamEntry(), findLastXRefOffset(), findStartXRef(), intFromModelObj(), readBEField() (+9 more)

### Community 34 - "contentTokenizer"
Cohesion: 0.20
Nodes (9): hexByte(), hexVal(), isHexChar(), isSpace(), newContentTokenizer(), contentTokenizer, ctoken, ctokenKind (+1 more)

### Community 36 - "Dict"
Cohesion: 0.06
Nodes (41): Support, Dict, outlineItemFirstRef(), isStandardTextFont(), buildLinearizationInfo(), intFromDict(), Decryptor, aes256Password() (+33 more)

### Community 37 - "Document"
Cohesion: 0.12
Nodes (13): dictInt(), main(), openPDF(), printDocumentInfo(), printMetadata(), printPages(), printSearch(), printTextPerPage() (+5 more)

### Community 38 - "gPDF — architecture map"
Cohesion: 0.05
Nodes (33): balancedMarkedContent(), newImageOpsWriter(), shapedTextArg(), symbolEncode(), syntheticBoldStroke(), winAnsiEncode(), zapfDingbatsEncode(), buildCIDSet() (+25 more)

### Community 39 - "structWalker"
Cohesion: 0.16
Nodes (9): ContentReader, ExtractStructuredText(), ExtractTextInReadingOrder(), TextElement, isCellRole(), JoinTextElements(), layoutOrderText(), structureTreeText() (+1 more)

### Community 40 - "testing.B"
Cohesion: 0.24
Nodes (17): BenchmarkParseObject(), benchBuilder(), benchDocument(), BenchmarkBuild200Pages(), BenchmarkBuildAndSave(), BenchmarkBuildAndSaveClassic(), BenchmarkCompressImages(), BenchmarkLayout() (+9 more)

### Community 41 - "writer/options.go"
Cohesion: 0.16
Nodes (11): docStream, TestStreamWriter(), DefaultWriterOptions(), normalizeWriterOptions(), NewStreamWriter(), newDanglingWriterTestDoc(), TestWriteRejectsDanglingReference(), writeEncryptedWithSeed() (+3 more)

### Community 42 - "NewPDFReader"
Cohesion: 0.12
Nodes (22): TestStrictModeRejectsMissingStartXRef(), TestTolerantModeRecoversMissingStartXRef(), TestValidateDocumentReportsStructuralStatus(), TestValidateRejectsInvalidLevel(), NewPDFReader(), NewPDFReaderWithOptions(), minimalPDFBytes(), safeByte() (+14 more)

### Community 43 - "OpenReaderWithPassword"
Cohesion: 0.16
Nodes (14): sample, Redact(), RedactionReport, basic(), corpus(), corpusDir(), save(), squarePNG() (+6 more)

### Community 44 - "EmbeddableFont"
Cohesion: 0.13
Nodes (15): DefaultFallback(), fileScore(), NewFallback(), parseFaces(), GetDefaultScanner(), loadFont(), StandardSubstitutes(), exercise() (+7 more)

### Community 46 - "steps"
Cohesion: 0.09
Nodes (21): env, GPDF_REQUIRE_TOOLS, VERAPDF_SHA256, VERAPDF_VERSION, jobs, test, name, on (+13 more)

### Community 48 - "registerPaths"
Cohesion: 0.27
Nodes (5): registerPaths(), scanDir(), LinuxScanner, MacOSScanner, WindowsScanner

### Community 50 - "ListObject"
Cohesion: 0.11
Nodes (5): toRoman(), ListObject, DefaultListStyle(), ListMarker, ListStyle

### Community 51 - "bidi.go"
Cohesion: 0.18
Nodes (17): TextDirection, BaseDirection(), classify(), Direction, isArabicLetterScript(), IsRTL(), Levels(), Mirror() (+9 more)

### Community 52 - "Decode"
Cohesion: 0.19
Nodes (16): FuzzFilters(), Encode(), encodePNG(), encodeTIFF(), pngFilter(), sign(), Decode(), decodePNG() (+8 more)

### Community 53 - "Font"
Cohesion: 0.06
Nodes (28): cffIndex(), cffIndexCount(), cffString(), parseCFFCharset(), parseCFFDict(), Font, coverageGlyphs(), forEachLookupSubtable() (+20 more)

### Community 54 - "EmbeddableFont"
Cohesion: 0.14
Nodes (5): aliasedFont, getRuneBlock(), fontConfig, EmbeddableFont, Run

### Community 55 - "Sign"
Cohesion: 0.31
Nodes (6): subjectName(), Sign(), SignFile(), Signer, SignOptions, SubFilter

### Community 56 - "Verify"
Cohesion: 0.20
Nodes (15): NewCMSSigner(), CMSSigner, popplerSig(), samplePDF(), testCert(), TestSecondSignatureKeepsFirstValid(), TestSignAndVerify(), TestTimestampAndValidationData() (+7 more)

### Community 57 - "decode.go"
Cohesion: 0.26
Nodes (11): DecodePNG(), DecodePNGToRaw(), DetectDimensions(), HasAlpha(), isGrayImage(), nrgbaAt(), ProcessImage(), Decoded (+3 more)

### Community 58 - "steps"
Cohesion: 0.11
Nodes (17): runs-on, steps, jobs, bench, name, on, pull_request, workflow_dispatch (+9 more)

### Community 59 - "FilterRegistry"
Cohesion: 0.18
Nodes (11): boundedWriter, applyFilterChain(), applyFilters(), applyFiltersWithLimits(), decodeStreamData(), decodeStreamDataWithLimits(), filterParams(), newBoundedWriter() (+3 more)

### Community 60 - "Still open"
Cohesion: 0.05
Nodes (28): GraphicRun, ImageRun, TextRun, rectArray(), jpegHasAdobeMarker(), Page, GeoViewport, Model3D (+20 more)

### Community 61 - "ExtractGeoViewports"
Cohesion: 0.19
Nodes (10): embeddedFileFromSpec(), ExtractEmbeddedFiles(), EmbeddedFile, textOf(), walkNameTree(), ExtractGeoViewports(), numberOf(), pairs() (+2 more)

### Community 63 - "TableCellSpec"
Cohesion: 0.24
Nodes (8): cellLayout, FloatingImage, ImageSide, ImageWrap, computeImageDimensions(), CellImageSpec, TableCellImageSpec, TableCellSpec

### Community 64 - "Support"
Cohesion: 0.13
Nodes (13): inSection(), mcidParents(), mcr(), setStructParents(), StructRow, ElemKind, ElemRef, List (+5 more)

### Community 65 - "packer"
Cohesion: 0.16
Nodes (7): objectGroup, offsetWriter, packer, StreamWriter, fileID(), outputTrailer(), xrefEntry

### Community 67 - "imageRun"
Cohesion: 0.22
Nodes (10): components(), decodeJPEGToRaw(), downsampleRun(), encodeJPEG(), placedSize(), resampleBox(), decodeImageFile(), jpegColorComponents() (+2 more)

### Community 68 - "arabic.go"
Cohesion: 0.32
Nodes (9): ArabicForms(), IsArabicLetter(), JoiningTypeOf(), LamAlefLigature(), PresentationForm(), TestPresentationForms(), nextBase(), Form (+1 more)

### Community 69 - "EncodeBytes"
Cohesion: 0.21
Nodes (9): EncodeBytes(), escapeContentName(), NewEncoder(), sortedKeys(), TestEncodeBytes(), writeHexString(), writeLiteralString(), Encoder (+1 more)

### Community 70 - "SearchPages"
Cohesion: 0.12
Nodes (8): ContentSearcher, SearchResult, ContentExtractor, SearchPages(), TestSearchPages_EmptyKeyword(), TestSearchPages_FindsKeywordOnCorrectPage(), TestSearchPages_MultipleKeywords(), TestSearchPages_RecordsIndices()

### Community 71 - "StandardDecryptor"
Cohesion: 0.19
Nodes (6): cryptMethod, aesDecrypt(), aesEncrypt(), pkcs7Pad(), pkcs7Unpad(), StandardDecryptor

### Community 72 - "runlength.go"
Cohesion: 0.31
Nodes (9): Filter, canEncodeRepeatRun(), copyLiteralRun(), copyRepeatedRun(), decode(), decodeLiteral(), encode(), writeLiteralRun() (+1 more)

### Community 73 - "time.Time"
Cohesion: 0.24
Nodes (7): Attachment, FileRelationship, PDFDate(), FormatDate(), ParseDate(), FuzzParseDate(), TestParseDate()

### Community 76 - "Usage"
Cohesion: 0.15
Nodes (12): Checking conformance, Extracting Pages, gPDF, Merging from Remote Sources (URL/S3), Merging PDFs, Reading a PDF, Reading order, updating in place, redacting, Searching and Replacing (+4 more)

### Community 77 - "buildSignatureUpdate"
Cohesion: 0.27
Nodes (8): AddValidationData(), buildSignatureUpdate(), copyDict(), docMDPLevel(), resolveArray(), resolveDict(), update, ValidationData

### Community 78 - "Handle"
Cohesion: 0.22
Nodes (4): Handle, Support, LayerBuilder, Spec

### Community 79 - "Pt"
Cohesion: 0.14
Nodes (4): At(), HeadingObject, ImageObject, Pt

### Community 80 - "ICCProfile"
Cohesion: 0.17
Nodes (11): pageOutputIntent, NewICCProfile(), SRGBProfile(), buildSRGBICC(), iccCurve(), iccDesc(), iccText(), iccXYZ() (+3 more)

### Community 82 - "TextStyle"
Cohesion: 0.12
Nodes (5): NewRichText(), RichText, RichTextSegment, DefaultTextStyle(), TextStyle

### Community 84 - "ExtractLayout"
Cohesion: 0.09
Nodes (23): LayoutReader, candidateRow, ExtractLayout(), clusterAnchors(), DetectTables(), detectTablesOnPage(), extractTableFromGroup(), inferTableBorder() (+15 more)

### Community 85 - "ReaderOptions"
Cohesion: 0.31
Nodes (8): DefaultReaderOptions(), normalizeReaderOptions(), ParseLimits, ParserMode, NewPDFReaderWithFilters(), NewPDFReaderWithFiltersAndOptions(), ReaderOptions, gPDF — code style and conventions

### Community 86 - "saslprep.go"
Cohesion: 0.38
Nodes (8): checkBidi(), isLCat(), isMappedToNothing(), isNonASCIISpace(), isProhibited(), isRandAL(), saslprep(), TestSASLprep()

### Community 87 - "project.yml"
Cohesion: 0.22
Nodes (8): encoding, excluded_tools, ignore_all_files_in_gitignore, ignored_paths, initial_prompt, languages, project_name, read_only

### Community 89 - "PDFALevel"
Cohesion: 0.32
Nodes (4): BuildXMP(), XMPInput, writePDFUAExtensionSchema(), PDFALevel

### Community 93 - "RectObject"
Cohesion: 0.33
Nodes (3): R(), Rect, RectObject

### Community 96 - "CollectionOptions"
Cohesion: 0.40
Nodes (4): CollectionField, CollectionFieldType, CollectionOptions, CollectionView

### Community 98 - ".pageHeight"
Cohesion: 0.14
Nodes (8): FlowOptions, toFloat(), pageConfig, DefaultImageLayout(), DefaultImageStyle(), ImageLayout, ImageStyle, ImageAlignment

### Community 99 - "objHeaderBefore"
Cohesion: 0.47
Nodes (6): isPDFWhitespace(), parseStartXRefOffset(), isDigit(), isRegular(), objHeaderBefore(), scanChunk()

### Community 100 - "CellStyle"
Cohesion: 0.33
Nodes (4): CellHorizontalAlign, CellSpec, CellStyle, CellVerticalAlign

### Community 101 - "TestSaveLinearizedWritesUsableHintTables"
Cohesion: 0.50
Nodes (3): linValue(), TestSaveLinearizedWritesUsableHintTables(), bitReader

### Community 106 - "Tokenizer"
Cohesion: 0.22
Nodes (6): BenchmarkTokenizer(), TestTokenizerCommentAdvancesOffset(), TestTokenizerStrayDelimitersMakeProgress(), NewTokenizer(), TestTokenizer_CurrentOffset_AfterTokens(), Tokenizer

### Community 145 - "Support"
Cohesion: 0.67
Nodes (3): Support, FieldKind, FieldSpec

## Knowledge Gaps
- **1 isolated node(s):** `github.com/gsoultan/gpdf`
  These have ≤1 connection - possible missing edges or undocumented components. (Counts symbols only; 441 node(s) total have ≤1 connection when file, concept and rationale nodes are included.)
- **64 thin communities (<3 nodes) omitted from report** — run `graphify query` to explore isolated nodes.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **Why does `Name` connect `Name` to `Page`, `Op`, `Document`, `Color`, `New`, `documentCore`, `testing.T`, `LineStyle`, `.EmbedStandardFonts`, `Integer`, `io.Writer`, `go_pkg_io`, `.layoutTextIntoPages`, `TextBoxObject`, `Open`, `FlowBuilder`, `pdftest.go`, `planLinearization`, `pdfDocument`, `OpenReaderWithPassword`, `ReplacesContent`, `validation.go`, `builtDocument`, `PDFWriter`, `io.ReaderAt`, `contentTokenizer`, `Dict`, `Document`, `gPDF — architecture map`, `structWalker`, `writer/options.go`, `NewPDFReader`, `Decode`, `Sign`, `Verify`, `FilterRegistry`, `Still open`, `ExtractGeoViewports`, `ParserImpl`, `TableCellSpec`, `Support`, `packer`, `TaggingAccess`, `EncodeBytes`, `StandardDecryptor`, `buildSignatureUpdate`, `Handle`, `structWalker`, `ExtractLayout`, `.pageHeight`, `TestSaveLinearizedWritesUsableHintTables`?**
  _High betweenness centrality (0.093) - this node is a cross-community bridge._
- **Are the 44 inferred relationships involving `Name` (e.g. with `.Fields()` and `.SigFlags()`) actually correct?**
  _`Name` has 44 INFERRED edges - model-reasoned connections that need verification._
- **What connects `github.com/gsoultan/gpdf` to the rest of the system?**
  _1 weakly-connected nodes found - possible documentation gaps or missing edges._
- **Should `Page` be split into smaller, more focused modules?**
  _Cohesion score 0.11692757772867324 - nodes in this community are weakly interconnected._
- **Why does `DocumentBuilder` connect `DocumentBuilder` to `pdf20_test.go`, `Name`, `Color`, `New`, `testing.T`, `LineStyle`, `PageAccess`, `ITableBuilder`, `.EmbedStandardFonts`, `ImageBuilder`, `Support`, `PageBuilder`, `TextBoxObject`, `pdftest.go`, `FlowBuilder`, `Font`, `Dict`, `gPDF — architecture map`, `testing.B`, `writer/options.go`, `OpenReaderWithPassword`, `EmbeddableFont`, `.setErr`, `.ApplyPageSize`, `ListObject`, `bidi.go`, `EmbeddableFont`, `Still open`, `Support`, `time.Time`, `Handle`, `ICCProfile`, `PDFALevel`, `MetadataBuilder`, `OutlineBuilder`, `CollectionOptions`, `.pageHeight`, `FormBuilder`?**
  _High betweenness centrality (0.092) - this node is a cross-community bridge._
- **Should `Name` be split into smaller, more focused modules?**
  _Cohesion score 0.039644565960355434 - nodes in this community are weakly interconnected._
- **Why does `Page` connect `pdf20_test.go` to `Page`, `.pageHeight`, `DocumentBuilder`, `time.Time`, `LineStyle`, `Pt`, `ICCProfile`, `ListObject`, `TextBoxObject`, `decode.go`, `Still open`, `RectObject`?**
  _High betweenness centrality (0.077) - this node is a cross-community bridge._