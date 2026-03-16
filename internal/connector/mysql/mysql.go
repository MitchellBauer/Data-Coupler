package mysql

import (
	"fmt"

	"github.com/mitchellbauer/data-coupler/internal/connector"

	// Register the mysql driver.
	_ "github.com/go-sql-driver/mysql"
)

func init() { connector.Register(&MySQLConnector{}) }

// MySQLConnector implements connector.Connector for MySQL databases.
type MySQLConnector struct{ connector.SQLBase }

func (c *MySQLConnector) Name() string { return "mysql" }

// Connect establishes a connection to MySQL using the provided config.
func (c *MySQLConnector) Connect(cfg connector.ConnectionConfig) error {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?parseTime=true",
		cfg.Username, cfg.Password, cfg.Host, cfg.Port, cfg.Database)
	base, err := connector.OpenSQL("mysql", dsn, "mysql")
	if err != nil {
		return err
	}
	c.SQLBase = base
	return nil
}
