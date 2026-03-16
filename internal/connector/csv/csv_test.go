package csv

import (
	"os"
	"reflect"
	"testing"

	"github.com/mitchellbauer/data-coupler/internal/connector"
)

func writeTempCSV(t *testing.T, content string) string {
	t.Helper()
	f, err := os.CreateTemp("", "test_*.csv")
	if err != nil {
		t.Fatalf("could not create temp file: %v", err)
	}
	if _, err := f.WriteString(content); err != nil {
		t.Fatalf("could not write temp file: %v", err)
	}
	f.Close()
	t.Cleanup(func() { os.Remove(f.Name()) })
	return f.Name()
}

func TestColumns(t *testing.T) {
	path := writeTempCSV(t, "ID,Name,Email\n1,Alice,a@example.com\n")

	c := &CSVConnector{}
	if err := c.Connect(connector.ConnectionConfig{FilePath: path}); err != nil {
		t.Fatalf("Connect failed: %v", err)
	}
	defer c.Disconnect()

	got, err := c.Columns("")
	if err != nil {
		t.Fatalf("Columns failed: %v", err)
	}

	want := []string{"ID", "Name", "Email"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Columns() = %v, want %v", got, want)
	}
}

func TestRows(t *testing.T) {
	path := writeTempCSV(t, "ID,Name\n1,Alice\n2,Bob\n")

	c := &CSVConnector{}
	if err := c.Connect(connector.ConnectionConfig{FilePath: path}); err != nil {
		t.Fatalf("Connect failed: %v", err)
	}
	defer c.Disconnect()

	// Consume headers first so Rows starts at the data.
	if _, err := c.Columns(""); err != nil {
		t.Fatalf("Columns failed: %v", err)
	}

	ch, err := c.Rows("")
	if err != nil {
		t.Fatalf("Rows failed: %v", err)
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

func TestBOMStripping(t *testing.T) {
	// Write a CSV with a UTF-8 BOM prefix.
	bom := "\xEF\xBB\xBF"
	path := writeTempCSV(t, bom+"ID,Name\n1,Alice\n")

	c := &CSVConnector{}
	if err := c.Connect(connector.ConnectionConfig{FilePath: path}); err != nil {
		t.Fatalf("Connect failed: %v", err)
	}
	defer c.Disconnect()

	headers, err := c.Columns("")
	if err != nil {
		t.Fatalf("Columns failed: %v", err)
	}

	if headers[0] != "ID" {
		t.Errorf("BOM not stripped: first header = %q, want %q", headers[0], "ID")
	}
}
