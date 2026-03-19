package doc

import (
	"fmt"

	"gpdf/doc/tagged"
	"gpdf/model"
)

// measureTableCellHeight returns an approximate height for the given cell content.
func measureTableCellHeight(b *DocumentBuilder, cell TableCellSpec, cellWidth float64) float64 {
	const fontSize = 10.0
	const lineSpacing = 1.2
	const fontName = "Helvetica"

	top, right, bottom, left := cell.Style.ResolvedPadding()
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
			height += h
		}
	}

	lines := collectCellLines(b, cell, cellWidth, fontName, fontSize)
	if len(lines) > 0 {
		lineHeight := fontSize * lineSpacing
		// For the first line, we only need the ascent.
		// For subsequent lines, we add the full line height.
		height += float64(len(lines)) * lineHeight
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

	// Styling options.
	// HeaderFillColor is the background fill for header rows (applied when HasHeaderFillColor is true).
	HeaderFillColor    Color
	HasHeaderFillColor bool
	// AlternateRowColor is applied to every odd-numbered data row (0-based) when HasAlternateRowColor is true.
	AlternateRowColor    Color
	HasAlternateRowColor bool
	dataRowIndex         int // counts data (non-header) rows for alternation
}

// BeginTable starts a tagged table region on the given page at (x, y) with the given size and number of columns.
func (b *DocumentBuilder) BeginTable(pageIndex int, x, y, width, height float64, cols int) *TableBuilder {
	if !b.pc.validPageIndex(pageIndex) || cols <= 0 {
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

// WithHeaderFillColor sets a background fill color applied to all header rows.
func (t *TableBuilder) WithHeaderFillColor(c Color) *TableBuilder {
	if t == nil {
		return t
	}
	t.HeaderFillColor = c
	t.HasHeaderFillColor = true
	return t
}

// WithAlternateRowColor sets a background fill color applied to every other data row.
func (t *TableBuilder) WithAlternateRowColor(c Color) *TableBuilder {
	if t == nil {
		return t
	}
	t.AlternateRowColor = c
	t.HasAlternateRowColor = true
	return t
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

	baseColWidth := t.width / float64(t.cols)

	rowHeight := 0.0
	cellHeights := make([]float64, len(cells))
	logicalCol := 0
	for idx, cell := range cells {
		if logicalCol >= t.cols {
			break
		}
		span := cell.ColSpan
		if span <= 0 {
			span = 1
		}
		cellWidth := baseColWidth * float64(span)
		h := measureTableCellHeight(t.builder, cell, cellWidth)
		cellHeights[idx] = h
		if h > rowHeight {
			rowHeight = h
		}
		logicalCol += span
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

	ps := &t.builder.pc.pages[t.pageIndex]

	// Determine the effective row fill color.
	var rowFillColor Color
	hasRowFill := false
	if isHeader && t.HasHeaderFillColor {
		rowFillColor = t.HeaderFillColor
		hasRowFill = true
	} else if !isHeader && t.HasAlternateRowColor && t.dataRowIndex%2 == 1 {
		rowFillColor = t.AlternateRowColor
		hasRowFill = true
	}

	logicalCol2 := 0
	for cellIdx, cell := range cells {
		if logicalCol2 >= t.cols {
			break
		}
		span := cell.ColSpan
		if span <= 0 {
			span = 1
		}
		cellColWidth := baseColWidth * float64(span)

		role := model.Name("TD")
		if isHeader || cell.IsHeader {
			role = model.Name("TH")
		}
		scope := cell.Scope
		if scope == "" && role == model.Name("TH") {
			scope = "Column"
		}

		left := t.x + float64(logicalCol2)*baseColWidth
		cellWidth := cellColWidth
		_ = cellHeights[cellIdx] // already computed above
		bottom := rowBottom

		// Draw cell background fill.
		effFill := rowFillColor
		effHasFill := hasRowFill
		if cell.Style.HasFillColor {
			effFill = cell.Style.FillColor
			effHasFill = true
		}
		if effHasFill {
			t.builder.FillRect(t.pageIndex, left, bottom, cellWidth, rowHeight, effFill)
		}

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

		topPad, rightPad, bottomPad, leftPad := cell.Style.ResolvedPadding()

		contentHeight := cellHeights[cellIdx] - (topPad + bottomPad)
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
			w, h := computeImageDimensions(cell.Image, cellWidth, rowHeight, leftPad, rightPad, topPad, bottomPad)
			imgY := currentY - h
			mcid := ps.NextMCID
			ps.NextMCID++
			sc.MCIDs = append(sc.MCIDs, mcid)
			ps.ImageRuns = append(ps.ImageRuns, imageRun{
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

		lines := collectCellLines(t.builder, cell, cellWidth, fontName, fontSize)

		lineHeight := fontSize * lineSpacing
		ascent := fontSize * 0.8
		for i, text := range lines {
			if text == "" {
				continue
			}
			mcid := ps.NextMCID
			ps.NextMCID++
			sc.MCIDs = append(sc.MCIDs, mcid)
			x := left + leftPad
			y := currentY
			if i == 0 {
				y -= ascent
				currentY -= ascent
			}
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
			ps.TextRuns = append(ps.TextRuns, tr)
			if currentY < bottom+bottomPad {
				break
			}
		}

		row.Cells = append(row.Cells, sc)
		logicalCol2 += span
	}
	if !isHeader {
		t.dataRowIndex++
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
	if t.pageIndex >= len(t.builder.pc.pages) {
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
		h := measureTableCellHeight(t.builder, cell, cellWidth)
		if h > rowHeight {
			rowHeight = h
		}
	}
	if rowHeight <= 0 {
		rowHeight = 16
	}

	rowBottom := t.currentY - rowHeight
	ps := &t.builder.pc.pages[t.pageIndex]

	const fontName = "Helvetica"
	const fontSize = 10.0
	const lineSpacing = 1.2

	logicalCol := 0
	for _, cell := range cells {
		if logicalCol >= t.cols {
			break
		}

		span := cell.ColSpan
		if span <= 0 {
			span = 1
		}
		cellColWidth := cellWidth * float64(span)

		left := t.x + float64(logicalCol)*cellWidth
		topPad, _, bottomPad, leftPad := cell.Style.ResolvedPadding()

		// Draw cell background fill for repeated headers.
		if isHeader && t.HasHeaderFillColor {
			t.builder.FillRect(t.pageIndex, left, rowBottom, cellColWidth, rowHeight, t.HeaderFillColor)
		} else if cell.Style.HasFillColor {
			t.builder.FillRect(t.pageIndex, left, rowBottom, cellColWidth, rowHeight, cell.Style.FillColor)
		}

		currentY := rowBottom + rowHeight - topPad

		lines := collectCellLines(t.builder, cell, cellColWidth, fontName, fontSize)
		lineHeight := fontSize * lineSpacing
		ascent := fontSize * 0.8
		for i, text := range lines {
			if text == "" {
				continue
			}
			y := currentY
			if i == 0 {
				y -= ascent
				currentY -= ascent
			}
			currentY -= lineHeight
			ps.TextRuns = append(ps.TextRuns, textRun{
				Text:            text,
				X:               left + leftPad,
				Y:               y,
				FontName:        fontName,
				FontSize:        fontSize,
				UseDefaultColor: true,
			})
			if currentY < rowBottom+bottomPad {
				break
			}
		}
		logicalCol += span
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

// CurrentY returns the current vertical position of the table.
// Useful for continuing drawing after the table.
func (t *TableBuilder) CurrentY() float64 {
	if t == nil {
		return 0
	}
	return t.currentY
}

// computeImageDimensions calculates the display size of a cell image,
// scaling to fit within the available cell area.
func computeImageDimensions(img *TableCellImageSpec, cellWidth, rowHeight, leftPad, rightPad, topPad, bottomPad float64) (w, h float64) {
	w, h = img.WidthPt, img.HeightPt
	if w > 0 && h > 0 {
		return w, h
	}
	maxWidth := cellWidth - leftPad - rightPad
	if maxWidth <= 0 {
		maxWidth = cellWidth
	}
	scale := maxWidth / float64(img.WidthPx)
	w = float64(img.WidthPx) * scale
	h = float64(img.HeightPx) * scale
	maxHeight := rowHeight - topPad - bottomPad
	if h > maxHeight {
		h = maxHeight
	}
	return w, h
}

// collectCellLines gathers all text lines from a cell (plain text, paragraphs, list items)
// and wraps them to fit the given cellWidth.
func collectCellLines(b *DocumentBuilder, cell TableCellSpec, cellWidth float64, fontName string, fontSize float64) []string {
	_, right, _, left := cell.Style.ResolvedPadding()
	contentWidth := cellWidth - left - right
	if contentWidth <= 0 {
		contentWidth = 0.1
	}

	var allLines []string
	if cell.Text != "" {
		allLines = append(allLines, b.wrapTextLines(cell.Text, fontSize, contentWidth, fontName)...)
	}
	for _, p := range cell.Paragraphs {
		if p != "" {
			allLines = append(allLines, b.wrapTextLines(p, fontSize, contentWidth, fontName)...)
		}
	}
	for i, item := range cell.ListItems {
		if item == "" {
			continue
		}
		prefix := ""
		switch cell.ListKind {
		case "ordered":
			prefix = fmt.Sprintf("%d. ", i+1)
		case "unordered":
			prefix = "\u2022 "
		default:
			prefix = "\u2022 "
		}
		wrapped := b.wrapTextLines(prefix+item, fontSize, contentWidth, fontName)
		allLines = append(allLines, wrapped...)
	}
	return allLines
}
