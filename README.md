# gPDF

A library for reading, searching, and modifying PDF documents.

## Features

- **Read Content**: Extract text, images, and layout information from PDF files.
- **Search**: Find keywords across pages.
- **Modify**: Replace text content and save back to PDF.
- **Table Detection**: Identify table-like structures in PDF layouts.
- **Image Extraction**: Extract images from PDF pages with color space conversion support.
- **Merge & Split**: Combine multiple PDFs into one or split a PDF into several documents.
- **PDF 2.0 (ISO 32000-2)**: `%PDF-2.0` output with UTF-8 text strings, 2.0 date form, XMP metadata, AES-256 (R 6)
  encryption with SASLprep passwords, embedded/associated files (`AttachFile`), page-level output intents,
  document parts (`DocumentPart`), black point compensation, and form fields with appearance streams.
  Reading accepts every 1.x and 2.0 feature the writer produces (text strings, dates, catalog `/Version`,
  attachments, `/Extensions`, `/DPartRoot`).
- **Digital signatures**: `sign.Sign` appends a PAdES (ETSI.CAdES.detached) or PKCS#7 signature as an incremental
  update, with optional certification (DocMDP), visible appearance and RFC 3161 time-stamp; `sign.Verify` checks the
  byte range, the CMS signature and the certificate chain; `sign.AddValidationData` stores certificates, CRLs and OCSP
  responses in /DSS (PAdES-B-LT).
- **Portable collections, 3D and geospatial**: `Collection` with a schema over the attachments, `Page.Add3D` (U3D/PRC
  with poster), `Page.AddGeoViewport` with `reader.ExtractGeoViewports` for coordinates.
- **Fonts**: TrueType and OpenType/CFF (`.otf`) embedding as subsets, `EmbedStandardFonts` to embed Liberation/URW
  substitutes for the standard 14 fonts (needed for PDF/A and PDF/UA), per-block and system-font fallback
  (`RegisterBlockFallbackFont`, `EnableSystemFontFallback`), and text shaping: Unicode bidi (UAX #9) reordering,
  bracket mirroring, Arabic joining through the font's GSUB init/medi/fina features (presentation forms as a
  fallback), ligatures and kerning, with a ToUnicode CMap that maps shaped glyphs back to their text.
- **Tagged PDF and PDF/UA**: structure tree in reading order (headings, paragraphs, lists with Lbl/LBody, tables
  with header scope, figures with alternative text, sections, Link elements for link annotations), and
  `PDFUA(doc.PDFUA1 | PDFUA2)` for ISO 14289 output: pdfuaid metadata, displayed title, language, artifacts for
  decorative graphics and headers/footers, the PDF 2.0 structure namespace for PDF/UA-2, and build-time checks
  (untagged content, missing alternative text, skipped heading levels, unembedded fonts).
- **PDF/A**: `PDFA(doc.PDFA2B | PDFA3B | PDFA4 | PDFA4E | PDFA4F)` with sRGB output intent and XMP identification.

## Usage

### Reading a PDF

```go
pdf, err := doc.Open("example.pdf")
if err != nil {
    log.Fatal(err)
}
defer pdf.Close()

text, err := pdf.ReadContent()
if err == nil {
    fmt.Println(text)
}
```

### Searching and Replacing

```go
err := pdf.Replace("Old Text", "New Text")
if err != nil {
    log.Fatal(err)
}

f, _ := os.Create("modified.pdf")
err = pdf.Save(f)
f.Close()
```

### Table Detection

```go
tables, err := pdf.ReadTables()
if err != nil {
    log.Fatal(err)
}

for _, pageTables := range tables {
    for _, table := range pageTables {
        fmt.Printf("Table at (%f, %f)\n", table.X, table.Y)
    }
}
```
 
### Merging PDFs

```go
doc1, _ := doc.Open("file1.pdf")
doc2, _ := doc.Open("file2.pdf")
defer doc1.Close()
defer doc2.Close()

merged, err := doc.Merge(doc1, doc2)
if err != nil {
    log.Fatal(err)
}
defer merged.Close()

f, _ := os.Create("merged.pdf")
merged.Save(f)
f.Close()
```

### Splitting PDFs

```go
src, _ := doc.Open("input.pdf")
defer src.Close()

// Split into multiple documents, each containing at most 1 page
splitDocs, err := doc.SplitEvery(src, 1)
if err != nil {
    log.Fatal(err)
}

for i, splitDoc := range splitDocs {
    defer splitDoc.Close()
    f, _ := os.Create(fmt.Sprintf("split_%d.pdf", i))
    splitDoc.Save(f)
    f.Close()
}
```

### Extracting Pages

```go
src, _ := doc.Open("input.pdf")
defer src.Close()

// Extract pages 1-3 (inclusive-exclusive range: [1, 3))
extracted, err := doc.Extract(src, 1, 3)
if err != nil {
    log.Fatal(err)
}
defer extracted.Close()

f, _ := os.Create("extracted.pdf")
extracted.Save(f)
f.Close()
```

### Merging from Remote Sources (URL/S3)

You can open PDFs from any `io.ReaderAt` (like a memory buffer, URL, or S3 object):

```go
resp, _ := http.Get("https://example.com/file.pdf")
data, _ := io.ReadAll(resp.Body)
resp.Body.Close()

// Open from memory
reader := bytes.NewReader(data)
pdf, _ := doc.OpenReader(reader, int64(len(data)))
defer pdf.Close()

// Merge with other documents
merged, _ := doc.Merge(doc1, doc2, pdf)
```