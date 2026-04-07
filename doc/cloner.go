package doc

import (
	"github.com/gsoultan/gpdf/model"
)

func (d *builtDocument) addObject(obj model.Object) model.Ref {
	num := d.size
	d.objects[num] = obj
	d.size++
	return model.Ref{ObjectNumber: num}
}

// objectCloner handles cloning objects between documents with remapping of references.
type objectCloner struct {
	src      Document
	dst      *builtDocument
	remapped map[int]int // src obj num -> dst obj num
}

func newObjectCloner(src Document, dst *builtDocument) *objectCloner {
	return &objectCloner{
		src:      src,
		dst:      dst,
		remapped: make(map[int]int),
	}
}

func (c *objectCloner) clone(obj model.Object) (model.Object, error) {
	if obj == nil {
		return nil, nil
	}

	switch v := obj.(type) {
	case model.Ref:
		// If already remapped, return the new reference
		if newNum, ok := c.remapped[v.ObjectNumber]; ok {
			return model.Ref{ObjectNumber: newNum}, nil
		}

		// Resolve the original object
		resolved, err := c.src.Resolve(v)
		if err != nil {
			return nil, err
		}

		// Reserve a new number in the destination
		newNum := c.dst.size
		c.dst.size++
		c.remapped[v.ObjectNumber] = newNum

		// Clone the resolved object and store it
		cloned, err := c.clone(resolved)
		if err != nil {
			return nil, err
		}
		c.dst.objects[newNum] = cloned
		return model.Ref{ObjectNumber: newNum}, nil

	case model.Dict:
		newDict := make(model.Dict, len(v))
		for k, val := range v {
			clonedVal, err := c.clone(val)
			if err != nil {
				return nil, err
			}
			newDict[k] = clonedVal
		}
		return newDict, nil

	case model.Array:
		newArr := make(model.Array, len(v))
		for i, val := range v {
			clonedVal, err := c.clone(val)
			if err != nil {
				return nil, err
			}
			newArr[i] = clonedVal
		}
		return newArr, nil

	case *model.Stream:
		clonedDict, err := c.clone(v.Dict)
		if err != nil {
			return nil, err
		}
		// Shallow copy the content bytes
		newContent := make([]byte, len(v.Content))
		copy(newContent, v.Content)
		return &model.Stream{
			Dict:    clonedDict.(model.Dict),
			Content: newContent,
		}, nil

	default:
		// Primitive types (Boolean, Integer, Real, String, Name, HexString) are copied by value
		return v, nil
	}
}
