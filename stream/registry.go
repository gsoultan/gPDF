package stream

// Registry is a mutable map of filter name to Filter. It implements FilterRegistry.
type Registry struct {
	m map[string]Filter
}

// NewRegistry returns an empty filter registry. Callers should register filters (e.g. FlateDecode).
func NewRegistry() *Registry {
	return &Registry{m: make(map[string]Filter)}
}

// Get returns the filter for the given name, or nil if not registered.
func (r *Registry) Get(name string) Filter {
	return r.m[name]
}

// Register adds a filter for the given name.
func (r *Registry) Register(name string, f Filter) {
	r.m[name] = f
}
