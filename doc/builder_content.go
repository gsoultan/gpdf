package doc

import (
	"fmt"
	"maps"
	"math"
	"slices"
	"strings"

	"github.com/gsoultan/gpdf/content"
	"github.com/gsoultan/gpdf/font"
	"github.com/gsoultan/gpdf/model"
)

// buildGraphicsOps returns content stream ops from graphic runs (backgrounds, lines, shapes).
func (b *DocumentBuilder) buildGraphicsOps(page *pageSpec, pageHeight float64) []content.Op {
	var ops []content.Op
	for _, gr := range page.GraphicRuns {
		ops = b.appendGraphicOps(ops, gr.Ops)
	}
	return ops
}

// appendGraphicOps appends a graphic run's ops. In a PDF/UA document vector graphics
// are decorative, so a run whose marked-content operators are balanced is wrapped in
// an Artifact sequence (runs that open or close a layer are left as they are).
func (b *DocumentBuilder) appendGraphicOps(ops []content.Op, run []content.Op) []content.Op {
	if b.pdfua == "" || len(run) == 0 || !balancedMarkedContent(run) {
		return append(ops, run...)
	}
	ops = append(ops, content.Op{Name: "BMC", Args: []model.Object{model.Name("Artifact")}})
	ops = append(ops, run...)
	return append(ops, content.Op{Name: "EMC"})
}

// balancedMarkedContent reports whether every BDC/BMC in ops has its EMC.
func balancedMarkedContent(ops []content.Op) bool {
	depth := 0
	for _, op := range ops {
		switch op.Name {
		case "BDC", "BMC":
			depth++
		case "EMC":
			depth--
			if depth < 0 {
				return false
			}
		}
	}
	return depth == 0
}

// buildTextOps returns content stream ops from text runs and a font resource map (baseName -> resName).
// If embeddedFonts is nil, uses b.fc.embeddedFonts.
func (b *DocumentBuilder) buildTextOps(page *pageSpec, pageHeight float64, embeddedFonts map[string]*embeddedFontUsage) ([]content.Op, map[string]model.Name) {
	tw := b.newTextOpsWriter(embeddedFonts)
	var ops []content.Op
	for _, r := range page.TextRuns {
		ops = tw.add(ops, r)
	}
	return ops, tw.fontRes
}

// textOpsWriter turns text runs into content stream ops. It tracks the fill color and the
// text state it last set, so consecutive runs only emit the operators that change.
type textOpsWriter struct {
	b               *DocumentBuilder
	ef              map[string]*embeddedFontUsage
	fontRes         map[string]model.Name
	currentColorSet bool
	currentColor    [3]float64
	// Tc, Tw, Tz and Tr are graphics-state parameters: they survive ET and apply to every
	// later text object. Track them and emit an operator whenever a run needs a different
	// value, so one run's letter spacing (or synthetic bold) never leaks into the next.
	// The state is unknown at the start because these ops may follow other content.
	ts textStateTracker
}

func (b *DocumentBuilder) newTextOpsWriter(embeddedFonts map[string]*embeddedFontUsage) *textOpsWriter {
	ef := b.fc.embeddedFonts
	if embeddedFonts != nil {
		ef = embeddedFonts
	}
	return &textOpsWriter{b: b, ef: ef, fontRes: make(map[string]model.Name)}
}

// forget drops the tracked fill color and text state. Call it after other content
// (graphics or images) that may have changed them.
func (w *textOpsWriter) forget() {
	w.currentColorSet = false
	w.ts = textStateTracker{}
}

// add appends the ops for one text run.
func (w *textOpsWriter) add(ops []content.Op, r textRun) []content.Op {
	b := w.b
	fontRes, ef := w.fontRes, w.ef
	baseName := r.FontName
	if baseName == "" {
		baseName = "Helvetica"
	}
	resName, ok := fontRes[baseName]
	if !ok {
		resName = model.Name(fmt.Sprintf("F%d", len(fontRes)+1))
		fontRes[baseName] = resName
	}
	size := r.FontSize
	if size <= 0 {
		size = 12
	}
	if !r.UseDefaultColor {
		if !w.currentColorSet || r.TextColorRGB != w.currentColor {
			ops = append(ops,
				content.Op{
					Name: "rg",
					Args: []model.Object{
						model.Real(r.TextColorRGB[0]),
						model.Real(r.TextColorRGB[1]),
						model.Real(r.TextColorRGB[2]),
					},
				},
			)
			w.currentColor = r.TextColorRGB
			w.currentColorSet = true
		}
	}
	if r.IsArtifact {
		ops = append(ops,
			content.Op{
				Name: "BDC",
				Args: []model.Object{
					model.Name("Artifact"),
					model.Dict{model.Name("Type"): model.Name("Pagination")},
				},
			},
		)
	} else if r.HasMCID {
		ops = append(ops,
			content.Op{
				Name: "BDC",
				Args: []model.Object{
					model.Name("Span"),
					model.Dict{model.Name("MCID"): model.Integer(int64(r.MCID))},
				},
			},
		)
	}

	var textArg model.Object
	useTJ := false
	if eu, ok := ef[baseName]; ok && len(r.Glyphs) > 0 {
		// Shaped glyphs: already in visual order with contextual forms and
		// ligatures applied; kerning becomes TJ adjustments.
		eu.markGlyphs(r.Glyphs)
		textArg, useTJ = shapedTextArg(eu, r.Glyphs)
	} else if eu, ok := ef[baseName]; ok {
		eu.markText(r.Text)
		runes := []rune(r.Text)
		scale := 1000.0 / float64(eu.font.UnitsPerEm())
		var tjArray model.Array
		var currentStr strings.Builder
		var lastR rune
		hasKern := false

		for i, currR := range runes {
			if i > 0 {
				k := eu.font.Kern(lastR, currR)
				if k != 0 {
					if currentStr.Len() > 0 {
						tjArray = append(tjArray, model.HexString(eu.font.Encode(currentStr.String())))
						currentStr.Reset()
					}
					// TJ displacement: positive moves left, negative moves right.
					// Font kern: negative moves closer (right for left char).
					// So -kern * scale.
					tjArray = append(tjArray, model.Real(-float64(k)*scale))
					hasKern = true
				}
			}
			currentStr.WriteRune(currR)
			lastR = currR
		}

		if hasKern {
			if currentStr.Len() > 0 {
				tjArray = append(tjArray, model.HexString(eu.font.Encode(currentStr.String())))
			}
			textArg = tjArray
			useTJ = true
		} else {
			textArg = model.HexString(eu.font.Encode(r.Text))
		}
	} else if r.FontName == "Symbol" {
		textArg = model.String(symbolEncode(r.Text))
	} else if r.FontName == "ZapfDingbats" {
		textArg = model.String(zapfDingbatsEncode(r.Text))
	} else {
		textArg = model.String(winAnsiEncode(r.Text))
	}

	btOps := []content.Op{
		{Name: "BT", Args: nil},
	}

	if r.Rotation != 0 {
		rad := r.Rotation * math.Pi / 180
		cosA, sinA := math.Cos(rad), math.Sin(rad)
		a, b, c, d := cosA, sinA, -sinA, cosA
		if r.SyntheticItalic {
			// Combine italic shear with rotation
			c += syntheticItalicSkew * a
			d += syntheticItalicSkew * b
		}
		btOps = append(btOps, content.Op{
			Name: "Tm",
			Args: []model.Object{
				model.Real(a), model.Real(b),
				model.Real(c), model.Real(d),
				model.Real(r.X), model.Real(r.Y),
			},
		})
	} else if r.SyntheticItalic {
		btOps = append(btOps, content.Op{
			Name: "Tm",
			Args: []model.Object{
				model.Real(1), model.Real(0),
				model.Real(syntheticItalicSkew), model.Real(1),
				model.Real(r.X), model.Real(r.Y),
			},
		})
	} else {
		btOps = append(btOps, content.Op{
			Name: "Td",
			Args: []model.Object{model.Real(r.X), model.Real(r.Y)},
		})
	}

	btOps = append(btOps, content.Op{Name: "Tf", Args: []model.Object{resName, model.Real(size)}})

	renderMode := int64(0)
	if r.SyntheticBold {
		renderMode = 2 // fill then stroke
	}
	hScale := r.HorizontalScale
	if hScale == 0 {
		hScale = 100
	}
	btOps = w.ts.apply(btOps, renderMode, hScale, r.LetterSpacing, r.WordSpacing)
	if r.SyntheticBold {
		// The outline is stroked in the text colour; q/Q keep the line width
		// and stroke colour from leaking into later drawing.
		btOps = append([]content.Op{{Name: "q"}, {Name: "RG", Args: []model.Object{
			model.Real(r.TextColorRGB[0]), model.Real(r.TextColorRGB[1]), model.Real(r.TextColorRGB[2]),
		}}}, btOps...)
		btOps = append(btOps, content.Op{Name: "w", Args: []model.Object{model.Real(syntheticBoldStroke(size))}})
	}
	opName := "Tj"
	if useTJ {
		opName = "TJ"
	}
	btOps = append(btOps,
		content.Op{Name: opName, Args: []model.Object{textArg}},
		content.Op{Name: "ET", Args: nil},
	)
	if r.SyntheticBold {
		// Q restores the text state and fill colour set inside q, so forget them
		btOps = append(btOps, content.Op{Name: "Q"})
		w.ts = textStateTracker{}
		w.currentColorSet = false
	}
	ops = append(ops, btOps...)

	if r.Underline {
		tw := b.textWidth(r.Text, size, baseName)
		uy := r.Y - size*0.1
		thick := size * 0.07
		ops = append(ops,
			content.Op{Name: "q"},
			content.Op{Name: "w", Args: []model.Object{model.Real(thick)}},
			content.Op{Name: "RG", Args: []model.Object{
				model.Real(r.TextColorRGB[0]), model.Real(r.TextColorRGB[1]), model.Real(r.TextColorRGB[2]),
			}},
			content.Op{Name: "m", Args: []model.Object{model.Real(r.X), model.Real(uy)}},
			content.Op{Name: "l", Args: []model.Object{model.Real(r.X + tw), model.Real(uy)}},
			content.Op{Name: "S"},
			content.Op{Name: "Q"},
		)
	}
	if r.Strikethrough {
		tw := b.textWidth(r.Text, size, baseName)
		sy := r.Y + size*0.3
		thick := size * 0.07
		ops = append(ops,
			content.Op{Name: "q"},
			content.Op{Name: "w", Args: []model.Object{model.Real(thick)}},
			content.Op{Name: "RG", Args: []model.Object{
				model.Real(r.TextColorRGB[0]), model.Real(r.TextColorRGB[1]), model.Real(r.TextColorRGB[2]),
			}},
			content.Op{Name: "m", Args: []model.Object{model.Real(r.X), model.Real(sy)}},
			content.Op{Name: "l", Args: []model.Object{model.Real(r.X + tw), model.Real(sy)}},
			content.Op{Name: "S"},
			content.Op{Name: "Q"},
		)
	}

	if r.IsArtifact || r.HasMCID {
		ops = append(ops, content.Op{Name: "EMC", Args: nil})
	}
	return ops
}

// buildImageOps returns content stream ops from image runs and XObject resources.
// Also returns imageExtGStates (opacity ExtGState dicts) to merge into page resources.
func (b *DocumentBuilder) buildImageOps(page *pageSpec, pageHeight float64, imageStreamNums []int) ([]content.Op, model.Dict, model.Dict) {
	iw := newImageOpsWriter(imageStreamNums)
	var ops []content.Op
	for i, im := range page.ImageRuns {
		ops = iw.add(ops, i, im)
	}
	return ops, iw.xobj, iw.gs
}

// imageOpsWriter turns image runs into content stream ops and collects the XObject and
// opacity ExtGState resources they use. Image i is always named /Im<i+1>.
type imageOpsWriter struct {
	streamNums []int
	xobj       model.Dict
	gs         model.Dict
	gsIndex    int
}

func newImageOpsWriter(streamNums []int) *imageOpsWriter {
	return &imageOpsWriter{streamNums: streamNums, xobj: model.Dict{}, gs: model.Dict{}}
}

// add appends the ops that draw image run i.
func (iw *imageOpsWriter) add(ops []content.Op, i int, im imageRun) []content.Op {
	imageStreamNums := iw.streamNums
	xobj, imageGS := iw.xobj, iw.gs
	imName := model.Name("Im" + fmt.Sprintf("%d", i+1))
	w, h := im.WidthPt, im.HeightPt
	if w <= 0 {
		w = float64(im.WidthPx)
	}
	if h <= 0 {
		h = float64(im.HeightPx)
	}
	if i < len(imageStreamNums) {
		xobj[imName] = model.Ref{ObjectNumber: imageStreamNums[i], Generation: 0}
	}
	if im.IsArtifact {
		ops = append(ops,
			content.Op{
				Name: "BDC",
				Args: []model.Object{
					model.Name("Artifact"),
					model.Dict{model.Name("Type"): model.Name("Pagination")},
				},
			},
		)
	} else if im.HasMCID {
		ops = append(ops,
			content.Op{
				Name: "BDC",
				Args: []model.Object{
					model.Name("Span"),
					model.Dict{model.Name("MCID"): model.Integer(int64(im.MCID))},
				},
			},
		)
	}
	ops = append(ops, content.Op{Name: "q", Args: nil})
	if im.Opacity > 0 && im.Opacity < 1 {
		iw.gsIndex++
		gsName := model.Name(fmt.Sprintf("IMGS%d", iw.gsIndex))
		imageGS[gsName] = model.Dict{
			model.Name("Type"): model.Name("ExtGState"),
			model.Name("ca"):   model.Real(im.Opacity),
			model.Name("CA"):   model.Real(im.Opacity),
		}
		ops = append(ops, content.Op{Name: "gs", Args: []model.Object{gsName}})
	}
	if im.ClipCircle {
		ops = append(ops, circlePathOps(im.ClipCX, im.ClipCY, im.ClipR)...)
		ops = append(ops,
			content.Op{Name: "W", Args: nil},
			content.Op{Name: "n", Args: nil},
		)
	}
	var a, bb, c, d, e, f float64
	// Prefer explicit matrix when provided
	if im.Matrix != ([6]float64{}) {
		a, bb, c, d, e, f = im.Matrix[0], im.Matrix[1], im.Matrix[2], im.Matrix[3], im.Matrix[4], im.Matrix[5]
	} else if im.RotateDeg != 0 {
		θ := im.RotateDeg * math.Pi / 180
		cosθ := math.Cos(θ)
		sinθ := math.Sin(θ)
		cx := im.X + w/2
		cy := im.Y + h/2
		a = w * cosθ
		bb = -w * sinθ
		c = h * sinθ
		d = h * cosθ
		e = cx - w/2*cosθ - h/2*sinθ
		f = cy + w/2*sinθ - h/2*cosθ
	} else {
		a, bb, c, d, e, f = w, 0, 0, h, im.X, im.Y
	}
	ops = append(ops,
		content.Op{Name: "cm", Args: []model.Object{
			model.Real(a), model.Real(bb), model.Real(c), model.Real(d),
			model.Real(e), model.Real(f),
		}},
		content.Op{Name: "Do", Args: []model.Object{imName}},
		content.Op{Name: "Q", Args: nil},
	)
	if im.IsArtifact || im.HasMCID {
		ops = append(ops, content.Op{Name: "EMC", Args: nil})
	}
	return ops
}

// buildPageContent returns content stream bytes and /Resources for graphics, text, and image runs.
// The paint order is set by SetPaintOrder: by default graphics first (backgrounds and
// borders), then text, then images; with PaintInCallOrder, runs paint in the order they were added.
// When compression is enabled and effective, returns errFlateCompressed as the error.
func (b *DocumentBuilder) buildPageContent(page *pageSpec, pageHeight float64, imageStreamNums []int) ([]byte, model.Dict, error) {
	if len(page.GraphicRuns) == 0 && len(page.TextRuns) == 0 && len(page.ImageRuns) == 0 {
		return nil, nil, fmt.Errorf("no content")
	}

	var ops []content.Op
	var fontRes map[string]model.Name
	var xobj, imageGS model.Dict
	if b.paintOrder == PaintInCallOrder {
		ops, fontRes, xobj, imageGS = b.buildOpsInCallOrder(page, imageStreamNums)
	} else {
		ops = b.buildGraphicsOps(page, pageHeight)
		textOps, fr := b.buildTextOps(page, pageHeight, nil)
		ops = append(ops, textOps...)
		imageOps, xo, gs := b.buildImageOps(page, pageHeight, imageStreamNums)
		ops = append(ops, imageOps...)
		fontRes, xobj, imageGS = fr, xo, gs
	}

	contentBytes, err := content.EncodeBytes(ops)
	if err != nil {
		return nil, nil, err
	}

	resources := model.Dict{}
	if len(page.TextRuns) > 0 {
		fontDict := model.Dict{}
		hasStandardFont := false
		for base, resName := range fontRes {
			if base == "" {
				base = "Helvetica"
			}
			if _, isEmbedded := b.fc.embeddedFonts[base]; isEmbedded {
				// Placeholder: will be replaced with a Ref to the Type0 font in Build().
				fontDict[resName] = model.Dict{
					model.Name("_embedded"): model.Name(base),
				}
			} else {
				dict := model.Dict{
					model.Name("Type"):     model.Name("Font"),
					model.Name("Subtype"):  model.Name("Type1"),
					model.Name("BaseFont"): model.Name(base),
				}
				if base != "Symbol" && base != "ZapfDingbats" {
					dict[model.Name("Encoding")] = model.Name("WinAnsiEncoding")
				}
				fontDict[resName] = dict
				hasStandardFont = true
			}
		}
		if hasStandardFont {
			// Signal that standard fonts need a ToUnicode map.
			// We use a sentinel that buildPageObjects will replace with a real Ref.
			resources[model.Name("_standardToUnicode")] = model.Boolean(true)
		}
		if len(fontDict) > 0 {
			resources[model.Name("Font")] = fontDict
		}
	}
	if len(page.ImageRuns) > 0 && len(xobj) > 0 {
		resources[model.Name("XObject")] = xobj
	}
	allGS := model.Dict{}
	for _, gr := range page.GraphicRuns {
		for name, dict := range gr.ExtGStates {
			allGS[name] = dict
		}
	}
	for name, dict := range imageGS {
		allGS[name] = dict
	}
	if len(allGS) > 0 {
		resources[model.Name("ExtGState")] = allGS
	}

	return contentBytes, resources, nil
}

// sentinel error kept for compatibility (no longer used internally).
var errFlateCompressed = fmt.Errorf("flate compressed")

func (b *DocumentBuilder) imageXObjectStream(im imageRun) *model.Stream {
	dict := model.Dict{
		model.Name("Type"):             model.Name("XObject"),
		model.Name("Subtype"):          model.Name("Image"),
		model.Name("Width"):            model.Integer(int64(im.WidthPx)),
		model.Name("Height"):           model.Integer(int64(im.HeightPx)),
		model.Name("BitsPerComponent"): model.Integer(int64(im.BitsPerComponent)),
		model.Name("ColorSpace"):       model.Name(im.ColorSpace),
		// Content holds the raw samples; the writer applies the filter when saving.
		model.Name("Filter"): model.Name("FlateDecode"),
		model.Name("Length"): model.Integer(int64(len(im.Raw))),
	}
	if p := predictorParms(im.ColorSpace, im.BitsPerComponent, im.WidthPx); p != nil {
		dict[model.Name("DecodeParms")] = p
	}
	return &model.Stream{Dict: dict, Content: im.Raw}
}

func (b *DocumentBuilder) jpegXObjectStream(im imageRun) *model.Stream {
	dict := model.Dict{
		model.Name("Type"):             model.Name("XObject"),
		model.Name("Subtype"):          model.Name("Image"),
		model.Name("Width"):            model.Integer(int64(im.WidthPx)),
		model.Name("Height"):           model.Integer(int64(im.HeightPx)),
		model.Name("BitsPerComponent"): model.Integer(int64(im.BitsPerComponent)),
		model.Name("ColorSpace"):       model.Name(im.ColorSpace),
		model.Name("Filter"):           model.Name("DCTDecode"),
		model.Name("Length"):           model.Integer(int64(len(im.Raw))),
	}
	if im.ColorSpace == "DeviceCMYK" && jpegHasAdobeMarker(im.Raw) {
		// Adobe (APP14) CMYK JPEGs store inverted samples; without this /Decode the
		// image renders as a negative.
		dict[model.Name("Decode")] = model.Array{
			model.Integer(1), model.Integer(0), model.Integer(1), model.Integer(0),
			model.Integer(1), model.Integer(0), model.Integer(1), model.Integer(0),
		}
	}
	return &model.Stream{Dict: dict, Content: im.Raw}
}

// jpegHasAdobeMarker reports whether the JPEG carries an Adobe APP14 segment.
func jpegHasAdobeMarker(data []byte) bool {
	for i := 2; i+4 <= len(data); {
		if data[i] != 0xFF {
			return false
		}
		marker := data[i+1]
		if marker == 0xD8 || (marker >= 0xD0 && marker <= 0xD7) || marker == 0x01 || marker == 0xFF {
			i += 2
			continue
		}
		if marker == 0xDA || marker == 0xD9 {
			return false // scan data or end: no more headers
		}
		segLen := int(data[i+2])<<8 | int(data[i+3])
		if marker == 0xEE && i+4+5 <= len(data) && string(data[i+4:i+9]) == "Adobe" {
			return true
		}
		i += 2 + segLen
	}
	return false
}

// shapedTextArg encodes shaped glyphs for Tj, or for TJ when kerning applies.
func shapedTextArg(eu *embeddedFontUsage, glyphs []font.Glyph) (model.Object, bool) {
	ge, _ := unwrapFont(eu.font).(font.GlyphEncoder)
	encode := func(ids []uint16) []byte {
		if ge != nil {
			return ge.EncodeGlyphs(ids)
		}
		out := make([]byte, 0, len(ids)*2)
		for _, id := range ids {
			out = append(out, byte(id>>8), byte(id))
		}
		return out
	}
	scale := 1000.0 / float64(eu.font.UnitsPerEm())
	var tj model.Array
	var pending []uint16
	hasKern := false
	for _, g := range glyphs {
		if g.Kern != 0 {
			if len(pending) > 0 {
				tj = append(tj, model.HexString(encode(pending)))
				pending = pending[:0]
			}
			tj = append(tj, model.Real(-float64(g.Kern)*scale))
			hasKern = true
		}
		pending = append(pending, g.ID)
	}
	if !hasKern {
		return model.HexString(encode(pending)), false
	}
	if len(pending) > 0 {
		tj = append(tj, model.HexString(encode(pending)))
	}
	return tj, true
}

func winAnsiEncode(s string) string {
	var out strings.Builder
	for _, r := range s {
		switch {
		case r < 128:
			out.WriteByte(byte(r))
		case r == '\u2022': // Bullet
			out.WriteByte(0x95)
		case r == '\u20ac': // Euro
			out.WriteByte(0x80)
		case r == '\u201a': // Single low-9 quotation mark
			out.WriteByte(0x82)
		case r == '\u0192': // Latin small letter f with hook
			out.WriteByte(0x83)
		case r == '\u201e': // Double low-9 quotation mark
			out.WriteByte(0x84)
		case r == '\u2026': // Horizontal ellipsis
			out.WriteByte(0x85)
		case r == '\u2020': // Dagger
			out.WriteByte(0x86)
		case r == '\u2021': // Double dagger
			out.WriteByte(0x87)
		case r == '\u02c6': // Modifier letter circumflex accent
			out.WriteByte(0x88)
		case r == '\u2030': // Per mille sign
			out.WriteByte(0x89)
		case r == '\u0160': // Latin capital letter S with caron
			out.WriteByte(0x8a)
		case r == '\u2039': // Single left-pointing angle quotation mark
			out.WriteByte(0x8b)
		case r == '\u0152': // Latin capital ligature OE
			out.WriteByte(0x8c)
		case r == '\u017d': // Latin capital letter Z with caron
			out.WriteByte(0x8e)
		case r == '\u2018': // Left single quotation mark
			out.WriteByte(0x91)
		case r == '\u2019': // Right single quotation mark
			out.WriteByte(0x92)
		case r == '\u201c': // Left double quotation mark
			out.WriteByte(0x93)
		case r == '\u201d': // Right double quotation mark
			out.WriteByte(0x94)
		case r == '\u2013': // En dash
			out.WriteByte(0x96)
		case r == '\u2014': // Em dash
			out.WriteByte(0x97)
		case r == '\u02dc': // Small tilde
			out.WriteByte(0x98)
		case r == '\u2122': // Trademark sign
			out.WriteByte(0x99)
		case r == '\u0161': // Latin small letter s with caron
			out.WriteByte(0x9a)
		case r == '\u203a': // Single right-pointing angle quotation mark
			out.WriteByte(0x9b)
		case r == '\u0153': // Latin small ligature oe
			out.WriteByte(0x9c)
		case r == '\u017e': // Latin small letter z with caron
			out.WriteByte(0x9e)
		case r == '\u0178': // Latin capital letter Y with diaeresis
			out.WriteByte(0x9f)
		case r >= 160 && r <= 255:
			out.WriteByte(byte(r))
		default:
			out.WriteRune('?')
		}
	}
	return out.String()
}

func symbolEncode(s string) string {
	var out strings.Builder
	for _, r := range s {
		switch {
		case r < 128:
			out.WriteByte(byte(r))
		case r == '\u221e': // Infinity
			out.WriteByte(0xa5)
		case r == '\u2022' || r == '\u22c5' || r == '\u2219' || r == '\u00b7': // Bullet / Middle Dot / Dot Operator
			out.WriteByte(0xb7) // middle dot in Symbol is 0xb7, used as bullet
		case r == '\u20ac': // Euro
			out.WriteByte(0xa0) // Symbol font doesn't really have Euro in standard, but some versions use 0xa0
		case r == '\u2212': // Minus
			out.WriteByte(0x2d) // Or 0xad? 0x2d is hyphen but looks like minus in Symbol
		default:
			// Minimal mapping for common symbols
			out.WriteByte('?')
		}
	}
	return out.String()
}

func zapfDingbatsEncode(s string) string {
	var out strings.Builder
	for _, r := range s {
		switch {
		case r < 128:
			out.WriteByte(byte(r))
		case r == '\u25cf' || r == '\u2022': // Solid circle
			out.WriteByte(0x6c)
		case r == '\u25a0': // Solid square
			out.WriteByte(0x6d)
		case r == '\u25cb' || r == 'o': // Open circle (approximate with 0x6d or something else if better)
			out.WriteByte(0x6d) // Still solid, maybe find an open one
		case r == '\u2713': // Checkmark
			out.WriteByte(0x34)
		case r == '\u2714': // Heavy checkmark
			out.WriteByte(0x35)
		case r == '\u2702': // Scissors
			out.WriteByte(0x22)
		case r == '\u2709': // Envelope
			out.WriteByte(0x29)
		case r == '\u260e': // Telephone
			out.WriteByte(0x28)
		default:
			out.WriteByte('?')
		}
	}
	return out.String()
}

func buildToUnicodeCMap(name string, special map[byte]rune) []byte {
	var b strings.Builder
	b.WriteString("/CIDInit /ProcSet findresource begin\n")
	b.WriteString("12 dict begin\n")
	b.WriteString("begincmap\n")
	b.WriteString("/CIDSystemInfo << /Registry (Adobe) /Ordering (UCS) /Supplement 0 >> def\n")
	fmt.Fprintf(&b, "/CMapName /Adobe-Identity-%s-UCS def\n", name)
	b.WriteString("/CMapType 2 def\n")
	b.WriteString("1 begincodespacerange\n")
	b.WriteString("<00> <FF>\n")
	b.WriteString("endcodespacerange\n")

	if len(special) > 0 {
		fmt.Fprintf(&b, "%d beginbfchar\n", len(special))
		keys := slices.Sorted(maps.Keys(special))
		for _, k := range keys {
			fmt.Fprintf(&b, "<%02X> <%04X>\n", k, special[k])
		}
		b.WriteString("endbfchar\n")
	}

	b.WriteString("endcmap\n")
	b.WriteString("CMapName currentdict /CMap defineresource pop\n")
	b.WriteString("end\nend\n")
	return []byte(b.String())
}

func buildToUnicodeSymbolCMap() []byte {
	return buildToUnicodeCMap("Symbol", map[byte]rune{
		0xa5: '\u221e', // Infinity
		0xb7: '\u2022', // Bullet
	})
}

func buildToUnicodeZapfDingbatsCMap() []byte {
	return buildToUnicodeCMap("ZapfDingbats", map[byte]rune{
		0x6c: '\u25cf', // Solid circle (often used for bullet)
		0x6d: '\u25a0', // Solid square
		0x34: '\u2713', // Checkmark
		0x35: '\u2714', // Heavy checkmark
	})
}

func buildToUnicodeWinAnsiCMap() []byte {
	return buildToUnicodeCMap("WinAnsi", map[byte]rune{
		0x80: '\u20ac', 0x82: '\u201a', 0x83: '\u0192', 0x84: '\u201e', 0x85: '\u2026',
		0x86: '\u2020', 0x87: '\u2021', 0x88: '\u02c6', 0x89: '\u2030', 0x8a: '\u0160',
		0x8b: '\u2039', 0x8c: '\u0152', 0x8e: '\u017d', 0x91: '\u2018', 0x92: '\u2019',
		0x93: '\u201c', 0x94: '\u201d', 0x95: '\u2022', 0x96: '\u2013', 0x97: '\u2014',
		0x98: '\u02dc', 0x99: '\u2122', 0x9a: '\u0161', 0x9b: '\u203a', 0x9c: '\u0153',
		0x9e: '\u017e', 0x9f: '\u0178',
	})
}

// textStateTracker remembers the text state parameters last set in a content stream.
type textStateTracker struct {
	known      bool
	tr         int64
	tz, tc, tw float64
}

// apply appends the Tr/Tz/Tc/Tw operators needed to reach the requested state.
func (t *textStateTracker) apply(ops []content.Op, tr int64, tz, tc, tw float64) []content.Op {
	if !t.known || t.tr != tr {
		ops = append(ops, content.Op{Name: "Tr", Args: []model.Object{model.Integer(tr)}})
	}
	if !t.known || t.tz != tz {
		ops = append(ops, content.Op{Name: "Tz", Args: []model.Object{model.Real(tz)}})
	}
	if !t.known || t.tc != tc {
		ops = append(ops, content.Op{Name: "Tc", Args: []model.Object{model.Real(tc)}})
	}
	if !t.known || t.tw != tw {
		ops = append(ops, content.Op{Name: "Tw", Args: []model.Object{model.Real(tw)}})
	}
	*t = textStateTracker{known: true, tr: tr, tz: tz, tc: tc, tw: tw}
	return ops
}

// syntheticItalicSkew is the horizontal shear of synthesised oblique text,
// the value Skia (and so Chromium) uses.
const syntheticItalicSkew = 0.25

// syntheticBoldStroke returns the outline stroke width (in points) for
// synthesised bold text of the given size, following Skia's fake-bold
// outset: 1/24 of the size at 9px or less, 1/32 at 36px or more,
// interpolated linearly in between.
func syntheticBoldStroke(size float64) float64 {
	px := size / 0.75
	ratio := 1.0 / 24
	switch {
	case px >= 36:
		ratio = 1.0 / 32
	case px > 9:
		ratio = 1.0/24 + (px-9)/(36-9)*(1.0/32-1.0/24)
	}
	return size * ratio
}
