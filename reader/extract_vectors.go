package reader

import (
	"slices"

	"gpdf/content"
	contentimpl "gpdf/content/impl"
	"gpdf/model"
)

type vectorExtractor struct {
	ctm         matrix
	ctmStack    []matrix
	strokeColor ColorRGB
	fillColor   ColorRGB
	pathLines   []vectorLine
	pathRects   []vectorRect
	currentX    float64
	currentY    float64
	hasCurrent  bool
	shapes      []VectorShape
}

type vectorLine struct {
	x1 float64
	y1 float64
	x2 float64
	y2 float64
}

type vectorRect struct {
	x float64
	y float64
	w float64
	h float64
}

func ExtractVectorsPerPage(src contentSource) ([][]VectorShape, error) {
	pages, err := src.Pages()
	if err != nil {
		return nil, err
	}
	parser := contentimpl.NewStreamParser()
	result := make([][]VectorShape, len(pages))
	for i, page := range pages {
		result[i] = extractVectorsFromPage(src, parser, page, 0, 0)
	}
	return result, nil
}

func extractVectorsFromPage(src contentSource, parser content.Parser, page model.Page, maxDecodedBytes int, maxOps int) []VectorShape {
	resources, _ := page.Resources()
	ops, err := pageContentOps(src, parser, page, maxDecodedBytes, maxOps)
	if err != nil || len(ops) == 0 {
		return nil
	}
	return extractVectorsFromOps(src, parser, ops, resources)
}

func extractVectorsFromOps(src contentSource, parser content.Parser, ops []content.Op, resources model.Dict) []VectorShape {
	ex := newVectorExtractor()
	return ex.extract(src, parser, ops, resources, make(map[model.Ref]struct{}, 4))
}

func newVectorExtractor() *vectorExtractor {
	return &vectorExtractor{
		ctm:         identityMatrix(),
		strokeColor: ColorRGB{},
		fillColor:   ColorRGB{},
	}
}

func (v *vectorExtractor) extract(src contentSource, parser content.Parser, ops []content.Op, resources model.Dict, visited map[model.Ref]struct{}) []VectorShape {
	for _, op := range ops {
		switch op.Name {
		case "q":
			v.ctmStack = append(v.ctmStack, v.ctm)
		case "Q":
			if len(v.ctmStack) == 0 {
				v.ctm = identityMatrix()
				continue
			}
			v.ctm = v.ctmStack[len(v.ctmStack)-1]
			v.ctmStack = v.ctmStack[:len(v.ctmStack)-1]
		case "cm":
			args := collectFloatArgs(op.Args)
			if len(args) >= 6 {
				v.ctm = v.ctm.multiply(matrixFromArgs(args))
			}
		case "RG":
			v.strokeColor = colorFromArgs(op.Args)
		case "rg":
			v.fillColor = colorFromArgs(op.Args)
		case "m":
			if len(op.Args) < 2 {
				continue
			}
			v.pathLines = nil
			v.pathRects = nil
			v.currentX = toFloat64(op.Args[0])
			v.currentY = toFloat64(op.Args[1])
			v.hasCurrent = true
		case "l":
			if len(op.Args) < 2 || !v.hasCurrent {
				continue
			}
			x1, y1 := v.currentX, v.currentY
			x2, y2 := toFloat64(op.Args[0]), toFloat64(op.Args[1])
			v.pathLines = append(v.pathLines, vectorLine{x1: x1, y1: y1, x2: x2, y2: y2})
			v.currentX = x2
			v.currentY = y2
		case "re":
			if len(op.Args) < 4 {
				continue
			}
			v.pathRects = append(v.pathRects, vectorRect{
				x: toFloat64(op.Args[0]),
				y: toFloat64(op.Args[1]),
				w: toFloat64(op.Args[2]),
				h: toFloat64(op.Args[3]),
			})
		case "S", "s":
			v.flushPath(true, false)
		case "f", "F", "f*":
			v.flushPath(false, true)
		case "B", "B*", "b", "b*":
			v.flushPath(true, true)
		case "n":
			v.pathLines = nil
			v.pathRects = nil
			v.hasCurrent = false
		case "Do":
			v.walkFormXObject(src, parser, op.Args, resources, visited)
		}
	}
	return v.shapes
}

func (v *vectorExtractor) walkFormXObject(src contentSource, parser content.Parser, args []model.Object, resources model.Dict, visited map[model.Ref]struct{}) {
	if len(args) == 0 {
		return
	}
	name, ok := args[0].(model.Name)
	if !ok {
		return
	}
	xObjectDict, ok := resources[model.Name("XObject")].(model.Dict)
	if !ok {
		return
	}
	xObj := xObjectDict[name]
	stream, ref, ok := resolveStreamObject(src, xObj)
	if !ok || stream == nil {
		return
	}
	if ref != nil {
		if _, seen := visited[*ref]; seen {
			return
		}
		visited[*ref] = struct{}{}
		defer delete(visited, *ref)
	}
	if subtype, ok := stream.Dict[model.Name("Subtype")].(model.Name); !ok || subtype != "Form" {
		return
	}
	childOps, err := parser.Parse(stream.Content)
	if err != nil {
		return
	}
	childResources := resources
	if dict, ok := resolveDictObject(src, stream.Dict[model.Name("Resources")]); ok {
		childResources = mergeResourceDict(resources, dict)
	}
	child := &vectorExtractor{
		ctm:         v.ctm,
		strokeColor: v.strokeColor,
		fillColor:   v.fillColor,
	}
	if matrixArray, ok := stream.Dict[model.Name("Matrix")].(model.Array); ok && len(matrixArray) >= 6 {
		args := []float64{toFloat64(matrixArray[0]), toFloat64(matrixArray[1]), toFloat64(matrixArray[2]), toFloat64(matrixArray[3]), toFloat64(matrixArray[4]), toFloat64(matrixArray[5])}
		child.ctm = child.ctm.multiply(matrixFromArgs(args))
	}
	v.shapes = append(v.shapes, child.extract(src, parser, childOps, childResources, visited)...)
}

func colorFromArgs(args []model.Object) ColorRGB {
	if len(args) < 3 {
		return ColorRGB{}
	}
	return ColorRGB{R: toFloat64(args[0]), G: toFloat64(args[1]), B: toFloat64(args[2])}
}

func (v *vectorExtractor) flushPath(stroke bool, fill bool) {
	for _, line := range v.pathLines {
		x1, y1 := v.ctm.apply(line.x1, line.y1)
		x2, y2 := v.ctm.apply(line.x2, line.y2)
		v.shapes = append(v.shapes, VectorShape{
			Kind:        "line",
			X1:          x1,
			Y1:          y1,
			X2:          x2,
			Y2:          y2,
			Stroke:      stroke,
			Fill:        fill,
			StrokeColor: v.strokeColor,
			FillColor:   v.fillColor,
		})
	}
	for _, rect := range v.pathRects {
		x1, y1 := v.ctm.apply(rect.x, rect.y)
		x2, y2 := v.ctm.apply(rect.x+rect.w, rect.y+rect.h)
		v.shapes = append(v.shapes, VectorShape{
			Kind:        "rect",
			X1:          min(x1, x2),
			Y1:          min(y1, y2),
			X2:          max(x1, x2),
			Y2:          max(y1, y2),
			Stroke:      stroke,
			Fill:        fill,
			StrokeColor: v.strokeColor,
			FillColor:   v.fillColor,
		})
	}
	v.pathLines = nil
	v.pathRects = nil
	v.hasCurrent = false
}

func collectFloatArgs(args []model.Object) []float64 {
	if len(args) == 0 {
		return nil
	}
	values := make([]float64, 0, len(args))
	for _, arg := range args {
		values = append(values, toFloat64(arg))
	}
	return values
}

func sortedShapes(shapes []VectorShape) []VectorShape {
	cloned := slices.Clone(shapes)
	slices.SortFunc(cloned, func(a, b VectorShape) int {
		if a.Y1 != b.Y1 {
			if a.Y1 < b.Y1 {
				return -1
			}
			return 1
		}
		if a.X1 < b.X1 {
			return -1
		}
		if a.X1 > b.X1 {
			return 1
		}
		if a.Kind < b.Kind {
			return -1
		}
		if a.Kind > b.Kind {
			return 1
		}
		return 0
	})
	return cloned
}
