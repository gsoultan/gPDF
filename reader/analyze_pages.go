package reader

// AnalyzePages collects the richer per-page extraction output used by GenerateCode.
func AnalyzePages(src contentSource) ([]AnalyzedPage, error) {
	pages, err := src.Pages()
	if err != nil {
		return nil, err
	}
	layouts, err := ExtractLayout(src)
	if err != nil {
		return nil, err
	}
	imagesPerPage, err := ExtractImagesPerPage(src)
	if err != nil {
		return nil, err
	}
	tables := DetectTables(layouts)
	analyzed := make([]AnalyzedPage, len(pages))
	for i, page := range pages {
		analyzed[i] = AnalyzedPage{
			Index:  i,
			Size:   resolvePageSize(page),
			Blocks: layouts[i].Blocks,
			Images: imagesPerPage[i],
			Tables: tables[i],
		}
	}
	return analyzed, nil
}
