package reader

import (
	"slices"
	"strings"
)

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
	rowTol  = 4.0 // points: Y values within this range belong to the same row
	colTol  = 6.0 // points: X values within this range belong to the same column
	minCols = 2   // minimum columns to consider a row part of a table
	minRows = 2   // minimum rows to form a table
)

// yBucket rounds y to the nearest rowTol bucket centre.
func yBucket(y float64) int {
	return int((y + rowTol/2) / rowTol)
}

func detectTablesOnPage(pl PageLayout) []Table {
	if len(pl.Blocks) == 0 {
		return nil
	}

	// ── Step 1: group blocks by Y bucket ────────────────────────────────────
	type rowEntry struct {
		yKey   int
		yVal   float64
		blocks []TextBlock
	}
	rowMap := make(map[int]*rowEntry)
	for _, b := range pl.Blocks {
		key := yBucket(b.Y)
		if rowMap[key] == nil {
			rowMap[key] = &rowEntry{yKey: key, yVal: b.Y}
		}
		rowMap[key].blocks = append(rowMap[key].blocks, b)
	}

	// Sort rows top-to-bottom (higher Y = higher on page in PDF coordinates).
	rows := make([]*rowEntry, 0, len(rowMap))
	for _, r := range rowMap {
		rows = append(rows, r)
	}
	slices.SortFunc(rows, func(a, b *rowEntry) int {
		// descending Y (top of page first)
		if a.yVal > b.yVal {
			return -1
		}
		if a.yVal < b.yVal {
			return 1
		}
		return 0
	})

	// ── Step 2: collect candidate rows (≥ minCols blocks) ───────────────────
	type candidateRow struct {
		yVal   float64
		blocks []TextBlock
	}
	var candidates []candidateRow
	for _, r := range rows {
		if len(r.blocks) >= minCols {
			// sort blocks left-to-right within the row
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

	// ── Step 3: build global column anchors from all candidate rows ──────────
	var allX []float64
	for _, row := range candidates {
		for _, b := range row.blocks {
			allX = append(allX, b.X)
		}
	}
	colAnchors := clusterAnchors(allX, colTol)
	if len(colAnchors) < minCols {
		return nil
	}

	// ── Step 4: assign blocks to (row, col) and group into contiguous tables ─
	type cellKey struct{ row, col int }
	cellMap := make(map[cellKey][]string)

	// Build one Table from all candidate rows.
	tbl := Table{
		Page: pl.Page,
		Rows: len(candidates),
		Cols: len(colAnchors),
	}

	minX, minY := 1e9, 1e9
	maxX, maxY := -1e9, -1e9

	for rowIdx, row := range candidates {
		for _, b := range row.blocks {
			minX = min(minX, b.X)
			minY = min(minY, b.Y)
			maxX = max(maxX, b.X+b.Width)
			maxY = max(maxY, b.Y+b.Height)

			colIdx := nearestAnchor(b.X, colAnchors, colTol)
			if colIdx < 0 {
				continue
			}
			k := cellKey{rowIdx, colIdx}
			cellMap[k] = append(cellMap[k], b.Text)
		}
	}

	if len(cellMap) > 0 {
		tbl.X = minX
		tbl.Y = minY
		tbl.Width = maxX - minX
		tbl.Height = maxY - minY
	}
	for k, texts := range cellMap {
		tbl.Cells = append(tbl.Cells, TableCell{
			Row:  k.row,
			Col:  k.col,
			Text: strings.TrimSpace(strings.Join(texts, " ")),
		})
	}
	// Sort cells for deterministic order (row-major).
	slices.SortFunc(tbl.Cells, func(a, b TableCell) int {
		if a.Row != b.Row {
			return a.Row - b.Row
		}
		return a.Col - b.Col
	})

	return []Table{tbl}
}

// clusterAnchors collapses nearby X values into representative column anchors.
func clusterAnchors(xs []float64, tol float64) []float64 {
	if len(xs) == 0 {
		return nil
	}
	sorted := make([]float64, len(xs))
	copy(sorted, xs)
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
