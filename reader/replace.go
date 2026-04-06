package reader

import (
	"bytes"
	"slices"
	"strings"

	"github.com/gsoultan/gpdf/content"
	contentimpl "github.com/gsoultan/gpdf/content/impl"
	"github.com/gsoultan/gpdf/model"
)

// ReplaceContent replaces all occurrences of old with new in the document's content streams.
// Mutates the resolved stream objects in place. Use Save to persist.
func ReplaceContent(src contentSource, old, new string) error {
	if old == "" || old == new {
		return nil
	}
	return ReplacesContent(src, map[string]string{old: new})
}

// ReplacesContent applies multiple replacements (old -> new) to the document's content streams.
// Replacements are applied in map iteration order; overlapping keys may have unexpected results.
func ReplacesContent(src contentSource, replacements map[string]string) error {
	pairs, needles := buildReplacementPlan(replacements)
	if len(pairs) == 0 {
		return nil
	}
	replacer := strings.NewReplacer(pairs...)

	pages, err := src.Pages()
	if err != nil {
		return err
	}
	parser := contentimpl.NewStreamParser()

	for _, page := range pages {
		contentRefs, err := pageContentRefs(page)
		if err != nil || len(contentRefs) == 0 {
			continue
		}
		for _, ref := range contentRefs {
			obj, err := src.Resolve(ref)
			if err != nil {
				continue
			}
			s, ok := obj.(*model.Stream)
			if !ok || s == nil {
				continue
			}
			if !containsAnyNeedle(s.Content, needles) {
				continue
			}
			ops, err := parser.Parse(s.Content)
			if err != nil {
				continue
			}
			modified := replaceInOps(ops, replacer)
			if !modified {
				continue
			}
			encoded, err := content.EncodeBytes(ops)
			if err != nil {
				return err
			}
			s.Content = encoded
		}
	}
	return nil
}

func buildReplacementPlan(replacements map[string]string) ([]string, [][]byte) {
	if len(replacements) == 0 {
		return nil, nil
	}

	keys := make([]string, 0, len(replacements))
	for old := range replacements {
		if old == "" {
			continue
		}
		keys = append(keys, old)
	}
	if len(keys) == 0 {
		return nil, nil
	}

	slices.Sort(keys)
	pairs := make([]string, 0, len(keys)*2)
	needles := make([][]byte, 0, len(keys))
	for _, old := range keys {
		pairs = append(pairs, old, replacements[old])
		needles = append(needles, []byte(old))
	}
	return pairs, needles
}

func containsAnyNeedle(content []byte, needles [][]byte) bool {
	for _, needle := range needles {
		if bytes.Contains(content, needle) {
			return true
		}
	}
	return false
}

func pageContentRefs(page model.Page) ([]model.Ref, error) {
	contentsObj := page.Contents()
	if contentsObj == nil {
		return nil, nil
	}
	switch v := contentsObj.(type) {
	case model.Ref:
		return []model.Ref{v}, nil
	case model.Array:
		refs := make([]model.Ref, 0, len(v))
		for _, item := range v {
			ref, ok := item.(model.Ref)
			if !ok {
				continue
			}
			refs = append(refs, ref)
		}
		return refs, nil
	}
	return nil, nil
}

func replaceInOps(ops []content.Op, replacer *strings.Replacer) bool {
	modified := false
	for i := range ops {
		switch ops[i].Name {
		case "Tj":
			if len(ops[i].Args) == 0 {
				continue
			}
			s, ok := ops[i].Args[0].(model.String)
			if !ok {
				continue
			}
			replaced := replacer.Replace(string(s))
			if replaced != string(s) {
				ops[i].Args[0] = model.String(replaced)
				modified = true
			}
		case "TJ":
			if len(ops[i].Args) == 0 {
				continue
			}
			arr, ok := ops[i].Args[0].(model.Array)
			if !ok {
				continue
			}
			arrModified := false
			for j, elem := range arr {
				s, ok := elem.(model.String)
				if !ok {
					continue
				}
				replaced := replacer.Replace(string(s))
				if replaced != string(s) {
					arr[j] = model.String(replaced)
					modified = true
					arrModified = true
				}
			}
			if arrModified {
				ops[i].Args[0] = arr
			}
		}
	}
	return modified
}
