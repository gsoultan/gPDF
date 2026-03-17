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
	var lines []string
	paragraphs := strings.Split(text, "\n")
	for _, para := range paragraphs {
		para = strings.TrimSpace(para)
		if para == "" {
			lines = append(lines, "")
			continue
		}
		words := strings.Fields(para)
		if len(words) == 0 {
			lines = append(lines, "")
			continue
		}
		var currentLine string
		var currentWidth float64
		spaceWidth := widthFn(" ", fontSize)
		for _, w := range words {
			wordWidth := widthFn(w, fontSize)
			if currentLine == "" {
				currentLine = w
				currentWidth = wordWidth
				continue
			}
			if currentWidth+spaceWidth+wordWidth <= width {
				currentLine += " " + w
				currentWidth += spaceWidth + wordWidth
				continue
			}
			lines = append(lines, currentLine)
			currentLine = w
			currentWidth = wordWidth
		}
		if currentLine != "" {
			lines = append(lines, currentLine)
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
