package tbl

import (
	"fmt"

	"gpdf/doc/builder"
	"gpdf/doc/style"
	tblspec "gpdf/doc/table"
	"gpdf/model"
)

const (
	defaultBottomMargin = 40.0
	defaultTopMargin    = 40.0
	defaultFontName     = "Helvetica"
	defaultFontSize     = 10.0
	defaultLineSpacing  = 1.2
)

type tableBuilder struct {
	pa       builder.PageAccess
	ta       builder.TaggingAccess
	fillRect FillRectFunc

	pageIndex  int
	x, y       float64
	width      float64
	height     float64
	cols       int
	currentRow int
	tableIndex int

	allowPageBreak bool
	headerCells    [][]tblspec.CellSpec
	currentY       float64
	inPageBreak    bool

	headerFillColor      style.Color
	hasHeaderFillColor   bool
	alternateRowColor    style.Color
	hasAlternateRowColor bool
	dataRowIndex         int
}

func (t *tableBuilder) WithHeaderFillColor(c style.Color) Builder {
	t.headerFillColor = c
	t.hasHeaderFillColor = true
	return t
}

func (t *tableBuilder) WithAlternateRowColor(c style.Color) Builder {
	t.alternateRowColor = c
	t.hasAlternateRowColor = true
	return t
}

func (t *tableBuilder) AllowPageBreak() Builder {
	t.allowPageBreak = true
	return t
}

func (t *tableBuilder) HeaderRow(cells ...tblspec.CellSpec) Builder {
	if t.allowPageBreak && !t.inPageBreak {
		cellsCopy := make([]tblspec.CellSpec, len(cells))
		copy(cellsCopy, cells)
		t.headerCells = append(t.headerCells, cellsCopy)
	}
	return t.addRow(true, cells...)
}

func (t *tableBuilder) Row(cells ...tblspec.CellSpec) Builder {
	return t.addRow(false, cells...)
}

func (t *tableBuilder) addRow(isHeader bool, cells ...tblspec.CellSpec) Builder {
	if len(cells) == 0 || t.cols <= 0 {
		return t
	}
	if t.ta == nil {
		return t
	}
	if !t.ta.TableAt(t.tableIndex) {
		return t
	}

	baseColWidth := t.width / float64(t.cols)
	rowHeight, cellHeights := t.measureRow(cells, baseColWidth)

	if t.allowPageBreak && !t.inPageBreak {
		if t.currentY-rowHeight < defaultBottomMargin {
			t.doPageBreak()
		}
	}

	t.ta.EnsureTableRow(t.tableIndex, t.currentRow, rowHeight)
	rowBottom := t.currentY - rowHeight
	ps := t.pa.PageAt(t.pageIndex)
	if ps == nil {
		return t
	}

	rowFillColor, hasRowFill := t.rowFill(isHeader)

	logicalCol := 0
	for cellIdx, cell := range cells {
		if logicalCol >= t.cols {
			break
		}
		span := cell.ColSpan
		if span <= 0 {
			span = 1
		}
		cellColWidth := baseColWidth * float64(span)
		left := t.x + float64(logicalCol)*baseColWidth

		role := model.Name("TD")
		if isHeader || cell.IsHeader {
			role = model.Name("TH")
		}
		scope := cell.Scope
		if scope == "" && role == model.Name("TH") {
			scope = "Column"
		}

		effFill, effHasFill := rowFillColor, hasRowFill
		if cell.Style.HasFillColor {
			effFill = cell.Style.FillColor
			effHasFill = true
		}
		if effHasFill && t.fillRect != nil {
			t.fillRect(t.pageIndex, left, rowBottom, cellColWidth, rowHeight, effFill)
		}

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
		case tblspec.CellVAlignMiddle:
			offset = freeSpace / 2
		case tblspec.CellVAlignBottom:
			offset = freeSpace
		}

		currentY := rowBottom + rowHeight - topPad - offset
		var cellMCIDs []int

		if cell.Image != nil && len(cell.Image.Raw) > 0 && cell.Image.WidthPx > 0 && cell.Image.HeightPx > 0 {
			w, h := computeImageDimensions(cell.Image, cellColWidth, rowHeight, leftPad, rightPad, topPad, bottomPad)
			imgY := currentY - h
			mcid := t.pa.NextMCID(t.pageIndex)
			cellMCIDs = append(cellMCIDs, mcid)
			ps.ImageRuns = append(ps.ImageRuns, builder.ImageRun{
				X: left + leftPad, Y: imgY,
				WidthPt: w, HeightPt: h,
				Raw: cell.Image.Raw, WidthPx: cell.Image.WidthPx, HeightPx: cell.Image.HeightPx,
				BitsPerComponent: cell.Image.BitsPerComponent, ColorSpace: cell.Image.ColorSpace,
				MCID: mcid, HasMCID: true,
			})
			currentY = imgY - topPad
		}

		lines := collectCellLines(cell)
		lineHeight := defaultFontSize * defaultLineSpacing
		for _, text := range lines {
			if text == "" {
				continue
			}
			mcid := t.pa.NextMCID(t.pageIndex)
			cellMCIDs = append(cellMCIDs, mcid)
			tr := builder.TextRun{
				Text: text, X: left + leftPad, Y: currentY,
				FontName: defaultFontName, FontSize: defaultFontSize,
				MCID: mcid, HasMCID: true, Role: role,
				UseDefaultColor: true,
			}
			if cell.Style.TextColorRGB != ([3]float64{}) {
				tr.TextColorRGB = cell.Style.TextColorRGB
				tr.UseDefaultColor = false
			}
			ps.TextRuns = append(ps.TextRuns, tr)
			currentY -= lineHeight
			if currentY < rowBottom+bottomPad {
				break
			}
		}

		t.ta.AddTableCell(t.tableIndex, t.currentRow, t.pageIndex, role, scope, cell.Alt, cell.Lang, cellMCIDs)
		logicalCol += span
	}
	if !isHeader {
		t.dataRowIndex++
	}
	t.currentY = rowBottom
	t.currentRow++
	return t
}

func (t *tableBuilder) measureRow(cells []tblspec.CellSpec, baseColWidth float64) (float64, []float64) {
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
		h := measureCellHeight(cell, cellWidth)
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
	return rowHeight, cellHeights
}

func (t *tableBuilder) rowFill(isHeader bool) (style.Color, bool) {
	if isHeader && t.hasHeaderFillColor {
		return t.headerFillColor, true
	}
	if !isHeader && t.hasAlternateRowColor && t.dataRowIndex%2 == 1 {
		return t.alternateRowColor, true
	}
	return style.Color{}, false
}

func (t *tableBuilder) doPageBreak() {
	t.pageIndex++
	if t.pageIndex >= t.pa.PageCount() {
		t.pa.AppendPage()
	}
	t.currentY = t.pa.PageHeight(t.pageIndex) - defaultTopMargin

	if len(t.headerCells) > 0 {
		t.inPageBreak = true
		for _, hdr := range t.headerCells {
			t.renderVisualRow(hdr)
		}
		t.inPageBreak = false
	}
}

func (t *tableBuilder) renderVisualRow(cells []tblspec.CellSpec) {
	cellWidth := t.width / float64(t.cols)
	rowHeight := 0.0
	for col, cell := range cells {
		if col >= t.cols {
			break
		}
		h := measureCellHeight(cell, cellWidth)
		if h > rowHeight {
			rowHeight = h
		}
	}
	if rowHeight <= 0 {
		rowHeight = 16
	}
	rowBottom := t.currentY - rowHeight
	ps := t.pa.PageAt(t.pageIndex)
	if ps == nil {
		return
	}

	for col, cell := range cells {
		if col >= t.cols {
			break
		}
		left := t.x + float64(col)*cellWidth
		topPad, _, bottomPad, leftPad := cell.Style.ResolvedPadding()
		currentY := rowBottom + rowHeight - topPad

		lines := collectCellLines(cell)
		lineHeight := defaultFontSize * defaultLineSpacing
		for _, text := range lines {
			if text == "" {
				continue
			}
			ps.TextRuns = append(ps.TextRuns, builder.TextRun{
				Text: text, X: left + leftPad, Y: currentY,
				FontName: defaultFontName, FontSize: defaultFontSize,
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

func measureCellHeight(cell tblspec.CellSpec, cellWidth float64) float64 {
	top, right, bottom, left := cell.Style.ResolvedPadding()
	height := top

	if cell.Image != nil && len(cell.Image.Raw) > 0 && cell.Image.WidthPx > 0 && cell.Image.HeightPx > 0 {
		w, h := cell.Image.WidthPt, cell.Image.HeightPt
		if w <= 0 || h <= 0 {
			maxWidth := cellWidth - left - right
			if maxWidth <= 0 {
				maxWidth = cellWidth
			}
			scale := maxWidth / float64(cell.Image.WidthPx)
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
		height += float64(lines) * defaultFontSize * defaultLineSpacing
	}
	height += bottom
	return height
}

func computeImageDimensions(img *tblspec.CellImageSpec, cellWidth, rowHeight, leftPad, rightPad, topPad, bottomPad float64) (w, h float64) {
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

func collectCellLines(cell tblspec.CellSpec) []string {
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
	return lines
}
