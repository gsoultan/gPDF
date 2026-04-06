package doc

import (
	"github.com/gsoultan/gpdf/doc/builder"
	"github.com/gsoultan/gpdf/doc/tagged"
	"github.com/gsoultan/gpdf/model"
)

var _ builder.TaggingAccess = (*DocumentBuilder)(nil)

func (b *DocumentBuilder) RecordBlock(pageIndex int, mcid int, role string) int {
	b.tagging.Blocks = append(b.tagging.Blocks, tagged.Block{
		PageIndex: pageIndex, MCID: mcid, Role: model.Name(role),
	})
	idx := len(b.tagging.Blocks) - 1
	b.tagging.RecordSectionBlock(idx)
	return idx
}

func (b *DocumentBuilder) RecordFigure(pageIndex int, mcid int, alt string) int {
	b.tagging.Figures = append(b.tagging.Figures, tagged.Figure{
		PageIndex: pageIndex, MCID: mcid, Alt: alt,
	})
	idx := len(b.tagging.Figures) - 1
	b.tagging.RecordSectionFigure(idx)
	return idx
}

func (b *DocumentBuilder) RecordList(pageIndex int, ordered bool, mcids []int) int {
	list := tagged.List{PageIndex: pageIndex, Ordered: ordered}
	for _, m := range mcids {
		list.Items = append(list.Items, tagged.ListItem{MCID: m})
	}
	b.tagging.Lists = append(b.tagging.Lists, list)
	idx := len(b.tagging.Lists) - 1
	b.tagging.RecordSectionList(idx)
	return idx
}

func (b *DocumentBuilder) RecordSectionBlock(idx int)  { b.tagging.RecordSectionBlock(idx) }
func (b *DocumentBuilder) RecordSectionFigure(idx int) { b.tagging.RecordSectionFigure(idx) }
func (b *DocumentBuilder) RecordSectionList(idx int)   { b.tagging.RecordSectionList(idx) }
func (b *DocumentBuilder) CurrentSection() int         { return b.tagging.CurrentSection }
func (b *DocumentBuilder) IsTagged() bool              { return b.useTagged }
func (b *DocumentBuilder) MarkTagged()                 { b.useTagged = true }

func (b *DocumentBuilder) AddTable(pageIndex int) int {
	b.tagging.Tables = append(b.tagging.Tables, tagged.StructTable{PageIndex: pageIndex})
	return len(b.tagging.Tables) - 1
}

func (b *DocumentBuilder) RecordSectionTable(idx int) { b.tagging.RecordSectionTable(idx) }

func (b *DocumentBuilder) TableAt(tableIndex int) bool {
	return tableIndex >= 0 && tableIndex < len(b.tagging.Tables)
}

func (b *DocumentBuilder) EnsureTableRow(tableIndex, rowIndex int, rowHeight float64) {
	if tableIndex < 0 || tableIndex >= len(b.tagging.Tables) {
		return
	}
	tbl := &b.tagging.Tables[tableIndex]
	if len(tbl.Rows) <= rowIndex {
		tbl.Rows = append(tbl.Rows, tagged.StructRow{})
	}
	if len(tbl.RowHeights) <= rowIndex {
		tbl.RowHeights = append(tbl.RowHeights, rowHeight)
	} else {
		tbl.RowHeights[rowIndex] = rowHeight
	}
}

func (b *DocumentBuilder) AddTableCell(tableIndex, rowIndex, pageIndex int, role model.Name, scope, alt, lang string, mcids []int) {
	if tableIndex < 0 || tableIndex >= len(b.tagging.Tables) {
		return
	}
	tbl := &b.tagging.Tables[tableIndex]
	if rowIndex < 0 || rowIndex >= len(tbl.Rows) {
		return
	}
	tbl.Rows[rowIndex].Cells = append(tbl.Rows[rowIndex].Cells, tagged.StructCell{
		PageIndex: pageIndex, Role: role, Scope: scope, Alt: alt, Lang: lang, MCIDs: mcids,
	})
}
