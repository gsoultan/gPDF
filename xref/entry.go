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
