package reader

// LayoutExtractor extracts positioned text blocks and detects tables from page content.
type LayoutExtractor interface {
	Layout() ([]PageLayout, error)
	Tables() ([][]Table, error)
}
