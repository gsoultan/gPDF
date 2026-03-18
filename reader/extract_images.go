package reader

import (
	"fmt"
	"math"

	"gpdf/content"
	contentimpl "gpdf/content/impl"
	"gpdf/model"
)

// ExtractImages returns all ImageInfo records across all pages (in page order).
func ExtractImages(src contentSource) ([]ImageInfo, error) {
	perPage, err := ExtractImagesPerPage(src)
	if err != nil {
		return nil, err
	}
	var all []ImageInfo
	for _, imgs := range perPage {
		all = append(all, imgs...)
	}
	return all, nil
}

// ExtractImagesPerPage returns a slice (one entry per page) of image lists.
// Pages with no images have an empty (nil) inner slice.
func ExtractImagesPerPage(src contentSource) ([][]ImageInfo, error) {
	pages, err := src.Pages()
	if err != nil {
		return nil, err
	}
	parser := contentimpl.NewStreamParser()
	result := make([][]ImageInfo, len(pages))
	for i, page := range pages {
		result[i] = extractImagesFromPage(src, parser, page, i, 0, 0, 0)
	}
	return result, nil
}

func extractImagesFromPage(src contentSource, parser content.Parser, page model.Page, pageIdx int, maxDecodedBytes int, maxOps int, maxImageBytes int) []ImageInfo {
	resources, ok := page.Resources()
	if !ok {
		return nil
	}
	ops, err := pageContentOps(src, parser, page, maxDecodedBytes, maxOps)
	if err != nil || len(ops) == 0 {
		return fallbackImagesFromResources(src, page, pageIdx, maxImageBytes)
	}
	images := extractImagesFromOps(ops, src, parser, resources, pageIdx, identityMatrix(), make(map[model.Ref]struct{}, 4), maxImageBytes)
	if len(images) == 0 {
		return fallbackImagesFromResources(src, page, pageIdx, maxImageBytes)
	}
	return images
}

func imageInfoFromStream(s *model.Stream, name string, pageIdx int, maxImageBytes int) (ImageInfo, bool) {
	if maxImageBytes > 0 && len(s.Content) > maxImageBytes {
		return ImageInfo{}, false
	}
	info := ImageInfo{
		Name: name,
		Page: pageIdx,
	}

	if w, ok := s.Dict[model.Name("Width")].(model.Integer); ok {
		info.Width = int(w)
	}
	if h, ok := s.Dict[model.Name("Height")].(model.Integer); ok {
		info.Height = int(h)
	}
	if bpc, ok := s.Dict[model.Name("BitsPerComponent")].(model.Integer); ok {
		info.BitsPerComponent = int(bpc)
	}

	// ColorSpace: may be a Name or an Array
	switch cs := s.Dict[model.Name("ColorSpace")].(type) {
	case model.Name:
		info.ColorSpace = string(cs)
	case model.Array:
		if len(cs) > 0 {
			if n, ok := cs[0].(model.Name); ok {
				info.ColorSpace = string(n)
			}
		}
	}

	// Filter: may be a Name or an Array of Names
	switch f := s.Dict[model.Name("Filter")].(type) {
	case model.Name:
		info.Filter = string(f)
	case model.Array:
		filters := make([]string, 0, len(f))
		for _, fv := range f {
			if fn, ok := fv.(model.Name); ok {
				filters = append(filters, string(fn))
			}
		}
		if len(filters) == 1 {
			info.Filter = filters[0]
		} else if len(filters) > 1 {
			info.Filter = fmt.Sprintf("%v", filters)
		}
	}
	info.Format = detectImageFormat(info.Filter, s.Content)

	info.Data = s.Content
	return info, true
}

func fallbackImagesFromResources(src contentSource, page model.Page, pageIdx int, maxImageBytes int) []ImageInfo {
	xObjects, ok := page.XObjects()
	if !ok {
		res, ok2 := page.Resources()
		if !ok2 {
			return nil
		}
		xoObj, exists := res[model.Name("XObject")]
		if !exists {
			return nil
		}
		resolved, err := resolveXObjectDict(src, xoObj)
		if err != nil || resolved == nil {
			return nil
		}
		xObjects = resolved
	}
	var images []ImageInfo
	for name, obj := range xObjects {
		stream, _, ok := resolveStreamObject(src, obj)
		if !ok || stream == nil {
			continue
		}
		subtype, _ := stream.Dict[model.Name("Subtype")].(model.Name)
		if subtype != "Image" {
			continue
		}
		if image, ok := imageInfoFromStream(stream, string(name), pageIdx, maxImageBytes); ok {
			images = append(images, image)
		}
	}
	return images
}

type imageState struct {
	ctm   matrix
	stack []matrix
}

func newImageState() *imageState {
	return &imageState{ctm: identityMatrix()}
}

func extractImagesFromOps(
	ops []content.Op,
	src contentSource,
	parser content.Parser,
	resources model.Dict,
	pageIdx int,
	baseCTM matrix,
	visited map[model.Ref]struct{},
	maxImageBytes int,
) []ImageInfo {
	state := newImageState()
	state.ctm = baseCTM
	var images []ImageInfo
	for _, op := range ops {
		switch op.Name {
		case "q":
			state.stack = append(state.stack, state.ctm)
		case "Q":
			if len(state.stack) == 0 {
				continue
			}
			state.ctm = state.stack[len(state.stack)-1]
			state.stack = state.stack[:len(state.stack)-1]
		case "cm":
			if len(op.Args) >= 6 {
				state.ctm = state.ctm.multiply(matrixFromObjects(op.Args))
			}
		case "Do":
			images = append(images, extractImagesFromXObject(op.Args, src, parser, resources, pageIdx, state.ctm, visited, maxImageBytes)...)
		}
	}
	return images
}

func extractImagesFromXObject(
	args model.Array,
	src contentSource,
	parser content.Parser,
	resources model.Dict,
	pageIdx int,
	ctm matrix,
	visited map[model.Ref]struct{},
	maxImageBytes int,
) []ImageInfo {
	if len(args) == 0 || resources == nil {
		return nil
	}
	name, ok := args[0].(model.Name)
	if !ok {
		return nil
	}
	xObjects, ok := resolveDictObject(src, resources[model.Name("XObject")])
	if !ok {
		return nil
	}
	xObject, ok := xObjects[name]
	if !ok {
		return nil
	}
	stream, ref, ok := resolveStreamObject(src, xObject)
	if !ok || stream == nil {
		return nil
	}
	subtype, _ := stream.Dict[model.Name("Subtype")].(model.Name)
	switch subtype {
	case "Image":
		image, ok := imageInfoFromStream(stream, string(name), pageIdx, maxImageBytes)
		if !ok {
			return nil
		}
		applyImagePlacement(&image, ctm)
		return []ImageInfo{image}
	case "Form":
		if ref != nil {
			if _, seen := visited[*ref]; seen {
				return nil
			}
			visited[*ref] = struct{}{}
			defer delete(visited, *ref)
		}
		nestedOps, err := parser.Parse(stream.Content)
		if err != nil {
			return nil
		}
		nestedResources := resources
		if formResources, ok := resolveDictObject(src, stream.Dict[model.Name("Resources")]); ok {
			nestedResources = mergeResourceDict(resources, formResources)
		}
		formMatrix := identityMatrix()
		if formArray, ok := stream.Dict[model.Name("Matrix")].(model.Array); ok {
			formMatrix = matrixFromObjects(formArray)
		}
		return extractImagesFromOps(nestedOps, src, parser, nestedResources, pageIdx, ctm.multiply(formMatrix), visited, maxImageBytes)
	}
	return nil
}

func matrixFromObjects(args []model.Object) matrix {
	values := make([]float64, 0, len(args))
	for _, arg := range args {
		values = append(values, toFloat64(arg))
	}
	return matrixFromArgs(values)
}

func applyImagePlacement(image *ImageInfo, ctm matrix) {
	image.X, image.Y = ctm.apply(0, 0)
	image.WidthPt = math.Hypot(ctm.a, ctm.b)
	image.HeightPt = math.Hypot(ctm.c, ctm.d)
	image.Rotation = ctm.rotationDegrees()
	if image.WidthPt == 0 {
		image.WidthPt = float64(image.Width)
	}
	if image.HeightPt == 0 {
		image.HeightPt = float64(image.Height)
	}
}

func detectImageFormat(filter string, data []byte) string {
	switch filter {
	case "DCTDecode":
		return "jpeg"
	case "JPXDecode":
		return "jpeg2000"
	}
	if len(data) >= 8 && string(data[:8]) == "\x89PNG\r\n\x1a\n" {
		return "png"
	}
	return ""
}

// resolveXObjectDict resolves an object to a model.Dict (handles Ref → Dict).
func resolveXObjectDict(src contentSource, obj model.Object) (model.Dict, error) {
	switch v := obj.(type) {
	case model.Dict:
		return v, nil
	case model.Ref:
		resolved, err := src.Resolve(v)
		if err != nil {
			return nil, err
		}
		if d, ok := resolved.(model.Dict); ok {
			return d, nil
		}
		if s, ok := resolved.(*model.Stream); ok && s != nil {
			return s.Dict, nil
		}
		return nil, nil
	}
	return nil, nil
}
