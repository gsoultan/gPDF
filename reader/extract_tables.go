package reader

import (
	"slices"
	"strings"
)

type candidateRow struct {
	yVal   float64
	blocks []TextBlock
}

// DetectTables analyses the positioned TextBlocks in each PageLayout and returns
// any table-like grids it finds (one slice of Table per page).
//
// Detection heuristic:
//  1. Group blocks into "rows" by rounding their Y coordinate to a grid.
//  2. Keep only rows that share similar X anchors with at least one other row
//     (i.e. at least 2 rows with ≥ 2 columns aligned within colTol points).
//  3. Build column buckets from the unique X anchors across all candidate rows.
//  4. Assign each block to (row, col) and emit a Table.
func DetectTables(layouts []PageLayout) [][]Table {
	result := make([][]Table, len(layouts))
	for i, pl := range layouts {
		result[i] = detectTablesOnPage(pl)
	}
	return result
}

const (
	colTol  = 6.0 // points: X values within this range belong to the same column
	minCols = 2   // minimum columns to consider a row part of a table
	minRows = 2   // minimum rows to form a table
)

// inferTableBorder searches pl.Shapes for rect or line shapes that coincide
// with tbl's bounding box and returns border styling when found.
func inferTableBorder(tbl Table, shapes []VectorShape) (hasBorder bool, color ColorRGB, width float64) {
	const snapTol = 12.0 // points: how close a shape edge must be to the table edge
	best := -1.0
	for _, s := range shapes {
		if !s.Stroke {
			continue
		}
		switch s.Kind {
		case "rect":
			// Accept rects that surround or closely match the table bounds
			dx1 := s.X1 - tbl.X
			if dx1 < 0 {
				dx1 = -dx1
			}
			dy1 := s.Y1 - tbl.Y
			if dy1 < 0 {
				dy1 = -dy1
			}
			dx2 := s.X2 - (tbl.X + tbl.Width)
			if dx2 < 0 {
				dx2 = -dx2
			}
			dy2 := s.Y2 - (tbl.Y + tbl.Height)
			if dy2 < 0 {
				dy2 = -dy2
			}
			if dx1 <= snapTol && dy1 <= snapTol && dx2 <= snapTol && dy2 <= snapTol {
				score := s.LineWidth
				if score > best {
					best = score
					color = s.StrokeColor
					width = s.LineWidth
					hasBorder = true
				}
			}
		case "line":
			// Accept horizontal or vertical lines that lie within the table area
			inX1 := s.X1 >= tbl.X-snapTol && s.X1 <= tbl.X+tbl.Width+snapTol
			inX2 := s.X2 >= tbl.X-snapTol && s.X2 <= tbl.X+tbl.Width+snapTol
			inY1 := s.Y1 >= tbl.Y-snapTol && s.Y1 <= tbl.Y+tbl.Height+snapTol
			inY2 := s.Y2 >= tbl.Y-snapTol && s.Y2 <= tbl.Y+tbl.Height+snapTol
			if inX1 && inX2 && inY1 && inY2 {
				score := s.LineWidth
				if score > best {
					best = score
					color = s.StrokeColor
					width = s.LineWidth
					hasBorder = true
				}
			}
		}
	}
	return
}

func detectTablesOnPage(pl PageLayout) []Table {
	if len(pl.Blocks) == 0 {
		return nil
	}

	// ── Step 1: group blocks by Y bucket ────────────────────────────────────
	type rowEntry struct {
		yVal   float64
		blocks []TextBlock
	}
	rowMap := make(map[int]*rowEntry)
	for _, b := range pl.Blocks {
		key := yBucket(b.Y)
		if rowMap[key] == nil {
			rowMap[key] = &rowEntry{yVal: b.Y}
		}
		rowMap[key].blocks = append(rowMap[key].blocks, b)
	}

	// Sort rows top-to-bottom
	rows := make([]*rowEntry, 0, len(rowMap))
	for _, r := range rowMap {
		rows = append(rows, r)
	}
	slices.SortFunc(rows, func(a, b *rowEntry) int {
		if a.yVal > b.yVal {
			return -1
		}
		if a.yVal < b.yVal {
			return 1
		}
		return 0
	})

	// ── Step 2: collect candidate rows (≥ minCols blocks) ───────────────────
	var candidates []candidateRow
	for _, r := range rows {
		if len(r.blocks) >= minCols {
			slices.SortFunc(r.blocks, func(a, b TextBlock) int {
				if a.X < b.X {
					return -1
				}
				if a.X > b.X {
					return 1
				}
				return 0
			})
			candidates = append(candidates, candidateRow{yVal: r.yVal, blocks: r.blocks})
		}
	}
	if len(candidates) < minRows {
		return nil
	}

	// ── Step 3: group contiguous rows into separate table candidates ──────
	var tables []Table
	var currentGroup []candidateRow
	for i, row := range candidates {
		if i > 0 {
			// If gap between rows is too large, finalize previous group.
			// Standard line spacing is around FontSize (10-12pts).
			// We use a threshold of 24pts to separate tables.
			if currentGroup[len(currentGroup)-1].yVal-row.yVal > 24.0 {
				if t := extractTableFromGroup(pl.Page, currentGroup); t != nil {
					tables = append(tables, *t)
				}
				currentGroup = nil
			}
		}
		currentGroup = append(currentGroup, row)
	}
	if t := extractTableFromGroup(pl.Page, currentGroup); t != nil {
		tables = append(tables, *t)
	}

	// Cross-reference vector shapes to detect border styling per table
	for i := range tables {
		if hasBorder, color, width := inferTableBorder(tables[i], pl.Shapes); hasBorder {
			tables[i].HasBorder = true
			tables[i].BorderColor = color
			tables[i].BorderWidth = width
		}
	}

	return tables
}

func extractTableFromGroup(page int, rows []candidateRow) *Table {
	if len(rows) < minRows {
		return nil
	}

	// Build anchors from all blocks in this group
	var allX []float64
	for _, r := range rows {
		for _, b := range r.blocks {
			allX = append(allX, b.X)
		}
	}
	anchors := clusterAnchors(allX, colTol)

	// Filter anchors: must be used by at least 2 rows to be a "column"
	usage := make([]int, len(anchors))
	for _, r := range rows {
		used := make(map[int]bool)
		for _, b := range r.blocks {
			if idx := nearestAnchor(b.X, anchors, colTol); idx >= 0 {
				used[idx] = true
			}
		}
		for idx := range used {
			usage[idx]++
		}
	}

	var filtered []float64
	for i, count := range usage {
		if count >= 2 {
			filtered = append(filtered, anchors[i])
		}
	}
	anchors = filtered

	if len(anchors) < minCols {
		return nil
	}

	// Assign cells and check density
	type cellKey struct{ row, col int }
	cellMap := make(map[cellKey][]string)
	styleMap := make(map[cellKey]TextStyle)
	filledCount := 0

	minX, minY := 1e9, 1e9
	maxX, maxY := -1e9, -1e9

	for rIdx, r := range rows {
		for _, b := range r.blocks {
			cIdx := nearestAnchor(b.X, anchors, colTol)
			if cIdx < 0 {
				continue
			}
			k := cellKey{rIdx, cIdx}
			if _, exists := cellMap[k]; !exists {
				filledCount++
				styleMap[k] = b.Style // capture first block's style
			}
			cellMap[k] = append(cellMap[k], b.Text)

			minX = min(minX, b.X)
			minY = min(minY, b.Y)
			maxX = max(maxX, b.X+b.Width)
			maxY = max(maxY, b.Y+b.Height)
		}
	}

	// Final check: do we still have enough valid rows and columns?
	validRows := 0
	for rIdx := range rows {
		rowCellCount := 0
		for cIdx := range anchors {
			if _, ok := cellMap[cellKey{rIdx, cIdx}]; ok {
				rowCellCount++
			}
		}
		if rowCellCount >= minCols {
			validRows++
		}
	}

	if validRows < minRows {
		return nil
	}

	// Density check: filledCount / (rows * cols)
	// Tables are expected to have a reasonable number of filled cells.
	density := float64(filledCount) / float64(len(rows)*len(anchors))
	if density < 0.4 {
		return nil
	}

	// Max column gap check: if any gap between columns is too large, it's likely not a single table.
	for i := 1; i < len(anchors); i++ {
		if anchors[i]-anchors[i-1] > 150.0 { // 150 points is a huge gap for a table
			return nil
		}
	}

	tbl := Table{
		Page:   page,
		Rows:   len(rows),
		Cols:   len(anchors),
		X:      minX,
		Y:      minY,
		Width:  maxX - minX,
		Height: maxY - minY,
	}
	for k, texts := range cellMap {
		tbl.Cells = append(tbl.Cells, TableCell{
			Row:   k.row,
			Col:   k.col,
			Text:  strings.TrimSpace(strings.Join(texts, " ")),
			Style: styleMap[k],
		})
	}
	slices.SortFunc(tbl.Cells, func(a, b TableCell) int {
		if a.Row != b.Row {
			return a.Row - b.Row
		}
		return a.Col - b.Col
	})

	return &tbl
}

// clusterAnchors collapses nearby X values into representative column anchors.
func clusterAnchors(xs []float64, tol float64) []float64 {
	if len(xs) == 0 {
		return nil
	}
	sorted := slices.Clone(xs)
	slices.Sort(sorted)

	var anchors []float64
	cur := sorted[0]
	count := 1
	sum := cur

	for _, x := range sorted[1:] {
		if x-cur <= tol {
			sum += x
			count++
		} else {
			anchors = append(anchors, sum/float64(count))
			cur = x
			sum = x
			count = 1
		}
	}
	anchors = append(anchors, sum/float64(count))
	return anchors
}

// nearestAnchor returns the index of the closest anchor within tol, or -1.
func nearestAnchor(x float64, anchors []float64, tol float64) int {
	best := -1
	bestDist := tol + 1
	for i, a := range anchors {
		d := x - a
		if d < 0 {
			d = -d
		}
		if d <= tol && d < bestDist {
			bestDist = d
			best = i
		}
	}
	return best
}
