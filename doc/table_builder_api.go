package doc

// TableBuilderAPI describes the fluent API for building tagged tables.
type TableBuilderAPI interface {
	BeginTable(pageIndex int, x, y, width, height float64, cols int) ITableBuilder
}
