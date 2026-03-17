package reader

import (
	"strings"

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
		raw, err := pageContentBytes(src, page)
		if err != nil || len(raw) == 0 {
			continue
		}
		ops, err := parser.Parse(raw)
		if err != nil {
			continue
		}
		var sb strings.Builder
		sb.Grow(len(raw) / 2)
		extractTextFromOps(&sb, ops)
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
	case model.Ref:
		obj, err := src.Resolve(v)
		if err != nil {
			return nil, err
		}
		s, ok := obj.(*model.Stream)
		if !ok || s == nil {
			return nil, nil
		}
		return s.Content, nil
	case model.Array:
		parts := make([][]byte, 0, len(v))
		total := 0
		for _, item := range v {
			ref, ok := item.(model.Ref)
			if !ok {
				continue
			}
			obj, err := src.Resolve(ref)
			if err != nil {
				continue
			}
			s, ok := obj.(*model.Stream)
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

func extractTextFromOps(out *strings.Builder, ops []content.Op) {
	inText := false
	for _, op := range ops {
		switch op.Name {
		case "BT":
			inText = true
		case "ET":
			inText = false
		case "Tj":
			if inText && len(op.Args) > 0 {
				if s, ok := op.Args[0].(model.String); ok {
					out.WriteString(string(s))
				}
			}
		case "TJ":
			if inText && len(op.Args) > 0 {
				if arr, ok := op.Args[0].(model.Array); ok {
					for _, elem := range arr {
						if s, ok := elem.(model.String); ok {
							out.WriteString(string(s))
						}
					}
				}
			}
		}
	}
}
