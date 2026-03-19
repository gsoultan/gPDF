package font

import (
	"sync"
)

// Registry manages a collection of discovered or registered fonts.
type Registry interface {
	// Register adds a font to the registry.
	Register(f Font)
	// Get returns the font with the given PostScript name, if found.
	Get(name string) (Font, bool)
	// List returns the PostScript names of all registered fonts.
	List() []string
}

type memoryRegistry struct {
	mu    sync.RWMutex
	fonts map[string]Font
}

// NewRegistry returns an empty in-memory font registry.
func NewRegistry() Registry {
	return &memoryRegistry{
		fonts: make(map[string]Font),
	}
}

func (r *memoryRegistry) Register(f Font) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.fonts[f.PostScriptName()] = f
}

func (r *memoryRegistry) Get(name string) (Font, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	f, ok := r.fonts[name]
	return f, ok
}

func (r *memoryRegistry) List() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	res := make([]string, 0, len(r.fonts))
	for name := range r.fonts {
		res = append(res, name)
	}
	return res
}
