package types

// IOConfig controls how we read/write the files.
type IOConfig struct {
	SkipHeader bool   `json:"skipHeader"` // If true, first row is treated as headers
	Delimiter  string `json:"delimiter"`  // e.g. "," or ";"
}

// Mapping defines a single column connection.
type Mapping struct {
	InputCol  string `json:"inputCol"`  // The exact header name in the Source CSV
	OutputCol string `json:"outputCol"` // The desired header name in the Destination CSV
	Transform string `json:"transform"` // Placeholder for Phase 2 (e.g. "upper", "date_fmt")
}

// Profile is the master configuration object.
type Profile struct {
	ID          string    `json:"id"`          // Unique filename (e.g. "payroll-export")
	Name        string    `json:"name"`        // Human readable name
	Description string    `json:"description"` // Optional notes
	Settings    IOConfig  `json:"settings"`    // Global file settings
	Mappings    []Mapping `json:"mappings"`    // List of column pairs
}


