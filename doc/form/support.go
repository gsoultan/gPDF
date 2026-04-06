package form

import "github.com/gsoultan/gpdf/model"

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
			case FieldCheckbox:
				fieldDict[model.Name("FT")] = model.Name("Btn")
				onName := model.Name("Yes")
				apDict := model.Dict{
					model.Name("N"): model.Dict{
						onName:            model.Name("Yes"),
						model.Name("Off"): model.Name("Off"),
					},
				}
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
			}

			if spec.Tooltip != "" {
				fieldDict[model.Name("TU")] = model.String(spec.Tooltip)
			}
			if spec.Required {
				fieldDict[model.Name("Ff")] = model.Integer(1)
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
			model.Name("T"):    model.String(name),
			model.Name("Kids"): group.kids,
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
	acroDict[model.Name("NeedAppearances")] = model.Boolean(true)
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
