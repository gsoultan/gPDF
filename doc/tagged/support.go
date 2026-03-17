package tagged

import "gpdf/model"

// Block describes one block-level tagged element (e.g. P, H1..H6) in the structure tree.
// It refers to a single marked-content sequence on a page via MCID.
type Block struct {
	PageIndex int
	MCID      int
	Role      model.Name

	Lang string
	Alt  string
}

// List represents a logical list (/L) on a page.
// Items correspond to /LI elements, each pointing to one marked-content sequence via MCID.
type List struct {
	PageIndex int
	Ordered   bool
	Items     []ListItem
}

// ListItem represents one item in a tagged list.
type ListItem struct {
	MCID int
}

// Figure describes a tagged image (Figure) with alternative text for accessibility.
type Figure struct {
	PageIndex int
	MCID      int
	Alt       string
}

// Section groups block, figure, table, and list indices for a Sect structure element.
type Section struct {
	BlockIndices  []int
	FigureIndices []int
	TableIndices  []int
	ListIndices   []int
}

// StructCell describes one logical table cell in the structure tree.
type StructCell struct {
	PageIndex int
	MCIDs     []int
	Role      model.Name

	Scope string
	Alt   string
	Lang  string
}

// StructRow describes one logical table row (TR) in the structure tree.
type StructRow struct {
	Cells []StructCell
}

// StructTable describes one logical table (Table) in the structure tree.
type StructTable struct {
	PageIndex  int
	Rows       []StructRow
	RowHeights []float64
}

// Support owns tagged-structure-related state for a document builder.
type Support struct {
	Tables         []StructTable
	Blocks         []Block
	Lists          []List
	Figures        []Figure
	Sections       []Section
	CurrentSection int
}

// RecordSectionBlock records a block index in the current section.
func (s *Support) RecordSectionBlock(blockIndex int) {
	if s.CurrentSection < 0 || s.CurrentSection >= len(s.Sections) {
		return
	}
	s.Sections[s.CurrentSection].BlockIndices = append(s.Sections[s.CurrentSection].BlockIndices, blockIndex)
}

// RecordSectionFigure records a figure index in the current section.
func (s *Support) RecordSectionFigure(figureIndex int) {
	if s.CurrentSection < 0 || s.CurrentSection >= len(s.Sections) {
		return
	}
	s.Sections[s.CurrentSection].FigureIndices = append(s.Sections[s.CurrentSection].FigureIndices, figureIndex)
}

// RecordSectionTable records a table index in the current section.
func (s *Support) RecordSectionTable(tableIndex int) {
	if s.CurrentSection < 0 || s.CurrentSection >= len(s.Sections) {
		return
	}
	s.Sections[s.CurrentSection].TableIndices = append(s.Sections[s.CurrentSection].TableIndices, tableIndex)
}

// RecordSectionList records a list index in the current section.
func (s *Support) RecordSectionList(listIndex int) {
	if s.CurrentSection < 0 || s.CurrentSection >= len(s.Sections) {
		return
	}
	s.Sections[s.CurrentSection].ListIndices = append(s.Sections[s.CurrentSection].ListIndices, listIndex)
}
