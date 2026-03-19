package reader

import (
	contentimpl "gpdf/content/impl"
	"gpdf/model"
)

// AnalyzePages collects the richer per-page extraction output used by GenerateCode.
func AnalyzePages(src contentSource) ([]AnalyzedPage, error) {
	opts := normalizeCodeGenOptions(CodeGenOptions{
		EmbedImages:        true,
		PreservePageSize:   true,
		PreserveTables:     true,
		PreserveTextStyles: true,
		PreservePositions:  true,
	})
	analyzed := make([]AnalyzedPage, 0, 16)
	err := AnalyzePagesWithOptions(src, opts, func(page AnalyzedPage) error {
		analyzed = append(analyzed, page)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return analyzed, nil
}

// AnalyzePagesWithOptions parses each page once and invokes visit with extracted page data.
func AnalyzePagesWithOptions(src contentSource, opts CodeGenOptions, visit func(AnalyzedPage) error) error {
	opts = normalizeCodeGenOptions(opts)
	pages, err := src.Pages()
	if err != nil {
		return err
	}
	parser := contentimpl.NewStreamParser()
	for pageIdx, page := range pages {
		resources, _ := page.Resources()
		ops, err := pageContentOps(src, parser, page, opts.MaxDecodedStreamBytes, opts.MaxOpsPerPage)
		if err != nil {
			return err
		}
		blocks := extractBlocksFromOps(
			ops,
			src,
			parser,
			resources,
			make(map[model.Ref]struct{}, 4),
			make(map[model.Ref]*toUnicodeDecoder, 4),
		)

		var images []ImageInfo
		if opts.EmbedImages {
			images = extractImagesFromOps(ops, src, parser, resources, pageIdx, identityMatrix(), make(map[model.Ref]struct{}, 4), opts.MaxImageBytes)
			pageSize := resolvePageSize(page)
			images = InferImageLayouts(images, blocks, pageSize.Width)
		}

		var shapes []VectorShape
		if opts.PreservePositions {
			shapes = sortedShapes(extractVectorsFromOps(src, parser, ops, resources))
		}

		var tables []Table
		pageSize := resolvePageSize(page)
		if opts.PreserveTables {
			layout := PageLayout{
				Page:   pageIdx,
				Width:  pageSize.Width,
				Height: pageSize.Height,
				Blocks: blocks,
			}
			tables = DetectTables([]PageLayout{layout})[0]
		}

		analyzed := AnalyzedPage{
			Index:  pageIdx,
			Size:   pageSize,
			Blocks: blocks,
			Images: images,
			Tables: tables,
			Shapes: shapes,
		}
		if visit != nil {
			if err := visit(analyzed); err != nil {
				return err
			}
		}
	}
	return nil
}
