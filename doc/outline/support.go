package outline

import "github.com/gsoultan/gpdf/model"

// Entry describes one document outline (bookmark).
type Entry struct {
	Title     string
	PageIndex int
	URL       string
	DestName  string
}

// LinkAnnotation describes a link annotation on a page.
type LinkAnnotation struct {
	PageIndex int
	Rect      [4]float64
	DestPage  int
	DestName  string
	URL       string
}

// Support owns outline, named-destination, and link-annotation state.
type Support struct {
	Entries    []Entry
	NamedDests map[string]int
	LinkAnnots []LinkAnnotation
}

// BuildOutlines writes outline (bookmark) objects into objs and wires them
// into the catalog. Returns the updated next object number.
func (s *Support) BuildOutlines(objs map[int]model.Object, pageNums []int, catalogDict model.Dict, nextNum int) int {
	if len(s.Entries) == 0 || len(pageNums) == 0 {
		return nextNum
	}
	var validEntries []Entry
	for _, e := range s.Entries {
		hasPage := e.PageIndex >= 0 && e.PageIndex < len(pageNums)
		if hasPage || e.URL != "" || e.DestName != "" {
			validEntries = append(validEntries, e)
		}
	}
	if len(validEntries) == 0 {
		return nextNum
	}

	outlineRootNum := nextNum
	nextNum++
	itemNums := make([]int, len(validEntries))
	for i := range validEntries {
		itemNums[i] = nextNum
		nextNum++
	}
	rootDict := model.Dict{
		model.Name("Type"):  model.Name("Outlines"),
		model.Name("First"): model.Ref{ObjectNumber: itemNums[0], Generation: 0},
		model.Name("Last"):  model.Ref{ObjectNumber: itemNums[len(itemNums)-1], Generation: 0},
		model.Name("Count"): model.Integer(int64(len(itemNums))),
	}
	objs[outlineRootNum] = rootDict

	for i, e := range validEntries {
		itemDict := model.Dict{
			model.Name("Title"):  model.String(e.Title),
			model.Name("Parent"): model.Ref{ObjectNumber: outlineRootNum, Generation: 0},
		}

		switch {
		case e.URL != "":
			itemDict[model.Name("A")] = model.Dict{
				model.Name("S"):   model.Name("URI"),
				model.Name("URI"): model.String(e.URL),
			}
		case e.DestName != "":
			itemDict[model.Name("A")] = model.Dict{
				model.Name("S"): model.Name("GoTo"),
				model.Name("D"): model.Name(e.DestName),
			}
		default:
			pageRef := model.Ref{ObjectNumber: pageNums[e.PageIndex], Generation: 0}
			itemDict[model.Name("Dest")] = model.Array{pageRef, model.Name("Fit")}
		}
		if i > 0 {
			itemDict[model.Name("Prev")] = model.Ref{ObjectNumber: itemNums[i-1], Generation: 0}
		}
		if i < len(validEntries)-1 {
			itemDict[model.Name("Next")] = model.Ref{ObjectNumber: itemNums[i+1], Generation: 0}
		}
		objs[itemNums[i]] = itemDict
	}
	catalogDict[model.Name("Outlines")] = model.Ref{ObjectNumber: outlineRootNum, Generation: 0}
	return nextNum
}

// BuildNamedDests writes the /Dests dictionary into objs and wires it into the
// catalog. Returns the updated next object number.
func (s *Support) BuildNamedDests(objs map[int]model.Object, pageNums []int, catalogDict model.Dict, nextNum int) int {
	if len(s.NamedDests) == 0 || len(pageNums) == 0 {
		return nextNum
	}
	destsDict := model.Dict{}
	for name, idx := range s.NamedDests {
		if idx >= 0 && idx < len(pageNums) {
			pageRef := model.Ref{ObjectNumber: pageNums[idx], Generation: 0}
			destsDict[model.Name(name)] = model.Array{pageRef, model.Name("Fit")}
		}
	}
	if len(destsDict) == 0 {
		return nextNum
	}
	destsNum := nextNum
	nextNum++
	objs[destsNum] = destsDict
	catalogDict[model.Name("Dests")] = model.Ref{ObjectNumber: destsNum, Generation: 0}
	return nextNum
}

// BuildLinkAnnotations builds link annotation dicts for the given page index
// and returns an array of annotation refs plus the updated next object number.
func (s *Support) BuildLinkAnnotations(objs map[int]model.Object, pageIndex int, pageNums []int, nextNum int) (model.Array, int) {
	var annotRefs model.Array
	for _, la := range s.LinkAnnots {
		if la.PageIndex != pageIndex {
			continue
		}
		rect := model.Array{
			model.Real(la.Rect[0]), model.Real(la.Rect[1]),
			model.Real(la.Rect[2]), model.Real(la.Rect[3]),
		}
		annotDict := model.Dict{
			model.Name("Type"):    model.Name("Annot"),
			model.Name("Subtype"): model.Name("Link"),
			model.Name("Rect"):    rect,
		}

		switch {
		case la.DestPage >= 0 && la.DestPage < len(pageNums):
			annotDict[model.Name("Dest")] = model.Array{
				model.Ref{ObjectNumber: pageNums[la.DestPage], Generation: 0},
				model.Name("Fit"),
			}
		case la.DestName != "":
			annotDict[model.Name("A")] = model.Dict{
				model.Name("S"): model.Name("GoTo"),
				model.Name("D"): model.Name(la.DestName),
			}
		case la.URL != "":
			annotDict[model.Name("A")] = model.Dict{
				model.Name("S"):   model.Name("URI"),
				model.Name("URI"): model.String(la.URL),
			}
		default:
			continue
		}

		annotNum := nextNum
		nextNum++
		objs[annotNum] = annotDict
		annotRefs = append(annotRefs, model.Ref{ObjectNumber: annotNum, Generation: 0})
	}
	return annotRefs, nextNum
}
