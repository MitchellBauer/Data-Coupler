package postgres

import (
	"fmt"

	"github.com/mitchellbauer/data-coupler/internal/connector"

	// Register the postgres driver.
	_ "github.com/lib/pq"
)

func init() { connector.Register(&PostgreSQLConnector{}) }

// PostgreSQLConnector implements connector.Connector for PostgreSQL databases.
type PostgreSQLConnector struct{ connector.SQLBase }

func (c *PostgreSQLConnector) Name() string { return "postgres" }

// Connect establishes a connection to PostgreSQL using the provided config.
func (c *PostgreSQLConnector) Connect(cfg connector.ConnectionConfig) error {
	dsn := fmt.Sprintf("host=%s port=%d dbname=%s user=%s password=%s sslmode=disable",
		cfg.Host, cfg.Port, cfg.Database, cfg.Username, cfg.Password)
	base, err := connector.OpenSQL("postgres", dsn, "postgres")
	if err != nil {
		return err
	}
	c.SQLBase = base
	return nil
}
