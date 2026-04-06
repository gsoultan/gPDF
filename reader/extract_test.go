package reader

import (
	"fmt"
	"strings"
	"testing"

	"github.com/gsoultan/gpdf/model"
)

// ── helpers ──────────────────────────────────────────────────────────────────

func makePageWithContent(contentRef model.Ref, resources model.Dict) model.Page {
	d := model.Dict{model.Name("Contents"): contentRef}
	if resources != nil {
		d[model.Name("Resources")] = resources
	}
	return model.Page{Dict: d}
}

func makePageWithMediaBox(contentRef model.Ref, resources model.Dict, w, h float64) model.Page {
	d := model.Dict{
		model.Name("Contents"): contentRef,
		model.Name("MediaBox"): model.Array{
			model.Real(0), model.Real(0), model.Real(w), model.Real(h),
		},
	}
	if resources != nil {
		d[model.Name("Resources")] = resources
	}
	return model.Page{Dict: d}
}

type stubContentSource struct {
	pages   []model.Page
	objects map[model.Ref]model.Object
}

// ── Image extraction tests ────────────────────────────────────────────────────

func TestExtractImages_ReturnsImageMetadata(t *testing.T) {
	imgRef := model.Ref{ObjectNumber: 10}
	contentRef := model.Ref{ObjectNumber: 11}

	imgStream := &model.Stream{
		Dict: model.Dict{
			model.Name("Type"):             model.Name("XObject"),
			model.Name("Subtype"):          model.Name("Image"),
			model.Name("Width"):            model.Integer(100),
			model.Name("Height"):           model.Integer(200),
			model.Name("BitsPerComponent"): model.Integer(8),
			model.Name("ColorSpace"):       model.Name("DeviceRGB"),
			model.Name("Filter"):           model.Name("DCTDecode"),
		},
		Content: []byte("fakejpegdata"),
	}

	src := stubContentSource{
		pages: []model.Page{
			makePageWithContent(contentRef, model.Dict{
				model.Name("XObject"): model.Dict{
					model.Name("Im0"): imgRef,
				},
			}),
		},
		objects: map[model.Ref]model.Object{
			contentRef: &model.Stream{Content: []byte("q /Im0 Do Q")},
			imgRef:     imgStream,
		},
	}

	images, err := ExtractImages(src)
	if err != nil {
		t.Fatalf("ExtractImages error: %v", err)
	}
	if len(images) != 1 {
		t.Fatalf("expected 1 image, got %d", len(images))
	}
	img := images[0]
	if img.Name != "Im0" {
		t.Errorf("Name: got %q, want %q", img.Name, "Im0")
	}
	if img.Width != 100 {
		t.Errorf("Width: got %d, want 100", img.Width)
	}
	if img.Height != 200 {
		t.Errorf("Height: got %d, want 200", img.Height)
	}
	if img.BitsPerComponent != 8 {
		t.Errorf("BitsPerComponent: got %d, want 8", img.BitsPerComponent)
	}
	if img.ColorSpace != "DeviceRGB" {
		t.Errorf("ColorSpace: got %q, want DeviceRGB", img.ColorSpace)
	}
	if img.Filter != "DCTDecode" {
		t.Errorf("Filter: got %q, want DCTDecode", img.Filter)
	}
	if string(img.Data) != "fakejpegdata" {
		t.Errorf("Data: got %q, want fakejpegdata", img.Data)
	}
	if img.Format != "jpeg" {
		t.Errorf("Format: got %q, want jpeg", img.Format)
	}
}

func TestExtractImages_TracksPlacementFromCTM(t *testing.T) {
	imgRef := model.Ref{ObjectNumber: 12}
	contentRef := model.Ref{ObjectNumber: 13}

	src := stubContentSource{
		pages: []model.Page{
			makePageWithContent(contentRef, model.Dict{
				model.Name("XObject"): model.Dict{model.Name("Im0"): imgRef},
			}),
		},
		objects: map[model.Ref]model.Object{
			contentRef: &model.Stream{Content: []byte("q 120 0 0 80 36 48 cm /Im0 Do Q")},
			imgRef: &model.Stream{Dict: model.Dict{
				model.Name("Type"):             model.Name("XObject"),
				model.Name("Subtype"):          model.Name("Image"),
				model.Name("Width"):            model.Integer(60),
				model.Name("Height"):           model.Integer(40),
				model.Name("BitsPerComponent"): model.Integer(8),
				model.Name("ColorSpace"):       model.Name("DeviceRGB"),
			}, Content: []byte("raw")},
		},
	}

	images, err := ExtractImages(src)
	if err != nil {
		t.Fatalf("ExtractImages error: %v", err)
	}
	if len(images) != 1 {
		t.Fatalf("expected 1 image, got %d", len(images))
	}
	img := images[0]
	if img.X != 36 || img.Y != 48 {
		t.Fatalf("placement: got (%.0f, %.0f), want (36, 48)", img.X, img.Y)
	}
	if img.WidthPt != 120 || img.HeightPt != 80 {
		t.Fatalf("size: got %.0fx%.0f, want 120x80", img.WidthPt, img.HeightPt)
	}
}

func TestExtractImages_SkipsFormXObjects(t *testing.T) {
	formRef := model.Ref{ObjectNumber: 20}
	contentRef := model.Ref{ObjectNumber: 21}

	src := stubContentSource{
		pages: []model.Page{
			makePageWithContent(contentRef, model.Dict{
				model.Name("XObject"): model.Dict{
					model.Name("F1"): formRef,
				},
			}),
		},
		objects: map[model.Ref]model.Object{
			contentRef: &model.Stream{Content: []byte("q /F1 Do Q")},
			formRef: &model.Stream{Dict: model.Dict{
				model.Name("Type"):    model.Name("XObject"),
				model.Name("Subtype"): model.Name("Form"),
			}, Content: []byte("BT (text) Tj ET")},
		},
	}

	images, err := ExtractImages(src)
	if err != nil {
		t.Fatalf("ExtractImages error: %v", err)
	}
	if len(images) != 0 {
		t.Errorf("expected 0 images (form XObjects excluded), got %d", len(images))
	}
}

func TestExtractImagesPerPage_GroupsByPage(t *testing.T) {
	img1Ref := model.Ref{ObjectNumber: 30}
	img2Ref := model.Ref{ObjectNumber: 31}
	c1Ref := model.Ref{ObjectNumber: 32}
	c2Ref := model.Ref{ObjectNumber: 33}

	makeImg := func(ref model.Ref, w, h int) *model.Stream {
		return &model.Stream{
			Dict: model.Dict{
				model.Name("Subtype"): model.Name("Image"),
				model.Name("Width"):   model.Integer(w),
				model.Name("Height"):  model.Integer(h),
			},
			Content: []byte("data"),
		}
	}

	src := stubContentSource{
		pages: []model.Page{
			makePageWithContent(c1Ref, model.Dict{
				model.Name("XObject"): model.Dict{model.Name("I1"): img1Ref},
			}),
			makePageWithContent(c2Ref, model.Dict{
				model.Name("XObject"): model.Dict{model.Name("I2"): img2Ref},
			}),
		},
		objects: map[model.Ref]model.Object{
			c1Ref:   &model.Stream{Content: []byte("")},
			c2Ref:   &model.Stream{Content: []byte("")},
			img1Ref: makeImg(img1Ref, 50, 60),
			img2Ref: makeImg(img2Ref, 70, 80),
		},
	}

	perPage, err := ExtractImagesPerPage(src)
	if err != nil {
		t.Fatalf("ExtractImagesPerPage error: %v", err)
	}
	if len(perPage) != 2 {
		t.Fatalf("expected 2 pages, got %d", len(perPage))
	}
	if len(perPage[0]) != 1 || perPage[0][0].Width != 50 {
		t.Errorf("page 0: expected 1 image with width 50, got %+v", perPage[0])
	}
	if len(perPage[1]) != 1 || perPage[1][0].Width != 70 {
		t.Errorf("page 1: expected 1 image with width 70, got %+v", perPage[1])
	}
}

// ── Layout extraction tests ───────────────────────────────────────────────────

func TestExtractLayout_CapturesPositionAndStyle(t *testing.T) {
	contentRef := model.Ref{ObjectNumber: 40}

	// Tm sets position (100,700); Tf sets font F1 size 12; rg sets colour; Tj emits text
	stream := []byte("BT /F1 12 Tf 1 0 0 0.5 0 0 rg 1 0 0 1 100 700 Tm (Hello) Tj ET")

	src := stubContentSource{
		pages: []model.Page{
			makePageWithMediaBox(contentRef, model.Dict{
				model.Name("Font"): model.Dict{},
			}, 595, 842),
		},
		objects: map[model.Ref]model.Object{
			contentRef: &model.Stream{Content: stream},
		},
	}

	layouts, err := ExtractLayout(src)
	if err != nil {
		t.Fatalf("ExtractLayout error: %v", err)
	}
	if len(layouts) != 1 {
		t.Fatalf("expected 1 layout, got %d", len(layouts))
	}
	pl := layouts[0]
	if pl.Width != 595 || pl.Height != 842 {
		t.Errorf("page size: got %.0fx%.0f, want 595x842", pl.Width, pl.Height)
	}
	if len(pl.Blocks) == 0 {
		t.Fatal("expected at least one TextBlock")
	}
	b := pl.Blocks[0]
	if !strings.Contains(b.Text, "Hello") {
		t.Errorf("TextBlock.Text: got %q, want to contain Hello", b.Text)
	}
	if b.X != 100 {
		t.Errorf("TextBlock.X: got %.1f, want 100", b.X)
	}
	if b.Y != 700 {
		t.Errorf("TextBlock.Y: got %.1f, want 700", b.Y)
	}
	if b.Style.FontSize != 12 {
		t.Errorf("TextBlock.Style.FontSize: got %.1f, want 12", b.Style.FontSize)
	}
}

func TestExtractLayout_UsesCropBoxRotateAndUserUnitForPageSize(t *testing.T) {
	contentRef := model.Ref{ObjectNumber: 41}
	src := stubContentSource{
		pages: []model.Page{{Dict: model.Dict{
			model.Name("Contents"): contentRef,
			model.Name("MediaBox"): model.Array{model.Integer(0), model.Integer(0), model.Integer(200), model.Integer(400)},
			model.Name("CropBox"):  model.Array{model.Integer(0), model.Integer(0), model.Integer(100), model.Integer(300)},
			model.Name("Rotate"):   model.Integer(90),
			model.Name("UserUnit"): model.Integer(2),
		}}},
		objects: map[model.Ref]model.Object{
			contentRef: &model.Stream{Content: []byte("BT ET")},
		},
	}

	layouts, err := ExtractLayout(src)
	if err != nil {
		t.Fatalf("ExtractLayout error: %v", err)
	}
	if len(layouts) != 1 {
		t.Fatalf("expected 1 layout, got %d", len(layouts))
	}
	if layouts[0].Width != 600 || layouts[0].Height != 200 {
		t.Fatalf("page size: got %.0fx%.0f, want 600x200", layouts[0].Width, layouts[0].Height)
	}
}

func TestExtractLayout_TracksColourChanges(t *testing.T) {
	contentRef := model.Ref{ObjectNumber: 50}

	// First text is black (default), second is red via rg operator
	stream := []byte("BT /F1 10 Tf 1 0 0 1 10 800 Tm (Black) Tj 1 0 0 rg 1 0 0 1 10 780 Tm (Red) Tj ET")

	src := stubContentSource{
		pages:   []model.Page{makePageWithContent(contentRef, nil)},
		objects: map[model.Ref]model.Object{contentRef: &model.Stream{Content: stream}},
	}

	layouts, err := ExtractLayout(src)
	if err != nil {
		t.Fatalf("ExtractLayout error: %v", err)
	}
	if len(layouts[0].Blocks) < 2 {
		t.Fatalf("expected ≥2 blocks, got %d", len(layouts[0].Blocks))
	}
	red := layouts[0].Blocks[1]
	if red.Style.ColorR != 1 || red.Style.ColorG != 0 || red.Style.ColorB != 0 {
		t.Errorf("expected red block rgb=(1,0,0), got (%.2f,%.2f,%.2f)",
			red.Style.ColorR, red.Style.ColorG, red.Style.ColorB)
	}
}

func TestExtractLayout_TdUpdatesPosition(t *testing.T) {
	contentRef := model.Ref{ObjectNumber: 60}

	// Start at (50,700), move down by 20 with Td
	stream := []byte("BT /F1 10 Tf 1 0 0 1 50 700 Tm (First) Tj 0 -20 Td (Second) Tj ET")

	src := stubContentSource{
		pages:   []model.Page{makePageWithContent(contentRef, nil)},
		objects: map[model.Ref]model.Object{contentRef: &model.Stream{Content: stream}},
	}

	layouts, err := ExtractLayout(src)
	if err != nil {
		t.Fatalf("ExtractLayout error: %v", err)
	}
	blocks := layouts[0].Blocks
	if len(blocks) < 2 {
		t.Fatalf("expected ≥2 blocks, got %d", len(blocks))
	}
	if blocks[0].Y != 700 {
		t.Errorf("first block Y: got %.1f, want 700", blocks[0].Y)
	}
	if blocks[1].Y != 680 {
		t.Errorf("second block Y after Td(0,-20): got %.1f, want 680", blocks[1].Y)
	}
}

func TestExtractVectors_TracksLinesAndRects(t *testing.T) {
	stream := []byte("0 0 1 RG 10 20 m 110 20 l S 1 0 0 rg 30 40 50 60 re f")
	contentRef := model.Ref{ObjectNumber: 140, Generation: 0}
	src := stubContentSource{
		pages:   []model.Page{makePageWithContent(contentRef, nil)},
		objects: map[model.Ref]model.Object{contentRef: &model.Stream{Content: stream}},
	}

	perPage, err := ExtractVectorsPerPage(src)
	if err != nil {
		t.Fatalf("ExtractVectorsPerPage error: %v", err)
	}
	if len(perPage) != 1 {
		t.Fatalf("expected one page result, got %d", len(perPage))
	}
	shapes := perPage[0]
	if len(shapes) != 2 {
		t.Fatalf("expected 2 shapes, got %d", len(shapes))
	}

	line := shapes[0]
	if line.Kind != "line" {
		t.Fatalf("first shape kind: got %q, want line", line.Kind)
	}
	if line.X1 != 10 || line.Y1 != 20 || line.X2 != 110 || line.Y2 != 20 {
		t.Fatalf("line coords: got (%v,%v)->(%v,%v), want (10,20)->(110,20)", line.X1, line.Y1, line.X2, line.Y2)
	}
	if !line.Stroke || line.Fill {
		t.Fatalf("line paint flags: stroke=%v fill=%v, want stroke=true fill=false", line.Stroke, line.Fill)
	}
	if line.StrokeColor.B != 1 {
		t.Fatalf("line stroke color B: got %v, want 1", line.StrokeColor.B)
	}

	rect := shapes[1]
	if rect.Kind != "rect" {
		t.Fatalf("second shape kind: got %q, want rect", rect.Kind)
	}
	if rect.X1 != 30 || rect.Y1 != 40 || rect.X2 != 80 || rect.Y2 != 100 {
		t.Fatalf("rect bounds: got (%v,%v)-(%v,%v), want (30,40)-(80,100)", rect.X1, rect.Y1, rect.X2, rect.Y2)
	}
	if rect.Stroke || !rect.Fill {
		t.Fatalf("rect paint flags: stroke=%v fill=%v, want stroke=false fill=true", rect.Stroke, rect.Fill)
	}
	if rect.FillColor.R != 1 {
		t.Fatalf("rect fill color R: got %v, want 1", rect.FillColor.R)
	}
}

func TestExtractVectors_AppliesCTM(t *testing.T) {
	stream := []byte("2 0 0 2 5 6 cm 1 1 m 3 1 l S")
	contentRef := model.Ref{ObjectNumber: 141, Generation: 0}
	src := stubContentSource{
		pages:   []model.Page{makePageWithContent(contentRef, nil)},
		objects: map[model.Ref]model.Object{contentRef: &model.Stream{Content: stream}},
	}

	perPage, err := ExtractVectorsPerPage(src)
	if err != nil {
		t.Fatalf("ExtractVectorsPerPage error: %v", err)
	}
	if len(perPage) != 1 || len(perPage[0]) != 1 {
		t.Fatalf("expected one transformed line shape, got %+v", perPage)
	}
	line := perPage[0][0]
	if line.X1 != 7 || line.Y1 != 8 || line.X2 != 11 || line.Y2 != 8 {
		t.Fatalf("transformed coords: got (%v,%v)->(%v,%v), want (7,8)->(11,8)", line.X1, line.Y1, line.X2, line.Y2)
	}
}

// ── Table detection tests ─────────────────────────────────────────────────────

func TestDetectTables_DetectsSimpleGrid(t *testing.T) {
	// 3 rows × 3 cols at consistent X anchors
	blocks := []TextBlock{
		{Text: "Name", X: 10, Y: 700, Style: TextStyle{FontSize: 10}},
		{Text: "Age", X: 150, Y: 700, Style: TextStyle{FontSize: 10}},
		{Text: "City", X: 290, Y: 700, Style: TextStyle{FontSize: 10}},

		{Text: "Alice", X: 10, Y: 680, Style: TextStyle{FontSize: 10}},
		{Text: "30", X: 150, Y: 680, Style: TextStyle{FontSize: 10}},
		{Text: "Paris", X: 290, Y: 680, Style: TextStyle{FontSize: 10}},

		{Text: "Bob", X: 10, Y: 660, Style: TextStyle{FontSize: 10}},
		{Text: "25", X: 150, Y: 660, Style: TextStyle{FontSize: 10}},
		{Text: "Rome", X: 290, Y: 660, Style: TextStyle{FontSize: 10}},
	}

	layouts := []PageLayout{{Page: 0, Width: 595, Height: 842, Blocks: blocks}}
	tables := DetectTables(layouts)

	if len(tables) != 1 {
		t.Fatalf("expected tables for 1 page, got %d", len(tables))
	}
	if len(tables[0]) != 1 {
		t.Fatalf("expected 1 table detected, got %d", len(tables[0]))
	}
	tbl := tables[0][0]
	if tbl.Rows != 3 {
		t.Errorf("Rows: got %d, want 3", tbl.Rows)
	}
	if tbl.Cols != 3 {
		t.Errorf("Cols: got %d, want 3", tbl.Cols)
	}
	if tbl.Cell(0, 0) != "Name" {
		t.Errorf("Cell(0,0): got %q, want Name", tbl.Cell(0, 0))
	}
	if tbl.Cell(1, 1) != "30" {
		t.Errorf("Cell(1,1): got %q, want 30", tbl.Cell(1, 1))
	}
	if tbl.Cell(2, 2) != "Rome" {
		t.Errorf("Cell(2,2): got %q, want Rome", tbl.Cell(2, 2))
	}
}

func TestDetectTables_ReturnsNilForSingleRow(t *testing.T) {
	blocks := []TextBlock{
		{Text: "A", X: 10, Y: 700},
		{Text: "B", X: 150, Y: 700},
		{Text: "C", X: 290, Y: 700},
	}
	layouts := []PageLayout{{Page: 0, Blocks: blocks}}
	tables := DetectTables(layouts)
	if len(tables[0]) != 0 {
		t.Errorf("expected no table for single row, got %d", len(tables[0]))
	}
}

func TestDetectTables_ReturnsNilForNoBlocks(t *testing.T) {
	layouts := []PageLayout{{Page: 0}}
	tables := DetectTables(layouts)
	if len(tables[0]) != 0 {
		t.Errorf("expected no table for empty page, got %d", len(tables[0]))
	}
}

func TestTable_CellLookup(t *testing.T) {
	tbl := Table{
		Rows: 2, Cols: 2,
		Cells: []TableCell{
			{Row: 0, Col: 0, Text: "A"},
			{Row: 0, Col: 1, Text: "B"},
			{Row: 1, Col: 0, Text: "C"},
		},
	}
	cases := []struct {
		r, c int
		want string
	}{
		{0, 0, "A"}, {0, 1, "B"}, {1, 0, "C"}, {1, 1, ""},
	}
	for _, tc := range cases {
		if got := tbl.Cell(tc.r, tc.c); got != tc.want {
			t.Errorf("Cell(%d,%d) = %q, want %q", tc.r, tc.c, got, tc.want)
		}
	}
}

func (s stubContentSource) Pages() ([]model.Page, error) {
	return s.pages, nil
}

func (s stubContentSource) Resolve(ref model.Ref) (model.Object, error) {
	obj, ok := s.objects[ref]
	if !ok {
		return nil, fmt.Errorf("missing object %v", ref)
	}
	return obj, nil
}

func TestExtractText_TraversesFormXObjects(t *testing.T) {
	pageContentRef := model.Ref{ObjectNumber: 1}
	formRef := model.Ref{ObjectNumber: 2}

	src := stubContentSource{
		pages: []model.Page{
			{Dict: model.Dict{
				model.Name("Contents"): pageContentRef,
				model.Name("Resources"): model.Dict{
					model.Name("XObject"): model.Dict{
						model.Name("X1"): formRef,
					},
				},
			}},
		},
		objects: map[model.Ref]model.Object{
			pageContentRef: &model.Stream{Content: []byte("q /X1 Do Q")},
			formRef: &model.Stream{Dict: model.Dict{
				model.Name("Type"):    model.Name("XObject"),
				model.Name("Subtype"): model.Name("Form"),
			}, Content: []byte("BT (Nested text) Tj ET")},
		},
	}

	text, err := ExtractText(src)
	if err != nil {
		t.Fatalf("ExtractText returned error: %v", err)
	}
	if !strings.Contains(text, "Nested text") {
		t.Fatalf("expected extracted text to contain nested form text, got %q", text)
	}
}

func TestExtractText_HandlesQuoteTextOperators(t *testing.T) {
	pageContentRef := model.Ref{ObjectNumber: 1}

	src := stubContentSource{
		pages: []model.Page{{Dict: model.Dict{model.Name("Contents"): pageContentRef}}},
		objects: map[model.Ref]model.Object{
			pageContentRef: &model.Stream{Content: []byte("BT (First) Tj (Second) ' 120 0 (Third) \" ET")},
		},
	}

	text, err := ExtractText(src)
	if err != nil {
		t.Fatalf("ExtractText returned error: %v", err)
	}
	for _, expected := range []string{"First", "Second", "Third"} {
		if !strings.Contains(text, expected) {
			t.Fatalf("expected extracted text to contain %q, got %q", expected, text)
		}
	}
	if strings.Contains(text, "FirstSecond") || strings.Contains(text, "SecondThird") {
		t.Fatalf("expected quote operators to preserve text boundaries, got %q", text)
	}
}

func TestExtractText_PreservesLineBoundariesForTextPositionOperators(t *testing.T) {
	pageContentRef := model.Ref{ObjectNumber: 1}

	src := stubContentSource{
		pages: []model.Page{{Dict: model.Dict{model.Name("Contents"): pageContentRef}}},
		objects: map[model.Ref]model.Object{
			pageContentRef: &model.Stream{Content: []byte("BT (Line1) Tj T* (Line2) Tj 0 -14 Td (Line3) Tj ET")},
		},
	}

	perPage, err := ExtractTextPerPage(src)
	if err != nil {
		t.Fatalf("ExtractTextPerPage returned error: %v", err)
	}
	if len(perPage) != 1 {
		t.Fatalf("expected 1 page, got %d", len(perPage))
	}
	if strings.Contains(perPage[0], "Line1Line2") || strings.Contains(perPage[0], "Line2Line3") {
		t.Fatalf("expected line boundaries to be preserved, got %q", perPage[0])
	}
}

func TestExtractText_InsertsSpaceForTJKerningAdjustments(t *testing.T) {
	pageContentRef := model.Ref{ObjectNumber: 1}

	src := stubContentSource{
		pages: []model.Page{{Dict: model.Dict{model.Name("Contents"): pageContentRef}}},
		objects: map[model.Ref]model.Object{
			pageContentRef: &model.Stream{Content: []byte("BT [(Hello) -250 (World)] TJ ET")},
		},
	}

	text, err := ExtractText(src)
	if err != nil {
		t.Fatalf("ExtractText returned error: %v", err)
	}
	if !strings.Contains(text, "Hello World") {
		t.Fatalf("expected TJ kerning to preserve spaces, got %q", text)
	}
}

func TestExtractText_ReadsDirectPageContentStreams(t *testing.T) {
	refContent := model.Ref{ObjectNumber: 2}

	src := stubContentSource{
		pages: []model.Page{{Dict: model.Dict{
			model.Name("Contents"): model.Array{
				model.Stream{Content: []byte("BT (Direct) Tj ET")},
				refContent,
			},
		}}},
		objects: map[model.Ref]model.Object{
			refContent: &model.Stream{Content: []byte("BT (Ref) Tj ET")},
		},
	}

	text, err := ExtractText(src)
	if err != nil {
		t.Fatalf("ExtractText returned error: %v", err)
	}
	if !strings.Contains(text, "Direct") || !strings.Contains(text, "Ref") {
		t.Fatalf("expected extracted text from direct and referenced contents, got %q", text)
	}
}

func TestExtractText_InsertsBoundaryBetweenTextBlocks(t *testing.T) {
	pageContentRef := model.Ref{ObjectNumber: 1}

	src := stubContentSource{
		pages: []model.Page{{Dict: model.Dict{model.Name("Contents"): pageContentRef}}},
		objects: map[model.Ref]model.Object{
			pageContentRef: &model.Stream{Content: []byte("BT (Block1) Tj ET BT (Block2) Tj ET")},
		},
	}

	text, err := ExtractText(src)
	if err != nil {
		t.Fatalf("ExtractText returned error: %v", err)
	}
	if strings.Contains(text, "Block1Block2") {
		t.Fatalf("expected boundary between text blocks, got %q", text)
	}
	if !strings.Contains(text, "Block1") || !strings.Contains(text, "Block2") {
		t.Fatalf("expected both blocks in output, got %q", text)
	}
}

func TestExtractText_DecodesTextUsingToUnicodeCMap(t *testing.T) {
	pageContentRef := model.Ref{ObjectNumber: 1}
	fontRef := model.Ref{ObjectNumber: 2}
	toUnicodeRef := model.Ref{ObjectNumber: 3}

	src := stubContentSource{
		pages: []model.Page{{Dict: model.Dict{
			model.Name("Contents"): pageContentRef,
			model.Name("Resources"): model.Dict{
				model.Name("Font"): model.Dict{
					model.Name("F1"): fontRef,
				},
			},
		}}},
		objects: map[model.Ref]model.Object{
			pageContentRef: &model.Stream{Content: []byte("BT /F1 12 Tf <00010002> Tj ET")},
			fontRef: model.Dict{
				model.Name("Type"):      model.Name("Font"),
				model.Name("Subtype"):   model.Name("Type0"),
				model.Name("Encoding"):  model.Name("Identity-H"),
				model.Name("ToUnicode"): toUnicodeRef,
			},
			toUnicodeRef: &model.Stream{Content: []byte(strings.Join([]string{
				"/CIDInit /ProcSet findresource begin",
				"12 dict begin",
				"begincmap",
				"1 begincodespacerange",
				"<0000> <FFFF>",
				"endcodespacerange",
				"2 beginbfchar",
				"<0001> <0048>",
				"<0002> <0069>",
				"endbfchar",
				"endcmap",
				"end",
				"end",
			}, "\n"))},
		},
	}

	text, err := ExtractText(src)
	if err != nil {
		t.Fatalf("ExtractText returned error: %v", err)
	}

	if !strings.Contains(text, "Hi") {
		t.Fatalf("expected extracted text to contain decoded text %q, got %q", "Hi", text)
	}
}
