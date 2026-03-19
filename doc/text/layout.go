package text

import (
	"math"
	"strings"
)

// FontWidthFunc measures the width of a string in points at the given font size.
type FontWidthFunc func(s string, fontSize float64) float64

// WrapLines splits text into lines that fit within width using the provided
// width measurement function.
func WrapLines(text string, fontSize, width float64, widthFn FontWidthFunc) []string {
	if width <= 0 {
		return nil
	}
	return WrapLinesDynamic(text, fontSize, widthFn, func(int) float64 { return width })
}

// LineWidthFunc returns the available width for the given line index.
type LineWidthFunc func(lineIdx int) float64

// LineRectFunc returns the X offset and available width for the given line index.
type LineRectFunc func(lineIdx int) (xOffset, width float64)

// WrapLinesDynamic splits text into lines using a potentially different width for each line.
func WrapLinesDynamic(text string, fontSize float64, widthFn FontWidthFunc, lineWidthFn LineWidthFunc) []string {
	return WrapLinesRect(text, fontSize, widthFn, func(lineIdx int) (float64, float64) {
		return 0, lineWidthFn(lineIdx)
	})
}

// WrapLinesRect splits text into lines using a potentially different X offset and width for each line.
func WrapLinesRect(text string, fontSize float64, widthFn FontWidthFunc, lineRectFn LineRectFunc) []string {
	var lines []string
	spaceWidth := widthFn(" ", fontSize)
	currentLineIdx := 0

	for para := range strings.SplitSeq(text, "\n") {
		para = strings.TrimSpace(para)
		if para == "" {
			lines = append(lines, "")
			currentLineIdx++
			continue
		}

		var currentLine string
		var currentWidth float64
		firstWordInPara := true

		for w := range strings.FieldsSeq(para) {
			wordWidth := widthFn(w, fontSize)
			_, targetWidth := lineRectFn(currentLineIdx)
			if targetWidth <= 0 {
				targetWidth = 0.1 // avoid zero or negative widths
			}

			if firstWordInPara {
				currentLine = w
				currentWidth = wordWidth
				firstWordInPara = false
				continue
			}

			if currentWidth+spaceWidth+wordWidth <= targetWidth {
				currentLine += " " + w
				currentWidth += spaceWidth + wordWidth
				continue
			}

			// Current word doesn't fit, finish current line
			lines = append(lines, currentLine)
			currentLineIdx++

			// New line for current word
			currentLine = w
			currentWidth = wordWidth
			// Check if word itself exceeds next targetWidth
			// This could loop if word is wider than any targetWidth,
			// but we'll just put it on its own line for now.
		}
		if currentLine != "" {
			lines = append(lines, currentLine)
			currentLineIdx++
		}
	}
	return lines
}

// ApproxWidth returns a simple approximation of text width in points for base-14 fonts.
func ApproxWidth(s string, fontSize float64) float64 {
	if s == "" || fontSize <= 0 {
		return 0
	}
	var width float64
	for _, r := range s {
		switch {
		case r == ' ':
			width += 0.33
		case r == 'i' || r == 'l' || r == 'I':
			width += 0.4
		case r == 'W' || r == 'M':
			width += 0.9
		default:
			width += 0.6
		}
	}
	return math.Max(width*fontSize, 0)
}
