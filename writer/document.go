package writer

import "github.com/gsoultan/gpdf/model"

// Document is the input to the writer: object graph and trailer (root, size).
type Document interface {
	Trailer() model.Trailer
	Resolve(ref model.Ref) (model.Object, error)
	ObjectNumbers() []int
}
