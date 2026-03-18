package reader

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"sort"
	"strings"
)

// GenerateCode reconstructs an existing PDF into Go source using the public doc builder API.
func GenerateCode(src contentSource, opts CodeGenOptions) (GeneratedCode, error) {
	pages, err := AnalyzePages(src)
	if err != nil {
		return GeneratedCode{}, err
	}
	options := normalizeCodeGenOptions(opts)
	var out strings.Builder
	writeCodeHeader(&out, options)
	writeCodeBody(&out, pages, options)
	return GeneratedCode{GoSource: out.String()}, nil
}

func normalizeCodeGenOptions(opts CodeGenOptions) CodeGenOptions {
	if opts.PackageName == "" {
		opts.PackageName = "main"
	}
	if opts.FunctionName == "" {
		opts.FunctionName = "BuildPDF"
	}
	if !opts.PreservePageSize && !opts.PreservePositions && !opts.PreserveTextStyles && !opts.PreserveTables && !opts.EmbedImages {
		opts.PreservePageSize = true
		opts.PreservePositions = true
		opts.PreserveTextStyles = true
		opts.PreserveTables = true
		opts.EmbedImages = true
	}
	return opts
}

func writeCodeHeader(out *strings.Builder, opts CodeGenOptions) {
	fmt.Fprintf(out, "package %s\n\n", opts.PackageName)
	out.WriteString("import (\n")
	out.WriteString("\t\"encoding/base64\"\n")
	out.WriteString("\t\"gpdf/doc\"\n")
	out.WriteString("\t\"gpdf/doc/style\"\n")
	out.WriteString("\t\"strings\"\n")
	out.WriteString(")\n\n")
	fmt.Fprintf(out, "func %s() *doc.DocumentBuilder {\n", opts.FunctionName)
	out.WriteString("\tb := doc.New()\n")
}

func writeCodeBody(out *strings.Builder, pages []AnalyzedPage, opts CodeGenOptions) {
	for _, page := range pages {
		if opts.PreservePageSize {
			fmt.Fprintf(out, "\tb.PageSize(%s, %s)\n", formatFloat(page.Size.Width), formatFloat(page.Size.Height))
		}
		out.WriteString("\tb.AddPage()\n")
		writeTextBlocks(out, page.Blocks, opts)
		writeImages(out, page.Images, opts)
		writeTables(out, page.Tables, opts)
		out.WriteByte('\n')
	}
	out.WriteString("\treturn b\n")
	out.WriteString("}\n")
}

func writeTextBlocks(out *strings.Builder, blocks []TextBlock, opts CodeGenOptions) {
	ordered := append([]TextBlock(nil), blocks...)
	sort.SliceStable(ordered, func(i, j int) bool {
		if ordered[i].Y != ordered[j].Y {
			return ordered[i].Y > ordered[j].Y
		}
		return ordered[i].X < ordered[j].X
	})
	for _, block := range ordered {
		text := quoteString(block.Text)
		fontName := block.Style.FontName
		if fontName == "" {
			fontName = block.Style.BaseFont
		}
		if fontName == "" {
			fontName = "Helvetica"
		}
		if opts.PreserveTextStyles {
			fmt.Fprintf(out,
				"\tb.DrawTextColored(%s, %s, %s, %s, %s, style.Color{R: %s, G: %s, B: %s})\n",
				text,
				formatFloat(block.X),
				formatFloat(block.Y),
				quoteString(fontName),
				formatFloat(block.Style.FontSize),
				formatFloat(block.Style.ColorR),
				formatFloat(block.Style.ColorG),
				formatFloat(block.Style.ColorB),
			)
			continue
		}
		fmt.Fprintf(out,
			"\tb.DrawText(%s, %s, %s, %s, %s)\n",
			text,
			formatFloat(block.X),
			formatFloat(block.Y),
			quoteString(fontName),
			formatFloat(block.Style.FontSize),
		)
	}
}

func writeImages(out *strings.Builder, images []ImageInfo, opts CodeGenOptions) {
	if !opts.EmbedImages {
		return
	}
	for idx, image := range images {
		dataName := fmt.Sprintf("imgData%d", idx)
		fmt.Fprintf(out, "\t%s, _ := base64.StdEncoding.DecodeString(%s)\n", dataName, joinBase64Chunks(image.Data))
		switch image.Format {
		case "png":
			fmt.Fprintf(out, "\t_ = b.DrawPNG(%s, %s, %s, %s, %s)\n",
				formatFloat(image.X), formatFloat(image.Y), formatFloat(nonZero(image.WidthPt, float64(image.Width))), formatFloat(nonZero(image.HeightPt, float64(image.Height))), dataName)
		case "jpeg":
			fmt.Fprintf(out, "\tb.DrawJPEG(%s, %s, %s, %s, %s, %d, %d, %s)\n",
				formatFloat(image.X), formatFloat(image.Y), formatFloat(nonZero(image.WidthPt, float64(image.Width))), formatFloat(nonZero(image.HeightPt, float64(image.Height))), dataName, image.Width, image.Height, quoteString(image.ColorSpace))
		default:
			fmt.Fprintf(out, "\tb.DrawImage(%s, %s, %s, %s, %s, %d, %d, %d, %s)\n",
				formatFloat(image.X), formatFloat(image.Y), formatFloat(nonZero(image.WidthPt, float64(image.Width))), formatFloat(nonZero(image.HeightPt, float64(image.Height))), dataName, image.Width, image.Height, image.BitsPerComponent, quoteString(image.ColorSpace))
		}
	}
}

func writeTables(out *strings.Builder, tables []Table, opts CodeGenOptions) {
	if !opts.PreserveTables || len(tables) == 0 {
		return
	}
	for _, table := range tables {
		fmt.Fprintf(out, "\t_ = strings.Join([]string{%s}, \" | \")\n", quoteString(tableSummary(table)))
	}
}

func tableSummary(table Table) string {
	if len(table.Cells) == 0 {
		return fmt.Sprintf("table rows=%d cols=%d", table.Rows, table.Cols)
	}
	parts := make([]string, 0, len(table.Cells))
	for _, cell := range table.Cells {
		parts = append(parts, cell.Text)
	}
	return strings.Join(parts, " | ")
}

func quoteString(value string) string {
	return fmt.Sprintf("%q", value)
}

func formatFloat(value float64) string {
	return strings.TrimRight(strings.TrimRight(fmt.Sprintf("%.3f", value), "0"), ".")
}

func nonZero(primary, fallback float64) float64 {
	if primary != 0 {
		return primary
	}
	return fallback
}

func chunkBase64(data []byte) []string {
	encoded := base64.StdEncoding.EncodeToString(data)
	if len(encoded) <= 80 {
		return []string{encoded}
	}
	var chunks []string
	for len(encoded) > 0 {
		n := min(len(encoded), 80)
		chunks = append(chunks, encoded[:n])
		encoded = encoded[n:]
	}
	return chunks
}

func joinBase64Chunks(data []byte) string {
	chunks := chunkBase64(data)
	if len(chunks) == 1 {
		return quoteString(chunks[0])
	}
	var buffer bytes.Buffer
	for i, chunk := range chunks {
		if i > 0 {
			buffer.WriteString(" +\n\t\t")
		}
		buffer.WriteString(quoteString(chunk))
	}
	return buffer.String()
}
