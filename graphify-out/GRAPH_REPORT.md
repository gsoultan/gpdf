# Graph Report - gpdf  (2026-11-01)

## Corpus Check
- 342 files · ~179,966 words
- Verdict: corpus is large enough that graph structure adds value.
- Unclassified: 7 file(s) not represented in the graph (top: (none) 7)

## Summary
- 3049 nodes · 17105 edges · 150 communities (82 shown, 68 thin omitted)
- Extraction: 96% EXTRACTED · 4% INFERRED · 0% AMBIGUOUS · INFERRED: 766 edges (avg confidence: 0.87)
- Token cost: 0 input · 0 output

## Graph Freshness
- Built from commit: `8b9a1c7f`
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
- EmbeddableFont
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
- ITableBuilder
- ExtractLayout
- builtDocument
- with
- cms.go
- .writeEncrypted
- io.ReaderAt
- contentTokenizer
- parser.go
- NewStandardDecryptor
- LayoutExtractor
- buildEmbeddedFontObjects
- .Validate
- testing.B
- writer/options.go
- NewPDFReader
- corpus
- DefaultStandardSubstitutes
- .setErr
- steps
- .ApplyPageSize
- memoryRegistry
- version_detector.go
- ListObject
- bidi.go
- Decode
- Font
- Pt
- Sign
- Verify
- decode.go
- steps
- applyFilterChain
- RedactOptions
- shape_test.go
- buildEmbeddedFontObjects
- PDFWriter
- Support
- packer
- TaggingAccess
- imageRun
- arabic.go
- .BuildFields
- SearchPages
- StreamWriter
- runlength.go
- Attachment
- Writer
- mockFont
- Usage
- buildSignatureUpdate
- Handle
- PageText
- Page
- MetadataBuilder
- TextStyle
- .Subset
- OutlineBuilder
- ReaderOptions
- saslprep.go
- project.yml
- CatalogReader
- PDFALevel
- BuildAES256EncryptDictForWriteWithRandom
- planLinearization
- TableImpl
- RectObject
- RichText
- Table
- CollectionOptions
- ContentReader
- .pageHeight
- objHeaderBefore
- CellStyle
- TestSaveLinearizedWritesUsableHintTables
- Encode
- ImageExtractor
- TextBuilder
- multiFlag
- ICCProfile
- CLAUDE.md
- suggested_commands.md
- task_completion_checklist.md
- github.com/gsoultan/gpdf
- detectLinearization
- LineObject
- DecodeTextString
- .EmbedStandardFonts
- StructureProvider
- Support
- PageBuilder
- PaintOrder

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
- `gPDF — architecture map` --references--> `textOpsWriter`  [INFERRED]
  .serena/memories/architecture.md → doc/builder_content.go
- `Medium` --references--> `jpegHasAdobeMarker()`  [INFERRED]
  .serena/memories/review_findings.md → doc/builder_content.go
- `gPDF — architecture map` --references--> `shapedTextArg()`  [INFERRED]
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

## Communities (150 total, 68 thin omitted)

### Community 0 - "Page"
Cohesion: 0.12
Nodes (59): dictInt(), floatFromObj(), main(), openPDF(), printDocumentInfo(), printMetadata(), printPages(), printSearch() (+51 more)

### Community 1 - "pdf20_test.go"
Cohesion: 0.17
Nodes (29): cleanup(), cleanup(), countWordSpaces(), Document, embeddedFontObjects, graphicRun, orderedRun, Page (+21 more)

### Community 2 - "syntax/impl/tokenizer.go"
Cohesion: 0.12
Nodes (11): ParserImpl, tokenizer, bitReader, hexByte(), hexVal(), isDelimiter(), isHex(), isWhitespace() (+3 more)

### Community 3 - "Name"
Cohesion: 0.03
Nodes (29): Support, OutlineItem, StructElem, StructTreeRoot, Array, Name, Entry, LinkAnnotation (+21 more)

### Community 4 - "DocumentBuilder"
Cohesion: 0.04
Nodes (7): annotBuilder, DocumentBuilder, documentPart, FormBuilder, ImageBuilder, TaggedContentBuilder, TaggedImageBuilder

### Community 5 - "Op"
Cohesion: 0.06
Nodes (76): Parser, writeUTF16BE(), Dict, ColorRGB, contentSource, abs64(), appendBoundary(), appendTextArg() (+68 more)

### Community 6 - "Object"
Cohesion: 0.03
Nodes (31): isPageTreeNode(), Support, ObjectEditor, patchDoc, checkPDFAFont(), resolveBuilt(), walkBuilt(), checkHeadingLevels() (+23 more)

### Community 7 - "Color"
Cohesion: 0.08
Nodes (13): Op, circlePathOps(), ellipsePathOps(), polygonPathOps(), roundedRectPathOps(), fillColorOps(), strokeStateOps(), PathBuilder (+5 more)

### Community 8 - "New"
Cohesion: 0.07
Nodes (70): main(), main(), main(), createLogo(), main(), NewDrawer(), New(), TestDrawImage() (+62 more)

### Community 9 - "Ref"
Cohesion: 0.04
Nodes (13): resolver, Ref, cryptFilterName(), documentCore, objectModifier, ObjectResolver, TestExtractFromObjectStreamChecksObjectNumber(), markModified() (+5 more)

### Community 10 - "testing.T"
Cohesion: 0.06
Nodes (63): TestStreamParser_ImplementsInterface(), TestTable_CellLookup(), baseEncoding(), TestEncodingTablesComplete(), sample(), TestDecodeEarlyChangeZeroMatchesStdlib(), TestDecodeSpecExample(), TestRoundTripAcrossWidthsAndResets() (+55 more)

### Community 11 - "LineStyle"
Cohesion: 0.08
Nodes (10): addGraphicRunWithState(), circlePathOps(), ellipsePathOps(), fillColorOps(), polygonPathOps(), roundedRectPathOps(), strokeStateOps(), Color (+2 more)

### Community 12 - "PageAccess"
Cohesion: 0.09
Nodes (9): PageAccess, alignLine(), defaultFont(), fontWidthFunc(), measureText(), wrapLines(), ApproxWidth(), Drawer (+1 more)

### Community 13 - "ITableBuilder"
Cohesion: 0.13
Nodes (5): cellLayout, computeImageDimensions(), TableBuilder, TableCellImageSpec, TableCellSpec

### Community 14 - "io.Writer"
Cohesion: 0.08
Nodes (15): Filter, Filter, Filter, Filter, Filter, Filter, Filter, Filter (+7 more)

### Community 15 - "Integer"
Cohesion: 0.05
Nodes (33): TestParseStrayDelimiterTerminates(), NewStreamParser(), TestStreamParser_InlineImage(), TestStreamParser_LineContinuation(), TestStreamParser_OctalEscape(), TestStreamParser_OperandTypes(), TestStreamParser_Parse(), rectArray() (+25 more)

### Community 16 - "EmbeddableFont"
Cohesion: 0.12
Nodes (8): aliasedFont, newEmbeddedFontUsage(), getRuneBlock(), reverseGlyphs(), fontConfig, EmbeddableFont, Shaper, Run

### Community 17 - "Open"
Cohesion: 0.07
Nodes (42): main(), createDummyPDF(), main(), openFromURL(), main(), main(), main(), main() (+34 more)

### Community 18 - ".layoutTextIntoPages"
Cohesion: 0.10
Nodes (11): countWordSpaces(), extraSpacing(), reverseRunes(), FontWidthFunc, LineRectFunc, LineWidthFunc, WrapLines(), WrapLinesDynamic() (+3 more)

### Community 19 - "TextBoxObject"
Cohesion: 0.10
Nodes (3): TextAlignment, TextBoxObject, TextObject

### Community 20 - "pdftest.go"
Cohesion: 0.09
Nodes (33): Flate(), Builder, New(), PNGUp(), obj, FuzzRoundTrip(), roundTripSeeds(), BenchmarkReadDocument() (+25 more)

### Community 22 - "gPDF — architecture map"
Cohesion: 0.05
Nodes (26): balancedMarkedContent(), buildToUnicodeCMap(), buildToUnicodeSymbolCMap(), buildToUnicodeWinAnsiCMap(), buildToUnicodeZapfDingbatsCMap(), newImageOpsWriter(), copyPageDict(), imageOpsWriter (+18 more)

### Community 24 - "pdfDocument"
Cohesion: 0.06
Nodes (15): String, mergeInherited(), outlineItemFirstRef(), MetadataProvider, pdfDocument, intFromModelObj(), makeReplaceStub(), TestReplaceContent_NoopWhenOldEqualsNew() (+7 more)

### Community 25 - "Drawer"
Cohesion: 0.11
Nodes (6): defaultBPC(), defaultColorSpace(), lastPage(), NewDrawer(), Figure, Drawer

### Community 26 - "steps"
Cohesion: 0.06
Nodes (33): name, permissions, runs-on, steps, strategy, build-mode, language, jobs (+25 more)

### Community 28 - "ExtractLayout"
Cohesion: 0.08
Nodes (24): LayoutReader, candidateRow, ExtractLayout(), clusterAnchors(), DetectTables(), detectTablesOnPage(), extractTableFromGroup(), inferTableBorder() (+16 more)

### Community 29 - "builtDocument"
Cohesion: 0.07
Nodes (16): builtDocument, newObjectCloner(), ImageReader, objectCloner, copyInfo(), copyPages(), newPageTree(), ImageInfo (+8 more)

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
Cohesion: 0.17
Nodes (11): PDFReader, decodeXRefStreamEntry(), findLastXRefOffset(), findStartXRef(), parseStartXRefOffset(), readBEField(), repairResult, detectPDFVersion() (+3 more)

### Community 34 - "contentTokenizer"
Cohesion: 0.20
Nodes (9): hexByte(), hexVal(), isHexChar(), isSpace(), newContentTokenizer(), contentTokenizer, ctoken, ctokenKind (+1 more)

### Community 36 - "NewStandardDecryptor"
Cohesion: 0.12
Nodes (24): Decryptor, ValidatePerms(), getBytes(), getIDBytes(), getInt(), authenticateLegacy(), checkUser(), computeFileKey() (+16 more)

### Community 38 - "buildEmbeddedFontObjects"
Cohesion: 0.17
Nodes (10): shapedTextArg(), symbolEncode(), winAnsiEncode(), zapfDingbatsEncode(), embeddedFontUsage, fontSegment, textOpsWriter, textRun (+2 more)

### Community 39 - ".Validate"
Cohesion: 0.22
Nodes (11): appendValidationError(), appendValidationWarn(), isValidationLevel(), validationError(), validationErrorOrNil(), validationWarn(), ValidationDiagnostic, ValidationLevel (+3 more)

### Community 40 - "testing.B"
Cohesion: 0.18
Nodes (21): BenchmarkTokenizer(), TestTokenizerCommentAdvancesOffset(), TestTokenizerStrayDelimitersMakeProgress(), NewTokenizer(), TestTokenizer_CurrentOffset_AfterTokens(), benchBuilder(), benchDocument(), BenchmarkBuild200Pages() (+13 more)

### Community 41 - "writer/options.go"
Cohesion: 0.13
Nodes (19): TestApplyFiltersWithLimitsRejectsOversizedDecode(), applyFilters(), applyFiltersWithLimits(), decodeStreamData(), decodeStreamDataWithLimits(), NewStandardRegistry(), FilterRegistry, FuzzFilters() (+11 more)

### Community 42 - "NewPDFReader"
Cohesion: 0.09
Nodes (27): TestStrictModeRejectsMissingStartXRef(), TestTolerantModeRecoversMissingStartXRef(), TestValidateDocumentReportsStructuralStatus(), TestValidateRejectsInvalidLevel(), NewPDFReader(), NewPDFReaderWithOptions(), minimalPDFBytes(), safeByte() (+19 more)

### Community 43 - "corpus"
Cohesion: 0.13
Nodes (22): sample, NewCMSSigner(), CMSSigner, popplerSig(), samplePDF(), testCert(), TestSecondSignatureKeepsFirstValid(), TestSignAndVerify() (+14 more)

### Community 44 - "DefaultStandardSubstitutes"
Cohesion: 0.12
Nodes (18): DefaultFallback(), fileScore(), NewFallback(), parseFaces(), GetDefaultScanner(), DefaultStandardSubstitutes(), loadFont(), StandardSubstitutes() (+10 more)

### Community 46 - "steps"
Cohesion: 0.09
Nodes (21): env, GPDF_REQUIRE_TOOLS, VERAPDF_SHA256, VERAPDF_VERSION, jobs, test, name, on (+13 more)

### Community 49 - "version_detector.go"
Cohesion: 0.12
Nodes (6): CatalogProvider, PDFVersion, defaultVersion(), digitAt(), parseVersionHeader(), parseVersionName()

### Community 50 - "ListObject"
Cohesion: 0.10
Nodes (5): toRoman(), ListObject, DefaultListStyle(), ListMarker, ListStyle

### Community 51 - "bidi.go"
Cohesion: 0.18
Nodes (17): TextDirection, BaseDirection(), classify(), Direction, isArabicLetterScript(), IsRTL(), Levels(), Mirror() (+9 more)

### Community 52 - "Decode"
Cohesion: 0.20
Nodes (15): Encode(), encodePNG(), encodeTIFF(), pngFilter(), sign(), Decode(), decodePNG(), decodeTIFF() (+7 more)

### Community 53 - "Font"
Cohesion: 0.09
Nodes (9): Font, allScriptFeatures(), featureLookups(), parseCoverageTable(), parseGSUBType1(), scriptFeatures(), cffInfo, GSUB (+1 more)

### Community 54 - "Pt"
Cohesion: 0.14
Nodes (4): At(), HeadingObject, ImageObject, Pt

### Community 55 - "Sign"
Cohesion: 0.31
Nodes (6): subjectName(), Sign(), SignFile(), Signer, SignOptions, SubFilter

### Community 56 - "Verify"
Cohesion: 0.22
Nodes (12): XMPInput, PDFDate(), FormatDate(), ParseDate(), FuzzParseDate(), TestParseDate(), SignatureInfo, textOf() (+4 more)

### Community 57 - "decode.go"
Cohesion: 0.21
Nodes (12): DecodePNG(), DecodePNGToRaw(), DetectDimensions(), HasAlpha(), isGrayImage(), nrgbaAt(), ProcessImage(), Decoded (+4 more)

### Community 58 - "steps"
Cohesion: 0.11
Nodes (17): runs-on, steps, jobs, bench, name, on, pull_request, workflow_dispatch (+9 more)

### Community 59 - "applyFilterChain"
Cohesion: 0.08
Nodes (22): cmapToken, codeRange, addBFRange(), bytesToInt(), decodeCMapString(), decodeHexString(), decodeRawTextBytes(), decodeUTF16BE() (+14 more)

### Community 61 - "shape_test.go"
Cohesion: 0.19
Nodes (12): cffIndex(), cffIndexCount(), cffString(), parseCFFCharset(), parseCFFDict(), loadFont(), TestCFFDictAndCharset(), TestParseCFF() (+4 more)

### Community 62 - "buildEmbeddedFontObjects"
Cohesion: 0.24
Nodes (9): buildCIDSet(), buildCIDWidthArray(), buildEmbeddedFontObjects(), embeddedProgramName(), fontBBoxArray(), pdfFontFlags(), programGlyphs(), randomSubsetTag() (+1 more)

### Community 63 - "PDFWriter"
Cohesion: 0.25
Nodes (5): Encryptor, errWriter, asErrWriter(), escapeName(), PDFWriter

### Community 64 - "Support"
Cohesion: 0.17
Nodes (9): StructRow, ElemKind, ElemRef, List, ListItem, Section, StructCell, StructTable (+1 more)

### Community 65 - "packer"
Cohesion: 0.29
Nodes (5): objectGroup, packable(), offsetWriter, packer, xrefEntry

### Community 67 - "imageRun"
Cohesion: 0.18
Nodes (11): components(), decodeJPEGToRaw(), downsampleRun(), encodeJPEG(), placedSize(), predictorParms(), resampleBox(), decodeImageFile() (+3 more)

### Community 68 - "arabic.go"
Cohesion: 0.32
Nodes (9): ArabicForms(), IsArabicLetter(), JoiningTypeOf(), LamAlefLigature(), PresentationForm(), TestPresentationForms(), nextBase(), Form (+1 more)

### Community 69 - ".BuildFields"
Cohesion: 0.11
Nodes (23): EncodeBytes(), escapeContentName(), NewEncoder(), sortedKeys(), TestEncodeBytes(), writeHexString(), writeLiteralString(), Encoder (+15 more)

### Community 70 - "SearchPages"
Cohesion: 0.12
Nodes (8): ContentSearcher, SearchResult, ContentExtractor, SearchPages(), TestSearchPages_EmptyKeyword(), TestSearchPages_FindsKeywordOnCorrectPage(), TestSearchPages_MultipleKeywords(), TestSearchPages_RecordsIndices()

### Community 72 - "runlength.go"
Cohesion: 0.23
Nodes (9): Filter, canEncodeRepeatRun(), copyLiteralRun(), copyRepeatedRun(), decode(), decodeLiteral(), encode(), writeLiteralRun() (+1 more)

### Community 76 - "Usage"
Cohesion: 0.15
Nodes (12): Checking conformance, Extracting Pages, gPDF, Merging from Remote Sources (URL/S3), Merging PDFs, Reading a PDF, Reading order, updating in place, redacting, Searching and Replacing (+4 more)

### Community 77 - "buildSignatureUpdate"
Cohesion: 0.27
Nodes (8): AddValidationData(), buildSignatureUpdate(), copyDict(), docMDPLevel(), resolveArray(), resolveDict(), update, ValidationData

### Community 79 - "PageText"
Cohesion: 0.33
Nodes (3): PageDecorator, PageText, TextAlign

### Community 80 - "Page"
Cohesion: 0.39
Nodes (4): GraphicRun, ImageRun, TextRun, Page

### Community 85 - "ReaderOptions"
Cohesion: 0.31
Nodes (8): DefaultReaderOptions(), normalizeReaderOptions(), ParseLimits, ParserMode, NewPDFReaderWithFilters(), NewPDFReaderWithFiltersAndOptions(), ReaderOptions, gPDF — code style and conventions

### Community 86 - "saslprep.go"
Cohesion: 0.38
Nodes (8): checkBidi(), isLCat(), isMappedToNothing(), isNonASCIISpace(), isProhibited(), isRandAL(), saslprep(), TestSASLprep()

### Community 87 - "project.yml"
Cohesion: 0.22
Nodes (8): encoding, excluded_tools, ignore_all_files_in_gitignore, ignored_paths, initial_prompt, languages, project_name, read_only

### Community 90 - "BuildAES256EncryptDictForWriteWithRandom"
Cohesion: 0.32
Nodes (8): aes256Password(), aesNoPadIV(), authenticateAES256(), hashR6(), aesNoPad(), TestHashR6Deterministic(), BuildAES256EncryptDictForWrite(), BuildAES256EncryptDictForWriteWithRandom()

### Community 91 - "planLinearization"
Cohesion: 0.17
Nodes (9): bitWriter, hintTableSet, linearizationPlan, bitsFor(), isPagesNode(), planLinearization(), remapRefs(), sortedPlace() (+1 more)

### Community 93 - "RectObject"
Cohesion: 0.33
Nodes (3): R(), Rect, RectObject

### Community 94 - "RichText"
Cohesion: 0.29
Nodes (4): NewRichText(), RichText, RichTextSegment, DefaultTextStyle()

### Community 96 - "CollectionOptions"
Cohesion: 0.40
Nodes (4): CollectionField, CollectionFieldType, CollectionOptions, CollectionView

### Community 98 - ".pageHeight"
Cohesion: 0.14
Nodes (5): FlowOptions, toFloat(), pageConfig, DefaultImageStyle(), ImageStyle

### Community 99 - "objHeaderBefore"
Cohesion: 0.60
Nodes (5): isPDFWhitespace(), isDigit(), isRegular(), objHeaderBefore(), scanChunk()

### Community 100 - "CellStyle"
Cohesion: 0.15
Nodes (11): FloatingImage, DefaultImageLayout(), ImageLayout, ImageSide, ImageWrap, CellImageSpec, ImageAlignment, CellHorizontalAlign (+3 more)

### Community 101 - "TestSaveLinearizedWritesUsableHintTables"
Cohesion: 0.50
Nodes (3): linValue(), TestSaveLinearizedWritesUsableHintTables(), bitReader

### Community 106 - "ICCProfile"
Cohesion: 0.33
Nodes (4): pageOutputIntent, NewICCProfile(), SRGBProfile(), ICCProfile

### Community 141 - "detectLinearization"
Cohesion: 0.40
Nodes (5): buildLinearizationInfo(), detectLinearization(), intFromDict(), isLinearized(), LinearizationInfo

### Community 143 - "DecodeTextString"
Cohesion: 0.40
Nodes (3): DecodeTextString(), TestDecodePDFDocEncoding(), TestEncryptedObjectStreams()

### Community 146 - "Support"
Cohesion: 0.67
Nodes (3): Support, FieldKind, FieldSpec

## Knowledge Gaps
- **1 isolated node(s):** `github.com/gsoultan/gpdf`
  These have ≤1 connection - possible missing edges or undocumented components. (Counts symbols only; 443 node(s) total have ≤1 connection when file, concept and rationale nodes are included.)
- **68 thin communities (<3 nodes) omitted from report** — run `graphify query` to explore isolated nodes.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **Why does `DocumentBuilder` connect `DocumentBuilder` to `pdf20_test.go`, `Name`, `Object`, `Color`, `New`, `testing.T`, `LineStyle`, `PageAccess`, `ITableBuilder`, `Integer`, `EmbeddableFont`, `.EmbedStandardFonts`, `Support`, `PageBuilder`, `.CurrentPage`, `PaintOrder`, `gPDF — architecture map`, `Font`, `FlowBuilder`, `Drawer`, `pdftest.go`, `ITableBuilder`, `testing.B`, `corpus`, `DefaultStandardSubstitutes`, `.setErr`, `.ApplyPageSize`, `ListObject`, `bidi.go`, `Support`, `StreamWriter`, `Attachment`, `Handle`, `PageText`, `MetadataBuilder`, `OutlineBuilder`, `PDFALevel`, `CollectionOptions`, `.pageHeight`, `TextBuilder`, `ICCProfile`?**
  _High betweenness centrality (0.093) - this node is a cross-community bridge._
- **Are the 44 inferred relationships involving `Name` (e.g. with `.Fields()` and `.SigFlags()`) actually correct?**
  _`Name` has 44 INFERRED edges - model-reasoned connections that need verification._
- **What connects `github.com/gsoultan/gpdf` to the rest of the system?**
  _1 weakly-connected nodes found - possible documentation gaps or missing edges._
- **Should `Page` be split into smaller, more focused modules?**
  _Cohesion score 0.12044979919678715 - nodes in this community are weakly interconnected._
- **Why does `Name` connect `Name` to `Page`, `syntax/impl/tokenizer.go`, `Op`, `Object`, `Color`, `New`, `Ref`, `testing.T`, `LineStyle`, `ITableBuilder`, `io.Writer`, `Integer`, `detectLinearization`, `Open`, `.layoutTextIntoPages`, `TextBoxObject`, `DecodeTextString`, `FlowBuilder`, `gPDF — architecture map`, `pdftest.go`, `pdfDocument`, `ExtractLayout`, `builtDocument`, `.writeEncrypted`, `io.ReaderAt`, `contentTokenizer`, `NewStandardDecryptor`, `buildEmbeddedFontObjects`, `.Validate`, `writer/options.go`, `NewPDFReader`, `corpus`, `Sign`, `Verify`, `decode.go`, `applyFilterChain`, `buildEmbeddedFontObjects`, `PDFWriter`, `Support`, `packer`, `TaggingAccess`, `imageRun`, `.BuildFields`, `StreamWriter`, `buildSignatureUpdate`, `Handle`, `Page`, `BuildAES256EncryptDictForWriteWithRandom`, `planLinearization`, `.pageHeight`, `TestSaveLinearizedWritesUsableHintTables`?**
  _High betweenness centrality (0.092) - this node is a cross-community bridge._
- **Should `syntax/impl/tokenizer.go` be split into smaller, more focused modules?**
  _Cohesion score 0.11960132890365449 - nodes in this community are weakly interconnected._
- **Why does `Page` connect `pdf20_test.go` to `Page`, `.pageHeight`, `DocumentBuilder`, `Object`, `Attachment`, `ICCProfile`, `Ref`, `LineObject`, `Integer`, `ListObject`, `TextBoxObject`, `Pt`, `decode.go`, `RectObject`?**
  _High betweenness centrality (0.079) - this node is a cross-community bridge._