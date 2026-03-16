package types

// IOConfig controls how a connector reads or writes data.
type IOConfig struct {
	Connector     string `json:"connector"`
	CredentialRef string `json:"credentialRef,omitempty"` // key into credential store (database connectors)
	Host          string `json:"host,omitempty"`          // database server hostname
	Port          int    `json:"port,omitempty"`          // database server port
	Database      string `json:"database,omitempty"`      // database name
	Username      string `json:"username,omitempty"`      // database username
	Query         string `json:"query,omitempty"`         // SQL query (database connectors)
	Path          string `json:"path,omitempty"`          // file path (csv, sqlite connectors)
	Template      string `json:"template,omitempty"`      // e.g., "fishbowl/parts"
}

// Transform describes a single value transformation to apply during column mapping.
type Transform struct {
	Type   string            `json:"type"`
	Params map[string]string `json:"params,omitempty"`
}

// Mapping defines a single column connection.
type Mapping struct {
	InputCol   string      `json:"inputCol"`   // The exact header name in the source
	OutputCol  string      `json:"outputCol"`  // The desired header name in the destination
	Transforms []Transform `json:"transforms"` // Ordered list of transforms to apply
}

// Profile is the master configuration object.
type Profile struct {
	ID          string    `json:"id"`          // Unique filename (e.g. "payroll-export")
	Version     int       `json:"version"`     // Schema version for future migrations
	Name        string    `json:"name"`        // Human readable name
	Description string    `json:"description"` // Optional notes
	Input       IOConfig  `json:"input"`       // Input connector config
	Output      IOConfig  `json:"output"`      // Output connector config
	Mappings    []Mapping `json:"mappings"`    // List of column pairs
}

// AppSettings is persisted to settings.json alongside the binary.
type AppSettings struct {
	LastConnector    string `json:"lastConnector"`
	LastProfilePath  string `json:"lastProfilePath,omitempty"`
	LastInputFolder  string `json:"lastInputFolder,omitempty"`
	LastOutputFolder string `json:"lastOutputFolder,omitempty"`
}
