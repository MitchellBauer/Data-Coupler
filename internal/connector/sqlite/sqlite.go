package sqlite

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/mitchellbauer/data-coupler/internal/connector"

	// Register the sqlite driver (CGO-free).
	_ "modernc.org/sqlite"
)

// SQLiteConnector implements connector.Connector for SQLite database files.
type SQLiteConnector struct {
	db *sql.DB
}

func (c *SQLiteConnector) Name() string { return "sqlite" }

// Connect opens the SQLite file at cfg.FilePath. No credentials are required.
func (c *SQLiteConnector) Connect(cfg connector.ConnectionConfig) error {
	db, err := sql.Open("sqlite", cfg.FilePath)
	if err != nil {
		return fmt.Errorf("sqlite connector: open: %w", err)
	}
	if err := db.PingContext(context.Background()); err != nil {
		db.Close()
		return fmt.Errorf("sqlite connector: ping: %w", err)
	}
	c.db = db
	return nil
}

// Disconnect closes the database connection.
func (c *SQLiteConnector) Disconnect() error {
	if c.db != nil {
		err := c.db.Close()
		c.db = nil
		return err
	}
	return nil
}

// Columns executes query and returns the column names without fetching any rows.
func (c *SQLiteConnector) Columns(query string) ([]string, error) {
	if c.db == nil {
		return nil, fmt.Errorf("sqlite connector: not connected")
	}
	rows, err := c.db.QueryContext(context.Background(), query)
	if err != nil {
		return nil, fmt.Errorf("sqlite connector: columns query: %w", err)
	}
	defer rows.Close()

	cols, err := rows.Columns()
	if err != nil {
		return nil, fmt.Errorf("sqlite connector: reading columns: %w", err)
	}
	return cols, nil
}

// Rows executes query and streams each row as a []string through the returned channel.
// The channel is closed when all rows have been sent or an error occurs.
func (c *SQLiteConnector) Rows(query string) (<-chan []string, error) {
	if c.db == nil {
		return nil, fmt.Errorf("sqlite connector: not connected")
	}
	rows, err := c.db.QueryContext(context.Background(), query)
	if err != nil {
		return nil, fmt.Errorf("sqlite connector: rows query: %w", err)
	}

	cols, err := rows.Columns()
	if err != nil {
		rows.Close()
		return nil, fmt.Errorf("sqlite connector: reading columns: %w", err)
	}

	ch := make(chan []string)
	go func() {
		defer close(ch)
		defer rows.Close()

		vals := make([]interface{}, len(cols))
		ptrs := make([]interface{}, len(cols))
		for i := range vals {
			ptrs[i] = &vals[i]
		}

		for rows.Next() {
			if err := rows.Scan(ptrs...); err != nil {
				return
			}
			row := make([]string, len(cols))
			for i, v := range vals {
				row[i] = fmt.Sprintf("%v", v)
			}
			ch <- row
		}
	}()
	return ch, nil
}
