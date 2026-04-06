package reader

import "github.com/gsoultan/gpdf/model"

func resolvePageSize(page model.Page) PageSize {
	box, boxName := effectivePageBox(page)
	userUnit := pageUserUnit(page)
	width := 0.0
	height := 0.0
	if len(box) >= 4 {
		width = (toFloat64(box[2]) - toFloat64(box[0])) * userUnit
		height = (toFloat64(box[3]) - toFloat64(box[1])) * userUnit
	}
	rotation := normalizeRotation(pageRotation(page))
	if rotation == 90 || rotation == 270 {
		width, height = height, width
	}
	return PageSize{
		Width:    width,
		Height:   height,
		Rotation: rotation,
		Box:      boxName,
		UserUnit: userUnit,
	}
}

func effectivePageBox(page model.Page) (model.Array, string) {
	if box, ok := page.CropBox(); ok && len(box) >= 4 {
		return box, "CropBox"
	}
	if box, ok := page.MediaBox(); ok && len(box) >= 4 {
		return box, "MediaBox"
	}
	return nil, ""
}

func pageRotation(page model.Page) int {
	rotation, _ := page.Rotate()
	return rotation
}

func normalizeRotation(rotation int) int {
	rotation %= 360
	if rotation < 0 {
		rotation += 360
	}
	return rotation
}

func pageUserUnit(page model.Page) float64 {
	v, ok := page.Dict[model.Name("UserUnit")]
	if !ok {
		return 1
	}
	switch unit := v.(type) {
	case model.Integer:
		if unit > 0 {
			return float64(unit)
		}
	case model.Real:
		if unit > 0 {
			return float64(unit)
		}
	}
	return 1
}
