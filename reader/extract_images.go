package reader

import (
	"bytes"
	"compress/zlib"
	"fmt"
	"io"
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
	// Flag CMYK images: consumer must convert CMYK→RGB to avoid color shift
	if info.ColorSpace == "DeviceCMYK" {
		info.NeedsColorConvert = true
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
	// Soft mask support: fully decode the SMask stream including FlateDecode +
	// PNG Predictor so that SMaskData always contains raw pixel bytes.
	if sm, ok := s.Dict[model.Name("SMask")]; ok {
		if smStream, ok3 := sm.(*model.Stream); ok3 {
			info.HasSMask = true
			decoded, wasDecoded := decodeSMaskFully(smStream)
			info.SMaskData = decoded
			info.SMaskDecoded = wasDecoded
			if w, ok := smStream.Dict[model.Name("Width")].(model.Integer); ok {
				info.SMaskWidth = int(w)
			}
			if h, ok := smStream.Dict[model.Name("Height")].(model.Integer); ok {
				info.SMaskHeight = int(h)
			}
		}
	}
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
	ctm     matrix
	opacity float64
	stack   []imageState
}

func newImageState() *imageState {
	return &imageState{ctm: identityMatrix(), opacity: 1.0}
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
			state.stack = append(state.stack, *state)
		case "Q":
			if len(state.stack) == 0 {
				continue
			}
			saved := state.stack[len(state.stack)-1]
			state.ctm = saved.ctm
			state.opacity = saved.opacity
			state.stack = state.stack[:len(state.stack)-1]
		case "cm":
			if len(op.Args) >= 6 {
				state.ctm = matrixFromObjects(op.Args).multiply(state.ctm)
			}
		case "gs":
			if len(op.Args) > 0 {
				if name, ok := op.Args[0].(model.Name); ok {
					if extGState, ok := resolveDictObject(src, resources[model.Name("ExtGState")]); ok {
						if gs, ok := resolveDictObject(src, extGState[name]); ok {
							if ca, ok := gs[model.Name("ca")].(model.Real); ok {
								state.opacity = float64(ca)
							} else if ca, ok := gs[model.Name("ca")].(model.Integer); ok {
								state.opacity = float64(ca)
							}
							if CA, ok := gs[model.Name("CA")].(model.Real); ok {
								state.opacity = float64(CA)
							} else if CA, ok := gs[model.Name("CA")].(model.Integer); ok {
								state.opacity = float64(CA)
							}
						}
					}
				}
			}
		case "Do":
			extracted := extractImagesFromXObject(op.Args, src, parser, resources, pageIdx, state.ctm, visited, maxImageBytes)
			for i := range extracted {
				extracted[i].Opacity = state.opacity
			}
			images = append(images, extracted...)
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
		// Resolve soft mask if present and indirect
		if !image.HasSMask {
			if sm, ok := stream.Dict[model.Name("SMask")]; ok {
				if smStream, _, ok2 := resolveStreamObject(src, sm); ok2 && smStream != nil {
					image.HasSMask = true
					image.SMaskData = smStream.Content
					if w, ok := smStream.Dict[model.Name("Width")].(model.Integer); ok {
						image.SMaskWidth = int(w)
					}
					if h, ok := smStream.Dict[model.Name("Height")].(model.Integer); ok {
						image.SMaskHeight = int(h)
					}
				}
			}
		}
		// Preserve exact placement matrix
		image.Matrix = [6]float64{ctm.a, ctm.b, ctm.c, ctm.d, ctm.e, ctm.f}
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

// decodeSMaskFully decodes an SMask stream by:
//  1. Decompressing FlateDecode (zlib) if present.
//  2. Un-filtering PNG Predictor rows (Predictor >= 10) if specified in DecodeParms.
//
// Returns the decoded bytes and wasDecoded=true when any processing was applied.
// Falls back to raw smStream.Content on any error.
func decodeSMaskFully(smStream *model.Stream) (data []byte, wasDecoded bool) {
	data = smStream.Content
	if len(data) == 0 {
		return data, false
	}

	// Resolve filter name
	var filterName string
	switch f := smStream.Dict[model.Name("Filter")].(type) {
	case model.Name:
		filterName = string(f)
	case model.Array:
		if len(f) > 0 {
			if n, ok := f[0].(model.Name); ok {
				filterName = string(n)
			}
		}
	}

	if filterName != "FlateDecode" {
		return data, false
	}

	// Decompress with zlib
	zr, err := zlib.NewReader(bytes.NewReader(data))
	if err != nil {
		return data, false
	}
	decompressed, err := io.ReadAll(zr)
	zr.Close()
	if err != nil {
		return data, false
	}
	wasDecoded = true
	data = decompressed

	// Check for PNG Predictor in DecodeParms
	var predictor, colors, bpc, columns int
	colors = 1
	bpc = 8
	var dpDict model.Dict
	switch dp := smStream.Dict[model.Name("DecodeParms")].(type) {
	case model.Dict:
		dpDict = dp
	case model.Array:
		if len(dp) > 0 {
			if d, ok := dp[0].(model.Dict); ok {
				dpDict = d
			}
		}
	}
	if dpDict != nil {
		if p, ok := dpDict[model.Name("Predictor")].(model.Integer); ok {
			predictor = int(p)
		}
		if c, ok := dpDict[model.Name("Colors")].(model.Integer); ok {
			colors = int(c)
		}
		if b, ok := dpDict[model.Name("BitsPerComponent")].(model.Integer); ok {
			bpc = int(b)
		}
		if col, ok := dpDict[model.Name("Columns")].(model.Integer); ok {
			columns = int(col)
		}
	}

	if predictor < 10 {
		return data, wasDecoded
	}

	// Fall back to stream Width when Columns not in DecodeParms
	if columns == 0 {
		if w, ok := smStream.Dict[model.Name("Width")].(model.Integer); ok {
			columns = int(w)
		}
	}
	if columns == 0 {
		return data, wasDecoded
	}

	// Apply PNG predictor un-filtering
	bytesPerPixel := (colors*bpc + 7) / 8
	rowSize := (columns*colors*bpc + 7) / 8
	rowStride := rowSize + 1 // +1 for per-row filter byte
	if len(data) < rowStride || rowStride == 0 {
		return data, wasDecoded
	}
	numRows := len(data) / rowStride
	output := make([]byte, numRows*rowSize)
	for row := range numRows {
		src := data[row*rowStride:]
		filterByte := src[0]
		rowData := src[1 : 1+rowSize]
		dst := output[row*rowSize:]
		var prevRow []byte
		if row > 0 {
			prevRow = output[(row-1)*rowSize:]
		}
		switch filterByte {
		case 0: // None
			copy(dst, rowData)
		case 1: // Sub
			for i := range rowSize {
				var a byte
				if i >= bytesPerPixel {
					a = dst[i-bytesPerPixel]
				}
				dst[i] = rowData[i] + a
			}
		case 2: // Up
			for i := range rowSize {
				var b byte
				if prevRow != nil {
					b = prevRow[i]
				}
				dst[i] = rowData[i] + b
			}
		case 3: // Average
			for i := range rowSize {
				var a, b byte
				if i >= bytesPerPixel {
					a = dst[i-bytesPerPixel]
				}
				if prevRow != nil {
					b = prevRow[i]
				}
				dst[i] = rowData[i] + byte((int(a)+int(b))/2)
			}
		case 4: // Paeth
			for i := range rowSize {
				var a, b, c byte
				if i >= bytesPerPixel {
					a = dst[i-bytesPerPixel]
				}
				if prevRow != nil {
					b = prevRow[i]
					if i >= bytesPerPixel {
						c = prevRow[i-bytesPerPixel]
					}
				}
				dst[i] = rowData[i] + paethPredictor(a, b, c)
			}
		default:
			copy(dst, rowData)
		}
	}
	return output, true
}

// paethPredictor implements the PNG Paeth predictor function (RFC 2083 §6.6).
func paethPredictor(a, b, c byte) byte {
	ia, ib, ic := int(a), int(b), int(c)
	p := ia + ib - ic
	pa := p - ia
	if pa < 0 {
		pa = -pa
	}
	pb := p - ib
	if pb < 0 {
		pb = -pb
	}
	pc := p - ic
	if pc < 0 {
		pc = -pc
	}
	if pa <= pb && pa <= pc {
		return a
	}
	if pb <= pc {
		return b
	}
	return c
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
