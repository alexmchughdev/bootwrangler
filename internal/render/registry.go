package render

import (
	"fmt"
	"sort"
	"sync"
)

var (
	mu        sync.RWMutex
	renderers = map[string]Renderer{}
)

// Register adds a renderer to the global registry.
// It panics if a renderer with the same family name is registered twice.
func Register(r Renderer) {
	mu.Lock()
	defer mu.Unlock()
	if _, exists := renderers[r.Family()]; exists {
		panic(fmt.Sprintf("render: duplicate renderer for family %q", r.Family()))
	}
	renderers[r.Family()] = r
}

// Lookup returns the renderer for the given OS family.
func Lookup(family string) (Renderer, error) {
	mu.RLock()
	defer mu.RUnlock()
	r, ok := renderers[family]
	if !ok {
		return nil, fmt.Errorf("unsupported os family: %s", family)
	}
	return r, nil
}

// Families returns the sorted list of registered OS families.
func Families() []string {
	mu.RLock()
	defer mu.RUnlock()
	families := make([]string, 0, len(renderers))
	for f := range renderers {
		families = append(families, f)
	}
	sort.Strings(families)
	return families
}
