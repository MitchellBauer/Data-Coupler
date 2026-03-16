package csv

import (
	"bufio"
	"bytes"
	"encoding/csv"
	"fmt"
	"os"

	"github.com/mitchellbauer/data-coupler/internal/connector"
)

// CSVConnector implements connector.Connector for delimited text files.
type CSVConnector struct {
	file    *os.File
	reader  *csv.Reader
	headers []string
}

func (c *CSVConnector) Name() string { return "csv" }

// Connect opens the file at cfg.FilePath and prepares the CSV reader.
func (c *CSVConnector) Connect(cfg connector.ConnectionConfig) error {
	f, err := os.Open(cfg.FilePath)
	if err != nil {
		return fmt.Errorf("csv connector: could not open file: %w", err)
	}

	br := bufio.NewReader(f)
	bom := []byte{0xEF, 0xBB, 0xBF}
	if peek, err := br.Peek(3); err == nil && bytes.Equal(peek, bom) {
		_, _ = br.Discard(3)
	}

	c.file = f
	c.reader = csv.NewReader(br)
	c.headers = nil
	return nil
}

// Disconnect closes the underlying file handle.
func (c *CSVConnector) Disconnect() error {
	if c.file != nil {
		err := c.file.Close()
		c.file = nil
		return err
	}
	return nil
}

// Columns reads and caches the header row. query is unused for CSV.
func (c *CSVConnector) Columns(_ string) ([]string, error) {
	if c.headers != nil {
		return c.headers, nil
	}
	if c.reader == nil {
		return nil, fmt.Errorf("csv connector: not connected")
	}
	headers, err := c.reader.Read()
	if err != nil {
		return nil, fmt.Errorf("csv connector: could not read headers: %w", err)
	}
	c.headers = headers
	return headers, nil
}

// Rows streams remaining data rows through a channel. query is unused for CSV.
// The caller must read the channel to completion to avoid a goroutine leak.
func (c *CSVConnector) Rows(_ string) (<-chan []string, error) {
	if c.reader == nil {
		return nil, fmt.Errorf("csv connector: not connected")
	}
	ch := make(chan []string)
	go func() {
		defer close(ch)
		for {
			record, err := c.reader.Read()
			if err != nil {
				break
			}
			ch <- record
		}
	}()
	return ch, nil
}

func init() { connector.Register(&CSVConnector{}) }

// WriteAll implements connector.Writer. It creates the file at path, writes
// headers as the first row, then writes all rows received from the channel.
func (c *CSVConnector) WriteAll(path string, headers []string, rows <-chan []string) (int, error) {
	f, err := os.Create(path)
	if err != nil {
		return 0, fmt.Errorf("csv: create %s: %w", path, err)
	}
	defer f.Close()

	w := csv.NewWriter(f)
	defer w.Flush()

	if err := w.Write(headers); err != nil {
		return 0, fmt.Errorf("csv: write headers: %w", err)
	}

	count := 0
	for row := range rows {
		if err := w.Write(row); err != nil {
			return count, fmt.Errorf("csv: write row %d: %w", count+1, err)
		}
		count++
	}
	return count, nil
}
