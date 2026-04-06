package doc

import (
	"github.com/gsoultan/gpdf/doc/builder"
	"github.com/gsoultan/gpdf/doc/table"
)

type imageRun = builder.ImageRun
type textRun = builder.TextRun

type CellVerticalAlign = table.CellVerticalAlign
type CellHorizontalAlign = table.CellHorizontalAlign
type CellStyle = table.CellStyle

type TableCellSpec = table.CellSpec
type TableCellImageSpec = table.CellImageSpec

const (
	CellVAlignTop    = table.CellVAlignTop
	CellVAlignMiddle = table.CellVAlignMiddle
	CellVAlignBottom = table.CellVAlignBottom

	CellHAlignLeft    = table.CellHAlignLeft
	CellHAlignCenter  = table.CellHAlignCenter
	CellHAlignRight   = table.CellHAlignRight
	CellHAlignJustify = table.CellHAlignJustify
)
