# Graph Report - gpdf  (2026-11-01)

## Corpus Check
- cluster-only mode — file stats not available

## Summary
- 3048 nodes · 17097 edges · 141 communities (87 shown, 54 thin omitted)
- Extraction: 96% EXTRACTED · 4% INFERRED · 0% AMBIGUOUS · INFERRED: 760 edges (avg confidence: 0.87)
- Token cost: 0 input · 0 output

## Graph Freshness
- Built from commit: `dff41a78`
- Run `git rev-parse HEAD` and compare to check if the graph is stale.
- Run `graphify update .` after code changes (no API cost).

## Community Hubs (Navigation)
- Page
- pdf20_test.go
- syntax/impl/tokenizer.go
- Name
- DocumentBuilder
- Op
- Object
- Color
- New
- Ref
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
- gPDF — architecture map
- Font
- pdfDocument
- Drawer
- steps
- ReplacesContent
- validation.go
- builtDocument
- with
- cms.go
- .writeEncrypted
- io.ReaderAt
- contentTokenizer
- parser.go
- NewStandardDecryptor
- Document
- buildEmbeddedFontObjects
- structWalker
- testing.B
- writer/options.go
- NewPDFReader
- OpenReaderWithPassword
- EmbeddableFont
- .setErr
- steps
- .ApplyPageSize
- Registry
- version_detector.go
- ListObject
- bidi.go
- Decode
- Font
- scanner.go
- Sign
- Verify
- decode.go
- steps
- applyFilterChain
- Features
- shape_test.go
- ParserImpl
- PDFWriter
- Support
- packer
- TaggingAccess
- imageRun
- arabic.go
- EncodeBytes
- SearchPages
- StreamWriter
- runlength.go
- time.Time
- Writer
- mockFont
- Usage
- buildSignatureUpdate
- Handle
- PageText
- Page
- appearanceStream
- TextStyle
- .Subset
- TextBlock
- ReaderOptions
- saslprep.go
- project.yml
- CatalogReader
- PDFALevel
- BuildAES256EncryptDictForWriteWithRandom
- .hintTables
- TableImpl
- RectObject
- ContentExtractor
- Table
- CollectionOptions
- ContentReader
- pageConfig
- objHeaderBefore
- CellStyle
- TestSaveLinearizedWritesUsableHintTables
- Encode
- ImageExtractor
- bitReader
- multiFlag
- Tokenizer
- CLAUDE.md
- suggested_commands.md
- task_completion_checklist.md
- github.com/gsoultan/gpdf

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
- `gPDF — architecture map` --references--> `DocumentBuilder`  [INFERRED]
  .serena/memories/architecture.md → doc/builder.go
- `gPDF — architecture map` --references--> `textOpsWriter`  [INFERRED]
  .serena/memories/architecture.md → doc/builder_content.go
- `Medium` --references--> `jpegHasAdobeMarker()`  [INFERRED]
  .serena/memories/review_findings.md → doc/builder_content.go
- `gPDF — architecture map` --references--> `shapedTextArg()`  [INFERRED]
  .serena/memories/architecture.md → doc/builder_content.go
- `gPDF — architecture map` --references--> `textStateTracker`  [INFERRED]
  .serena/memories/architecture.md → doc/builder_content.go

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

## Communities (141 total, 54 thin omitted)

### Community 0 - "Page"
Cohesion: 0.18
Nodes (37): emptyAppearance(), formXObject(), frameOps(), symbolAppearance(), AcroForm, Action, Annotation, Boolean (+29 more)

### Community 1 - "pdf20_test.go"
Cohesion: 0.23
Nodes (20): cleanup(), cleanup(), Document, embeddedFontObjects, graphicRun, LayoutReader, orderedRun, Page (+12 more)

### Community 2 - "syntax/impl/tokenizer.go"
Cohesion: 0.23
Nodes (9): tokenizer, hexByte(), hexVal(), isDelimiter(), isHex(), isWhitespace(), numberToken(), Token (+1 more)

### Community 3 - "Name"
Cohesion: 0.03
Nodes (39): asciiFileName(), md5sum(), collectionItem(), buttonAppearance(), fontResources(), num(), pdfLiteral(), textFieldAppearance() (+31 more)

### Community 4 - "DocumentBuilder"
Cohesion: 0.03
Nodes (19): annotBuilder, DocumentBuilder, documentPart, Support, FormBuilder, ImageBuilder, MetadataBuilder, OutlineBuilder (+11 more)

### Community 5 - "Op"
Cohesion: 0.06
Nodes (78): Op, Parser, balancedMarkedContent(), ColorRGB, contentSource, abs64(), appendBoundary(), appendTextArg() (+70 more)

### Community 6 - "Object"
Cohesion: 0.03
Nodes (30): isPageTreeNode(), ObjectEditor, patchDoc, checkPDFAFont(), resolveBuilt(), walkBuilt(), checkHeadingLevels(), clip() (+22 more)

### Community 7 - "Color"
Cohesion: 0.06
Nodes (13): circlePathOps(), ellipsePathOps(), polygonPathOps(), roundedRectPathOps(), fillColorOps(), strokeStateOps(), PathBuilder, ColorFromHex() (+5 more)

### Community 8 - "New"
Cohesion: 0.07
Nodes (72): main(), main(), main(), createLogo(), main(), New(), DefaultStandardSubstitutes(), TestDrawImage() (+64 more)

### Community 9 - "Ref"
Cohesion: 0.05
Nodes (10): resolver, Ref, cryptFilterName(), documentCore, ObjectResolver, TestExtractFromObjectStreamChecksObjectNumber(), CryptFilterDecryptor, cryptMethod (+2 more)

### Community 10 - "testing.T"
Cohesion: 0.06
Nodes (65): TestStreamParser_ImplementsInterface(), NewRichText(), DefaultTextStyle(), TestTable_CellLookup(), baseEncoding(), TestEncodingTablesComplete(), sample(), TestDecodeEarlyChangeZeroMatchesStdlib() (+57 more)

### Community 11 - "LineStyle"
Cohesion: 0.08
Nodes (14): addGraphicRunWithState(), circlePathOps(), ellipsePathOps(), fillColorOps(), NewDrawer(), polygonPathOps(), roundedRectPathOps(), strokeStateOps() (+6 more)

### Community 12 - "PageAccess"
Cohesion: 0.09
Nodes (11): PageAccess, alignLine(), countWordSpaces(), defaultFont(), fontWidthFunc(), measureText(), NewDrawer(), wrapLines() (+3 more)

### Community 13 - "ITableBuilder"
Cohesion: 0.07
Nodes (11): cellLayout, ITableBuilder, ImageSide, ImageWrap, computeImageDimensions(), CellImageSpec, TableBuilder, TableBuilderAPI (+3 more)

### Community 14 - "io.Writer"
Cohesion: 0.07
Nodes (15): Filter, Filter, Filter, Filter, Filter, Filter, Filter, Filter (+7 more)

### Community 15 - "Integer"
Cohesion: 0.06
Nodes (31): TestParseStrayDelimiterTerminates(), NewStreamParser(), TestStreamParser_InlineImage(), TestStreamParser_LineContinuation(), TestStreamParser_OctalEscape(), TestStreamParser_OperandTypes(), TestStreamParser_Parse(), buildToUnicodeCMap() (+23 more)

### Community 16 - "go_pkg_io"
Cohesion: 0.21
Nodes (13): NewFilter(), NewFilter(), NewFilter(), NewFilter(), NewFilter(), RegisterStandardFilters(), Filter, NewFilter() (+5 more)

### Community 17 - "Open"
Cohesion: 0.07
Nodes (41): main(), createDummyPDF(), main(), openFromURL(), main(), main(), main(), main() (+33 more)

### Community 18 - ".layoutTextIntoPages"
Cohesion: 0.08
Nodes (14): countWordSpaces(), extraSpacing(), reverseRunes(), fontSegment, RichText, RichTextSegment, FontWidthFunc, LineRectFunc (+6 more)

### Community 19 - "TextBoxObject"
Cohesion: 0.05
Nodes (7): At(), HeadingObject, ImageObject, Pt, TextAlignment, TextBoxObject, TextObject

### Community 20 - "pdftest.go"
Cohesion: 0.09
Nodes (30): Flate(), Builder, New(), PNGUp(), obj, fuzzSeeds(), contentOf(), onePageWithContentObj() (+22 more)

### Community 21 - "FlowBuilder"
Cohesion: 0.10
Nodes (9): FloatingImage, FlowBuilder, FlowOptions, DefaultImageLayout(), DefaultImageStyle(), ImageLayout, ImageStyle, Align (+1 more)

### Community 22 - "gPDF — architecture map"
Cohesion: 0.09
Nodes (19): newImageOpsWriter(), imageOpsWriter, BuildXMP(), writePDFUAExtensionSchema(), pageSpec, PDFUALevel, SRGBICCProfile(), encodeTextStrings() (+11 more)

### Community 23 - "Font"
Cohesion: 0.08
Nodes (8): aliasedFont, newEmbeddedFontUsage(), getRuneBlock(), reverseGlyphs(), fontConfig, Font, Shaper, Run

### Community 24 - "pdfDocument"
Cohesion: 0.09
Nodes (6): String, mergeInherited(), outlineItemFirstRef(), MetadataProvider, pdfDocument, intFromModelObj()

### Community 25 - "Drawer"
Cohesion: 0.11
Nodes (6): defaultBPC(), defaultColorSpace(), lastPage(), NewDrawer(), Figure, Drawer

### Community 26 - "steps"
Cohesion: 0.06
Nodes (33): name, permissions, runs-on, steps, strategy, build-mode, language, jobs (+25 more)

### Community 27 - "ReplacesContent"
Cohesion: 0.07
Nodes (23): ExtractText(), ExtractTextPerPage(), TestExtractText_DecodesTextUsingToUnicodeCMap(), TestExtractText_HandlesQuoteTextOperators(), TestExtractText_InsertsBoundaryBetweenTextBlocks(), TestExtractText_InsertsSpaceForTJKerningAdjustments(), TestExtractText_PreservesLineBoundariesForTextPositionOperators(), TestExtractText_ReadsDirectPageContentStreams() (+15 more)

### Community 28 - "validation.go"
Cohesion: 0.11
Nodes (22): candidateRow, clusterAnchors(), DetectTables(), detectTablesOnPage(), extractTableFromGroup(), inferTableBorder(), nearestAnchor(), TestDetectTables_DetectsSimpleGrid() (+14 more)

### Community 29 - "builtDocument"
Cohesion: 0.09
Nodes (8): builtDocument, newObjectCloner(), ImageReader, objectCloner, copyInfo(), copyPages(), newPageTree(), ImageInfo

### Community 30 - "with"
Cohesion: 0.06
Nodes (30): runs-on, steps, strategy, default, description, fuzztime, jobs, fuzz (+22 more)

### Community 31 - "cms.go"
Cohesion: 0.16
Nodes (20): algorithmIdentifier, attribute, digestOID(), explicit0(), hashFromOID(), parseCertificateSet(), parseCMS(), pick() (+12 more)

### Community 32 - ".writeEncrypted"
Cohesion: 0.11
Nodes (18): HexString, BuildEncryptDictForWrite(), headerFor(), generationProvider, validateDocumentGraph(), validateDocumentGraphWithOptions(), validateIncrementalDocumentGraph(), validateObjectRefs() (+10 more)

### Community 33 - "io.ReaderAt"
Cohesion: 0.15
Nodes (14): PDFReader, decodeXRefStreamEntry(), findLastXRefOffset(), findStartXRef(), parseStartXRefOffset(), readBEField(), repairResult, readScanChunk() (+6 more)

### Community 34 - "contentTokenizer"
Cohesion: 0.21
Nodes (9): hexByte(), hexVal(), isHexChar(), isSpace(), newContentTokenizer(), contentTokenizer, ctoken, ctokenKind (+1 more)

### Community 35 - "parser.go"
Cohesion: 0.13
Nodes (9): parseFaces(), registerPaths(), loadFont(), exercise(), FuzzParse(), decodeUTF16BE(), Parse(), parseAtOffset() (+1 more)

### Community 36 - "NewStandardDecryptor"
Cohesion: 0.13
Nodes (24): Decryptor, ValidatePerms(), getBytes(), getIDBytes(), getInt(), authenticateLegacy(), checkUser(), computeFileKey() (+16 more)

### Community 37 - "Document"
Cohesion: 0.10
Nodes (14): dictInt(), floatFromObj(), main(), openPDF(), printDocumentInfo(), printMetadata(), printPages(), printSearch() (+6 more)

### Community 38 - "buildEmbeddedFontObjects"
Cohesion: 0.12
Nodes (17): shapedTextArg(), symbolEncode(), winAnsiEncode(), zapfDingbatsEncode(), buildCIDSet(), buildCIDWidthArray(), buildEmbeddedFontObjects(), fontBBoxArray() (+9 more)

### Community 39 - "structWalker"
Cohesion: 0.17
Nodes (10): ExtractStructuredText(), ExtractTextInReadingOrder(), TextElement, isCellRole(), JoinTextElements(), layoutOrderText(), nameOf(), structureTreeText() (+2 more)

### Community 40 - "testing.B"
Cohesion: 0.17
Nodes (22): BenchmarkParseObject(), BenchmarkTokenizer(), TestTokenizerCommentAdvancesOffset(), TestTokenizerStrayDelimitersMakeProgress(), NewTokenizer(), TestTokenizer_CurrentOffset_AfterTokens(), benchBuilder(), benchDocument() (+14 more)

### Community 41 - "writer/options.go"
Cohesion: 0.13
Nodes (18): TestApplyFiltersWithLimitsRejectsOversizedDecode(), applyFilters(), applyFiltersWithLimits(), decodeStreamData(), decodeStreamDataWithLimits(), NewStandardRegistry(), FilterRegistry, NewRegistry() (+10 more)

### Community 42 - "NewPDFReader"
Cohesion: 0.11
Nodes (23): TestStrictModeRejectsMissingStartXRef(), TestTolerantModeRecoversMissingStartXRef(), TestValidateDocumentReportsStructuralStatus(), TestValidateRejectsInvalidLevel(), NewPDFReader(), NewPDFReaderWithOptions(), minimalPDFBytes(), safeByte() (+15 more)

### Community 43 - "OpenReaderWithPassword"
Cohesion: 0.10
Nodes (20): sample, OpenReaderWithPassword(), DecodeTextString(), TestDecodePDFDocEncoding(), basic(), corpus(), corpusDir(), reopen() (+12 more)

### Community 44 - "EmbeddableFont"
Cohesion: 0.13
Nodes (10): embeddedProgramName(), EmbeddableFont, DefaultFallback(), fileScore(), NewFallback(), GetDefaultScanner(), StandardSubstitutes(), Fallback (+2 more)

### Community 45 - ".setErr"
Cohesion: 0.16
Nodes (4): decodeImageFile(), jpegColorComponents(), ImageCompression, ImageOptions

### Community 46 - "steps"
Cohesion: 0.09
Nodes (21): env, GPDF_REQUIRE_TOOLS, VERAPDF_SHA256, VERAPDF_VERSION, jobs, test, name, on (+13 more)

### Community 48 - "Registry"
Cohesion: 0.14
Nodes (6): memoryRegistry, Registry, scanDir(), LinuxScanner, MacOSScanner, WindowsScanner

### Community 49 - "version_detector.go"
Cohesion: 0.14
Nodes (7): CatalogProvider, PDFVersion, defaultVersion(), detectPDFVersion(), digitAt(), parseVersionHeader(), parseVersionName()

### Community 50 - "ListObject"
Cohesion: 0.12
Nodes (5): toRoman(), ListObject, DefaultListStyle(), ListMarker, ListStyle

### Community 51 - "bidi.go"
Cohesion: 0.19
Nodes (16): TextDirection, BaseDirection(), classify(), Direction, isArabicLetterScript(), IsRTL(), Levels(), Mirror() (+8 more)

### Community 52 - "Decode"
Cohesion: 0.19
Nodes (15): Encode(), encodePNG(), encodeTIFF(), pngFilter(), sign(), Decode(), decodePNG(), decodeTIFF() (+7 more)

### Community 53 - "Font"
Cohesion: 0.13
Nodes (3): Font, GSUB, tableRecord

### Community 54 - "scanner.go"
Cohesion: 0.12
Nodes (10): CFFFont, GlyphEncoder, Metrics, allScriptFeatures(), featureLookups(), parseCoverageTable(), parseGSUBType1(), scriptFeatures() (+2 more)

### Community 55 - "Sign"
Cohesion: 0.31
Nodes (6): subjectName(), Sign(), SignFile(), Signer, SignOptions, SubFilter

### Community 56 - "Verify"
Cohesion: 0.19
Nodes (15): NewCMSSigner(), popplerSig(), samplePDF(), testCert(), TestSecondSignatureKeepsFirstValid(), TestSignAndVerify(), TestTimestampAndValidationData(), SignatureInfo (+7 more)

### Community 57 - "decode.go"
Cohesion: 0.20
Nodes (12): DecodePNG(), DecodePNGToRaw(), DetectDimensions(), HasAlpha(), isGrayImage(), nrgbaAt(), ProcessImage(), Decoded (+4 more)

### Community 58 - "steps"
Cohesion: 0.11
Nodes (17): runs-on, steps, jobs, bench, name, on, pull_request, workflow_dispatch (+9 more)

### Community 59 - "applyFilterChain"
Cohesion: 0.13
Nodes (14): codeRange, addBFRange(), bytesToInt(), decodeCMapString(), decodeRawTextBytes(), decodeUTF16BE(), decodeWinAnsi(), intToBytes() (+6 more)

### Community 60 - "Features"
Cohesion: 0.15
Nodes (9): Redact(), RedactArea, RedactOptions, RedactionReport, Features, Critical / high, gPDF — deep review findings (2026-10-31, commit 036d17f; fixes 2026-10-31 on branch claude/project-thread-tv91nd), Repo / process (+1 more)

### Community 61 - "shape_test.go"
Cohesion: 0.19
Nodes (13): cffIndex(), cffIndexCount(), cffString(), parseCFFCharset(), parseCFFDict(), loadFont(), TestCFFDictAndCharset(), TestParseCFF() (+5 more)

### Community 63 - "PDFWriter"
Cohesion: 0.29
Nodes (5): Encryptor, errWriter, asErrWriter(), escapeName(), PDFWriter

### Community 64 - "Support"
Cohesion: 0.17
Nodes (9): StructRow, ElemKind, ElemRef, List, ListItem, Section, StructCell, StructTable (+1 more)

### Community 65 - "packer"
Cohesion: 0.29
Nodes (5): objectGroup, packable(), offsetWriter, packer, xrefEntry

### Community 67 - "imageRun"
Cohesion: 0.19
Nodes (8): jpegHasAdobeMarker(), components(), decodeJPEGToRaw(), downsampleRun(), encodeJPEG(), placedSize(), resampleBox(), imageRun

### Community 68 - "arabic.go"
Cohesion: 0.25
Nodes (10): ArabicForms(), IsArabicLetter(), JoiningTypeOf(), LamAlefLigature(), PresentationForm(), TestJoiningTypes(), TestPresentationForms(), nextBase() (+2 more)

### Community 69 - "EncodeBytes"
Cohesion: 0.23
Nodes (8): EncodeBytes(), escapeContentName(), NewEncoder(), sortedKeys(), TestEncodeBytes(), writeHexString(), writeLiteralString(), Encoder

### Community 70 - "SearchPages"
Cohesion: 0.18
Nodes (7): ContentSearcher, SearchResult, SearchPages(), TestSearchPages_EmptyKeyword(), TestSearchPages_FindsKeywordOnCorrectPage(), TestSearchPages_MultipleKeywords(), TestSearchPages_RecordsIndices()

### Community 72 - "runlength.go"
Cohesion: 0.31
Nodes (9): Filter, canEncodeRepeatRun(), copyLiteralRun(), copyRepeatedRun(), decode(), decodeLiteral(), encode(), writeLiteralRun() (+1 more)

### Community 73 - "time.Time"
Cohesion: 0.24
Nodes (7): Attachment, FileRelationship, PDFDate(), FormatDate(), ParseDate(), FuzzParseDate(), TestParseDate()

### Community 76 - "Usage"
Cohesion: 0.17
Nodes (11): Extracting Pages, gPDF, Merging from Remote Sources (URL/S3), Merging PDFs, Reading a PDF, Reading order, updating in place, redacting, Searching and Replacing, Splitting PDFs (+3 more)

### Community 77 - "buildSignatureUpdate"
Cohesion: 0.27
Nodes (8): AddValidationData(), buildSignatureUpdate(), copyDict(), docMDPLevel(), resolveArray(), resolveDict(), update, ValidationData

### Community 78 - "Handle"
Cohesion: 0.22
Nodes (4): Handle, Support, LayerBuilder, Spec

### Community 79 - "PageText"
Cohesion: 0.33
Nodes (3): PageDecorator, PageText, TextAlign

### Community 80 - "Page"
Cohesion: 0.39
Nodes (4): GraphicRun, ImageRun, TextRun, Page

### Community 81 - "appearanceStream"
Cohesion: 0.22
Nodes (7): rectArray(), GeoViewport, Model3D, FormatReal(), appearanceStream(), escapeLiteral(), FormatReal()

### Community 84 - "TextBlock"
Cohesion: 0.33
Nodes (7): InferImageLayouts(), medianHeight(), OrderBlocks(), regionParagraphs(), splitAtGap(), xyCut(), TextBlock

### Community 85 - "ReaderOptions"
Cohesion: 0.31
Nodes (8): DefaultReaderOptions(), normalizeReaderOptions(), ParseLimits, ParserMode, NewPDFReaderWithFilters(), NewPDFReaderWithFiltersAndOptions(), ReaderOptions, gPDF — code style and conventions

### Community 86 - "saslprep.go"
Cohesion: 0.44
Nodes (8): checkBidi(), isLCat(), isMappedToNothing(), isNonASCIISpace(), isProhibited(), isRandAL(), saslprep(), TestSASLprep()

### Community 87 - "project.yml"
Cohesion: 0.22
Nodes (8): encoding, excluded_tools, ignore_all_files_in_gitignore, ignored_paths, initial_prompt, languages, project_name, read_only

### Community 90 - "BuildAES256EncryptDictForWriteWithRandom"
Cohesion: 0.32
Nodes (8): aes256Password(), aesNoPadIV(), authenticateAES256(), hashR6(), aesNoPad(), TestHashR6Deterministic(), BuildAES256EncryptDictForWrite(), BuildAES256EncryptDictForWriteWithRandom()

### Community 91 - ".hintTables"
Cohesion: 0.36
Nodes (4): contentInfo, bitWriter, hintTableSet, bitsFor()

### Community 93 - "RectObject"
Cohesion: 0.33
Nodes (3): R(), Rect, RectObject

### Community 96 - "CollectionOptions"
Cohesion: 0.40
Nodes (4): CollectionField, CollectionFieldType, CollectionOptions, CollectionView

### Community 99 - "objHeaderBefore"
Cohesion: 0.60
Nodes (5): isPDFWhitespace(), isDigit(), isRegular(), objHeaderBefore(), scanChunk()

### Community 100 - "CellStyle"
Cohesion: 0.40
Nodes (3): CellHorizontalAlign, CellStyle, CellVerticalAlign

### Community 101 - "TestSaveLinearizedWritesUsableHintTables"
Cohesion: 0.50
Nodes (3): linValue(), TestSaveLinearizedWritesUsableHintTables(), bitReader

## Knowledge Gaps
- **1 isolated node(s):** `github.com/gsoultan/gpdf`
  These have ≤1 connection - possible missing edges or undocumented components. (Counts symbols only; 442 node(s) total have ≤1 connection when file, concept and rationale nodes are included.)
- **54 thin communities (<3 nodes) omitted from report** — run `graphify query` to explore isolated nodes.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **Why does `Name` connect `Name` to `Page`, `Op`, `Object`, `Color`, `New`, `Ref`, `testing.T`, `LineStyle`, `ITableBuilder`, `io.Writer`, `Integer`, `Open`, `.layoutTextIntoPages`, `TextBoxObject`, `pdftest.go`, `FlowBuilder`, `gPDF — architecture map`, `pdfDocument`, `ReplacesContent`, `validation.go`, `builtDocument`, `.writeEncrypted`, `io.ReaderAt`, `contentTokenizer`, `NewStandardDecryptor`, `Document`, `buildEmbeddedFontObjects`, `structWalker`, `writer/options.go`, `NewPDFReader`, `OpenReaderWithPassword`, `Sign`, `Verify`, `decode.go`, `applyFilterChain`, `ParserImpl`, `PDFWriter`, `Support`, `packer`, `TaggingAccess`, `imageRun`, `EncodeBytes`, `StreamWriter`, `buildSignatureUpdate`, `Handle`, `Page`, `appearanceStream`, `BuildAES256EncryptDictForWriteWithRandom`, `.hintTables`, `pageConfig`, `TestSaveLinearizedWritesUsableHintTables`?**
  _High betweenness centrality (0.092) - this node is a cross-community bridge._
- **Are the 44 inferred relationships involving `Name` (e.g. with `.Fields()` and `.SigFlags()`) actually correct?**
  _`Name` has 44 INFERRED edges - model-reasoned connections that need verification._
- **What connects `github.com/gsoultan/gpdf` to the rest of the system?**
  _1 weakly-connected nodes found - possible documentation gaps or missing edges._
- **Should `Name` be split into smaller, more focused modules?**
  _Cohesion score 0.031692406692406694 - nodes in this community are weakly interconnected._
- **Why does `DocumentBuilder` connect `DocumentBuilder` to `pdf20_test.go`, `Name`, `Color`, `New`, `testing.T`, `LineStyle`, `PageAccess`, `ITableBuilder`, `TextBoxObject`, `pdftest.go`, `FlowBuilder`, `gPDF — architecture map`, `Font`, `Drawer`, `testing.B`, `OpenReaderWithPassword`, `EmbeddableFont`, `.setErr`, `.ApplyPageSize`, `bidi.go`, `Features`, `Support`, `StreamWriter`, `time.Time`, `Handle`, `PageText`, `PDFALevel`, `CollectionOptions`, `pageConfig`?**
  _High betweenness centrality (0.092) - this node is a cross-community bridge._
- **Should `DocumentBuilder` be split into smaller, more focused modules?**
  _Cohesion score 0.028035456606885178 - nodes in this community are weakly interconnected._
- **Why does `Page` connect `pdf20_test.go` to `Page`, `DocumentBuilder`, `Object`, `time.Time`, `Ref`, `LineStyle`, `appearanceStream`, `ListObject`, `TextBoxObject`, `FlowBuilder`, `decode.go`, `RectObject`?**
  _High betweenness centrality (0.079) - this node is a cross-community bridge._