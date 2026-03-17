package doc

import "gpdf/model"

// pageSpec holds a page dict and optional content (text, image, and graphic runs).
type pageSpec struct {
	dict        model.Dict
	textRuns    []textRun
	imageRuns   []imageRun
	graphicRuns []graphicRun

	// nextMCID is the next marked-content ID to assign on this page when tagging content.
	nextMCID int
	// nextGSIndex is a counter for generating unique ExtGState resource names (GS1, GS2, ...).
	nextGSIndex int
}
