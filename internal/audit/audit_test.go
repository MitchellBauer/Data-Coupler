package audit

import (
	"bufio"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestAppendEntry(t *testing.T) {
	dir := t.TempDir()
	SetLogPath(dir)

	entry := AuditEntry{
		Timestamp:   time.Now(),
		ProfileName: "Test Profile",
		ProfileID:   "test-profile",
		InputSource: "csv: input.csv",
		OutputPath:  "output.csv",
		RowsOut:     42,
		DurationMs:  150,
	}

	if err := AppendEntry(entry); err != nil {
		t.Fatalf("AppendEntry() error: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(dir, "audit.log"))
	if err != nil {
		t.Fatalf("could not read audit.log: %v", err)
	}

	var got AuditEntry
	if err := json.Unmarshal([]byte(strings.TrimSpace(string(data))), &got); err != nil {
		t.Fatalf("audit.log line is not valid JSON: %v\nLine: %s", err, data)
	}

	if got.ProfileName != entry.ProfileName {
		t.Errorf("ProfileName = %q, want %q", got.ProfileName, entry.ProfileName)
	}
	if got.RowsOut != entry.RowsOut {
		t.Errorf("RowsOut = %d, want %d", got.RowsOut, entry.RowsOut)
	}
}

func TestTrimLog_ReducesLines(t *testing.T) {
	dir := t.TempDir()
	SetLogPath(dir)

	// Write 1200 lines.
	f, err := os.Create(filepath.Join(dir, "audit.log"))
	if err != nil {
		t.Fatal(err)
	}
	w := bufio.NewWriter(f)
	for i := 0; i < 1200; i++ {
		// Use line number as a sentinel so we can verify the newest are kept.
		_, _ = w.WriteString(`{"rowsOut":` + string(rune('0'+i%10)) + `}` + "\n")
	}
	_ = w.Flush()
	f.Close()

	if err := TrimLog(1000); err != nil {
		t.Fatalf("TrimLog() error: %v", err)
	}

	data, _ := os.ReadFile(filepath.Join(dir, "audit.log"))
	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	if len(lines) != 1000 {
		t.Errorf("after TrimLog(1000), got %d lines, want 1000", len(lines))
	}
}

func TestTrimLog_MissingFile(t *testing.T) {
	dir := t.TempDir()
	SetLogPath(dir)
	// audit.log does not exist — TrimLog must return nil without creating the file.
	if err := TrimLog(1000); err != nil {
		t.Errorf("TrimLog() on missing file returned error: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, "audit.log")); !os.IsNotExist(err) {
		t.Error("TrimLog should not create audit.log when it doesn't exist")
	}
}
