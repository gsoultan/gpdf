package doc

import (
	"fmt"
	"os"
	"sort"

	"github.com/gsoultan/gpdf/model"
)

// Build returns a Document that can be saved. The document is in-memory; see
// StreamTo and BuildTo for output that is written as it is built.
func (b *DocumentBuilder) Build() (Document, error) {
	if b.stream != nil {
		return nil, fmt.Errorf("%w: a builder with StreamTo is completed with Finish", ErrStreaming)
	}
	return b.build()
}

func (b *DocumentBuilder) build() (*builtDocument, error) {
	if b.err != nil {
		return nil, b.err
	}
	defer b.applyDecorators()()
	if b.err != nil {
		return nil, b.err
	}
	b.prepareMetadata()
	objs := make(map[int]model.Object)
	nextNum := 1
	if b.stream != nil {
		// Numbers below this belong to the images already written.
		sw, err := b.stream.writer(b)
		if err != nil {
			return nil, err
		}
		nextNum = sw.NextObjectNumber()
	}

	catalogNum := nextNum
	nextNum++
	pagesNum := nextNum
	nextNum++

	pageNums, pageRefs := b.allocatePageNums(nextNum)
	nextNum += len(b.pc.pages)

	nextNum = b.buildMinimalStream(objs, nextNum)
	minimalStreamNum := nextNum - 1

	nextNum, err := b.buildPageObjects(objs, pageNums, pagesNum, minimalStreamNum, nextNum)
	if err != nil {
		return nil, err
	}

	// Add shared standard ToUnicode stream if used
	standardToUnicodeRef := model.Ref{ObjectNumber: 0, Generation: 0}
	symbolToUnicodeRef := model.Ref{ObjectNumber: 0, Generation: 0}
	zapfToUnicodeRef := model.Ref{ObjectNumber: 0, Generation: 0}
	for _, obj := range objs {
		if pd, ok := obj.(model.Dict); ok {
			if res, ok := pd[model.Name("Resources")].(model.Dict); ok {
				if _, hasSent := res[model.Name("_standardToUnicode")]; hasSent {
					if standardToUnicodeRef.ObjectNumber == 0 {
						standardToUnicodeRef = model.Ref{ObjectNumber: nextNum, Generation: 0}
						nextNum++
						cmap := buildToUnicodeWinAnsiCMap()
						objs[standardToUnicodeRef.ObjectNumber] = &model.Stream{
							Dict:    model.Dict{model.Name("Length"): model.Integer(len(cmap))},
							Content: cmap,
						}

						symbolToUnicodeRef = model.Ref{ObjectNumber: nextNum, Generation: 0}
						nextNum++
						scmap := buildToUnicodeSymbolCMap()
						objs[symbolToUnicodeRef.ObjectNumber] = &model.Stream{
							Dict:    model.Dict{model.Name("Length"): model.Integer(len(scmap))},
							Content: scmap,
						}

						zapfToUnicodeRef = model.Ref{ObjectNumber: nextNum, Generation: 0}
						nextNum++
						zcmap := buildToUnicodeZapfDingbatsCMap()
						objs[zapfToUnicodeRef.ObjectNumber] = &model.Stream{
							Dict:    model.Dict{model.Name("Length"): model.Integer(len(zcmap))},
							Content: zcmap,
						}
					}
					delete(res, model.Name("_standardToUnicode"))
					if fontDict, ok := res[model.Name("Font")].(model.Dict); ok {
						for _, fontObj := range fontDict {
							if fd, ok := fontObj.(model.Dict); ok {
								if subtype, _ := fd[model.Name("Subtype")].(model.Name); subtype == model.Name("Type1") {
									baseFont, _ := fd[model.Name("BaseFont")].(model.Name)
									if baseFont == "Symbol" {
										fd[model.Name("ToUnicode")] = symbolToUnicodeRef
									} else if baseFont == "ZapfDingbats" {
										fd[model.Name("ToUnicode")] = zapfToUnicodeRef
									} else {
										fd[model.Name("ToUnicode")] = standardToUnicodeRef
									}
								}
							}
						}
					}
				}
			}
		}
	}

	pagesDict := model.Dict{
		model.Name("Type"):  model.Name("Pages"),
		model.Name("Kids"):  pageRefs,
		model.Name("Count"): model.Integer(len(b.pc.pages)),
	}
	objs[pagesNum] = pagesDict

	catalogDict := model.Dict{
		model.Name("Type"):  model.Name("Catalog"),
		model.Name("Pages"): model.Ref{ObjectNumber: pagesNum, Generation: 0},
	}
	if b.metadata.Lang != "" {
		catalogDict[model.Name("Lang")] = model.String(b.metadata.Lang)
	}
	if b.pdfua != "" {
		// PDF/UA: viewers show the document title, not the file name.
		catalogDict[model.Name("ViewerPreferences")] = model.Dict{model.Name("DisplayDocTitle"): model.Boolean(true)}
	}

	nextNum = b.buildCatalogExtras(objs, catalogDict, pageNums, nextNum)
	nextNum = b.buildAttachments(objs, catalogDict, pageNums, nextNum)
	nextNum = b.buildPageOutputIntents(objs, pageNums, nextNum)
	nextNum = b.buildCollection(objs, catalogDict, nextNum)
	if err := b.buildDocumentParts(objs, catalogDict, pageNums, nextNum); err != nil {
		return nil, err
	}
	if b.partsNextNum > nextNum {
		nextNum = b.partsNextNum
	}
	nextNum = b.buildEmbeddedFonts(objs, nextNum)

	objs[catalogNum] = catalogDict

	infoNum := nextNum
	nextNum++
	objs[infoNum] = b.metadata.BuildInfoDict()

	nextNum = b.metadata.BuildMetadataStream(objs, catalogDict, nextNum)

	trailer := model.Trailer{
		Dict: model.Dict{
			model.Name("Root"): model.Ref{ObjectNumber: catalogNum, Generation: 0},
			model.Name("Size"): model.Integer(nextNum),
			model.Name("Info"): model.Ref{ObjectNumber: infoNum, Generation: 0},
		},
	}
	encodeTextStrings(objs, b.metadata.PDF20)
	if err := b.checkPDFA(objs); err != nil {
		return nil, err
	}
	if err := b.checkPDFUA(objs); err != nil {
		return nil, err
	}
	d := &builtDocument{
		trailer:         trailer,
		objects:         objs,
		size:            nextNum,
		noObjectStreams: b.noObjectStreams,
	}
	if b.pdfa != "" {
		d.pdfa = true
		d.version = b.pdfa.PDFVersion()
	} else if v := b.pdfuaVersion(); v != "" {
		d.version = v
	}
	return d, nil
}

// BuildAndSave builds the document and saves it to the specified path.
func (b *DocumentBuilder) BuildAndSave(path string) error {
	doc, err := b.Build()
	if err != nil {
		return err
	}
	defer doc.Close()

	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	if err := doc.Save(f); err != nil {
		return err
	}
	return f.Close()
}

// allocatePageNums returns object numbers and refs for all pages.
func (b *DocumentBuilder) allocatePageNums(startNum int) (pageNums []int, pageRefs model.Array) {
	pageNums = make([]int, 0, len(b.pc.pages))
	for i := range b.pc.pages {
		num := startNum + i
		pageNums = append(pageNums, num)
		pageRefs = append(pageRefs, model.Ref{ObjectNumber: num, Generation: 0})
	}
	return
}

// buildMinimalStream creates a shared minimal content stream for blank pages.
func (b *DocumentBuilder) buildMinimalStream(objs map[int]model.Object, nextNum int) int {
	minimalContent := []byte("n\n")
	objs[nextNum] = &model.Stream{
		Dict:    model.Dict{model.Name("Length"): model.Integer(len(minimalContent))},
		Content: minimalContent,
	}
	return nextNum + 1
}

// buildPageObjects builds page dicts, content streams, image XObjects, and link annotations.
func (b *DocumentBuilder) buildPageObjects(objs map[int]model.Object, pageNums []int, pagesNum, minimalStreamNum, nextNum int) (int, error) {
	for idx, pageNum := range pageNums {
		spec := b.pc.pages[idx]
		pageDict := copyPageDict(spec.Dict)
		hasContent := len(spec.TextRuns) > 0 || len(spec.ImageRuns) > 0 || len(spec.GraphicRuns) > 0

		if !hasContent {
			pageDict[model.Name("Contents")] = model.Ref{ObjectNumber: minimalStreamNum, Generation: 0}
			if _, ok := pageDict[model.Name("Resources")]; !ok {
				pageDict[model.Name("Resources")] = model.Dict{}
			}
		} else {
			contentStreamNum := nextNum
			nextNum++
			imageStreamNums := make([]int, len(spec.ImageRuns))
			for i, im := range spec.ImageRuns {
				if im.ObjNum > 0 {
					imageStreamNums[i] = im.ObjNum // written already (streaming output)
					continue
				}
				imageStreamNums[i] = nextNum
				nextNum++
			}
			contentBytes, resources, buildErr := b.buildPageContent(&b.pc.pages[idx], b.pc.height(idx), imageStreamNums)
			if buildErr != nil {
				return nextNum, fmt.Errorf("page %d: %w", idx+1, buildErr)
			}
			pageDict[model.Name("Contents")] = model.Ref{ObjectNumber: contentStreamNum, Generation: 0}
			if existing, ok := pageDict[model.Name("Resources")].(model.Dict); ok && existing != nil {
				for k, v := range resources {
					existing[k] = v
				}
				pageDict[model.Name("Resources")] = existing
			} else {
				pageDict[model.Name("Resources")] = resources
			}
			streamDict := model.Dict{model.Name("Length"): model.Integer(len(contentBytes))}
			if !b.noCompressContent {
				streamDict[model.Name("Filter")] = model.Name("FlateDecode")
			}
			contentStream := &model.Stream{Dict: streamDict, Content: contentBytes}
			if b.stream != nil {
				if err := b.stream.writeObject(b, contentStreamNum, contentStream); err != nil {
					return nextNum, err
				}
			} else {
				objs[contentStreamNum] = contentStream
			}
			for i, im := range spec.ImageRuns {
				if im.ObjNum > 0 {
					continue
				}
				imgStream, maskStream := b.imageStreams(im)
				if maskStream != nil {
					maskNum := nextNum
					nextNum++
					objs[maskNum] = maskStream
					imgStream.Dict[model.Name("SMask")] = model.Ref{ObjectNumber: maskNum}
				}
				objs[imageStreamNums[i]] = imgStream
			}
		}

		annotRefs, updatedNext := b.outlines.BuildLinkAnnotations(objs, idx, pageNums, nextNum)
		nextNum = updatedNext
		for _, build := range b.annots[idx] {
			ref := build(objs, model.Ref{ObjectNumber: pageNum}, func() int { n := nextNum; nextNum++; return n })
			annotRefs = append(annotRefs, ref)
		}
		if len(annotRefs) > 0 {
			pageDict[model.Name("Annots")] = annotRefs
			if b.useTagged {
				// Tab order follows the structure tree (PDF/UA requires /Tabs on pages with annotations).
				pageDict[model.Name("Tabs")] = model.Name("S")
			}
		}

		objs[pageNum] = pageDict
		pageDict[model.Name("Parent")] = model.Ref{ObjectNumber: pagesNum, Generation: 0}
	}
	return nextNum, nil
}

// buildCatalogExtras adds named dests, outlines, form fields, tagged structure, OCGs, and ICC to the catalog.
func (b *DocumentBuilder) buildCatalogExtras(objs map[int]model.Object, catalogDict model.Dict, pageNums []int, nextNum int) int {
	nextNum = b.outlines.BuildNamedDests(objs, pageNums, catalogDict, nextNum)

	if b.forms.UseAcroForm || len(b.forms.Fields) > 0 {
		_, nextNum = b.forms.BuildFields(objs, pageNums, catalogDict, nextNum)
	}

	if b.useTagged {
		nextNum = b.tagging.BuildStructure(objs, catalogDict, pageNums, nextNum)
	}

	nextNum = b.layers.BuildOCProperties(objs, catalogDict, nextNum)

	nextNum = b.outlines.BuildOutlines(objs, pageNums, catalogDict, nextNum)

	nextNum = b.buildICCOutputIntent(objs, catalogDict, nextNum)

	return nextNum
}

// buildICCOutputIntent writes the ICC output intent objects.
func (b *DocumentBuilder) buildICCOutputIntent(objs map[int]model.Object, catalogDict model.Dict, nextNum int) int {
	if b.iccProfile == nil || len(b.iccProfile.Data) == 0 {
		return nextNum
	}
	iccStreamNum := nextNum
	nextNum++
	objs[iccStreamNum] = &model.Stream{
		Dict: model.Dict{
			model.Name("N"):         model.Integer(int64(b.iccProfile.N)),
			model.Name("Alternate"): model.Name(b.iccProfile.Alternate),
			model.Name("Length"):    model.Integer(int64(len(b.iccProfile.Data))),
			model.Name("Filter"):    model.Name("FlateDecode"),
		},
		Content: b.iccProfile.Data,
	}
	outputIntent := model.Dict{
		model.Name("Type"):                      model.Name("OutputIntent"),
		model.Name("S"):                         model.Name("GTS_PDFA1"),
		model.Name("OutputConditionIdentifier"): model.String("sRGB IEC61966-2.1"),
		model.Name("Info"):                      model.String("sRGB IEC61966-2.1"),
		model.Name("RegistryName"):              model.String("http://www.color.org"),
		model.Name("DestOutputProfile"):         model.Ref{ObjectNumber: iccStreamNum},
	}
	intentNum := nextNum
	nextNum++
	objs[intentNum] = outputIntent
	catalogDict[model.Name("OutputIntents")] = model.Array{
		model.Ref{ObjectNumber: intentNum},
	}
	return nextNum
}

// buildEmbeddedFonts creates CID font objects and replaces placeholder refs in page resources.
func (b *DocumentBuilder) buildEmbeddedFonts(objs map[int]model.Object, nextNum int) int {
	embeddedFontObjNums := make(map[string]int)
	for psName, usage := range b.fc.embeddedFonts {
		if len(usage.usedGIDs) == 0 {
			continue
		}
		type0Num := buildEmbeddedFontObjects(usage, objs, &nextNum, b.pdfa != "" && !b.pdfa.pdf20())
		embeddedFontObjNums[psName] = type0Num
	}
	if len(embeddedFontObjNums) > 0 {
		b.replaceEmbeddedFontPlaceholders(objs, embeddedFontObjNums)
	}
	return b.indirectStandardFonts(objs, nextNum)
}

// indirectStandardFonts moves the simple font dictionaries that pages hold
// inline in their /Resources into shared indirect objects, one per distinct
// dictionary. Readers expect font dictionaries to be indirect (poppler's font
// scanner reports an internal error for a direct one), and sharing them keeps
// multi-page files small.
func (b *DocumentBuilder) indirectStandardFonts(objs map[int]model.Object, nextNum int) int {
	shared := map[string]int{}
	pageNums := make([]int, 0, len(objs))
	for num := range objs {
		pageNums = append(pageNums, num)
	}
	sort.Ints(pageNums)
	for _, num := range pageNums {
		pageDict, ok := objs[num].(model.Dict)
		if !ok || pageDict[model.Name("Type")] != model.Name("Page") {
			continue
		}
		resDict, _ := pageDict[model.Name("Resources")].(model.Dict)
		fontDict, _ := resDict[model.Name("Font")].(model.Dict)
		for resName, val := range fontDict {
			d, ok := val.(model.Dict)
			if !ok {
				continue
			}
			if _, placeholder := d[model.Name("_embedded")]; placeholder {
				continue
			}
			key := fmt.Sprint(d) // fmt prints map keys sorted, so equal dicts share a key
			objNum, seen := shared[key]
			if !seen {
				objNum = nextNum
				nextNum++
				objs[objNum] = d
				shared[key] = objNum
			}
			fontDict[resName] = model.Ref{ObjectNumber: objNum}
		}
	}
	return nextNum
}
