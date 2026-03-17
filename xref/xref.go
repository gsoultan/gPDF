package xref

// Entry represents one cross-reference table entry (offset, generation, in-use flag).
// For type-2 (compressed) entries, Compressed is true and StreamObjectNumber/IndexInStream
// identify the containing object stream and the object's index within it.
type Entry struct {
	Offset             int64
	Generation         int
	InUse              bool
	Compressed         bool
	StreamObjectNumber int
	IndexInStream      int
}

// Table provides access to the cross-reference table for resolving indirect object locations.
type Table interface {
	// Get returns the entry for the given object number, if present.
	Get(objectNumber int) (Entry, bool)
	// Set records or updates an entry for the given object number.
	Set(objectNumber int, e Entry)
	// Size returns the total number of entries (as required by trailer /Size).
	Size() int
	// ObjectNumbers returns all object numbers with in-use entries (for iteration).
	ObjectNumbers() []int
}
