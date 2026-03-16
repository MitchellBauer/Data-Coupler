package audit

import (
	"bufio"
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// AuditEntry records a single conversion run.
type AuditEntry struct {
	Timestamp   time.Time `json:"timestamp"`
	ProfileName string    `json:"profileName"`
	ProfileID   string    `json:"profileId"`
	InputSource string    `json:"inputSource"` // e.g. "csv: input.csv" or "mssql: host"
	OutputPath  string    `json:"outputPath"`
	RowsOut     int       `json:"rowsOut"`
	DurationMs  int64     `json:"durationMs"`
	Error       string    `json:"error,omitempty"`
}

var (
	logPath string
	mu      sync.Mutex
)

// SetLogPath sets the directory in which audit.log will be written.
// Must be called once at startup before any AppendEntry calls.
func SetLogPath(dir string) {
	mu.Lock()
	defer mu.Unlock()
	logPath = filepath.Join(dir, "audit.log")
}

// AppendEntry marshals e as a single JSON line and appends it to the log file.
// It opens the file in append mode each time so concurrent processes are safe.
func AppendEntry(e AuditEntry) error {
	mu.Lock()
	defer mu.Unlock()

	if logPath == "" {
		return nil // no-op if SetLogPath was never called
	}

	line, err := json.Marshal(e)
	if err != nil {
		return err
	}

	f, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = f.Write(append(line, '\n'))
	return err
}

// TrimLog ensures the log file has at most maxEntries lines, keeping the newest.
// If the file is missing or has fewer lines than maxEntries, it does nothing.
func TrimLog(maxEntries int) error {
	mu.Lock()
	defer mu.Unlock()

	if logPath == "" {
		return nil
	}

	f, err := os.Open(logPath)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}

	var lines []string
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		if line := sc.Text(); line != "" {
			lines = append(lines, line)
		}
	}
	f.Close()
	if err := sc.Err(); err != nil {
		return err
	}

	if len(lines) <= maxEntries {
		return nil // nothing to trim
	}

	// Keep only the last maxEntries lines.
	lines = lines[len(lines)-maxEntries:]

	out, err := os.Create(logPath)
	if err != nil {
		return err
	}
	defer out.Close()

	w := bufio.NewWriter(out)
	for _, l := range lines {
		if _, err := w.WriteString(l + "\n"); err != nil {
			return err
		}
	}
	return w.Flush()
}
