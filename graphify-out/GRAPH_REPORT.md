# Graph Report - gpdf  (2026-10-31)

## Corpus Check
- 252 files · ~99,460 words
- Verdict: corpus is large enough that graph structure adds value.
- Unclassified: 2 file(s) not represented in the graph (top: (none) 2)

## Summary
- 2093 nodes · 9405 edges · 110 communities (58 shown, 52 thin omitted)
- Extraction: 97% EXTRACTED · 3% INFERRED · 0% AMBIGUOUS · INFERRED: 309 edges (avg confidence: 0.85)
- Token cost: 0 input · 0 output

## Graph Freshness
- Built from commit: `036d17fb`
- Run `git rev-parse HEAD` and compare to check if the graph is stale.
- Run `graphify update .` after code changes (no API cost).

## Community Hubs (Navigation)
- Ref
- read_document_test.go
- io.ReaderAt
- Name
- Color
- contentSource
- LineStyle
- PageAccess
- Dict
- DocumentBuilder
- testing.T
- TextBoxObject
- Drawer
- extractBlocksFromOps
- steps
- TableBuilder
- go_pkg_io
- io.Reader
- Integer
- FlowBuilder
- io.Writer
- pdfDocument
- layout.go
- FilterRegistry
- Font
- New
- contentTokenizer
- .layoutTextIntoPages
- ITableBuilder
- ExtractText
- builtDocument
- NewPDFReader
- Open
- ListObject
- ExtractLayout
- ExtractTextPerPage
- Document
- fontConfig
- .ImageWithLayout
- TextStyle
- validation.go
- Font
- extract_tables.go
- TaggingAccess
- standardR4Decryptor
- Document
- Support
- Registry
- runlength.go
- .WriteLinearized
- mockFont
- EncodeBytes
- scanner.go
- Signer
- version_detector.go
- Usage
- Handle
- Table
- .Scan
- MetadataBuilder
- OpenReaderWithPassword
- OutlineBuilder
- parser.go
- FormBuilder
- TextBuilder
- CatalogProvider
- StandardDecryptor
- Writer
- RectObject
- pagesize.go
- ContentExtractor
- Table
- CellStyle
- ICCProfile
- go_pkg_sort
- ImageExtractor
- LayoutExtractor
- StructureProvider
- validateDocumentGraph
- ImageOptions
- Support
- FormProvider
- multiFlag
- github.com/gsoultan/gpdf

## God Nodes (most connected - your core abstractions)
1. `Name` - 215 edges
2. `Ref` - 176 edges
3. `Object` - 165 edges
4. `PageAccess` - 129 edges
5. `Page` - 117 edges
6. `DocumentBuilder` - 114 edges
7. `Stream` - 103 edges
8. `Trailer` - 103 edges
9. `Catalog` - 99 edges
10. `Outlines` - 91 edges

## Surprising Connections (you probably didn't know these)
- `main()` --calls--> `New()`  [EXTRACTED]
  cmd/cv/main.go → doc/doc.go
- `main()` --calls--> `New()`  [EXTRACTED]
  cmd/example/main.go → doc/doc.go
- `main()` --calls--> `Open()`  [EXTRACTED]
  cmd/example_merge/main.go → doc/doc.go
- `main()` --calls--> `Merge()`  [EXTRACTED]
  cmd/example_merge_remote/main.go → doc/merge.go
- `openFromURL()` --references--> `Document`  [EXTRACTED]
  cmd/example_merge_remote/main.go → doc/document.go

## Import Cycles
- 2-file cycle: `doc/example_split_test.go -> doc/merge_test.go -> doc/example_split_test.go`
- 2-file cycle: `doc/example_split_test.go -> doc/split_test.go -> doc/example_split_test.go`
- 2-file cycle: `doc/merge_test.go -> doc/split_test.go -> doc/merge_test.go`
- 3-file cycle: `doc/example_split_test.go -> doc/merge_test.go -> doc/split_test.go -> doc/example_split_test.go`
- 3-file cycle: `doc/example_split_test.go -> doc/split_test.go -> doc/merge_test.go -> doc/example_split_test.go`

## Communities (110 total, 52 thin omitted)

### Community 0 - "Ref"
Cohesion: 0.10
Nodes (43): floatFromObj(), Op, graphicRun, patchDoc, AcroForm, Action, Annotation, Boolean (+35 more)

### Community 1 - "read_document_test.go"
Cohesion: 0.18
Nodes (18): cleanup(), main(), cleanup(), main(), pdfFontFlags(), ContentReader, Document, embeddedFontObjects (+10 more)

### Community 2 - "io.ReaderAt"
Cohesion: 0.05
Nodes (40): ParserImpl, tokenizer, detectLinearization(), isLinearized(), DefaultReaderOptions(), normalizeReaderOptions(), ParseLimits, ParserMode (+32 more)

### Community 3 - "Name"
Cohesion: 0.04
Nodes (13): attachWidgetToPage(), Support, OutlineItem, StructElem, StructTreeRoot, Array, Name, Entry (+5 more)

### Community 4 - "Color"
Cohesion: 0.08
Nodes (13): circlePathOps(), ellipsePathOps(), polygonPathOps(), roundedRectPathOps(), fillColorOps(), strokeStateOps(), PathBuilder, ColorFromHex() (+5 more)

### Community 5 - "contentSource"
Cohesion: 0.08
Nodes (47): Parser, ColorRGB, contentSource, abs64(), appendBoundary(), appendTextArg(), extractTextFromOps(), extractTextFromXObject() (+39 more)

### Community 6 - "LineStyle"
Cohesion: 0.07
Nodes (11): addGraphicRunWithState(), circlePathOps(), ellipsePathOps(), fillColorOps(), polygonPathOps(), roundedRectPathOps(), strokeStateOps(), LineStyle (+3 more)

### Community 7 - "PageAccess"
Cohesion: 0.09
Nodes (10): PageAccess, alignLine(), countWordSpaces(), defaultFont(), fontWidthFunc(), measureText(), NewDrawer(), wrapLines() (+2 more)

### Community 8 - "Dict"
Cohesion: 0.06
Nodes (32): GraphicRun, ImageRun, TextRun, Page, CatalogReader, Support, Dict, buildLinearizationInfo() (+24 more)

### Community 9 - "DocumentBuilder"
Cohesion: 0.05
Nodes (5): DocumentBuilder, ImageBuilder, PageBuilder, TaggedContentBuilder, TaggedImageBuilder

### Community 10 - "testing.T"
Cohesion: 0.11
Nodes (39): NewStreamParser(), TestStreamParser_ImplementsInterface(), TestStreamParser_InlineImage(), TestStreamParser_LineContinuation(), TestStreamParser_OctalEscape(), TestStreamParser_OperandTypes(), TestStreamParser_Parse(), TestTable_CellLookup() (+31 more)

### Community 11 - "TextBoxObject"
Cohesion: 0.06
Nodes (7): At(), HeadingObject, LineObject, Pt, TextAlignment, TextBoxObject, TextObject

### Community 12 - "Drawer"
Cohesion: 0.10
Nodes (8): defaultBPC(), defaultColorSpace(), lastPage(), NewDrawer(), DecodePNGToRaw(), HasAlpha(), Figure, Drawer

### Community 13 - "extractBlocksFromOps"
Cohesion: 0.12
Nodes (31): bytesToInt(), decodeCMapString(), decodeHexString(), decodeRawTextBytes(), decodeTextBytes(), decodeUTF16BE(), decodeWinAnsi(), intToBytes() (+23 more)

### Community 14 - "steps"
Cohesion: 0.06
Nodes (33): name, permissions, runs-on, steps, strategy, build-mode, language, jobs (+25 more)

### Community 15 - "TableBuilder"
Cohesion: 0.12
Nodes (9): cellLayout, ImageSide, ImageWrap, computeImageDimensions(), CellImageSpec, TableBuilder, TableCellImageSpec, TableCellSpec (+1 more)

### Community 16 - "go_pkg_io"
Cohesion: 0.19
Nodes (13): NewFilter(), NewFilter(), NewFilter(), NewFilter(), NewFilter(), RegisterStandardFilters(), Filter, NewFilter() (+5 more)

### Community 17 - "io.Reader"
Cohesion: 0.09
Nodes (9): Filter, Filter, Filter, Filter, Filter, Filter, Filter, Filter (+1 more)

### Community 18 - "Integer"
Cohesion: 0.09
Nodes (16): buildToUnicodeCMap(), buildToUnicodeSymbolCMap(), buildToUnicodeWinAnsiCMap(), buildToUnicodeZapfDingbatsCMap(), copyPageDict(), buildCIDWidthArray(), buildEmbeddedFontObjects(), fontBBoxArray() (+8 more)

### Community 19 - "FlowBuilder"
Cohesion: 0.15
Nodes (4): FloatingImage, FlowBuilder, FlowOptions, Align

### Community 20 - "io.Writer"
Cohesion: 0.21
Nodes (4): BuildAES256EncryptDictForWrite(), Encryptor, PDFWriter, escapeName()

### Community 21 - "pdfDocument"
Cohesion: 0.10
Nodes (5): String, mergeInherited(), outlineItemFirstRef(), MetadataProvider, pdfDocument

### Community 22 - "layout.go"
Cohesion: 0.11
Nodes (9): countWordSpaces(), ApproxWidth(), FontWidthFunc, LineRectFunc, LineWidthFunc, WrapLines(), WrapLinesDynamic(), WrapLinesRect() (+1 more)

### Community 23 - "FilterRegistry"
Cohesion: 0.10
Nodes (18): boundedWriter, TestApplyFiltersWithLimitsRejectsOversizedDecode(), applyFilters(), applyFiltersWithLimits(), decodeStreamData(), decodeStreamDataWithLimits(), newBoundedWriter(), NewStandardRegistry() (+10 more)

### Community 25 - "New"
Cohesion: 0.07
Nodes (27): main(), main(), main(), createLogo(), main(), NewDrawer(), New(), R() (+19 more)

### Community 26 - "contentTokenizer"
Cohesion: 0.22
Nodes (9): hexByte(), hexVal(), isHexChar(), isSpace(), newContentTokenizer(), contentTokenizer, ctoken, ctokenKind (+1 more)

### Community 27 - ".layoutTextIntoPages"
Cohesion: 0.12
Nodes (9): symbolEncode(), winAnsiEncode(), zapfDingbatsEncode(), reverseRunes(), toFloat(), pageConfig, pageSpec, TextLayoutOptions (+1 more)

### Community 29 - "ExtractText"
Cohesion: 0.10
Nodes (16): ExtractText(), TestExtractText_DecodesTextUsingToUnicodeCMap(), TestExtractText_HandlesQuoteTextOperators(), TestExtractText_InsertsBoundaryBetweenTextBlocks(), TestExtractText_InsertsSpaceForTJKerningAdjustments(), TestExtractText_ReadsDirectPageContentStreams(), TestExtractText_TraversesFormXObjects(), makeReplaceStub() (+8 more)

### Community 30 - "builtDocument"
Cohesion: 0.12
Nodes (5): builtDocument, newObjectCloner(), ImageReader, objectCloner, ImageInfo

### Community 31 - "NewPDFReader"
Cohesion: 0.12
Nodes (21): TestStrictModeRejectsMissingStartXRef(), TestTolerantModeRecoversMissingStartXRef(), TestValidateDocumentReportsStructuralStatus(), TestValidateRejectsInvalidLevel(), NewPDFReader(), NewPDFReaderWithOptions(), minimalPDFBytes(), safeByte() (+13 more)

### Community 32 - "Open"
Cohesion: 0.15
Nodes (18): main(), main(), main(), Open(), ExampleExtract(), ExampleSplit(), ExampleSplitEvery(), Extract() (+10 more)

### Community 33 - "ListObject"
Cohesion: 0.12
Nodes (5): toRoman(), ListObject, DefaultListStyle(), ListMarker, ListStyle

### Community 34 - "ExtractLayout"
Cohesion: 0.14
Nodes (16): ExtractImages(), ExtractLayout(), makePageWithContent(), TestExtractImages_ReturnsImageMetadata(), TestExtractImages_SkipsFormXObjects(), TestExtractImages_TracksPlacementFromCTM(), TestExtractImagesPerPage_GroupsByPage(), TestExtractLayout_CapturesPositionAndStyle() (+8 more)

### Community 35 - "ExtractTextPerPage"
Cohesion: 0.12
Nodes (11): ContentSearcher, SearchResult, ExtractTextPerPage(), SearchPages(), TestExtractText_PreservesLineBoundariesForTextPositionOperators(), TestExtractTextPerPage_EmptyPageYieldsEmptyString(), TestExtractTextPerPage_ReturnsOneEntryPerPage(), TestSearchPages_EmptyKeyword() (+3 more)

### Community 37 - "fontConfig"
Cohesion: 0.14
Nodes (6): newEmbeddedFontUsage(), embeddedFontUsage, getRuneBlock(), fontConfig, fontSegment, EmbeddableFont

### Community 38 - ".ImageWithLayout"
Cohesion: 0.12
Nodes (9): DetectDimensions(), ProcessImage(), ImageObject, DefaultImageLayout(), DefaultImageStyle(), ImageLayout, ImageStyle, ImageInfo (+1 more)

### Community 39 - "TextStyle"
Cohesion: 0.12
Nodes (6): NewRichText(), RichText, RichTextSegment, DefaultTextStyle(), TextStyle, TestRichText()

### Community 40 - "validation.go"
Cohesion: 0.26
Nodes (11): appendValidationError(), appendValidationWarn(), isValidationLevel(), validationError(), validationErrorOrNil(), validationWarn(), ValidationDiagnostic, ValidationLevel (+3 more)

### Community 42 - "extract_tables.go"
Cohesion: 0.20
Nodes (11): candidateRow, clusterAnchors(), DetectTables(), detectTablesOnPage(), extractTableFromGroup(), inferTableBorder(), nearestAnchor(), TestDetectTables_DetectsSimpleGrid() (+3 more)

### Community 46 - "standardR4Decryptor"
Cohesion: 0.22
Nodes (5): pkcs7Unpad(), aes256Decryptor, r4Method, resolveCFMethod(), standardR4Decryptor

### Community 47 - "Document"
Cohesion: 0.20
Nodes (11): dictInt(), main(), openPDF(), printDocumentInfo(), printMetadata(), printPages(), printSearch(), printTextPerPage() (+3 more)

### Community 48 - "Support"
Cohesion: 0.15
Nodes (9): inSection(), StructRow, Block, List, ListItem, Section, StructCell, StructTable (+1 more)

### Community 50 - "runlength.go"
Cohesion: 0.31
Nodes (9): Filter, canEncodeRepeatRun(), copyLiteralRun(), copyRepeatedRun(), decode(), decodeLiteral(), encode(), writeLiteralRun() (+1 more)

### Community 53 - "EncodeBytes"
Cohesion: 0.24
Nodes (7): EncodeBytes(), escapeContentName(), NewEncoder(), TestEncodeBytes(), writeHexString(), writeLiteralString(), Encoder

### Community 54 - "scanner.go"
Cohesion: 0.24
Nodes (4): GetDefaultScanner(), tableChecksum(), FontScanner, GSUB

### Community 55 - "Signer"
Cohesion: 0.31
Nodes (4): Sign(), SignFile(), Signer, SignOptions

### Community 56 - "version_detector.go"
Cohesion: 0.29
Nodes (5): PDFVersion, defaultVersion(), detectPDFVersion(), digitAt(), parseVersionHeader()

### Community 57 - "Usage"
Cohesion: 0.18
Nodes (10): Extracting Pages, Features, gPDF, Merging from Remote Sources (URL/S3), Merging PDFs, Reading a PDF, Searching and Replacing, Splitting PDFs (+2 more)

### Community 59 - "Handle"
Cohesion: 0.22
Nodes (4): Handle, Support, LayerBuilder, Spec

### Community 60 - "Table"
Cohesion: 0.20
Nodes (3): LayoutReader, Table, TableCell

### Community 61 - ".Scan"
Cohesion: 0.27
Nodes (4): scanDir(), LinuxScanner, MacOSScanner, WindowsScanner

### Community 63 - "OpenReaderWithPassword"
Cohesion: 0.36
Nodes (7): createDummyPDF(), main(), openFromURL(), OpenReader(), OpenReaderWithPassword(), OpenWithPassword(), TestOpenWithPassword_Empty()

### Community 65 - "parser.go"
Cohesion: 0.39
Nodes (6): registerPaths(), decodeUTF16BE(), Parse(), parseAtOffset(), ParseCollection(), tableRecord

### Community 72 - "pagesize.go"
Cohesion: 0.40
Nodes (3): Custom(), Size, TestPageSizeAPI()

### Community 75 - "CellStyle"
Cohesion: 0.40
Nodes (3): CellHorizontalAlign, CellStyle, CellVerticalAlign

### Community 77 - "ICCProfile"
Cohesion: 0.50
Nodes (3): NewICCProfile(), SRGBProfile(), ICCProfile

### Community 82 - "validateDocumentGraph"
Cohesion: 0.83
Nodes (3): validateDocumentGraph(), validateDocumentGraphWithOptions(), validateIncrementalDocumentGraph()

### Community 84 - "Support"
Cohesion: 0.67
Nodes (3): Support, FieldKind, FieldSpec

## Knowledge Gaps
- **1 isolated node(s):** `github.com/gsoultan/gpdf`
  These have ≤1 connection - possible missing edges or undocumented components. (Counts symbols only; 319 node(s) total have ≤1 connection when file, concept and rationale nodes are included.)
- **52 thin communities (<3 nodes) omitted from report** — run `graphify query` to explore isolated nodes.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **Why does `DocumentBuilder` connect `DocumentBuilder` to `read_document_test.go`, `Name`, `Color`, `LineStyle`, `PageAccess`, `Dict`, `testing.T`, `Drawer`, `TableBuilder`, `FlowBuilder`, `New`, `.layoutTextIntoPages`, `ITableBuilder`, `fontConfig`, `.ImageWithLayout`, `Font`, `.ApplyPageSize`, `.addImageRun`, `Support`, `Handle`, `MetadataBuilder`, `OutlineBuilder`, `FormBuilder`, `TextBuilder`, `.getEffectiveStyle`, `ICCProfile`, `Support`, `.CurrentPage`?**
  _High betweenness centrality (0.127) - this node is a cross-community bridge._
- **Are the 39 inferred relationships involving `Name` (e.g. with `.Fields()` and `.SigFlags()`) actually correct?**
  _`Name` has 39 INFERRED edges - model-reasoned connections that need verification._
- **What connects `github.com/gsoultan/gpdf` to the rest of the system?**
  _1 weakly-connected nodes found - possible documentation gaps or missing edges._
- **Should `Ref` be split into smaller, more focused modules?**
  _Cohesion score 0.10485324033711131 - nodes in this community are weakly interconnected._
- **Why does `Name` connect `Name` to `Ref`, `read_document_test.go`, `io.ReaderAt`, `Color`, `contentSource`, `LineStyle`, `Dict`, `testing.T`, `TextBoxObject`, `extractBlocksFromOps`, `TableBuilder`, `Integer`, `FlowBuilder`, `io.Writer`, `pdfDocument`, `FilterRegistry`, `New`, `contentTokenizer`, `.layoutTextIntoPages`, `ExtractText`, `builtDocument`, `NewPDFReader`, `Open`, `ExtractLayout`, `ExtractTextPerPage`, `validation.go`, `TaggingAccess`, `standardR4Decryptor`, `Document`, `Support`, `.WriteLinearized`, `EncodeBytes`, `Handle`?**
  _High betweenness centrality (0.090) - this node is a cross-community bridge._
- **Should `io.ReaderAt` be split into smaller, more focused modules?**
  _Cohesion score 0.050883898709985664 - nodes in this community are weakly interconnected._
- **Why does `PageAccess` connect `PageAccess` to `Ref`, `Color`, `LineStyle`, `Font`, `Drawer`?**
  _High betweenness centrality (0.066) - this node is a cross-community bridge._