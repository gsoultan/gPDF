package builder

import (
	"gpdf/content"
	"gpdf/model"
)

// GraphicRun holds pre-built content stream operators for one vector drawing operation.
type GraphicRun struct {
	Ops        []content.Op
	ExtGStates map[model.Name]model.Dict
}
