package doc

import (
	"fmt"

	"github.com/gsoultan/gpdf/model"
)

// Model3D is a 3D artwork stream for a 3D annotation (ISO 32000-2 §13.6).
type Model3D struct {
	// Format is "U3D" or "PRC".
	Format string
	Data   []byte
	// Name labels the artwork (annotation /Contents and the default view's name).
	Name string
	// Poster, when set, is a PNG shown before the 3D content is activated.
	Poster []byte
	// ActivateOnOpen makes the viewer start the 3D artwork when the page opens;
	// otherwise it starts when clicked.
	ActivateOnOpen bool
}

// Add3D places a 3D annotation on the page. Viewers without 3D support show the
// poster image, if any, or nothing.
func (p *Page) Add3D(rect [4]float64, m Model3D) *Page {
	switch m.Format {
	case "U3D", "PRC":
	default:
		p.builder.setErr(fmt.Errorf("3D model format %q: want U3D or PRC", m.Format))
		return p
	}
	if len(m.Data) == 0 {
		p.builder.setErr(fmt.Errorf("3D model has no data"))
		return p
	}
	p.builder.addAnnotation(p.pageIndex, func(objs map[int]model.Object, pageRef model.Ref, alloc func() int) model.Ref {
		streamNum := alloc()
		stream := &model.Stream{
			Dict: model.Dict{
				model.Name("Type"):    model.Name("3D"),
				model.Name("Subtype"): model.Name(m.Format),
				model.Name("Length"):  model.Integer(len(m.Data)),
				model.Name("Filter"):  model.Name("FlateDecode"),
			},
			Content: m.Data,
		}
		objs[streamNum] = stream
		name := m.Name
		if name == "" {
			name = "Default"
		}
		view := model.Dict{
			model.Name("Type"): model.Name("3DView"),
			model.Name("XN"):   model.String(name),
			model.Name("IN"):   model.String(name),
			model.Name("MS"):   model.Name("M"),
			model.Name("C2W"):  model.Array{model.Integer(1), model.Integer(0), model.Integer(0), model.Integer(0), model.Integer(1), model.Integer(0), model.Integer(0), model.Integer(0), model.Integer(1), model.Integer(0), model.Integer(0), model.Integer(0)},
		}
		stream.Dict[model.Name("VA")] = model.Array{view}
		stream.Dict[model.Name("DV")] = model.Integer(0)
		activation := model.Name("PC")
		if m.ActivateOnOpen {
			activation = "PO"
		}
		annot := model.Dict{
			model.Name("Type"):     model.Name("Annot"),
			model.Name("Subtype"):  model.Name("3D"),
			model.Name("Rect"):     rectArray(rect),
			model.Name("P"):        pageRef,
			model.Name("F"):        model.Integer(4),
			model.Name("Contents"): model.String(name),
			model.Name("3DD"):      model.Ref{ObjectNumber: streamNum},
			model.Name("3DV"):      model.Integer(0),
			model.Name("3DA"): model.Dict{
				model.Name("A"):   activation,
				model.Name("D"):   model.Name("PC"),
				model.Name("DIS"): model.Name("I"),
			},
		}
		w, h := rect[2]-rect[0], rect[3]-rect[1]
		if len(m.Poster) > 0 && w > 0 && h > 0 {
			if ref, ok := p.builder.pngXObject(objs, alloc, m.Poster); ok {
				apNum := alloc()
				content := fmt.Sprintf("q %s 0 0 %s 0 0 cm /Im0 Do Q", model.FormatReal(w), model.FormatReal(h))
				objs[apNum] = &model.Stream{
					Dict: model.Dict{
						model.Name("Type"):      model.Name("XObject"),
						model.Name("Subtype"):   model.Name("Form"),
						model.Name("BBox"):      model.Array{model.Integer(0), model.Integer(0), model.Real(w), model.Real(h)},
						model.Name("Resources"): model.Dict{model.Name("XObject"): model.Dict{model.Name("Im0"): ref}},
						model.Name("Length"):    model.Integer(len(content)),
					},
					Content: []byte(content),
				}
				annot[model.Name("AP")] = model.Dict{model.Name("N"): model.Ref{ObjectNumber: apNum}}
			}
		}
		num := alloc()
		objs[num] = annot
		return model.Ref{ObjectNumber: num}
	})
	return p
}
