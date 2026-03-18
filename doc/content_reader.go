package doc

// ContentReader extracts text content from a PDF document.
type ContentReader interface {
	ReadContent() (string, error)
	ReadContentPerPage() ([]string, error)
}
