package doc

import (
	"fmt"

	"gpdf/doc/tagged"
	"gpdf/model"
)

// CellVerticalAlign controls vertical placement of content inside a table cell.
type CellVerticalAlign string

const (
	CellVAlignTop    CellVerticalAlign = "top"
	CellVAlignMiddle CellVerticalAlign = "middle"
	CellVAlignBottom CellVerticalAlign = "bottom"
)

// CellStyle controls padding and vertical alignment for a table cell.
// Zero value uses sensible defaults (padding 4pt on all sides, top alignment).
type CellStyle struct {
	PaddingTop    float64
	PaddingRight  float64
	PaddingBottom float64
	PaddingLeft   float64

	VAlign CellVerticalAlign

	// TextColorRGB is optional RGB text color in [0,1]. When all components are zero,
	// the document default color is used.
	TextColorRGB [3]float64
}

func (s CellStyle) resolvedPadding() (top, right, bottom, left float64) {
	const defaultPad = 4.0
	if s.PaddingTop == 0 && s.PaddingRight == 0 && s.PaddingBottom == 0 && s.PaddingLeft == 0 {
		return defaultPad, defaultPad, defaultPad, defaultPad
	}
	return s.PaddingTop, s.PaddingRight, s.PaddingBottom, s.PaddingLeft
}

// TableCellImageSpec describes an image to be placed inside a table cell.
type TableCellImageSpec struct {
	Raw               []byte
	WidthPx, HeightPx int
	BitsPerComponent  int
	ColorSpace        string

	WidthPt, HeightPt float64
}

// TableCellSpec describes a single cell in a tagged table.
type TableCellSpec struct {
	Text       string
	Paragraphs []string

	ListItems []string
	ListKind  string // "ordered", "unordered", or empty

	Image *TableCellImageSpec

	Style CellStyle

	ColSpan  int
	RowSpan  int
	IsHeader bool

	Scope string
	Alt   string
	Lang  string
}

// measureTableCellHeight returns an approximate height for the given cell content.
func measureTableCellHeight(cell TableCellSpec, cellWidth float64) float64 {
	const fontSize = 10.0
	const lineSpacing = 1.2

	top, right, bottom, left := cell.Style.resolvedPadding()
	height := top

	if cell.Image != nil && len(cell.Image.Raw) > 0 && cell.Image.WidthPx > 0 && cell.Image.HeightPx > 0 {
		w := cell.Image.WidthPt
		h := cell.Image.HeightPt
		if w <= 0 || h <= 0 {
			maxWidth := cellWidth - left - right
			if maxWidth <= 0 {
				maxWidth = cellWidth
			}
			scale := maxWidth / float64(cell.Image.WidthPx)
			w = float64(cell.Image.WidthPx) * scale
			h = float64(cell.Image.HeightPx) * scale
		}
		if h > 0 {
			height += h + top
		}
	}

	lines := 0
	if cell.Text != "" {
		lines++
	}
	for _, p := range cell.Paragraphs {
		if p != "" {
			lines++
		}
	}
	for _, item := range cell.ListItems {
		if item != "" {
			lines++
		}
	}
	if lines > 0 {
		lineHeight := fontSize * lineSpacing
		height += float64(lines) * lineHeight
	}
	height += bottom
	return height
}

const (
	tableDefaultBottomMargin = 40.0
	tableDefaultTopMargin    = 40.0
)

// TableBuilder builds a tagged table that can optionally span multiple pages.
type TableBuilder struct {
	builder   *DocumentBuilder
	pageIndex int

	x, y          float64
	width, height float64
	cols          int
	currentRow    int

	tableIndex int

	// Multi-page support fields.
	allowPageBreak bool
	headerCells    [][]TableCellSpec
	currentY       float64
	inPageBreak    bool
}

// BeginTable starts a tagged table region on the given page at (x, y) with the given size and number of columns.
func (b *DocumentBuilder) BeginTable(pageIndex int, x, y, width, height float64, cols int) *TableBuilder {
	if pageIndex < 0 || pageIndex >= len(b.pages) || cols <= 0 {
		return nil
	}
	b.useTagged = true
	b.tagging.Tables = append(b.tagging.Tables, tagged.StructTable{
		PageIndex: pageIndex,
	})
	tableIndex := len(b.tagging.Tables) - 1
	b.tagging.RecordSectionTable(tableIndex)
	return &TableBuilder{
		builder:    b,
		pageIndex:  pageIndex,
		x:          x,
		y:          y,
		width:      width,
		height:     height,
		cols:       cols,
		tableIndex: tableIndex,
		currentY:   y + height,
	}
}

// AllowPageBreak enables automatic page breaks when rows exceed available space.
// When enabled, header rows are repeated on each continuation page.
func (t *TableBuilder) AllowPageBreak() *TableBuilder {
	if t == nil {
		return t
	}
	t.allowPageBreak = true
	return t
}

// HeaderRow adds a header row to the table (cells treated as TH).
// When AllowPageBreak is enabled, header rows are repeated on continuation pages.
func (t *TableBuilder) HeaderRow(cells ...TableCellSpec) *TableBuilder {
	if t != nil && t.allowPageBreak && !t.inPageBreak {
		cellsCopy := make([]TableCellSpec, len(cells))
		copy(cellsCopy, cells)
		t.headerCells = append(t.headerCells, cellsCopy)
	}
	return t.addRow(true, cells...)
}

// Row adds a data row to the table (cells treated as TD).
func (t *TableBuilder) Row(cells ...TableCellSpec) *TableBuilder {
	return t.addRow(false, cells...)
}

func (t *TableBuilder) addRow(isHeader bool, cells ...TableCellSpec) *TableBuilder {
	if t == nil || t.builder == nil || len(cells) == 0 || t.cols <= 0 {
		return t
	}
	if t.tableIndex < 0 || t.tableIndex >= len(t.builder.tagging.Tables) {
		return t
	}
	table := &t.builder.tagging.Tables[t.tableIndex]

	cellWidth := t.width / float64(t.cols)

	rowHeight := 0.0
	cellHeights := make([]float64, len(cells))
	for col, cell := range cells {
		if col >= t.cols {
			break
		}
		h := measureTableCellHeight(cell, cellWidth)
		cellHeights[col] = h
		if h > rowHeight {
			rowHeight = h
		}
	}
	if rowHeight <= 0 {
		rowHeight = t.height / float64(t.cols)
		if rowHeight <= 0 {
			rowHeight = 16
		}
	}

	if t.allowPageBreak && !t.inPageBreak {
		if t.currentY-rowHeight < t.tableBottomMargin() {
			t.doPageBreak()
		}
	}

	if len(table.Rows) <= t.currentRow {
		table.Rows = append(table.Rows, tagged.StructRow{})
	}
	row := &table.Rows[t.currentRow]

	if len(table.RowHeights) <= t.currentRow {
		table.RowHeights = append(table.RowHeights, rowHeight)
	} else {
		table.RowHeights[t.currentRow] = rowHeight
	}

	rowBottom := t.currentY - rowHeight

	ps := &t.builder.pages[t.pageIndex]
	for col, cell := range cells {
		if col >= t.cols {
			break
		}
		if cell.ColSpan <= 0 {
			cell.ColSpan = 1
		}
		role := model.Name("TD")
		if isHeader || cell.IsHeader {
			role = model.Name("TH")
		}
		scope := cell.Scope
		if scope == "" && role == model.Name("TH") {
			scope = "Column"
		}

		left := t.x + float64(col)*cellWidth
		bottom := rowBottom

		sc := tagged.StructCell{
			PageIndex: t.pageIndex,
			Role:      role,
			Scope:     scope,
			Alt:       cell.Alt,
			Lang:      cell.Lang,
		}

		const fontName = "Helvetica"
		const fontSize = 10.0
		const lineSpacing = 1.2

		topPad, rightPad, bottomPad, leftPad := cell.Style.resolvedPadding()

		contentHeight := cellHeights[col] - (topPad + bottomPad)
		if contentHeight < 0 {
			contentHeight = 0
		}
		freeSpace := rowHeight - (contentHeight + topPad + bottomPad)
		if freeSpace < 0 {
			freeSpace = 0
		}

		var offset float64
		switch cell.Style.VAlign {
		case CellVAlignMiddle:
			offset = freeSpace / 2
		case CellVAlignBottom:
			offset = freeSpace
		default:
			offset = 0
		}

		currentY := bottom + rowHeight - topPad - offset

		if cell.Image != nil && len(cell.Image.Raw) > 0 && cell.Image.WidthPx > 0 && cell.Image.HeightPx > 0 {
			w := cell.Image.WidthPt
			h := cell.Image.HeightPt
			if w <= 0 || h <= 0 {
				maxWidth := cellWidth - leftPad - rightPad
				if maxWidth <= 0 {
					maxWidth = cellWidth
				}
				scale := maxWidth / float64(cell.Image.WidthPx)
				w = float64(cell.Image.WidthPx) * scale
				h = float64(cell.Image.HeightPx) * scale
				maxHeight := rowHeight - topPad - bottomPad
				if h > maxHeight {
					h = maxHeight
				}
			}
			imgY := currentY - h
			mcid := ps.nextMCID
			ps.nextMCID++
			sc.MCIDs = append(sc.MCIDs, mcid)
			ps.imageRuns = append(ps.imageRuns, imageRun{
				X:                left + leftPad,
				Y:                imgY,
				WidthPt:          w,
				HeightPt:         h,
				Raw:              cell.Image.Raw,
				WidthPx:          cell.Image.WidthPx,
				HeightPx:         cell.Image.HeightPx,
				BitsPerComponent: cell.Image.BitsPerComponent,
				ColorSpace:       cell.Image.ColorSpace,
				MCID:             mcid,
				HasMCID:          true,
			})
			currentY = imgY - topPad
		}

		var lines []string
		if cell.Text != "" {
			lines = append(lines, cell.Text)
		}
		if len(cell.Paragraphs) > 0 {
			lines = append(lines, cell.Paragraphs...)
		}
		if len(cell.ListItems) > 0 {
			for i, item := range cell.ListItems {
				prefix := ""
				switch cell.ListKind {
				case "ordered":
					prefix = fmt.Sprintf("%d. ", i+1)
				case "unordered":
					prefix = "• "
				}
				lines = append(lines, prefix+item)
			}
		}

		lineHeight := fontSize * lineSpacing
		for _, text := range lines {
			if text == "" {
				continue
			}
			mcid := ps.nextMCID
			ps.nextMCID++
			sc.MCIDs = append(sc.MCIDs, mcid)
			x := left + leftPad
			y := currentY
			currentY -= lineHeight
			tr := textRun{
				Text:            text,
				X:               x,
				Y:               y,
				FontName:        fontName,
				FontSize:        fontSize,
				MCID:            mcid,
				HasMCID:         true,
				Role:            role,
				UseDefaultColor: true,
			}
			if cell.Style.TextColorRGB != ([3]float64{}) {
				tr.TextColorRGB = cell.Style.TextColorRGB
				tr.UseDefaultColor = false
			}
			ps.textRuns = append(ps.textRuns, tr)
			if currentY < bottom+bottomPad {
				break
			}
		}

		row.Cells = append(row.Cells, sc)
	}
	t.currentY = rowBottom
	t.currentRow++
	return t
}

// tableBottomMargin returns the minimum y coordinate before a page break triggers.
func (t *TableBuilder) tableBottomMargin() float64 {
	return tableDefaultBottomMargin
}

// doPageBreak moves the table to the next page and re-renders header rows.
func (t *TableBuilder) doPageBreak() {
	t.pageIndex++
	if t.pageIndex >= len(t.builder.pages) {
		t.builder.AddPage()
	}
	pageHeight := t.builder.pageHeight(t.pageIndex)
	t.currentY = pageHeight - tableDefaultTopMargin

	if len(t.headerCells) > 0 {
		t.inPageBreak = true
		for _, hdr := range t.headerCells {
			t.renderVisualRow(true, hdr)
		}
		t.inPageBreak = false
	}
}

// renderVisualRow renders a row's content without creating tagged structure entries.
// Used for repeating header rows on continuation pages.
func (t *TableBuilder) renderVisualRow(isHeader bool, cells []TableCellSpec) {
	cellWidth := t.width / float64(t.cols)

	rowHeight := 0.0
	for col, cell := range cells {
		if col >= t.cols {
			break
		}
		h := measureTableCellHeight(cell, cellWidth)
		if h > rowHeight {
			rowHeight = h
		}
	}
	if rowHeight <= 0 {
		rowHeight = 16
	}

	rowBottom := t.currentY - rowHeight
	ps := &t.builder.pages[t.pageIndex]

	const fontName = "Helvetica"
	const fontSize = 10.0
	const lineSpacing = 1.2

	for col, cell := range cells {
		if col >= t.cols {
			break
		}

		left := t.x + float64(col)*cellWidth
		topPad, _, bottomPad, leftPad := cell.Style.resolvedPadding()
		currentY := rowBottom + rowHeight - topPad

		var lines []string
		if cell.Text != "" {
			lines = append(lines, cell.Text)
		}
		lines = append(lines, cell.Paragraphs...)
		for i, item := range cell.ListItems {
			prefix := ""
			switch cell.ListKind {
			case "ordered":
				prefix = fmt.Sprintf("%d. ", i+1)
			case "unordered":
				prefix = "• "
			}
			lines = append(lines, prefix+item)
		}

		lineHeight := fontSize * lineSpacing
		for _, text := range lines {
			if text == "" {
				continue
			}
			ps.textRuns = append(ps.textRuns, textRun{
				Text:            text,
				X:               left + leftPad,
				Y:               currentY,
				FontName:        fontName,
				FontSize:        fontSize,
				UseDefaultColor: true,
			})
			currentY -= lineHeight
			if currentY < rowBottom+bottomPad {
				break
			}
		}
	}
	t.currentY = rowBottom
}

// EndTable finishes the table and returns the underlying DocumentBuilder.
func (t *TableBuilder) EndTable() *DocumentBuilder {
	if t == nil {
		return nil
	}
	return t.builder
}
