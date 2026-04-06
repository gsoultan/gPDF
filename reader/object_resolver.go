package reader

import "github.com/gsoultan/gpdf/model"

// ObjectResolver resolves indirect object references and enumerates object numbers.
type ObjectResolver interface {
	Resolve(ref model.Ref) (model.Object, error)
	ObjectNumbers() []int
}
