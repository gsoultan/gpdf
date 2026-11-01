package form

import (
	"fmt"

	"github.com/gsoultan/gpdf/model"
)

// FieldKind identifies the general kind of AcroForm field.
type FieldKind int

const (
	FieldText FieldKind = iota
	FieldCheckbox
	FieldRadio
	FieldSubmit
)

// FieldSpec describes a form field and its widget annotation.
type FieldSpec struct {
	PageIndex int
	Rect      [4]float64

	Name      string
	Kind      FieldKind
	Value     string
	Checked   bool
	Tooltip   string
	Required  bool
	GroupName string
	SubmitURL string
}

// Support owns AcroForm-related state for a document builder.
type Support struct {
	UseAcroForm bool
	SigFlags    int
	Fields      []FieldSpec
}

// BuildFields builds AcroForm field dictionaries and widget annotations,
// wiring them into Catalog /AcroForm and page /Annots. Returns the AcroForm
// object number (0 if none) and the updated next object number.
func (s *Support) BuildFields(objs map[int]model.Object, pageNums []int, catalogDict model.Dict, nextNum int) (acroFormNum int, outNext int) {
	outNext = nextNum
	if !s.UseAcroForm && len(s.Fields) == 0 {
		return 0, outNext
	}
	fieldsArray := model.Array{}
	helvRef, zadbRef, outNext := fontResources(objs, outNext)
	apNum := func(s *model.Stream) model.Ref {
		n := outNext
		outNext++
		objs[n] = s
		return model.Ref{ObjectNumber: n}
	}

	radioGroups := make(map[string]struct {
		fieldNum int
		kids     model.Array
		onName   model.Name
	})

	for _, spec := range s.Fields {
		if spec.PageIndex < 0 || spec.PageIndex >= len(pageNums) {
			continue
		}
		pageRef := model.Ref{ObjectNumber: pageNums[spec.PageIndex], Generation: 0}
		w := spec.Rect[2] - spec.Rect[0]
		h := spec.Rect[3] - spec.Rect[1]
		rect := model.Array{
			model.Real(spec.Rect[0]),
			model.Real(spec.Rect[1]),
			model.Real(spec.Rect[2]),
			model.Real(spec.Rect[3]),
		}

		switch spec.Kind {
		case FieldText, FieldCheckbox, FieldSubmit:
			fieldNum := outNext
			outNext++
			widgetNum := outNext
			outNext++

			fieldDict := model.Dict{
				model.Name("T"): model.String(spec.Name),
			}
			widgetDict := model.Dict{
				model.Name("Type"):    model.Name("Annot"),
				model.Name("Subtype"): model.Name("Widget"),
				model.Name("Rect"):    rect,
				model.Name("P"):       pageRef,
				model.Name("Parent"):  model.Ref{ObjectNumber: fieldNum, Generation: 0},
			}

			switch spec.Kind {
			case FieldText:
				fieldDict[model.Name("FT")] = model.Name("Tx")
				if spec.Value != "" {
					fieldDict[model.Name("V")] = model.String(spec.Value)
				}
				ap, size := textFieldAppearance(spec.Value, w, h, helvRef)
				fieldDict[model.Name("DA")] = model.String(fmt.Sprintf("/%s %s Tf 0 g", helvName, num(size)))
				widgetDict[model.Name("AP")] = model.Dict{model.Name("N"): apNum(ap)}
			case FieldCheckbox:
				fieldDict[model.Name("FT")] = model.Name("Btn")
				onName := model.Name("Yes")
				apDict := model.Dict{
					model.Name("N"): model.Dict{
						onName:            apNum(symbolAppearance("4", w, h)),
						model.Name("Off"): apNum(emptyAppearance("4", w, h)),
					},
				}
				widgetDict[model.Name("MK")] = model.Dict{model.Name("CA"): model.String("4")}
				fieldDict[model.Name("DA")] = model.String("/" + zadbName + " 0 Tf 0 g")
				widgetDict[model.Name("AS")] = model.Name("Off")
				if spec.Checked {
					fieldDict[model.Name("V")] = onName
					widgetDict[model.Name("AS")] = onName
				} else {
					fieldDict[model.Name("V")] = model.Name("Off")
				}
				widgetDict[model.Name("AP")] = apDict
			case FieldSubmit:
				fieldDict[model.Name("FT")] = model.Name("Btn")
				if spec.Value != "" {
					fieldDict[model.Name("TU")] = model.String(spec.Value)
				}
				widgetDict[model.Name("A")] = model.Dict{
					model.Name("S"): model.Name("SubmitForm"),
					model.Name("F"): model.String(spec.SubmitURL),
				}
				fieldDict[model.Name("Ff")] = model.Integer(1 << 16) // Pushbutton
				widgetDict[model.Name("MK")] = model.Dict{model.Name("CA"): model.String(spec.Value), model.Name("BG"): model.Array{model.Real(0.75)}}
				widgetDict[model.Name("AP")] = model.Dict{model.Name("N"): apNum(buttonAppearance(spec.Value, w, h, helvRef))}
			}

			if spec.Tooltip != "" {
				fieldDict[model.Name("TU")] = model.String(spec.Tooltip)
			}
			if spec.Required {
				ff, _ := fieldDict[model.Name("Ff")].(model.Integer)
				fieldDict[model.Name("Ff")] = ff | 1
			}

			objs[fieldNum] = fieldDict
			objs[widgetNum] = widgetDict
			fieldsArray = append(fieldsArray, model.Ref{ObjectNumber: fieldNum, Generation: 0})

			attachWidgetToPage(objs, pageNums[spec.PageIndex], widgetNum)

		case FieldRadio:
			group, ok := radioGroups[spec.GroupName]
			if !ok {
				group = struct {
					fieldNum int
					kids     model.Array
					onName   model.Name
				}{fieldNum: outNext}
				outNext++
				radioGroups[spec.GroupName] = group
				fieldsArray = append(fieldsArray, model.Ref{ObjectNumber: group.fieldNum, Generation: 0})
			}
			widgetNum := outNext
			outNext++
			onName := model.Name(spec.Name)
			widgetDict := model.Dict{
				model.Name("Type"):    model.Name("Annot"),
				model.Name("Subtype"): model.Name("Widget"),
				model.Name("Rect"):    rect,
				model.Name("P"):       pageRef,
				model.Name("Parent"):  model.Ref{ObjectNumber: group.fieldNum, Generation: 0},
				model.Name("AS"):      model.Name("Off"),
				model.Name("MK"):      model.Dict{model.Name("CA"): model.String("l")},
				model.Name("AP"): model.Dict{model.Name("N"): model.Dict{
					onName:            apNum(symbolAppearance("l", w, h)),
					model.Name("Off"): apNum(emptyAppearance("l", w, h)),
				}},
			}
			if spec.Checked {
				group.onName = onName
				widgetDict[model.Name("AS")] = onName
			}
			objs[widgetNum] = widgetDict
			group.kids = append(group.kids, model.Ref{ObjectNumber: widgetNum, Generation: 0})
			radioGroups[spec.GroupName] = group

			attachWidgetToPage(objs, pageNums[spec.PageIndex], widgetNum)
		}
	}

	for name, group := range radioGroups {
		fieldDict := model.Dict{
			model.Name("FT"):   model.Name("Btn"),
			model.Name("Ff"):   model.Integer(1 << 15), // Radio
			model.Name("T"):    model.String(name),
			model.Name("Kids"): group.kids,
			model.Name("DA"):   model.String("/" + zadbName + " 0 Tf 0 g"),
		}
		if group.onName != "" {
			fieldDict[model.Name("V")] = group.onName
		}
		objs[group.fieldNum] = fieldDict
	}

	if len(fieldsArray) == 0 && !s.UseAcroForm {
		return 0, outNext
	}

	acroFormNum = outNext
	outNext++
	acroDict := model.Dict{
		model.Name("Fields"): fieldsArray,
	}
	if s.SigFlags != 0 {
		acroDict[model.Name("SigFlags")] = model.Integer(int64(s.SigFlags))
	}
	// Every widget has an appearance stream, so NeedAppearances (deprecated in
	// PDF 2.0 and forbidden in PDF/A) is not needed.
	acroDict[model.Name("DR")] = model.Dict{model.Name("Font"): model.Dict{
		model.Name(helvName): helvRef,
		model.Name(zadbName): zadbRef,
	}}
	acroDict[model.Name("DA")] = model.String("/" + helvName + " 0 Tf 0 g")
	objs[acroFormNum] = acroDict
	catalogDict[model.Name("AcroForm")] = model.Ref{ObjectNumber: acroFormNum, Generation: 0}
	return acroFormNum, outNext
}

func attachWidgetToPage(objs map[int]model.Object, pageObjNum int, widgetNum int) {
	pd, ok := objs[pageObjNum].(model.Dict)
	if !ok {
		return
	}
	var annots model.Array
	if existing, ok := pd[model.Name("Annots")].(model.Array); ok {
		annots = existing
	}
	annots = append(annots, model.Ref{ObjectNumber: widgetNum, Generation: 0})
	pd[model.Name("Annots")] = annots
	objs[pageObjNum] = pd
}
