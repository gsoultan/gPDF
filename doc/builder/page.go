package builder

import "github.com/gsoultan/gpdf/model"

// Page holds a page dict and collected content runs.
type Page struct {
	Dict        model.Dict
	TextRuns    []TextRun
	ImageRuns   []ImageRun
	GraphicRuns []GraphicRun
	NextMCID    int
	NextGSIndex int
	CurrX       float64
	CurrY       float64
}
