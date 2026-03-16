package mssql

import (
	"fmt"
	"net/url"

	"github.com/mitchellbauer/data-coupler/internal/connector"

	// Register the sqlserver driver.
	_ "github.com/microsoft/go-mssqldb"
)

func init() { connector.Register(&MSSQLConnector{}) }

// MSSQLConnector implements connector.Connector for Microsoft SQL Server.
type MSSQLConnector struct{ connector.SQLBase }

func (c *MSSQLConnector) Name() string { return "mssql" }

// Connect establishes a connection to SQL Server using the provided config.
func (c *MSSQLConnector) Connect(cfg connector.ConnectionConfig) error {
	u := &url.URL{
		Scheme: "sqlserver",
		User:   url.UserPassword(cfg.Username, cfg.Password),
		Host:   fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
	}
	q := u.Query()
	q.Set("database", cfg.Database)
	u.RawQuery = q.Encode()

	base, err := connector.OpenSQL("sqlserver", u.String(), "mssql")
	if err != nil {
		return err
	}
	c.SQLBase = base
	return nil
}
