package reader

import "gpdf/model"

// ObjectResolver resolves indirect object references and enumerates object numbers.
type ObjectResolver interface {
	Resolve(ref model.Ref) (model.Object, error)
	ObjectNumbers() []int
}
