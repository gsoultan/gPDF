package reader

import (
	"math"
	"slices"
)

// TextBlock is a positioned text fragment extracted from a page content stream.
type TextBlock struct {
	Text   string
	X      float64
	Y      float64
	Width  float64
	Height float64
	Style  TextStyle
}

const (
	rowTol = 4.0 // points: Y values within this range belong to the same row
)

// yBucket rounds y to the nearest rowTol bucket centre.
func yBucket(y float64) int {
	return int((y + rowTol/2) / rowTol)
}

// MergeBlocks combines adjacent or overlapping TextBlocks on the same line
// into single runs of text. This helps avoid false-positive table detection
// and makes layout output more readable.
func MergeBlocks(blocks []TextBlock) []TextBlock {
	if len(blocks) <= 1 {
		return blocks
	}

	// Sort by Y (descending) and then X (ascending)
	sorted := make([]TextBlock, len(blocks))
	copy(sorted, blocks)
	slices.SortStableFunc(sorted, func(a, b TextBlock) int {
		if math.Abs(a.Y-b.Y) > 1.0 { // Line threshold
			if a.Y > b.Y {
				return -1
			}
			return 1
		}
		if a.X < b.X {
			return -1
		}
		if a.X > b.X {
			return 1
		}
		return 0
	})

	merged := make([]TextBlock, 0, len(sorted))
	if len(sorted) == 0 {
		return merged
	}

	current := sorted[0]
	for i := 1; i < len(sorted); i++ {
		next := sorted[i]

		// Conditions to merge:
		// 1. Same line (Y difference <= 1.0)
		// 2. Same style (Font, Size, Color)
		// 3. Close enough (Distance <= current.Style.FontSize * 0.4)

		sameLine := math.Abs(current.Y-next.Y) <= 1.0
		sameStyle := current.Style.FontName == next.Style.FontName &&
			math.Abs(current.Style.FontSize-next.Style.FontSize) < 0.1 &&
			math.Abs(current.Style.ColorR-next.Style.ColorR) < 0.01 &&
			math.Abs(current.Style.ColorG-next.Style.ColorG) < 0.01 &&
			math.Abs(current.Style.ColorB-next.Style.ColorB) < 0.01 &&
			math.Abs(current.Style.CharSpacing-next.Style.CharSpacing) < 0.01 &&
			math.Abs(current.Style.WordSpacing-next.Style.WordSpacing) < 0.01 &&
			math.Abs(current.Style.HorizontalScale-next.Style.HorizontalScale) < 0.5

		dist := next.X - (current.X + current.Width)
		// Standard space is around 0.3 * FontSize.
		// Kerning can be small negative value.
		// Allow up to 0.4 * FontSize to be merged into a single run.
		threshold := current.Style.FontSize * 0.4
		closeEnough := dist > -current.Style.FontSize*0.5 && dist < threshold

		if sameLine && sameStyle && closeEnough {
			// Merge
			if dist > current.Style.FontSize*0.1 {
				// Significant gap -> space
				current.Text += " " + next.Text
			} else {
				current.Text += next.Text
			}
			// Update width
			current.Width = (next.X + next.Width) - current.X
		} else {
			merged = append(merged, current)
			current = next
		}
	}
	merged = append(merged, current)

	return merged
}
