package connector

import (
	"context"
	"database/sql"
	"fmt"
)

// SQLBase provides shared database/sql implementations of Disconnect, Columns, and Rows.
// Embed it in SQL connector structs and call OpenSQL from Connect() to initialise it.
type SQLBase struct {
	db    *sql.DB
	cname string // connector name used in error messages
}

// OpenSQL opens and pings a database connection, returning an initialised SQLBase.
// driverName is passed to sql.Open (e.g. "sqlserver", "mysql", "postgres", "sqlite").
// connectorName is used only in error messages (e.g. "mssql").
func OpenSQL(driverName, dsn, connectorName string) (SQLBase, error) {
	db, err := sql.Open(driverName, dsn)
	if err != nil {
		return SQLBase{}, fmt.Errorf("%s connector: open: %w", connectorName, err)
	}
	if err := db.PingContext(context.Background()); err != nil {
		db.Close()
		return SQLBase{}, fmt.Errorf("%s connector: ping: %w", connectorName, err)
	}
	return SQLBase{db: db, cname: connectorName}, nil
}

// Disconnect closes the database connection.
func (b *SQLBase) Disconnect() error {
	if b.db != nil {
		err := b.db.Close()
		b.db = nil
		return err
	}
	return nil
}

// Columns executes query and returns the column names without fetching any rows.
func (b *SQLBase) Columns(query string) ([]string, error) {
	if b.db == nil {
		return nil, fmt.Errorf("%s connector: not connected", b.cname)
	}
	rows, err := b.db.QueryContext(context.Background(), query)
	if err != nil {
		return nil, fmt.Errorf("%s connector: columns query: %w", b.cname, err)
	}
	defer rows.Close()

	cols, err := rows.Columns()
	if err != nil {
		return nil, fmt.Errorf("%s connector: reading columns: %w", b.cname, err)
	}
	return cols, nil
}

// Rows executes query and streams each row as a []string through the returned channel.
// The channel is closed when all rows have been sent or an error occurs.
// NULL values are rendered as the string "<nil>" via fmt.Sprintf("%v", v).
func (b *SQLBase) Rows(query string) (<-chan []string, error) {
	if b.db == nil {
		return nil, fmt.Errorf("%s connector: not connected", b.cname)
	}
	rows, err := b.db.QueryContext(context.Background(), query)
	if err != nil {
		return nil, fmt.Errorf("%s connector: rows query: %w", b.cname, err)
	}

	cols, err := rows.Columns()
	if err != nil {
		rows.Close()
		return nil, fmt.Errorf("%s connector: reading columns: %w", b.cname, err)
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
