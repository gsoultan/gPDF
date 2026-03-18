package stream

// FilterRegistry maps PDF filter names to Filter implementations.
type FilterRegistry interface {
	// Get returns the filter for the given name, or nil if not supported.
	Get(name string) Filter
	// Register adds a filter for the given name.
	Register(name string, f Filter)
}
