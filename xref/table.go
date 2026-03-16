package xref

import "sort"

// TableImpl is an in-memory cross-reference table.
type TableImpl struct {
	entries map[int]Entry
	maxNum  int
}

// NewTable returns a new empty cross-reference table.
func NewTable() *TableImpl {
	return &TableImpl{entries: make(map[int]Entry)}
}

// Get returns the entry for the given object number.
func (t *TableImpl) Get(objectNumber int) (Entry, bool) {
	e, ok := t.entries[objectNumber]
	return e, ok
}

// Set records or updates an entry for the given object number.
func (t *TableImpl) Set(objectNumber int, e Entry) {
	t.entries[objectNumber] = e
	if objectNumber > t.maxNum {
		t.maxNum = objectNumber
	}
}

// Size returns the total number of entries. PDF requires Size to be max object number + 1.
func (t *TableImpl) Size() int {
	return t.maxNum + 1
}

// ObjectNumbers returns all object numbers that have in-use entries, sorted.
func (t *TableImpl) ObjectNumbers() []int {
	var nums []int
	for n, e := range t.entries {
		if e.InUse {
			nums = append(nums, n)
		}
	}
	sort.Ints(nums)
	return nums
}
