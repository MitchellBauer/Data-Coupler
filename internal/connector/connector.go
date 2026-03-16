package connector

// ConnectionConfig holds the parameters needed to establish any connector connection.
// For CSV connectors, only FilePath is used. Other fields are reserved for Phase 2 database connectors.
type ConnectionConfig struct {
	Host     string
	Port     int
	Database string
	Username string
	Password string
	FilePath string
	Extra    map[string]string
}

// Connector is the interface every data source and destination must satisfy.
type Connector interface {
	Name() string
	Connect(cfg ConnectionConfig) error
	Disconnect() error
	Columns(query string) ([]string, error)
	Rows(query string) (<-chan []string, error)
}
