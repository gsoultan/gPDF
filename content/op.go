package content

import "github.com/gsoultan/gpdf/model"

// Op represents a single PDF content stream operator: name plus operands (PDF objects).
// Operands are in order: e.g. for "1 0 0 1 0 0 cm", Name is "cm", Args is six numbers.
type Op struct {
	Name string
	Args []model.Object
}
