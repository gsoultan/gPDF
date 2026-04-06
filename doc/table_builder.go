package doc

import (
	"fmt"
	"math"

	"github.com/gsoultan/gpdf/doc/image"
	"github.com/gsoultan/gpdf/doc/style"
	"github.com/gsoultan/gpdf/doc/table"
	"github.com/gsoultan/gpdf/doc/tagged"
	"github.com/gsoultan/gpdf/model"
)

const (
	tableDefaultBottomMargin = 40.0
	tableDefaultTopMargin    = 40.0
)

// ITableBuilder describes the fluent API for building tagged tables.
type ITableBuilder interface {
	WithHeaderFillColor(c Color) ITableBuilder
	WithAlternateRowColor(c Color) ITableBuilder
	WithMargins(top, bottom float64) ITableBuilder
	WithColumnWidths(widths ...float64) ITableBuilder
	WithFlow(f *FlowBuilder) ITableBuilder
	AllowPageBreak() ITableBuilder
	HeaderSpec(cells ...TableCellSpec) ITableBuilder
	RowSpec(cells ...TableCellSpec) ITableBuilder
	FooterRow(cells ...TableCellSpec) ITableBuilder
	EndTable() *DocumentBuilder
	Done() *FlowBuilder
	CurrentY() float64

	// New improved methods
	At(x, y float64) ITableBuilder
	Width(w float64) ITableBuilder
	Header(texts ...string) ITableBuilder
	Row(texts ...string) ITableBuilder
	Draw() *DocumentBuilder
}

// cellLayout contains measurements for a single cell.
type cellLayout struct {
	lines                                []string
	imageHeight                          float64
	imageWidth                           float64
	totalHeight                          float64
	cellWidth                            float64
	fontName                             string
	fontSize                             float64
	ascent                               float64
	lineHeight                           float64
	topPad, rightPad, bottomPad, leftPad float64

	// New fields for image wrapping
	numLinesBesideImage int
	imageSide           style.ImageSide
	imageWrap           style.ImageWrap
}

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
	footerCells    [][]TableCellSpec
	currentY       float64
	inPageBreak    bool

	// Styling options.
	HeaderFillColor      Color
	HasHeaderFillColor   bool
	AlternateRowColor    Color
	HasAlternateRowColor bool
	dataRowIndex         int

	// Margins and column widths.
	topMargin    float64
	bottomMargin float64
	colWidths    []float64

	cachedColWidths    []float64
	cachedFooterHeight float64
	footerHeightValid  bool

	flow *FlowBuilder
}

// BeginTable starts a tagged table region on the given page at (x, y) with the given size and number of columns.
func (b *DocumentBuilder) BeginTable(pageIndex int, x, y, width, height float64, cols int) ITableBuilder {
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
		builder:      b,
		pageIndex:    pageIndex,
		x:            x,
		y:            y,
		width:        width,
		height:       height,
		cols:         cols,
		tableIndex:   tableIndex,
		currentY:     y + height,
		topMargin:    tableDefaultTopMargin,
		bottomMargin: tableDefaultBottomMargin,
	}
}

// WithHeaderFillColor sets a background fill color applied to all header rows.
func (t *TableBuilder) WithHeaderFillColor(c Color) ITableBuilder {
	if t != nil {
		t.HeaderFillColor = c
		t.HasHeaderFillColor = true
	}
	return t
}

// WithAlternateRowColor sets a background fill color applied to every other data row.
func (t *TableBuilder) WithAlternateRowColor(c Color) ITableBuilder {
	if t != nil {
		t.AlternateRowColor = c
		t.HasAlternateRowColor = true
	}
	return t
}

// WithMargins sets the top and bottom margins for page breaks.
func (t *TableBuilder) WithMargins(top, bottom float64) ITableBuilder {
	if t != nil {
		t.topMargin = top
		t.bottomMargin = bottom
	}
	return t
}

// WithColumnWidths sets relative weights for columns.
func (t *TableBuilder) WithColumnWidths(widths ...float64) ITableBuilder {
	if t != nil {
		t.colWidths = widths
		t.cachedColWidths = nil
		t.footerHeightValid = false
	}
	return t
}

// AllowPageBreak enables automatic page breaks when rows exceed available space.
func (t *TableBuilder) AllowPageBreak() ITableBuilder {
	if t != nil {
		t.allowPageBreak = true
	}
	return t
}

// HeaderSpec adds a header row to the table.
func (t *TableBuilder) HeaderSpec(cells ...TableCellSpec) ITableBuilder {
	if t != nil && t.allowPageBreak && !t.inPageBreak {
		cellsCopy := make([]TableCellSpec, len(cells))
		copy(cellsCopy, cells)
		t.headerCells = append(t.headerCells, cellsCopy)
	}
	return t.addRowInternal(true, false, false, cells)
}

// RowSpec adds a data row to the table.
func (t *TableBuilder) RowSpec(cells ...TableCellSpec) ITableBuilder {
	return t.addRowInternal(false, false, false, cells)
}

// FooterRow adds a footer row that repeats on continuation pages.
func (t *TableBuilder) FooterRow(cells ...TableCellSpec) ITableBuilder {
	if t != nil && t.allowPageBreak && !t.inPageBreak {
		cellsCopy := make([]TableCellSpec, len(cells))
		copy(cellsCopy, cells)
		t.footerCells = append(t.footerCells, cellsCopy)
		t.footerHeightValid = false
	}
	return t.addRowInternal(false, true, false, cells)
}

// WithFlow sets the flow builder to be updated after the table is finished.
func (t *TableBuilder) WithFlow(f *FlowBuilder) ITableBuilder {
	if t != nil {
		t.flow = f
	}
	return t
}

// EndTable finishes the table and returns the underlying DocumentBuilder.
func (t *TableBuilder) EndTable() *DocumentBuilder {
	if t == nil {
		return nil
	}
	if t.flow != nil {
		t.flow.pageIndex = t.pageIndex
		t.flow.currY = t.currentY
		t.flow.syncCursor()
	} else {
		// Update page cursor directly if not in flow
		if t.builder.pc.validPageIndex(t.pageIndex) {
			ps := &t.builder.pc.pages[t.pageIndex]
			ps.CurrY = t.currentY
		}
	}
	return t.builder
}

// Done finishes the table and returns the parent FlowBuilder if available.
func (t *TableBuilder) Done() *FlowBuilder {
	t.EndTable()
	return t.flow
}

// CurrentY returns the current vertical position of the table.
func (t *TableBuilder) CurrentY() float64 {
	if t == nil {
		return 0
	}
	return t.currentY
}

// At sets the starting position of the table.
func (t *TableBuilder) At(x, y float64) ITableBuilder {
	if t != nil {
		t.x = x
		t.y = y
		t.currentY = y + t.height
	}
	return t
}

// Width sets the total width of the table.
func (t *TableBuilder) Width(w float64) ITableBuilder {
	if t != nil {
		t.width = w
		t.cachedColWidths = nil
	}
	return t
}

// Header adds a header row with simple text cells.
func (t *TableBuilder) Header(texts ...string) ITableBuilder {
	if t == nil {
		return nil
	}
	cells := make([]TableCellSpec, len(texts))
	for i, txt := range texts {
		cells[i] = TableCellSpec{Text: txt}
	}
	return t.HeaderSpec(cells...)
}

// Row adds a data row with simple text cells.
func (t *TableBuilder) Row(texts ...string) ITableBuilder {
	if t == nil {
		return nil
	}
	cells := make([]TableCellSpec, len(texts))
	for i, txt := range texts {
		cells[i] = TableCellSpec{Text: txt}
	}
	return t.RowSpec(cells...)
}

// Draw ends the table building and returns the DocumentBuilder.
func (t *TableBuilder) Draw() *DocumentBuilder {
	return t.EndTable()
}

func (t *TableBuilder) addRowInternal(isHeader, isFooter, isArtifact bool, cells []TableCellSpec) ITableBuilder {
	if t == nil || t.builder == nil || len(cells) == 0 || t.cols <= 0 {
		return t
	}
	if t.tableIndex < 0 || t.tableIndex >= len(t.builder.tagging.Tables) {
		return t
	}

	colWidths := t.calculateColumnWidths()
	layouts, rowHeight := t.measureRow(cells, colWidths)

	if t.allowPageBreak && !t.inPageBreak && !isArtifact {
		footerH := t.getFooterHeight(colWidths)
		if t.currentY-rowHeight < t.bottomMargin+footerH {
			pageH := t.builder.pageHeight(t.pageIndex)
			isAtTop := t.currentY >= pageH-t.topMargin-0.1
			if !isAtTop {
				t.doPageBreak()
				return t.addRowInternal(isHeader, isFooter, isArtifact, cells)
			}
			// Already at top, must split if it still doesn't fit
			available := t.currentY - t.bottomMargin - footerH
			if rowHeight > available && available > 20 { // 20pt minimum to split
				return t.splitAndAdd(isHeader, isFooter, cells, layouts, colWidths, available)
			}
		}
	}

	t.renderFullRow(isHeader, isFooter, isArtifact, cells, layouts, colWidths, rowHeight)
	return t
}

func (t *TableBuilder) splitAndAdd(isHeader, isFooter bool, cells []TableCellSpec, layouts []cellLayout, colWidths []float64, available float64) ITableBuilder {
	p1Cells := make([]TableCellSpec, len(cells))
	p2Cells := make([]TableCellSpec, len(cells))

	for i, cell := range cells {
		p1, p2 := t.splitCellContent(cell, layouts[i], available)
		p1Cells[i] = p1
		p2Cells[i] = p2
	}

	t.addRowInternal(isHeader, isFooter, false, p1Cells)
	t.doPageBreak()
	return t.addRowInternal(isHeader, isFooter, false, p2Cells)
}

func (t *TableBuilder) splitCellContent(cell TableCellSpec, l cellLayout, available float64) (p1, p2 TableCellSpec) {
	p1, p2 = cell, cell
	allLines := t.collectCellLines(cell, l.cellWidth, l.fontName, l.fontSize, l.imageWidth, l.imageHeight, l.imageWrap, cell.Image)

	availContent := available - l.topPad - l.bottomPad
	if l.imageHeight > 0 {
		if l.imageHeight <= availContent {
			p2.Image = nil
			availContent -= l.imageHeight
		} else {
			p1.Image = nil
		}
	}

	numLines := int(availContent / l.lineHeight)
	if numLines < 0 {
		numLines = 0
	}
	if numLines > len(allLines) {
		numLines = len(allLines)
	}

	p1.Text = ""
	p1.ListItems = nil
	p1.Paragraphs = allLines[:numLines]

	p2.Text = ""
	p2.ListItems = nil
	p2.Paragraphs = allLines[numLines:]
	return
}

func (t *TableBuilder) calculateColumnWidths() []float64 {
	if t.cachedColWidths != nil {
		return t.cachedColWidths
	}
	var res []float64
	if len(t.colWidths) == t.cols {
		total := 0.0
		for _, w := range t.colWidths {
			total += w
		}
		res = make([]float64, t.cols)
		for i, w := range t.colWidths {
			res[i] = (w / total) * t.width
		}
	} else {
		w := t.width / float64(t.cols)
		res = make([]float64, t.cols)
		for i := range res {
			res[i] = w
		}
	}
	t.cachedColWidths = res
	return res
}

func (t *TableBuilder) getFooterHeight(colWidths []float64) float64 {
	if len(t.footerCells) == 0 {
		return 0
	}
	if t.footerHeightValid {
		return t.cachedFooterHeight
	}
	h := 0.0
	for _, footer := range t.footerCells {
		_, rowH := t.measureRow(footer, colWidths)
		h += rowH
	}
	t.cachedFooterHeight = h
	t.footerHeightValid = true
	return h
}

func (t *TableBuilder) measureRow(cells []TableCellSpec, colWidths []float64) ([]cellLayout, float64) {
	layouts := make([]cellLayout, len(cells))
	maxH := 0.0
	logicalCol := 0
	for i, cell := range cells {
		if logicalCol >= t.cols {
			break
		}
		span := cell.ColSpan
		if span <= 0 {
			span = 1
		}
		width := 0.0
		for j := 0; j < span && logicalCol+j < t.cols; j++ {
			width += colWidths[logicalCol+j]
		}
		layouts[i] = t.measureCell(cell, width)
		if layouts[i].totalHeight > maxH {
			maxH = layouts[i].totalHeight
		}
		logicalCol += span
	}
	return layouts, maxH
}

func (t *TableBuilder) measureCell(cell TableCellSpec, width float64) cellLayout {
	name, size := cell.Style.ResolvedFont()
	top, right, bottom, left := cell.Style.ResolvedPadding()
	l := cellLayout{
		fontName: name, fontSize: size,
		topPad: top, rightPad: right, bottomPad: bottom, leftPad: left,
		cellWidth: width,
	}

	if f, ok := t.builder.fc.fonts[name]; ok {
		l.ascent = float64(f.Ascent()) / float64(f.UnitsPerEm()) * size
		l.lineHeight = size * 1.2
	} else {
		l.ascent = size * 0.8
		l.lineHeight = size * 1.2
	}

	img := cell.Image
	var imgW, imgH float64
	if img != nil && len(img.Raw) > 0 {
		if img.WidthPx == 0 {
			if info, err := image.ProcessImage(img.Raw); err == nil {
				img.Raw = info.Raw
				img.WidthPx = info.WidthPx
				img.HeightPx = info.HeightPx
				img.ColorSpace = info.ColorSpace
				img.BitsPerComponent = info.BitsPerComponent
				img.IsJPEG = info.IsJPEG
			}
		}
		imgW, imgH = computeImageDimensions(img, width, 10000, left, right, top, bottom)
		l.imageWidth, l.imageHeight = imgW, imgH
		l.imageSide = img.Side
		l.imageWrap = img.Wrap
	}

	l.lines = t.collectCellLines(cell, width, name, size, imgW, imgH, l.imageWrap, img)

	height := top + bottom
	switch l.imageWrap {
	case style.ImageWrapSquare, style.ImageWrapTight:
		padding := 0.0
		if img != nil {
			padding = img.PaddingPt
		}
		imageAreaHeight := imgH + padding
		l.numLinesBesideImage = int(math.Ceil(imageAreaHeight / l.lineHeight))

		textHeight := float64(len(l.lines)) * l.lineHeight
		height += math.Max(imageAreaHeight, textHeight)
	case style.ImageWrapThrough, style.ImageWrapTopBottom:
		if imgH > 0 {
			height += imgH
		}
		height += float64(len(l.lines)) * l.lineHeight
	default:
		if imgH > 0 {
			height += imgH
		}
		height += float64(len(l.lines)) * l.lineHeight
	}

	l.totalHeight = height
	return l
}

func (t *TableBuilder) renderFullRow(isHeader, isFooter, isArtifact bool, cells []TableCellSpec, layouts []cellLayout, colWidths []float64, rowHeight float64) {
	table := &t.builder.tagging.Tables[t.tableIndex]
	var row *tagged.StructRow
	if !isArtifact {
		if len(table.Rows) <= t.currentRow {
			table.Rows = append(table.Rows, tagged.StructRow{})
		}
		row = &table.Rows[t.currentRow]
		if len(table.RowHeights) <= t.currentRow {
			table.RowHeights = append(table.RowHeights, rowHeight)
		} else {
			table.RowHeights[t.currentRow] = rowHeight
		}
	}

	rowBottom := t.currentY - rowHeight
	fillColor, hasFill := t.getRowFillColor(isHeader)

	logicalCol := 0
	for i, cell := range cells {
		if logicalCol >= t.cols {
			break
		}
		span := cell.ColSpan
		if span <= 0 {
			span = 1
		}
		cellWidth := 0.0
		for j := 0; j < span && logicalCol+j < t.cols; j++ {
			cellWidth += colWidths[logicalCol+j]
		}

		left := t.x
		for j := 0; j < logicalCol; j++ {
			left += colWidths[j]
		}

		t.renderCell(isHeader, isArtifact, row, cell, layouts[i], left, cellWidth, rowHeight, rowBottom, fillColor, hasFill)
		logicalCol += span
	}

	if !isHeader && !isFooter && !isArtifact {
		t.dataRowIndex++
	}
	t.currentY = rowBottom
	if !isArtifact {
		t.currentRow++
	}
}

func (t *TableBuilder) getRowFillColor(isHeader bool) (Color, bool) {
	if isHeader && t.HasHeaderFillColor {
		return t.HeaderFillColor, true
	}
	if !isHeader && t.HasAlternateRowColor && t.dataRowIndex%2 == 1 {
		return t.AlternateRowColor, true
	}
	return Color{}, false
}

func (t *TableBuilder) renderCell(isHeader, isArtifact bool, row *tagged.StructRow, cell TableCellSpec, l cellLayout, left, width, rowHeight, bottom float64, rowFill Color, hasRowFill bool) {
	// 1. Fill background
	effFill, effHasFill := rowFill, hasRowFill
	if cell.Style.HasFillColor {
		effFill, effHasFill = cell.Style.FillColor, true
	}
	if effHasFill {
		t.builder.FillRect(t.pageIndex, left, bottom, width, rowHeight, effFill)
	}

	// 2. Setup tagging
	var sc *tagged.StructCell
	role := model.Name("TD")
	if isHeader || cell.IsHeader {
		role = model.Name("TH")
	}
	if !isArtifact && row != nil {
		scope := cell.Scope
		if scope == "" && role == model.Name("TH") {
			scope = "Column"
		}
		sc = &tagged.StructCell{
			PageIndex: t.pageIndex,
			Role:      role,
			Scope:     scope,
			Alt:       cell.Alt,
			Lang:      cell.Lang,
		}
	}

	// 3. Calculate alignment offset
	free := rowHeight - l.totalHeight
	if free < 0 {
		free = 0
	}
	var offset float64
	switch cell.Style.VAlign {
	case CellVAlignMiddle:
		offset = free / 2
	case CellVAlignBottom:
		offset = free
	}

	currY := bottom + rowHeight - l.topPad - offset
	startY := currY
	ps := &t.builder.pc.pages[t.pageIndex]

	// 4. Render image
	imageX := left + l.leftPad
	cellContentWidth := width - l.leftPad - l.rightPad
	paddingPt := 0.0
	if cell.Image != nil {
		paddingPt = cell.Image.PaddingPt
	}

	if cell.Image != nil && l.imageHeight > 0 {
		mcid := 0
		if !isArtifact {
			mcid = ps.NextMCID
			ps.NextMCID++
			if sc != nil {
				sc.MCIDs = append(sc.MCIDs, mcid)
			}
		}

		if l.imageWrap == style.ImageWrapSquare || l.imageWrap == style.ImageWrapTight {
			if l.imageSide == style.ImageSideRight {
				imageX = left + width - l.rightPad - l.imageWidth
			} else {
				imageX = left + l.leftPad
			}
		} else {
			switch cell.Style.HAlign {
			case CellHAlignCenter:
				if cellContentWidth > l.imageWidth {
					imageX = left + l.leftPad + (cellContentWidth-l.imageWidth)/2
				}
			case CellHAlignRight:
				if cellContentWidth > l.imageWidth {
					imageX = left + width - l.rightPad - l.imageWidth
				}
			}
		}

		ps.ImageRuns = append(ps.ImageRuns, imageRun{
			X: imageX, Y: currY - l.imageHeight,
			WidthPt: l.imageWidth, HeightPt: l.imageHeight,
			Raw: cell.Image.Raw, WidthPx: cell.Image.WidthPx, HeightPx: cell.Image.HeightPx,
			BitsPerComponent: cell.Image.BitsPerComponent, ColorSpace: cell.Image.ColorSpace,
			IsJPEG: cell.Image.IsJPEG,
			MCID:   mcid, HasMCID: !isArtifact, IsArtifact: isArtifact,
		})
		if l.imageWrap == style.ImageWrapTopBottom || l.imageWrap == style.ImageWrapThrough {
			currY -= l.imageHeight
		}
	}

	// 5. Render lines
	for i, txt := range l.lines {
		if txt == "" {
			// Even for empty lines, we should advance currY if it's not the first line or if it represents a paragraph break.
			if i > 0 {
				currY -= l.lineHeight
			}
			continue
		}
		mcid := 0
		if !isArtifact {
			mcid = ps.NextMCID
			ps.NextMCID++
			if sc != nil {
				sc.MCIDs = append(sc.MCIDs, mcid)
			}
		}
		y := currY
		if i == 0 {
			y -= l.ascent
			currY -= l.ascent
		}
		currY -= l.lineHeight

		// Horizontal alignment and wrapping adjustment
		lineX := left + l.leftPad
		effCellContentWidth := cellContentWidth
		wordSpacing := 0.0

		if (l.imageWrap == style.ImageWrapSquare || l.imageWrap == style.ImageWrapTight) && l.imageHeight > 0 {
			if i < l.numLinesBesideImage {
				effCellContentWidth = cellContentWidth - l.imageWidth - paddingPt
				if l.imageSide == style.ImageSideLeft {
					lineX = left + l.leftPad + l.imageWidth + paddingPt
				}
			} else if i == l.numLinesBesideImage {
				// If the image was taller than the lines beside it, we might need to adjust currY
				// but in our measurement we already account for the max(imageAreaHeight, textHeight).
				// So if we are here, we are below the wrapped lines.
				imageAreaHeight := l.imageHeight + paddingPt
				expectedY := startY - imageAreaHeight
				if currY+l.lineHeight > expectedY { // currY already subtracted l.lineHeight
					// This line should start after the image area
					y = expectedY - l.ascent
					currY = expectedY - l.lineHeight
				}
			}
		}

		lineWidth := t.builder.textWidth(txt, l.fontSize, l.fontName)

		switch cell.Style.HAlign {
		case CellHAlignCenter:
			if effCellContentWidth > lineWidth {
				lineX += (effCellContentWidth - lineWidth) / 2
			}
		case CellHAlignRight:
			if effCellContentWidth > lineWidth {
				lineX += effCellContentWidth - lineWidth
			}
		case CellHAlignJustify:
			isLastLine := i == len(l.lines)-1
			if !isLastLine && effCellContentWidth > lineWidth {
				numSpaces := countWordSpaces(txt)
				if numSpaces > 0 {
					wordSpacing = (effCellContentWidth - lineWidth) / float64(numSpaces)
				}
			}
		}

		tr := textRun{
			Text: txt, X: lineX, Y: y,
			FontName: l.fontName, FontSize: l.fontSize,
			MCID: mcid, HasMCID: !isArtifact, IsArtifact: isArtifact,
			Role: role, UseDefaultColor: true,
			WordSpacing: wordSpacing,
		}
		if cell.Style.HasTextColor {
			tr.TextColorRGB = [3]float64{cell.Style.TextColor.R, cell.Style.TextColor.G, cell.Style.TextColor.B}
			tr.UseDefaultColor = false
		} else if cell.Style.TextColorRGB != ([3]float64{}) {
			tr.TextColorRGB = cell.Style.TextColorRGB
			tr.UseDefaultColor = false
		}
		ps.TextRuns = append(ps.TextRuns, tr)
	}

	if sc != nil {
		row.Cells = append(row.Cells, *sc)
	}
}

func (t *TableBuilder) doPageBreak() {
	// Footers
	if len(t.footerCells) > 0 {
		t.inPageBreak = true
		for _, footer := range t.footerCells {
			t.addRowInternal(false, true, true, footer)
		}
		t.inPageBreak = false
	}

	t.pageIndex++
	if t.pageIndex >= len(t.builder.pc.pages) {
		t.builder.AddPage()
	}
	t.currentY = t.builder.pageHeight(t.pageIndex) - t.topMargin

	// Headers
	if len(t.headerCells) > 0 {
		t.inPageBreak = true
		for _, hdr := range t.headerCells {
			t.addRowInternal(true, false, true, hdr)
		}
		t.inPageBreak = false
	}
}

func (t *TableBuilder) collectCellLines(cell TableCellSpec, cellWidth float64, fontName string, fontSize float64, imgW, imgH float64, wrap style.ImageWrap, img *table.CellImageSpec) []string {
	_, right, _, left := cell.Style.ResolvedPadding()
	contentWidth := cellWidth - left - right
	if contentWidth <= 0 {
		contentWidth = 0.1
	}

	lineHeight := fontSize * 1.2
	paddingPt := 0.0
	if img != nil {
		paddingPt = img.PaddingPt
	}

	lineWidthFn := func(lineIdx int) float64 {
		if (wrap == style.ImageWrapSquare || wrap == style.ImageWrapTight) && imgW > 0 {
			if float64(lineIdx)*lineHeight < imgH+paddingPt {
				return contentWidth - imgW - paddingPt
			}
		}
		return contentWidth
	}

	var allLines []string
	collect := func(textStr string) {
		if textStr == "" {
			return
		}
		// We need to account for existing lines when calculating line index for lineWidthFn
		currentLineOffset := len(allLines)
		wrapped := t.builder.wrapTextLinesDynamic(textStr, fontSize, fontName, func(idx int) float64 {
			return lineWidthFn(currentLineOffset + idx)
		})
		allLines = append(allLines, wrapped...)
	}

	if cell.Text != "" {
		collect(cell.Text)
	}
	for _, p := range cell.Paragraphs {
		collect(p)
	}
	for i, item := range cell.ListItems {
		if item == "" {
			continue
		}
		prefix := "\u2022 "
		if cell.ListKind == "ordered" {
			prefix = fmt.Sprintf("%d. ", i+1)
		}
		collect(prefix + item)
	}
	return allLines
}

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
