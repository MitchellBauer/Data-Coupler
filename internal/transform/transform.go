package transform

import "sort"

// Transformer is the interface every value transform must satisfy.
type Transformer interface {
	Name() string
	Apply(value string, params map[string]string) (string, error)
}

// RowTransformer is an optional interface for transforms that need the full
// input row (e.g. Concatenate). The engine checks for this and calls ApplyRow
// instead of Apply when it is implemented.
type RowTransformer interface {
	Transformer
	ApplyRow(inputRow []string, headerMap map[string]int, params map[string]string) (string, error)
}

var registry = map[string]Transformer{}

// Register adds a transformer to the global registry, keyed by its name.
func Register(t Transformer) {
	registry[t.Name()] = t
}

// Get retrieves a transformer by name. Returns false if not found.
func Get(name string) (Transformer, bool) {
	t, ok := registry[name]
	return t, ok
}

// List returns the names of all registered transforms, sorted alphabetically.
func List() []string {
	names := make([]string, 0, len(registry))
	for name := range registry {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}
