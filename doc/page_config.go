package doc

import "gpdf/model"

// pageConfig holds page-related state for the builder.
type pageConfig struct {
	pageSize [2]float64
	pages    []pageSpec
}

func (pc *pageConfig) validPageIndex(idx int) bool {
	return idx >= 0 && idx < len(pc.pages)
}

func (pc *pageConfig) addPage(w, h float64) {
	dict := model.Dict{
		model.Name("Type"):     model.Name("Page"),
		model.Name("MediaBox"): model.Array{model.Integer(0), model.Integer(0), model.Real(w), model.Real(h)},
	}
	pc.pages = append(pc.pages, pageSpec{
		Dict:  dict,
		CurrX: 72,
		CurrY: h - 72,
	})
}

func (pc *pageConfig) height(pageIndex int) float64 {
	if !pc.validPageIndex(pageIndex) {
		return 0
	}
	spec := pc.pages[pageIndex]
	if mb, ok := spec.Dict[model.Name("MediaBox")].(model.Array); ok && len(mb) == 4 {
		return toFloat(mb[3])
	}
	if pc.pageSize[1] > 0 {
		return pc.pageSize[1]
	}
	return 842
}

func (pc *pageConfig) width(pageIndex int) float64 {
	if !pc.validPageIndex(pageIndex) {
		return 0
	}
	spec := pc.pages[pageIndex]
	if mb, ok := spec.Dict[model.Name("MediaBox")].(model.Array); ok && len(mb) == 4 {
		return toFloat(mb[2])
	}
	if pc.pageSize[0] > 0 {
		return pc.pageSize[0]
	}
	return 595
}

func toFloat(val model.Object) float64 {
	switch v := val.(type) {
	case model.Real:
		return float64(v)
	case model.Integer:
		return float64(v)
	}
	return 0
}
