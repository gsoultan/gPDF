package doc_test

import (
	"bytes"
	"strings"
	"testing"

	"gpdf/doc"
)

func TestDrawTextBox_AllowsSimplePagination(t *testing.T) {
	buf := new(bytes.Buffer)
	builder := doc.New().
		Title("Paginate").
		PageSize(200, 120).
		AddPage().
		AddPage()

	// Enough text to require many lines.
	var textBuilder strings.Builder
	for i := 0; i < 40; i++ {
		textBuilder.WriteString("line ")
	}
	text := textBuilder.String()

	opts := doc.TextLayoutOptions{
		Width:          80,
		LineHeight:     10,
		AllowPageBreak: true,
	}

	d, err := builder.DrawTextBox(0, text, 20, 100, "Helvetica", 9, opts).Build()
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	defer d.Close()

	if err := d.Save(buf); err != nil {
		t.Fatalf("Save: %v", err)
	}

	pages, err := d.Pages()
	if err != nil {
		t.Fatalf("Pages: %v", err)
	}
	if len(pages) != 2 {
		t.Fatalf("expected 2 pages, got %d", len(pages))
	}

	// Basic sanity check that second page has a content reference.
	contents := pages[1].Contents()
	if contents == nil {
		t.Fatalf("expected second page to have Contents")
	}
}
