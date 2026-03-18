package reader

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"io"
	"sort"
	"strings"
)

const (
	defaultInlineImageLimit      = 256 * 1024
	defaultMaxDecodedStreamBytes = 256 * 1024 * 1024
	defaultMaxImageBytes         = 32 * 1024 * 1024
	defaultMaxOpsPerPage         = 1_000_000
)

type codeOutput struct {
	out        io.Writer
	assets     []GeneratedAsset
	imageVar   int
	imageAsset int
}

func (out *codeOutput) writeString(value string) error {
	_, err := io.WriteString(out.out, value)
	return err
}

func (out *codeOutput) writef(format string, args ...any) error {
	_, err := fmt.Fprintf(out.out, format, args...)
	return err
}

// GenerateCode reconstructs an existing PDF into Go source using the public doc builder API.
func GenerateCode(src contentSource, opts CodeGenOptions) (GeneratedCode, error) {
	if err := ValidateCodeGenOptions(opts); err != nil {
		return GeneratedCode{}, err
	}
	options := normalizeCodeGenOptions(opts)
	var out strings.Builder
	assets, err := generateCodeTo(src, &out, options)
	if err != nil {
		return GeneratedCode{}, err
	}
	return GeneratedCode{GoSource: out.String(), Assets: assets}, nil
}

// GenerateCodeTo writes generated Go source incrementally to out and returns optional binary assets.
func GenerateCodeTo(src contentSource, out io.Writer, opts CodeGenOptions) ([]GeneratedAsset, error) {
	if err := ValidateCodeGenOptions(opts); err != nil {
		return nil, err
	}
	return generateCodeTo(src, out, normalizeCodeGenOptions(opts))
}

func generateCodeTo(src contentSource, out io.Writer, opts CodeGenOptions) ([]GeneratedAsset, error) {
	writer := &codeOutput{out: out}
	if err := writeCodeHeader(writer, opts); err != nil {
		return nil, err
	}
	err := AnalyzePagesWithOptions(src, opts, func(page AnalyzedPage) error {
		return writeCodePage(writer, page, opts)
	})
	if err != nil {
		return nil, err
	}
	if err := writeCodeFooter(writer); err != nil {
		return nil, err
	}
	return writer.assets, nil
}

func normalizeCodeGenOptions(opts CodeGenOptions) CodeGenOptions {
	if opts.PackageName == "" {
		opts.PackageName = "main"
	}
	if opts.FunctionName == "" {
		opts.FunctionName = "BuildPDF"
	}
	if opts.InlineImageLimit <= 0 {
		opts.InlineImageLimit = defaultInlineImageLimit
	}
	if opts.MaxDecodedStreamBytes <= 0 {
		opts.MaxDecodedStreamBytes = defaultMaxDecodedStreamBytes
	}
	if opts.MaxImageBytes <= 0 {
		opts.MaxImageBytes = defaultMaxImageBytes
	}
	if opts.MaxOpsPerPage <= 0 {
		opts.MaxOpsPerPage = defaultMaxOpsPerPage
	}
	return opts
}

func writeCodeHeader(out *codeOutput, opts CodeGenOptions) error {
	if err := out.writef("package %s\n\n", opts.PackageName); err != nil {
		return err
	}
	if err := out.writeString("import (\n"); err != nil {
		return err
	}
	if err := out.writeString("\t\"encoding/base64\"\n"); err != nil {
		return err
	}
	if err := out.writeString("\t\"fmt\"\n"); err != nil {
		return err
	}
	if err := out.writeString("\t\"gpdf/doc\"\n"); err != nil {
		return err
	}
	if err := out.writeString("\t\"gpdf/doc/style\"\n"); err != nil {
		return err
	}
	if err := out.writeString(")\n\n"); err != nil {
		return err
	}
	if err := out.writeString("var generatedAssets = map[string][]byte{}\n\n"); err != nil {
		return err
	}
	if err := out.writeString("func generatedAsset(name string) ([]byte, error) {\n"); err != nil {
		return err
	}
	if err := out.writeString("\tdata, ok := generatedAssets[name]\n"); err != nil {
		return err
	}
	if err := out.writeString("\tif !ok {\n"); err != nil {
		return err
	}
	if err := out.writeString("\t\treturn nil, fmt.Errorf(\"missing generated asset %q\", name)\n"); err != nil {
		return err
	}
	if err := out.writeString("\t}\n"); err != nil {
		return err
	}
	if err := out.writeString("\treturn data, nil\n"); err != nil {
		return err
	}
	if err := out.writeString("}\n\n"); err != nil {
		return err
	}
	if err := out.writeString("var _ = base64.StdEncoding\n"); err != nil {
		return err
	}
	if err := out.writeString("var _ = fmt.Sprintf\n\n"); err != nil {
		return err
	}
	if err := out.writef("func %s() *doc.DocumentBuilder {\n", opts.FunctionName); err != nil {
		return err
	}
	return out.writeString("\tb := doc.New()\n")
}

func writeCodeFooter(out *codeOutput) error {
	if err := out.writeString("\treturn b\n"); err != nil {
		return err
	}
	return out.writeString("}\n")
}

func writeCodePage(out *codeOutput, page AnalyzedPage, opts CodeGenOptions) error {
	pageIndex := page.Index
	if opts.PreservePageSize {
		if err := out.writef("\tb.PageSize(%s, %s)\n", formatFloat(page.Size.Width), formatFloat(page.Size.Height)); err != nil {
			return err
		}
	}
	if err := out.writeString("\tb.AddPage()\n"); err != nil {
		return err
	}
	if err := writeTextBlocks(out, page.Blocks, opts); err != nil {
		return err
	}
	if err := writeImages(out, page.Images, opts); err != nil {
		return err
	}
	if err := writeShapes(out, page.Shapes, pageIndex); err != nil {
		return err
	}
	if err := writeTables(out, page.Tables, opts, pageIndex); err != nil {
		return err
	}
	return out.writeString("\n")

}

func writeTextBlocks(out *codeOutput, blocks []TextBlock, opts CodeGenOptions) error {
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
			if err := out.writef(
				"\tb.DrawTextColored(%s, %s, %s, %s, %s, style.Color{R: %s, G: %s, B: %s})\n",
				text,
				formatFloat(block.X),
				formatFloat(block.Y),
				quoteString(fontName),
				formatFloat(block.Style.FontSize),
				formatFloat(block.Style.ColorR),
				formatFloat(block.Style.ColorG),
				formatFloat(block.Style.ColorB),
			); err != nil {
				return err
			}
			continue
		}
		if err := out.writef(
			"\tb.DrawText(%s, %s, %s, %s, %s)\n",
			text,
			formatFloat(block.X),
			formatFloat(block.Y),
			quoteString(fontName),
			formatFloat(block.Style.FontSize),
		); err != nil {
			return err
		}
	}
	return nil
}

func writeImages(out *codeOutput, images []ImageInfo, opts CodeGenOptions) error {
	if !opts.EmbedImages {
		return nil
	}
	for _, image := range images {
		if opts.MaxImageBytes > 0 && len(image.Data) > opts.MaxImageBytes {
			continue
		}
		dataName := fmt.Sprintf("imgData%d", out.imageVar)
		out.imageVar++
		if opts.InlineImageLimit > 0 && len(image.Data) > opts.InlineImageLimit {
			assetName := fmt.Sprintf("image_%06d.%s", out.imageAsset, imageExtension(image.Format))
			out.imageAsset++
			out.assets = append(out.assets, GeneratedAsset{Name: assetName, Data: bytes.Clone(image.Data)})
			if err := out.writef("\t%s, err := generatedAsset(%s)\n", dataName, quoteString(assetName)); err != nil {
				return err
			}
		} else {
			if err := out.writef("\t%s, err := base64.StdEncoding.DecodeString(%s)\n", dataName, quoteString(base64.StdEncoding.EncodeToString(image.Data))); err != nil {
				return err
			}
		}
		if err := out.writeString("\tif err != nil {\n"); err != nil {
			return err
		}
		if err := out.writeString("\t\tpanic(fmt.Errorf(\"decode embedded image: %w\", err))\n"); err != nil {
			return err
		}
		if err := out.writeString("\t}\n"); err != nil {
			return err
		}
		switch image.Format {
		case "png":
			if err := out.writef("\t_ = b.DrawPNG(%s, %s, %s, %s, %s)\n",
				formatFloat(image.X), formatFloat(image.Y), formatFloat(nonZero(image.WidthPt, float64(image.Width))), formatFloat(nonZero(image.HeightPt, float64(image.Height))), dataName); err != nil {
				return err
			}
		case "jpeg":
			if err := out.writef("\tb.DrawJPEG(%s, %s, %s, %s, %s, %d, %d, %s)\n",
				formatFloat(image.X), formatFloat(image.Y), formatFloat(nonZero(image.WidthPt, float64(image.Width))), formatFloat(nonZero(image.HeightPt, float64(image.Height))), dataName, image.Width, image.Height, quoteString(image.ColorSpace)); err != nil {
				return err
			}
		default:
			if err := out.writef("\tb.DrawImage(%s, %s, %s, %s, %s, %d, %d, %d, %s)\n",
				formatFloat(image.X), formatFloat(image.Y), formatFloat(nonZero(image.WidthPt, float64(image.Width))), formatFloat(nonZero(image.HeightPt, float64(image.Height))), dataName, image.Width, image.Height, image.BitsPerComponent, quoteString(image.ColorSpace)); err != nil {
				return err
			}
		}
	}
	return nil
}

func imageExtension(format string) string {
	switch format {
	case "png", "jpeg", "jpg", "jpx", "jbig2":
		return format
	default:
		return "bin"
	}
}

func writeShapes(out *codeOutput, shapes []VectorShape, pageIndex int) error {
	for _, shape := range shapes {
		switch shape.Kind {
		case "line":
			if !shape.Stroke {
				continue
			}
			if err := out.writef("\tb.DrawLine(%d, %s, %s, %s, %s, doc.LineStyle{Width: 1, Color: doc.Color{R: %s, G: %s, B: %s}})\n",
				pageIndex,
				formatFloat(shape.X1),
				formatFloat(shape.Y1),
				formatFloat(shape.X2),
				formatFloat(shape.Y2),
				formatFloat(shape.StrokeColor.R),
				formatFloat(shape.StrokeColor.G),
				formatFloat(shape.StrokeColor.B),
			); err != nil {
				return err
			}
		case "rect":
			x := min(shape.X1, shape.X2)
			y := min(shape.Y1, shape.Y2)
			width := max(shape.X1, shape.X2) - x
			height := max(shape.Y1, shape.Y2) - y
			if shape.Stroke && shape.Fill {
				if err := out.writef("\tb.FillStrokeRect(%d, %s, %s, %s, %s, doc.Color{R: %s, G: %s, B: %s}, doc.LineStyle{Width: 1, Color: doc.Color{R: %s, G: %s, B: %s}})\n",
					pageIndex,
					formatFloat(x), formatFloat(y), formatFloat(width), formatFloat(height),
					formatFloat(shape.FillColor.R), formatFloat(shape.FillColor.G), formatFloat(shape.FillColor.B),
					formatFloat(shape.StrokeColor.R), formatFloat(shape.StrokeColor.G), formatFloat(shape.StrokeColor.B),
				); err != nil {
					return err
				}
				continue
			}
			if shape.Fill {
				if err := out.writef("\tb.FillRect(%d, %s, %s, %s, %s, doc.Color{R: %s, G: %s, B: %s})\n",
					pageIndex,
					formatFloat(x), formatFloat(y), formatFloat(width), formatFloat(height),
					formatFloat(shape.FillColor.R), formatFloat(shape.FillColor.G), formatFloat(shape.FillColor.B),
				); err != nil {
					return err
				}
				continue
			}
			if shape.Stroke {
				if err := out.writef("\tb.DrawRect(%d, %s, %s, %s, %s, doc.LineStyle{Width: 1, Color: doc.Color{R: %s, G: %s, B: %s}})\n",
					pageIndex,
					formatFloat(x), formatFloat(y), formatFloat(width), formatFloat(height),
					formatFloat(shape.StrokeColor.R), formatFloat(shape.StrokeColor.G), formatFloat(shape.StrokeColor.B),
				); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func writeTables(out *codeOutput, tables []Table, opts CodeGenOptions, pageIndex int) error {
	if !opts.PreserveTables || len(tables) == 0 {
		return nil
	}
	for _, table := range tables {
		if table.Rows <= 0 || table.Cols <= 0 {
			continue
		}
		if err := out.writeString("\t{\n"); err != nil {
			return err
		}
		if err := out.writef("\t\ttbl := b.BeginTable(%d, 0, 0, 400, 200, %d)\n", pageIndex, table.Cols); err != nil {
			return err
		}
		if err := out.writeString("\t\tif tbl != nil {\n"); err != nil {
			return err
		}
		for row := range table.Rows {
			cells := make([]string, table.Cols)
			for col := range table.Cols {
				cells[col] = fmt.Sprintf("doc.TableCellSpec{Text: %s}", quoteString(table.Cell(row, col)))
			}
			if err := out.writef("\t\t\ttbl.Row(%s)\n", strings.Join(cells, ", ")); err != nil {
				return err
			}
		}
		if err := out.writeString("\t\t\ttbl.EndTable()\n"); err != nil {
			return err
		}
		if err := out.writeString("\t\t}\n"); err != nil {
			return err
		}
		if err := out.writeString("\t}\n"); err != nil {
			return err
		}
	}
	return nil
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
