package table

// CellSpec describes a single cell in a tagged table.
type CellSpec struct {
	Text       string
	Paragraphs []string

	ListItems []string
	ListKind  string

	Image *CellImageSpec

	Style CellStyle

	ColSpan  int
	RowSpan  int
	IsHeader bool

	Scope string
	Alt   string
	Lang  string
}
