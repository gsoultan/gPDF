package builder

import (
	"github.com/gsoultan/gpdf/font"
	"github.com/gsoultan/gpdf/model"
)

// PageAccess provides sub-builders with controlled access to page state.
// DocumentBuilder implements this interface to satisfy the Dependency Inversion principle.
type PageAccess interface {
	ValidPageIndex(idx int) bool
	PageCount() int
	PageHeight(pageIndex int) float64
	PageWidth(pageIndex int) float64
	PageAt(pageIndex int) *Page
	AppendPage()
	NextMCID(pageIndex int) int
	FontByName(name string) font.Font
}

// TaggingAccess provides sub-builders with access to tagged PDF structure.
type TaggingAccess interface {
	RecordBlock(pageIndex int, mcid int, role string) int
	RecordFigure(pageIndex int, mcid int, alt string) int
	RecordList(pageIndex int, ordered bool, mcids []int) int
	RecordSectionBlock(blockIndex int)
	RecordSectionFigure(figureIndex int)
	RecordSectionList(listIndex int)
	CurrentSection() int
	IsTagged() bool
	MarkTagged()

	AddTable(pageIndex int) int
	RecordSectionTable(tableIndex int)
	TableAt(tableIndex int) bool
	EnsureTableRow(tableIndex, rowIndex int, rowHeight float64)
	AddTableCell(tableIndex, rowIndex, pageIndex int, role model.Name, scope, alt, lang string, mcids []int)
}
