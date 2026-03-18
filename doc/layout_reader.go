package doc

import "gpdf/reader"

// LayoutReader extracts positioned text blocks and detects tables from page content.
type LayoutReader interface {
	ReadLayout() ([]reader.PageLayout, error)
	ReadTables() ([][]reader.Table, error)
}
