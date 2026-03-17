package doc

import (
	"gpdf/content"
	"gpdf/doc/layer"
	"gpdf/model"
)

// BeginLayer starts drawing into a named optional content group (OCG). All
// subsequent drawing operations on the given pageIndex will be wrapped in a
// marked-content block associated with the OCG until EndLayer is called.
// When the layer does not yet exist, it is created and added to the document's
// OCProperties; repeated calls with the same name reuse the same OCG.
func (b *DocumentBuilder) BeginLayer(name string, onByDefault bool) *layer.Handle {
	return b.layers.BeginLayer(name, onByDefault)
}

// DrawInLayer wraps a drawing function so that all content it emits on the
// given page is associated with the specified layer.
func (b *DocumentBuilder) DrawInLayer(lh *layer.Handle, pageIndex int, draw func(db *DocumentBuilder)) *DocumentBuilder {
	if lh == nil || draw == nil {
		return b
	}
	if pageIndex < 0 || pageIndex >= len(b.pages) {
		return b
	}
	ps := &b.pages[pageIndex]

	mcid := ps.nextMCID
	ps.nextMCID++

	ps.graphicRuns = append(ps.graphicRuns, graphicRun{
		ops: []content.Op{
			{
				Name: "BDC",
				Args: []model.Object{
					model.Name("OC"),
					model.Dict{
						model.Name("Name"): model.String(lh.Name),
						model.Name("MCID"): model.Integer(int64(mcid)),
					},
				},
			},
		},
	})

	draw(b)

	ps.graphicRuns = append(ps.graphicRuns, graphicRun{
		ops: []content.Op{
			{Name: "EMC"},
		},
	})
	return b
}
