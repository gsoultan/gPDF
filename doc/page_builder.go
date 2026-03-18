package doc

// PageBuilder configures page size and manages pages.
type PageBuilder interface {
	PageSize(width, height float64) *DocumentBuilder
	AddPage() *DocumentBuilder
}
