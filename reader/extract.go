package reader

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"maps"
	"slices"
	"strconv"
	"strings"
	"unicode/utf16"
	"unicode/utf8"

	"gpdf/content"
	contentimpl "gpdf/content/impl"
	"gpdf/model"
)

// contentSource is the minimal interface needed to extract text from a PDF document.
type contentSource interface {
	Pages() ([]model.Page, error)
	Resolve(ref model.Ref) (model.Object, error)
}

// ExtractText returns all text from a document-like source (Pages + Resolve).
func ExtractText(src contentSource) (string, error) {
	perPage, err := ExtractTextPerPage(src)
	if err != nil {
		return "", err
	}
	trimmed := make([]string, len(perPage))
	totalLen := 0
	nonEmpty := 0
	for i, text := range perPage {
		t := strings.TrimSpace(text)
		trimmed[i] = t
		if t == "" {
			continue
		}
		totalLen += len(t)
		nonEmpty++
	}
	if nonEmpty == 0 {
		return "", nil
	}

	var out strings.Builder
	out.Grow(totalLen + max(nonEmpty-1, 0))
	first := true
	for _, text := range trimmed {
		if text == "" {
			continue
		}
		if !first {
			out.WriteByte(' ')
		}
		first = false
		out.WriteString(text)
	}
	return out.String(), nil
}

// ExtractTextPerPage returns text for each page in order. Empty string for pages with no extractable text.
func ExtractTextPerPage(src contentSource) ([]string, error) {
	pages, err := src.Pages()
	if err != nil {
		return nil, err
	}
	out := make([]string, len(pages))
	parser := contentimpl.NewStreamParser()
	for i, page := range pages {
		ops, err := pageContentOps(src, parser, page, 0, 0)
		if err != nil || len(ops) == 0 {
			continue
		}
		var sb strings.Builder
		sb.Grow(len(ops) * 4)
		resources, _ := page.Resources()
		extractTextFromOps(
			&sb,
			ops,
			src,
			parser,
			resources,
			make(map[model.Ref]struct{}, 4),
			make(map[model.Ref]*toUnicodeDecoder, 4),
		)
		out[i] = strings.TrimSpace(sb.String())
	}
	return out, nil
}

func pageContentBytes(src contentSource, page model.Page) ([]byte, error) {
	contentsObj := page.Contents()
	if contentsObj == nil {
		return nil, nil
	}
	switch v := contentsObj.(type) {
	case model.Ref, *model.Stream, model.Stream:
		s, _, ok := resolveStreamObject(src, v)
		if !ok || s == nil || len(s.Content) == 0 {
			return nil, nil
		}
		return s.Content, nil
	case model.Array:
		parts := make([][]byte, 0, len(v))
		total := 0
		for _, item := range v {
			s, _, ok := resolveStreamObject(src, item)
			if !ok || s == nil || len(s.Content) == 0 {
				continue
			}
			parts = append(parts, s.Content)
			total += len(s.Content)
		}
		if len(parts) == 0 {
			return nil, nil
		}
		if len(parts) == 1 {
			return parts[0], nil
		}
		raw := make([]byte, 0, total+len(parts)-1)
		for i, part := range parts {
			if i > 0 {
				raw = append(raw, '\n')
			}
			raw = append(raw, part...)
		}
		return raw, nil
	}
	return nil, nil
}

func pageContentOps(src contentSource, parser content.Parser, page model.Page, maxDecodedBytes int, maxOps int) ([]content.Op, error) {
	contentsObj := page.Contents()
	if contentsObj == nil {
		return nil, nil
	}
	ctx := pageOpsContext{src: src, parser: parser, maxDecodedBytes: maxDecodedBytes, maxOps: maxOps}
	return ctx.parse(contentsObj)
}

type pageOpsContext struct {
	src             contentSource
	parser          content.Parser
	maxDecodedBytes int
	maxOps          int
	decodedBytes    int
	opsCount        int
}

func (ctx *pageOpsContext) parse(obj model.Object) ([]content.Op, error) {
	switch value := obj.(type) {
	case model.Ref, *model.Stream, model.Stream:
		streamObj, _, ok := resolveStreamObject(ctx.src, value)
		if !ok || streamObj == nil || len(streamObj.Content) == 0 {
			return nil, nil
		}
		if ctx.maxDecodedBytes > 0 {
			ctx.decodedBytes += len(streamObj.Content)
			if ctx.decodedBytes > ctx.maxDecodedBytes {
				return nil, fmt.Errorf("decoded page content exceeds limit (%d bytes)", ctx.maxDecodedBytes)
			}
		}
		ops, err := ctx.parser.Parse(streamObj.Content)
		if err != nil {
			return nil, err
		}
		if ctx.maxOps > 0 {
			ctx.opsCount += len(ops)
			if ctx.opsCount > ctx.maxOps {
				return nil, fmt.Errorf("page operation count exceeds limit (%d)", ctx.maxOps)
			}
		}
		return ops, nil
	case model.Array:
		all := make([]content.Op, 0, len(value)*32)
		for _, item := range value {
			ops, err := ctx.parse(item)
			if err != nil {
				return nil, err
			}
			all = append(all, ops...)
		}
		return all, nil
	default:
		return nil, nil
	}
}

// SearchPages finds keywords in per-page text and returns SearchResults.
// Indices maps page index to byte offsets where the keyword starts on that page.
func SearchPages(perPage []string, keywords ...string) []model.SearchResult {
	results := make([]model.SearchResult, len(keywords))
	for i, kw := range keywords {
		results[i] = model.SearchResult{Keyword: kw, Indices: make(map[int][]int)}
		if kw == "" {
			continue
		}
		for pageIdx, text := range perPage {
			if !strings.Contains(text, kw) {
				continue
			}
			indices := make([]int, 0, strings.Count(text, kw))
			pos := 0
			for {
				idx := strings.Index(text[pos:], kw)
				if idx < 0 {
					break
				}
				indices = append(indices, pos+idx)
				pos += idx + len(kw)
			}
			if len(indices) > 0 {
				results[i].Pages = append(results[i].Pages, pageIdx)
				results[i].Indices[pageIdx] = indices
			}
		}
	}
	return results
}

func extractTextFromOps(
	out *strings.Builder,
	ops []content.Op,
	src contentSource,
	parser content.Parser,
	resources model.Dict,
	visited map[model.Ref]struct{},
	fontDecoders map[model.Ref]*toUnicodeDecoder,
) {
	inText := false
	var activeDecoder *toUnicodeDecoder
	var tmY float64
	tmSet := false
	hadText := false // tracks whether any text was emitted before this BT
	for _, op := range ops {
		switch op.Name {
		case "BT":
			inText = true
			tmSet = false
			if hadText {
				appendBoundary(out, '\n')
			}
		case "ET":
			inText = false
		case "Tf":
			if len(op.Args) > 0 {
				activeDecoder = resolveFontDecoder(src, resources, op.Args[0], fontDecoders)
			}
		case "Tj":
			if inText && len(op.Args) > 0 {
				appendTextArg(out, op.Args[0], activeDecoder)
				hadText = true
			}
		case "TJ":
			if inText && len(op.Args) > 0 {
				appendTextArg(out, op.Args[0], activeDecoder)
				hadText = true
			}
		case "T*":
			if inText {
				appendBoundary(out, '\n')
			}
		case "Td", "TD":
			// Td tx ty — move to next line offset by (tx, ty).
			// Insert newline only for vertical moves. Horizontal-only moves (ty=0) are
			// character-level advances; explicit space chars in strings provide word separation.
			if inText && len(op.Args) >= 2 {
				ty := toFloat64(op.Args[1])
				if ty < -0.5 || ty > 0.5 {
					appendBoundary(out, '\n')
				}
			}
		case "Tm":
			// Tm a b c d e f — sets text matrix; e=x, f=y in user space.
			// Insert newline when y changes significantly.
			// Same-y repositioning is character-level layout; explicit space chars in strings
			// provide word separation, so no space is inserted here.
			if inText && len(op.Args) >= 6 {
				f := toFloat64(op.Args[5])
				if tmSet && abs64(f-tmY) > 0.5 {
					appendBoundary(out, '\n')
				}
				tmY = f
				tmSet = true
			}
		case "'":
			if inText && len(op.Args) > 0 {
				appendBoundary(out, '\n')
				appendTextArg(out, op.Args[0], activeDecoder)
				hadText = true
			}
		case "\"":
			if inText && len(op.Args) > 0 {
				appendBoundary(out, '\n')
				appendTextArg(out, op.Args[len(op.Args)-1], activeDecoder)
				hadText = true
			}
		case "Do":
			extractTextFromXObject(out, op.Args, src, parser, resources, visited, fontDecoders)
		}
	}
}

func toFloat64(obj model.Object) float64 {
	switch v := obj.(type) {
	case model.Real:
		return float64(v)
	case model.Integer:
		return float64(v)
	}
	return 0
}

func abs64(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}

func appendTextArg(out *strings.Builder, arg model.Object, decoder *toUnicodeDecoder) {
	switch v := arg.(type) {
	case model.String:
		out.WriteString(decodeTextBytes([]byte(v), decoder))
	case model.Array:
		for _, elem := range v {
			switch item := elem.(type) {
			case model.String:
				out.WriteString(decodeTextBytes([]byte(item), decoder))
			case model.Integer:
				if item <= model.Integer(-120) {
					appendBoundary(out, ' ')
				}
			case model.Real:
				if item <= model.Real(-120) {
					appendBoundary(out, ' ')
				}
			}
		}
	}
}

func appendBoundary(out *strings.Builder, boundary byte) {
	if out.Len() == 0 {
		return
	}
	last := out.String()[out.Len()-1]
	switch boundary {
	case '\n':
		if last == '\n' {
			return
		}
	case ' ':
		if isWhitespaceByte(last) {
			return
		}
	}
	out.WriteByte(boundary)
}

func isWhitespaceByte(b byte) bool {
	return b == ' ' || b == '\n' || b == '\r' || b == '\t'
}

func extractTextFromXObject(
	out *strings.Builder,
	args model.Array,
	src contentSource,
	parser content.Parser,
	resources model.Dict,
	visited map[model.Ref]struct{},
	fontDecoders map[model.Ref]*toUnicodeDecoder,
) {
	if len(args) == 0 || resources == nil {
		return
	}
	name, ok := args[0].(model.Name)
	if !ok {
		return
	}
	xObjects, ok := resources[model.Name("XObject")].(model.Dict)
	if !ok {
		return
	}
	xObjectObj, ok := xObjects[name]
	if !ok {
		return
	}

	xObject, ref, ok := resolveStreamObject(src, xObjectObj)
	if !ok || len(xObject.Content) == 0 {
		return
	}
	if subtype, ok := xObject.Dict[model.Name("Subtype")].(model.Name); !ok || subtype != model.Name("Form") {
		return
	}

	if ref != nil {
		if _, seen := visited[*ref]; seen {
			return
		}
		visited[*ref] = struct{}{}
		defer delete(visited, *ref)
	}

	nestedOps, err := parser.Parse(xObject.Content)
	if err != nil {
		return
	}
	nestedResources := resources
	if r, ok := resolveDictObject(src, xObject.Dict[model.Name("Resources")]); ok {
		nestedResources = mergeResourceDict(resources, r)
	}

	extractTextFromOps(out, nestedOps, src, parser, nestedResources, visited, fontDecoders)
}

type toUnicodeDecoder struct {
	mapping    map[string]string
	maxCodeLen int
}

func (d *toUnicodeDecoder) decode(data []byte) string {
	if len(data) == 0 {
		return ""
	}
	if d == nil || len(d.mapping) == 0 || d.maxCodeLen <= 0 {
		return decodeRawTextBytes(data)
	}

	var out strings.Builder
	out.Grow(len(data))
	for i := 0; i < len(data); {
		matched := false
		limit := min(d.maxCodeLen, len(data)-i)
		for codeLen := limit; codeLen >= 1; codeLen-- {
			if decoded, ok := d.mapping[string(data[i:i+codeLen])]; ok {
				out.WriteString(decoded)
				i += codeLen
				matched = true
				break
			}
		}
		if !matched {
			out.WriteRune(rune(data[i]))
			i++
		}
	}
	return out.String()
}

func decodeTextBytes(data []byte, decoder *toUnicodeDecoder) string {
	if decoder != nil {
		return decoder.decode(data)
	}
	return decodeRawTextBytes(data)
}

func decodeRawTextBytes(data []byte) string {
	if len(data) == 0 {
		return ""
	}
	if bytes.IndexByte(data, 0) >= 0 && len(data)%2 == 0 {
		return decodeUTF16BE(data)
	}
	if utf8.Valid(data) {
		return string(data)
	}
	return decodeWinAnsi(data)
}

func decodeWinAnsi(data []byte) string {
	var out strings.Builder
	out.Grow(len(data))
	for _, b := range data {
		switch b {
		case 0x80:
			out.WriteRune('\u20ac')
		case 0x82:
			out.WriteRune('\u201a')
		case 0x83:
			out.WriteRune('\u0192')
		case 0x84:
			out.WriteRune('\u201e')
		case 0x85:
			out.WriteRune('\u2026')
		case 0x86:
			out.WriteRune('\u2020')
		case 0x87:
			out.WriteRune('\u2021')
		case 0x88:
			out.WriteRune('\u02c6')
		case 0x89:
			out.WriteRune('\u2030')
		case 0x8a:
			out.WriteRune('\u0160')
		case 0x8b:
			out.WriteRune('\u2039')
		case 0x8c:
			out.WriteRune('\u0152')
		case 0x8e:
			out.WriteRune('\u017d')
		case 0x91:
			out.WriteRune('\u2018')
		case 0x92:
			out.WriteRune('\u2019')
		case 0x93:
			out.WriteRune('\u201c')
		case 0x94:
			out.WriteRune('\u201d')
		case 0x95:
			out.WriteRune('\u2022')
		case 0x96:
			out.WriteRune('\u2013')
		case 0x97:
			out.WriteRune('\u2014')
		case 0x98:
			out.WriteRune('\u02dc')
		case 0x99:
			out.WriteRune('\u2122')
		case 0x9a:
			out.WriteRune('\u0161')
		case 0x9b:
			out.WriteRune('\u203a')
		case 0x9c:
			out.WriteRune('\u0153')
		case 0x9e:
			out.WriteRune('\u017e')
		case 0x9f:
			out.WriteRune('\u0178')
		default:
			out.WriteRune(rune(b))
		}
	}
	return out.String()
}

func resolveFontName(src contentSource, resources model.Dict, fontOperand model.Object) string {
	fontName, ok := fontOperand.(model.Name)
	if !ok || resources == nil {
		return ""
	}

	fontResourceObj, ok := resources[model.Name("Font")]
	if !ok {
		return ""
	}
	fontResources, ok := resolveDictObject(src, fontResourceObj)
	if !ok {
		return ""
	}
	fontObj, ok := fontResources[fontName]
	if !ok {
		return ""
	}

	fontDict, _, ok := resolveDictWithRef(src, fontObj)
	if !ok {
		return ""
	}

	baseFont, ok := fontDict[model.Name("BaseFont")].(model.Name)
	if ok {
		name := string(baseFont)
		if strings.Contains(name, "+") {
			// Subset font name like ABCD+Helvetica-Bold
			parts := strings.Split(name, "+")
			return parts[len(parts)-1]
		}
		return name
	}
	return ""
}

func resolveFontDecoder(
	src contentSource,
	resources model.Dict,
	fontOperand model.Object,
	cache map[model.Ref]*toUnicodeDecoder,
) *toUnicodeDecoder {
	fontName, ok := fontOperand.(model.Name)
	if !ok || resources == nil {
		return nil
	}

	fontResourceObj, ok := resources[model.Name("Font")]
	if !ok {
		return nil
	}
	fontResources, ok := resolveDictObject(src, fontResourceObj)
	if !ok {
		return nil
	}
	fontObj, ok := fontResources[fontName]
	if !ok {
		return nil
	}

	fontDict, fontRef, ok := resolveDictWithRef(src, fontObj)
	if !ok {
		return nil
	}
	if fontRef != nil {
		if decoder, exists := cache[*fontRef]; exists {
			return decoder
		}
	}

	toUnicodeObj, ok := fontDict[model.Name("ToUnicode")]
	if !ok {
		if fontRef != nil {
			cache[*fontRef] = nil
		}
		return nil
	}

	toUnicodeStream, _, ok := resolveStreamObject(src, toUnicodeObj)
	if !ok || toUnicodeStream == nil {
		if fontRef != nil {
			cache[*fontRef] = nil
		}
		return nil
	}

	decoder := parseToUnicodeDecoder(toUnicodeStream.Content)
	if fontRef != nil {
		cache[*fontRef] = decoder
	}
	return decoder
}

func parseToUnicodeDecoder(content []byte) *toUnicodeDecoder {
	if len(content) == 0 {
		return nil
	}

	decoder := &toUnicodeDecoder{mapping: make(map[string]string)}
	mode := ""
	remaining := 0

	for line := range strings.SplitSeq(string(content), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		if n, ok := parseCMapBeginCount(line, "beginbfchar"); ok {
			mode = "bfchar"
			remaining = n
			continue
		}
		if n, ok := parseCMapBeginCount(line, "beginbfrange"); ok {
			mode = "bfrange"
			remaining = n
			continue
		}
		if line == "endbfchar" || line == "endbfrange" {
			mode = ""
			remaining = 0
			continue
		}
		if remaining <= 0 {
			continue
		}

		switch mode {
		case "bfchar":
			parseBFCharLine(decoder, line)
			remaining--
		case "bfrange":
			parseBFRangeLine(decoder, line)
			remaining--
		}
	}

	if len(decoder.mapping) == 0 {
		return nil
	}
	return decoder
}

func parseCMapBeginCount(line string, marker string) (int, bool) {
	if !strings.HasSuffix(line, marker) {
		return 0, false
	}
	fields := slices.Collect(strings.FieldsSeq(line))
	if len(fields) == 0 {
		return 0, false
	}
	n, err := strconv.Atoi(fields[0])
	if err != nil || n <= 0 {
		return 0, false
	}
	return n, true
}

func parseBFCharLine(decoder *toUnicodeDecoder, line string) {
	codes := parseHexCodes(line)
	if len(codes) < 2 {
		return
	}
	decoder.mapping[string(codes[0])] = decodeCMapString(codes[1])
	decoder.maxCodeLen = max(decoder.maxCodeLen, len(codes[0]))
}

func parseBFRangeLine(decoder *toUnicodeDecoder, line string) {
	codes := parseHexCodes(line)
	if len(codes) < 3 || len(codes[0]) == 0 || len(codes[0]) != len(codes[1]) {
		return
	}

	start, ok := bytesToInt(codes[0])
	if !ok {
		return
	}
	end, ok := bytesToInt(codes[1])
	if !ok || end < start {
		return
	}
	count := end - start + 1
	codeLen := len(codes[0])
	decoder.maxCodeLen = max(decoder.maxCodeLen, codeLen)

	if strings.Contains(line, "[") {
		available := len(codes) - 2
		for i := range min(count, available) {
			src := intToBytes(start+i, codeLen)
			decoder.mapping[string(src)] = decodeCMapString(codes[2+i])
		}
		return
	}

	baseRunes := []rune(decodeCMapString(codes[2]))
	if len(baseRunes) == 0 {
		return
	}
	base := baseRunes[0]
	for i := range count {
		src := intToBytes(start+i, codeLen)
		decoder.mapping[string(src)] = string(base + rune(i))
	}
}

func parseHexCodes(line string) [][]byte {
	out := make([][]byte, 0, 4)
	for {
		start := strings.IndexByte(line, '<')
		if start < 0 {
			break
		}
		line = line[start+1:]
		end := strings.IndexByte(line, '>')
		if end < 0 {
			break
		}
		if decoded, ok := decodeHexString(line[:end]); ok {
			out = append(out, decoded)
		}
		line = line[end+1:]
	}
	return out
}

func decodeHexString(hexText string) ([]byte, bool) {
	hexText = strings.TrimSpace(hexText)
	if hexText == "" {
		return nil, false
	}
	if len(hexText)%2 != 0 {
		hexText += "0"
	}
	decoded := make([]byte, hex.DecodedLen(len(hexText)))
	if _, err := hex.Decode(decoded, []byte(hexText)); err != nil {
		return nil, false
	}
	return decoded, true
}

func decodeCMapString(data []byte) string {
	if len(data) == 0 {
		return ""
	}
	if len(data)%2 == 0 {
		return decodeUTF16BE(data)
	}
	return decodeRawTextBytes(data)
}

func decodeUTF16BE(data []byte) string {
	if len(data) == 0 || len(data)%2 != 0 {
		return ""
	}
	units := make([]uint16, len(data)/2)
	for i := range len(units) {
		units[i] = uint16(data[2*i])<<8 | uint16(data[2*i+1])
	}
	return string(utf16.Decode(units))
}

func bytesToInt(data []byte) (int, bool) {
	if len(data) == 0 || len(data) > 4 {
		return 0, false
	}
	v := 0
	for _, b := range data {
		v = (v << 8) | int(b)
	}
	return v, true
}

func intToBytes(v int, width int) []byte {
	if width <= 0 {
		return nil
	}
	out := make([]byte, width)
	for i := width - 1; i >= 0; i-- {
		out[i] = byte(v & 0xff)
		v >>= 8
	}
	return out
}

func resolveDictObject(src contentSource, obj model.Object) (model.Dict, bool) {
	switch v := obj.(type) {
	case model.Dict:
		return v, true
	case model.Ref:
		resolved, err := src.Resolve(v)
		if err != nil {
			return nil, false
		}
		switch d := resolved.(type) {
		case model.Dict:
			return d, true
		case *model.Stream:
			if d == nil {
				return nil, false
			}
			return d.Dict, true
		case model.Stream:
			return d.Dict, true
		}
	case *model.Stream:
		if v == nil {
			return nil, false
		}
		return v.Dict, true
	case model.Stream:
		return v.Dict, true
	}
	return nil, false
}

func resolveDictWithRef(src contentSource, obj model.Object) (model.Dict, *model.Ref, bool) {
	switch v := obj.(type) {
	case model.Ref:
		resolved, err := src.Resolve(v)
		if err != nil {
			return nil, nil, false
		}
		dict, ok := resolveDictObject(src, resolved)
		if !ok {
			return nil, nil, false
		}
		return dict, &v, true
	default:
		dict, ok := resolveDictObject(src, v)
		if !ok {
			return nil, nil, false
		}
		return dict, nil, true
	}
}

func mergeResourceDict(parent model.Dict, child model.Dict) model.Dict {
	if parent == nil {
		return child
	}
	if child == nil {
		return parent
	}

	merged := maps.Clone(parent)
	for key, childValue := range child {
		if parentValue, exists := merged[key]; exists {
			parentDict, parentIsDict := parentValue.(model.Dict)
			childDict, childIsDict := childValue.(model.Dict)
			if parentIsDict && childIsDict {
				nested := maps.Clone(parentDict)
				maps.Copy(nested, childDict)
				merged[key] = nested
				continue
			}
		}
		merged[key] = childValue
	}
	return merged
}

func resolveStreamObject(src contentSource, obj model.Object) (*model.Stream, *model.Ref, bool) {
	switch v := obj.(type) {
	case model.Ref:
		resolved, err := src.Resolve(v)
		if err != nil {
			return nil, nil, false
		}
		s, ok := resolved.(*model.Stream)
		if !ok || s == nil {
			return nil, nil, false
		}
		return s, &v, true
	case *model.Stream:
		if v == nil {
			return nil, nil, false
		}
		return v, nil, true
	case model.Stream:
		s := v
		return &s, nil, true
	}
	return nil, nil, false
}
