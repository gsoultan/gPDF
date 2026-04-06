// readcontent demonstrates Open → ReadContent / ReadImages / ReadLayout / ReadTables → Close.
//
// Usage (run from repo root so "CV CONTOH.pdf" is found):
//
//	go run ./cmd/readcontent
//	go run ./cmd/readcontent "CV CONTOH.pdf"
//	go run ./cmd/readcontent ./path/to/file.pdf
package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/gsoultan/gpdf/doc"
)

func main() {
	path := "CV CONTOH.pdf"
	if len(os.Args) >= 2 {
		path = os.Args[1]
	}

	d, err := doc.Open(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Open %q: %v\n", path, err)
		os.Exit(1)
	}
	defer d.Close()

	// ── Text ────────────────────────────────────────────────────────────────
	text, err := d.ReadContent()
	if err != nil {
		fmt.Fprintf(os.Stderr, "ReadContent: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("=== ReadContent ===")
	if text == "" {
		fmt.Println("(no text extracted)")
	} else {
		const maxPreview = 2000
		preview := text
		if len(preview) > maxPreview {
			preview = preview[:maxPreview] + fmt.Sprintf("\n... (%d more characters)", len(text)-maxPreview)
		}
		fmt.Println(preview)
	}

	// ── Search ──────────────────────────────────────────────────────────────
	keywords := []string{"email", "experience", "skill"}
	results, err := d.Search(keywords...)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Search: %v\n", err)
	} else {
		fmt.Println("\n=== Search ===")
		for _, r := range results {
			if len(r.Pages) == 0 {
				fmt.Printf("%q: not found\n", r.Keyword)
				continue
			}
			fmt.Printf("%q: pages %v indices %v\n", r.Keyword, r.Pages, r.Indices)
		}
	}

	// ── Images ──────────────────────────────────────────────────────────────
	images, err := d.ReadImages()
	if err != nil {
		fmt.Fprintf(os.Stderr, "ReadImages: %v\n", err)
	} else {
		fmt.Printf("\n=== Images (%d total) ===\n", len(images))
		for _, img := range images {
			fmt.Printf("  page=%d name=%-6s %dx%d pos=(%.0f,%.0f) size=(%.0f,%.0f) bpc=%d cs=%-12s filter=%s bytes=%d\n",
				img.Page, img.Name, img.Width, img.Height,
				img.X, img.Y, img.WidthPt, img.HeightPt,
				img.BitsPerComponent, img.ColorSpace, img.Filter, len(img.Data))
		}
		if len(images) == 0 {
			fmt.Println("  (no images found)")
		}
	}

	// ── Layout ──────────────────────────────────────────────────────────────
	layouts, err := d.ReadLayout()
	if err != nil {
		fmt.Fprintf(os.Stderr, "ReadLayout: %v\n", err)
	} else {
		fmt.Printf("\n=== Layout (%d pages) ===\n", len(layouts))
		for _, pl := range layouts {
			fmt.Printf("  page=%d size=%.0fx%.0f blocks=%d\n",
				pl.Page, pl.Width, pl.Height, len(pl.Blocks))
			const maxBlocks = 100
			shown := 0
			for _, b := range pl.Blocks {
				if shown >= maxBlocks {
					fmt.Printf("    ... (%d more blocks)\n", len(pl.Blocks)-maxBlocks)
					break
				}
				preview := b.Text
				if len(preview) > 40 {
					preview = preview[:40] + "…"
				}
				fmt.Printf("    (%.0f,%.0f) font=%-10s size=%.1f rgb=(%.2f,%.2f,%.2f) %q\n",
					b.X, b.Y,
					b.Style.FontName, b.Style.FontSize,
					b.Style.ColorR, b.Style.ColorG, b.Style.ColorB,
					preview)
				shown++
			}
		}
	}

	// ── Tables ──────────────────────────────────────────────────────────────
	tables, err := d.ReadTables()
	if err != nil {
		fmt.Fprintf(os.Stderr, "ReadTables: %v\n", err)
	} else {
		totalTables := 0
		for _, pt := range tables {
			totalTables += len(pt)
		}
		fmt.Printf("\n=== Tables (%d detected) ===\n", totalTables)
		for pageIdx, pageTables := range tables {
			for tIdx, tbl := range pageTables {
				fmt.Printf("  page=%d table=%d rows=%d cols=%d\n",
					pageIdx, tIdx, tbl.Rows, tbl.Cols)
				for r := range tbl.Rows {
					rowCells := make([]string, tbl.Cols)
					for c := range tbl.Cols {
						cell := tbl.Cell(r, c)
						if len(cell) > 20 {
							cell = cell[:20] + "…"
						}
						rowCells[c] = fmt.Sprintf("%-22s", cell)
					}
					fmt.Printf("    row %d: %s\n", r, strings.Join(rowCells, " | "))
				}
			}
		}
		if totalTables == 0 {
			fmt.Println("  (no tables detected)")
		}
	}
}
