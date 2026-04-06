package layer

import "github.com/gsoultan/gpdf/model"

// Handle is a lightweight handle to an optional content group (OCG).
// Store and reuse when drawing into the same layer across multiple pages.
type Handle struct {
	Name string
	Idx  int
}

// Spec describes one optional content group (OCG) to be written into
// Catalog /OCProperties.
type Spec struct {
	Name        string
	OnByDefault bool
}

// Support owns optional content group (OCG) state for a document builder.
type Support struct {
	Layers []Spec
}

// BeginLayer registers a named OCG and returns a Handle.
// If a layer with the same name already exists, the existing handle is returned.
func (s *Support) BeginLayer(name string, onByDefault bool) *Handle {
	if name == "" {
		return nil
	}
	if s.Layers == nil {
		s.Layers = make([]Spec, 0, 4)
	}
	for i, l := range s.Layers {
		if l.Name == name {
			return &Handle{Name: name, Idx: i}
		}
	}
	idx := len(s.Layers)
	s.Layers = append(s.Layers, Spec{Name: name, OnByDefault: onByDefault})
	return &Handle{Name: name, Idx: idx}
}

// BuildOCProperties adds Catalog /OCProperties and related configuration
// objects for all layers. Returns the updated next object number.
func (s *Support) BuildOCProperties(objs map[int]model.Object, catalogDict model.Dict, nextNum int) int {
	if len(s.Layers) == 0 {
		return nextNum
	}
	var ocgRefs model.Array
	var onArray model.Array
	for _, l := range s.Layers {
		ocgNum := nextNum
		nextNum++
		ocgDict := model.Dict{
			model.Name("Type"): model.Name("OCG"),
			model.Name("Name"): model.String(l.Name),
		}
		objs[ocgNum] = ocgDict
		ref := model.Ref{ObjectNumber: ocgNum, Generation: 0}
		ocgRefs = append(ocgRefs, ref)
		if l.OnByDefault {
			onArray = append(onArray, ref)
		}
	}
	defNum := nextNum
	nextNum++
	defDict := model.Dict{
		model.Name("Name"): model.String("DefaultOCConfig"),
	}
	if len(onArray) > 0 {
		defDict[model.Name("ON")] = onArray
	}
	objs[defNum] = defDict

	ocProps := model.Dict{
		model.Name("OCGs"): ocgRefs,
		model.Name("D"):    model.Ref{ObjectNumber: defNum, Generation: 0},
	}
	ocPropsNum := nextNum
	nextNum++
	objs[ocPropsNum] = ocProps
	catalogDict[model.Name("OCProperties")] = model.Ref{ObjectNumber: ocPropsNum, Generation: 0}
	return nextNum
}
