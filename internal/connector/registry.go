package connector

var registry = map[string]Connector{}

// Register adds a connector to the global registry, keyed by its name.
func Register(c Connector) {
	registry[c.Name()] = c
}

// Get retrieves a connector by name. Returns false if not found.
func Get(name string) (Connector, bool) {
	c, ok := registry[name]
	return c, ok
}
