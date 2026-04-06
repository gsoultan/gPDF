package doc

import "github.com/gsoultan/gpdf/doc/layer"

// LayerBuilder controls optional content groups (layers).
type LayerBuilder interface {
	BeginLayer(name string, onByDefault bool) *layer.Handle
	DrawInLayer(lh *layer.Handle, pageIndex int, draw func(db *DocumentBuilder)) *DocumentBuilder
}
