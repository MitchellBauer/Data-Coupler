package odbc

import (
	"fmt"

	"github.com/mitchellbauer/data-coupler/internal/connector"

	// Register the ODBC driver (Windows built-in ODBC API via CGO).
	_ "github.com/alexbrainman/odbc"
)

func init() { connector.Register(&ODBCConnector{}) }

// ODBCConnector implements connector.Connector for Windows ODBC data sources.
// cfg.Database holds the DSN name registered in the Windows ODBC Data Source Administrator.
// cfg.Username and cfg.Password are optional credentials.
type ODBCConnector struct{ connector.SQLBase }

func (c *ODBCConnector) Name() string { return "odbc" }

// Connect builds an ODBC connection string from the DSN name and optional credentials,
// then opens and pings the data source.
func (c *ODBCConnector) Connect(cfg connector.ConnectionConfig) error {
	dsn := fmt.Sprintf("DSN=%s", cfg.Database)
	if cfg.Username != "" {
		dsn += fmt.Sprintf(";UID=%s", cfg.Username)
	}
	if cfg.Password != "" {
		dsn += fmt.Sprintf(";PWD=%s", cfg.Password)
	}
	base, err := connector.OpenSQL("odbc", dsn, "odbc")
	if err != nil {
		return err
	}
	c.SQLBase = base
	return nil
}
