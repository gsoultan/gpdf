package doc

import "github.com/gsoultan/gpdf/model"

// annotBuilder writes one annotation (and the objects it needs) and returns its
// reference. alloc hands out object numbers.
type annotBuilder func(objs map[int]model.Object, pageRef model.Ref, alloc func() int) model.Ref

func (b *DocumentBuilder) addAnnotation(pageIndex int, build annotBuilder) {
	if b.annots == nil {
		b.annots = map[int][]annotBuilder{}
	}
	b.annots[pageIndex] = append(b.annots[pageIndex], build)
}

func rectArray(r [4]float64) model.Array {
	return model.Array{model.Real(r[0]), model.Real(r[1]), model.Real(r[2]), model.Real(r[3])}
}

// pngXObject writes a PNG as an image XObject (with /SMask for alpha) and returns
// its reference.
func (b *DocumentBuilder) pngXObject(objs map[int]model.Object, alloc func() int, png []byte) (model.Ref, bool) {
	var im imageRun
	if err := im.SetPNG(png); err != nil {
		b.logWarning("poster image: " + err.Error())
		return model.Ref{}, false
	}
	stream := b.imageXObjectStream(im)
	if im.HasMask && len(im.Mask) > 0 {
		maskNum := alloc()
		objs[maskNum] = &model.Stream{
			Dict: model.Dict{
				model.Name("Type"):             model.Name("XObject"),
				model.Name("Subtype"):          model.Name("Image"),
				model.Name("Width"):            model.Integer(int64(im.MaskWidth)),
				model.Name("Height"):           model.Integer(int64(im.MaskHeight)),
				model.Name("BitsPerComponent"): model.Integer(8),
				model.Name("ColorSpace"):       model.Name("DeviceGray"),
				model.Name("Filter"):           model.Name("FlateDecode"),
				model.Name("Length"):           model.Integer(int64(len(im.Mask))),
			},
			Content: im.Mask,
		}
		stream.Dict[model.Name("SMask")] = model.Ref{ObjectNumber: maskNum}
	}
	num := alloc()
	objs[num] = stream
	return model.Ref{ObjectNumber: num}, true
}
