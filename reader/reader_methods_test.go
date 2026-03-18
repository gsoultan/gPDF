package reader

import (
	"strings"
	"testing"

	"gpdf/model"
)

// ── SearchPages ───────────────────────────────────────────────────────────────

func TestSearchPages_FindsKeywordOnCorrectPage(t *testing.T) {
	perPage := []string{"hello world", "foo bar", "hello again"}
	results := SearchPages(perPage, "hello")
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	r := results[0]
	if r.Keyword != "hello" {
		t.Errorf("keyword mismatch: got %q", r.Keyword)
	}
	if len(r.Pages) != 2 || r.Pages[0] != 0 || r.Pages[1] != 2 {
		t.Errorf("unexpected pages: %v", r.Pages)
	}
}

func TestSearchPages_MultipleKeywords(t *testing.T) {
	perPage := []string{"the quick brown fox", "lazy dog"}
	results := SearchPages(perPage, "quick", "dog", "missing")
	if len(results) != 3 {
		t.Fatalf("expected 3 results, got %d", len(results))
	}
	if len(results[0].Pages) != 1 || results[0].Pages[0] != 0 {
		t.Errorf("'quick' should be on page 0, got %v", results[0].Pages)
	}
	if len(results[1].Pages) != 1 || results[1].Pages[0] != 1 {
		t.Errorf("'dog' should be on page 1, got %v", results[1].Pages)
	}
	if len(results[2].Pages) != 0 {
		t.Errorf("'missing' should not be found, got %v", results[2].Pages)
	}
}

func TestSearchPages_EmptyKeyword(t *testing.T) {
	perPage := []string{"some text"}
	results := SearchPages(perPage, "")
	if len(results) != 1 {
		t.Fatalf("expected 1 result entry, got %d", len(results))
	}
	if len(results[0].Pages) != 0 {
		t.Errorf("empty keyword should return no pages, got %v", results[0].Pages)
	}
}

func TestSearchPages_RecordsIndices(t *testing.T) {
	perPage := []string{"ab cd ab"}
	results := SearchPages(perPage, "ab")
	if len(results[0].Indices[0]) != 2 {
		t.Errorf("expected 2 indices for 'ab', got %v", results[0].Indices[0])
	}
	if results[0].Indices[0][0] != 0 || results[0].Indices[0][1] != 6 {
		t.Errorf("unexpected indices: %v", results[0].Indices[0])
	}
}

// ── ReplaceContent / ReplacesContent ─────────────────────────────────────────

func makeReplaceStub(content string) stubContentSource {
	ref := model.Ref{ObjectNumber: 1}
	return stubContentSource{
		pages: []model.Page{{Dict: model.Dict{
			model.Name("Contents"): ref,
		}}},
		objects: map[model.Ref]model.Object{
			ref: &model.Stream{Content: []byte(content)},
		},
	}
}

func TestReplaceContent_ReplacesTextInStream(t *testing.T) {
	src := makeReplaceStub("BT (Hello World) Tj ET")
	if err := ReplaceContent(src, "Hello", "Hi"); err != nil {
		t.Fatalf("ReplaceContent returned error: %v", err)
	}
	text, err := ExtractText(src)
	if err != nil {
		t.Fatalf("ExtractText returned error: %v", err)
	}
	if !strings.Contains(text, "Hi World") {
		t.Errorf("expected replaced text %q, got %q", "Hi World", text)
	}
	if strings.Contains(text, "Hello") {
		t.Errorf("old text should be replaced, got %q", text)
	}
}

func TestReplaceContent_NoopWhenOldEqualsNew(t *testing.T) {
	original := "BT (Hello) Tj ET"
	src := makeReplaceStub(original)
	if err := ReplaceContent(src, "Hello", "Hello"); err != nil {
		t.Fatalf("ReplaceContent returned error: %v", err)
	}
	text, _ := ExtractText(src)
	if !strings.Contains(text, "Hello") {
		t.Errorf("text should be unchanged, got %q", text)
	}
}

func TestReplacesContent_AppliesMultipleReplacements(t *testing.T) {
	src := makeReplaceStub("BT (foo bar baz) Tj ET")
	err := ReplacesContent(src, map[string]string{
		"foo": "one",
		"bar": "two",
	})
	if err != nil {
		t.Fatalf("ReplacesContent returned error: %v", err)
	}
	text, _ := ExtractText(src)
	if !strings.Contains(text, "one") || !strings.Contains(text, "two") {
		t.Errorf("expected both replacements applied, got %q", text)
	}
}

// ── ExtractTextPerPage ────────────────────────────────────────────────────────

func TestExtractTextPerPage_ReturnsOneEntryPerPage(t *testing.T) {
	ref1 := model.Ref{ObjectNumber: 1}
	ref2 := model.Ref{ObjectNumber: 2}
	src := stubContentSource{
		pages: []model.Page{
			{Dict: model.Dict{model.Name("Contents"): ref1}},
			{Dict: model.Dict{model.Name("Contents"): ref2}},
		},
		objects: map[model.Ref]model.Object{
			ref1: &model.Stream{Content: []byte("BT (Page one) Tj ET")},
			ref2: &model.Stream{Content: []byte("BT (Page two) Tj ET")},
		},
	}
	perPage, err := ExtractTextPerPage(src)
	if err != nil {
		t.Fatalf("ExtractTextPerPage returned error: %v", err)
	}
	if len(perPage) != 2 {
		t.Fatalf("expected 2 pages, got %d", len(perPage))
	}
	if !strings.Contains(perPage[0], "Page one") {
		t.Errorf("page 0: expected 'Page one', got %q", perPage[0])
	}
	if !strings.Contains(perPage[1], "Page two") {
		t.Errorf("page 1: expected 'Page two', got %q", perPage[1])
	}
}

func TestExtractTextPerPage_EmptyPageYieldsEmptyString(t *testing.T) {
	ref := model.Ref{ObjectNumber: 1}
	src := stubContentSource{
		pages: []model.Page{
			{Dict: model.Dict{model.Name("Contents"): ref}},
			{Dict: model.Dict{}}, // no contents
		},
		objects: map[model.Ref]model.Object{
			ref: &model.Stream{Content: []byte("BT (Text) Tj ET")},
		},
	}
	perPage, err := ExtractTextPerPage(src)
	if err != nil {
		t.Fatalf("ExtractTextPerPage returned error: %v", err)
	}
	if len(perPage) != 2 {
		t.Fatalf("expected 2 pages, got %d", len(perPage))
	}
	if perPage[1] != "" {
		t.Errorf("empty page should return empty string, got %q", perPage[1])
	}
}
