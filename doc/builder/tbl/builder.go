package tbl

import (
	"gpdf/doc/builder"
	"gpdf/doc/style"
	tblspec "gpdf/doc/table"
)

// FillRectFunc draws a filled rectangle on a page.
type FillRectFunc func(pageIndex int, x, y, width, height float64, c style.Color)

// Builder is the interface for constructing tagged tables.
type Builder interface {
	WithHeaderFillColor(c style.Color) Builder
	WithAlternateRowColor(c style.Color) Builder
	AllowPageBreak() Builder
	HeaderRow(cells ...tblspec.CellSpec) Builder
	Row(cells ...tblspec.CellSpec) Builder
}

// NewBuilder creates a table builder on the given page region.
func NewBuilder(
	pa builder.PageAccess,
	ta builder.TaggingAccess,
	fillRect FillRectFunc,
	pageIndex int,
	x, y, width, height float64,
	cols int,
	tableIndex int,
) Builder {
	return &tableBuilder{
		pa:         pa,
		ta:         ta,
		fillRect:   fillRect,
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
