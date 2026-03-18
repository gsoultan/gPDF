package reader

import (
	"strings"

	"gpdf/model"
)

type fontInfo struct {
	ResourceName string
	BaseFont     string
	Bold         bool
	Italic       bool
	Monospace    bool
	Serif        bool
	Widths       map[int]float64
	DefaultWidth float64
}

func resolveFontInfo(src contentSource, resources model.Dict, fontOperand model.Object) fontInfo {
	name, ok := fontOperand.(model.Name)
	if !ok || resources == nil {
		return fontInfo{}
	}
	info := fontInfo{ResourceName: string(name), DefaultWidth: 600}
	fontResources, ok := resolveDictObject(src, resources[model.Name("Font")])
	if !ok {
		return info
	}
	fontObj, ok := fontResources[name]
	if !ok {
		return info
	}
	fontDict, _, ok := resolveDictWithRef(src, fontObj)
	if !ok {
		return info
	}
	if baseFont, ok := fontDict[model.Name("BaseFont")].(model.Name); ok {
		info.BaseFont = string(baseFont)
		lower := strings.ToLower(info.BaseFont)
		info.Bold = strings.Contains(lower, "bold")
		info.Italic = strings.Contains(lower, "italic") || strings.Contains(lower, "oblique")
		info.Monospace = strings.Contains(lower, "courier") || strings.Contains(lower, "mono")
		info.Serif = strings.Contains(lower, "times") || strings.Contains(lower, "serif")
	}
	widths, ok := fontDict[model.Name("Widths")].(model.Array)
	if !ok {
		return info
	}
	firstChar, _ := fontDict[model.Name("FirstChar")].(model.Integer)
	info.Widths = make(map[int]float64, len(widths))
	for i, widthObj := range widths {
		info.Widths[int(firstChar)+i] = toFloat64(widthObj)
	}
	return info
}
