package sqlite

import (
	"github.com/mitchellbauer/data-coupler/internal/connector"

	// Register the sqlite driver (CGO-free).
	_ "modernc.org/sqlite"
)

func init() { connector.Register(&SQLiteConnector{}) }

// SQLiteConnector implements connector.Connector for SQLite database files.
type SQLiteConnector struct{ connector.SQLBase }

func (c *SQLiteConnector) Name() string { return "sqlite" }

// Connect opens the SQLite file at cfg.FilePath. No credentials are required.
func (c *SQLiteConnector) Connect(cfg connector.ConnectionConfig) error {
	base, err := connector.OpenSQL("sqlite", cfg.FilePath, "sqlite")
	if err != nil {
		return err
	}
	c.SQLBase = base
	return nil
}
