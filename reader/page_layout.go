package reader

// PageLayout groups all positioned text blocks extracted from a single page,
// together with the page dimensions.
type PageLayout struct {
	Page   int
	Width  float64
	Height float64
	Blocks []TextBlock
}
