package mssql

import (
	"context"
	"database/sql"
	"fmt"
	"net/url"

	"github.com/mitchellbauer/data-coupler/internal/connector"

	// Register the sqlserver driver.
	_ "github.com/microsoft/go-mssqldb"
)

// MSSQLConnector implements connector.Connector for Microsoft SQL Server.
type MSSQLConnector struct {
	db *sql.DB
}

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

	db, err := sql.Open("sqlserver", u.String())
	if err != nil {
		return fmt.Errorf("mssql connector: open: %w", err)
	}
	if err := db.PingContext(context.Background()); err != nil {
		db.Close()
		return fmt.Errorf("mssql connector: ping: %w", err)
	}
	c.db = db
	return nil
}

// Disconnect closes the database connection.
func (c *MSSQLConnector) Disconnect() error {
	if c.db != nil {
		err := c.db.Close()
		c.db = nil
		return err
	}
	return nil
}

// Columns executes query and returns the column names without fetching any rows.
func (c *MSSQLConnector) Columns(query string) ([]string, error) {
	if c.db == nil {
		return nil, fmt.Errorf("mssql connector: not connected")
	}
	rows, err := c.db.QueryContext(context.Background(), query)
	if err != nil {
		return nil, fmt.Errorf("mssql connector: columns query: %w", err)
	}
	defer rows.Close()

	cols, err := rows.Columns()
	if err != nil {
		return nil, fmt.Errorf("mssql connector: reading columns: %w", err)
	}
	return cols, nil
}

// Rows executes query and streams each row as a []string through the returned channel.
// The channel is closed when all rows have been sent or an error occurs.
func (c *MSSQLConnector) Rows(query string) (<-chan []string, error) {
	if c.db == nil {
		return nil, fmt.Errorf("mssql connector: not connected")
	}
	rows, err := c.db.QueryContext(context.Background(), query)
	if err != nil {
		return nil, fmt.Errorf("mssql connector: rows query: %w", err)
	}

	cols, err := rows.Columns()
	if err != nil {
		rows.Close()
		return nil, fmt.Errorf("mssql connector: reading columns: %w", err)
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
