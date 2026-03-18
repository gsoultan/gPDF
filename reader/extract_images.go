package reader

import (
	"fmt"

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
	result := make([][]ImageInfo, len(pages))
	for i, page := range pages {
		result[i] = extractImagesFromPage(src, page, i)
	}
	return result, nil
}

func extractImagesFromPage(src contentSource, page model.Page, pageIdx int) []ImageInfo {
	xObjects, ok := page.XObjects()
	if !ok {
		// XObjects might be behind a Ref; try resolving /Resources first.
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
		img := imageInfoFromStream(stream, string(name), pageIdx)
		images = append(images, img)
	}
	return images
}

func imageInfoFromStream(s *model.Stream, name string, pageIdx int) ImageInfo {
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

	info.Data = s.Content
	return info
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
