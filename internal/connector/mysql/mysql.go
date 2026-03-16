package mysql

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/mitchellbauer/data-coupler/internal/connector"

	// Register the mysql driver.
	_ "github.com/go-sql-driver/mysql"
)

// MySQLConnector implements connector.Connector for MySQL databases.
type MySQLConnector struct {
	db *sql.DB
}

func (c *MySQLConnector) Name() string { return "mysql" }

// Connect establishes a connection to MySQL using the provided config.
func (c *MySQLConnector) Connect(cfg connector.ConnectionConfig) error {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?parseTime=true",
		cfg.Username, cfg.Password, cfg.Host, cfg.Port, cfg.Database)

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return fmt.Errorf("mysql connector: open: %w", err)
	}
	if err := db.PingContext(context.Background()); err != nil {
		db.Close()
		return fmt.Errorf("mysql connector: ping: %w", err)
	}
	c.db = db
	return nil
}

// Disconnect closes the database connection.
func (c *MySQLConnector) Disconnect() error {
	if c.db != nil {
		err := c.db.Close()
		c.db = nil
		return err
	}
	return nil
}

// Columns executes query and returns the column names without fetching any rows.
func (c *MySQLConnector) Columns(query string) ([]string, error) {
	if c.db == nil {
		return nil, fmt.Errorf("mysql connector: not connected")
	}
	rows, err := c.db.QueryContext(context.Background(), query)
	if err != nil {
		return nil, fmt.Errorf("mysql connector: columns query: %w", err)
	}
	defer rows.Close()

	cols, err := rows.Columns()
	if err != nil {
		return nil, fmt.Errorf("mysql connector: reading columns: %w", err)
	}
	return cols, nil
}

// Rows executes query and streams each row as a []string through the returned channel.
// The channel is closed when all rows have been sent or an error occurs.
func (c *MySQLConnector) Rows(query string) (<-chan []string, error) {
	if c.db == nil {
		return nil, fmt.Errorf("mysql connector: not connected")
	}
	rows, err := c.db.QueryContext(context.Background(), query)
	if err != nil {
		return nil, fmt.Errorf("mysql connector: rows query: %w", err)
	}

	cols, err := rows.Columns()
	if err != nil {
		rows.Close()
		return nil, fmt.Errorf("mysql connector: reading columns: %w", err)
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
