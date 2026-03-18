package reader

// AnalyzedPage groups the richer per-page extraction output used by GenerateCode.
type AnalyzedPage struct {
	Index  int
	Size   PageSize
	Blocks []TextBlock
	Images []ImageInfo
	Tables []Table
	Shapes []VectorShape
}
