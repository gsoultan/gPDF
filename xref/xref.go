package xref

// Table provides access to the cross-reference table for resolving indirect object locations.
type Table interface {
	Get(objectNumber int) (Entry, bool)
	Set(objectNumber int, e Entry)
	Size() int
	ObjectNumbers() []int
}
