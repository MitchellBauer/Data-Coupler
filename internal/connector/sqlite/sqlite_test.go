package sqlite

import (
	"database/sql"
	"os"
	"reflect"
	"testing"

	"github.com/mitchellbauer/data-coupler/internal/connector"
	_ "modernc.org/sqlite"
)

// newTempDB creates a temp SQLite file, runs setup SQL, and returns the path.
func newTempDB(t *testing.T, setup ...string) string {
	t.Helper()
	f, err := os.CreateTemp("", "sqlite_test_*.db")
	if err != nil {
		t.Fatalf("could not create temp file: %v", err)
	}
	f.Close()

	db, err := sql.Open("sqlite", f.Name())
	if err != nil {
		t.Fatalf("sql.Open() error: %v", err)
	}
	for _, stmt := range setup {
		if _, err := db.Exec(stmt); err != nil {
			t.Fatalf("setup Exec(%q) error: %v", stmt, err)
		}
	}
	db.Close()

	t.Cleanup(func() { os.Remove(f.Name()) })
	return f.Name()
}

func TestSQLiteConnect_Disconnect(t *testing.T) {
	path := newTempDB(t)
	c := &SQLiteConnector{}
	if err := c.Connect(connector.ConnectionConfig{FilePath: path}); err != nil {
		t.Fatalf("Connect() error: %v", err)
	}
	if err := c.Disconnect(); err != nil {
		t.Errorf("Disconnect() error: %v", err)
	}
}

func TestSQLiteColumns(t *testing.T) {
	path := newTempDB(t,
		"CREATE TABLE items (id INTEGER, name TEXT, price REAL)",
	)
	c := &SQLiteConnector{}
	if err := c.Connect(connector.ConnectionConfig{FilePath: path}); err != nil {
		t.Fatalf("Connect() error: %v", err)
	}
	defer c.Disconnect()

	cols, err := c.Columns("SELECT * FROM items")
	if err != nil {
		t.Fatalf("Columns() error: %v", err)
	}
	want := []string{"id", "name", "price"}
	if !reflect.DeepEqual(cols, want) {
		t.Errorf("Columns() = %v, want %v", cols, want)
	}
}

func TestSQLiteRows(t *testing.T) {
	path := newTempDB(t,
		"CREATE TABLE t (id INTEGER, name TEXT)",
		"INSERT INTO t VALUES (1, 'Alice')",
		"INSERT INTO t VALUES (2, 'Bob')",
	)
	c := &SQLiteConnector{}
	if err := c.Connect(connector.ConnectionConfig{FilePath: path}); err != nil {
		t.Fatalf("Connect() error: %v", err)
	}
	defer c.Disconnect()

	ch, err := c.Rows("SELECT id, name FROM t ORDER BY id")
	if err != nil {
		t.Fatalf("Rows() error: %v", err)
	}

	var got [][]string
	for row := range ch {
		got = append(got, row)
	}

	want := [][]string{{"1", "Alice"}, {"2", "Bob"}}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Rows() = %v, want %v", got, want)
	}
}

func TestSQLiteRows_NullValue(t *testing.T) {
	path := newTempDB(t,
		"CREATE TABLE t (id INTEGER, name TEXT)",
		"INSERT INTO t VALUES (1, NULL)",
	)
	c := &SQLiteConnector{}
	if err := c.Connect(connector.ConnectionConfig{FilePath: path}); err != nil {
		t.Fatalf("Connect() error: %v", err)
	}
	defer c.Disconnect()

	ch, err := c.Rows("SELECT id, name FROM t")
	if err != nil {
		t.Fatalf("Rows() error: %v", err)
	}

	var rows [][]string
	for row := range ch {
		rows = append(rows, row)
	}

	if len(rows) != 1 {
		t.Fatalf("got %d rows, want 1", len(rows))
	}
	if rows[0][1] != "<nil>" {
		t.Errorf("NULL value = %q, want %q", rows[0][1], "<nil>")
	}
}
