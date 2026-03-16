package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/mitchellbauer/data-coupler/internal/types"
)

// ── LoadProfile ───────────────────────────────────────────────────────────────

func TestLoadProfile_NewFormat(t *testing.T) {
	json := `{
		"id": "test-profile",
		"version": 1,
		"name": "Test",
		"description": "desc",
		"input":  {"connector": "csv", "path": "in.csv"},
		"output": {"connector": "csv", "path": "out.csv"},
		"mappings": [{"inputCol": "A", "outputCol": "B", "transforms": []}]
	}`
	path := writeTempFile(t, json)

	p, err := LoadProfile(path)
	if err != nil {
		t.Fatalf("LoadProfile() error: %v", err)
	}
	if p.ID != "test-profile" {
		t.Errorf("ID = %q, want %q", p.ID, "test-profile")
	}
	if p.Input.Connector != "csv" {
		t.Errorf("Input.Connector = %q, want %q", p.Input.Connector, "csv")
	}
	if len(p.Mappings) != 1 || p.Mappings[0].InputCol != "A" {
		t.Errorf("Mappings = %v, want [{A → B}]", p.Mappings)
	}
}

func TestLoadProfile_OldFormat(t *testing.T) {
	// Phase 1A format: has "settings", no "input"/"output" blocks.
	json := `{
		"id": "legacy",
		"name": "Legacy Profile",
		"description": "old style",
		"settings": {},
		"mappings": [
			{"inputCol": "ID", "outputCol": "TargetID", "transform": "TrimSpace"}
		]
	}`
	path := writeTempFile(t, json)

	p, err := LoadProfile(path)
	if err != nil {
		t.Fatalf("LoadProfile() error: %v", err)
	}
	if p.ID != "legacy" {
		t.Errorf("ID = %q, want %q", p.ID, "legacy")
	}
	if p.Input.Connector != "csv" {
		t.Errorf("Input.Connector = %q, want %q", p.Input.Connector, "csv")
	}
	if p.Output.Connector != "csv" {
		t.Errorf("Output.Connector = %q, want %q", p.Output.Connector, "csv")
	}
	if p.Version != 1 {
		t.Errorf("Version = %d, want 1", p.Version)
	}
}

func TestMigrateV0_TransformCarried(t *testing.T) {
	json := `{
		"id": "x", "name": "X",
		"settings": {},
		"mappings": [
			{"inputCol": "A", "outputCol": "B", "transform": "TrimSpace"}
		]
	}`
	path := writeTempFile(t, json)

	p, err := LoadProfile(path)
	if err != nil {
		t.Fatalf("LoadProfile() error: %v", err)
	}
	if len(p.Mappings) != 1 {
		t.Fatalf("len(Mappings) = %d, want 1", len(p.Mappings))
	}
	if len(p.Mappings[0].Transforms) != 1 || p.Mappings[0].Transforms[0].Type != "TrimSpace" {
		t.Errorf("Transforms = %v, want [{TrimSpace}]", p.Mappings[0].Transforms)
	}
}

func TestMigrateV0_NoneTransformDropped(t *testing.T) {
	json := `{
		"id": "x", "name": "X",
		"settings": {},
		"mappings": [
			{"inputCol": "A", "outputCol": "B", "transform": "none"}
		]
	}`
	path := writeTempFile(t, json)

	p, err := LoadProfile(path)
	if err != nil {
		t.Fatalf("LoadProfile() error: %v", err)
	}
	if len(p.Mappings[0].Transforms) != 0 {
		t.Errorf("Transforms = %v, want empty", p.Mappings[0].Transforms)
	}
}

func TestLoadProfile_MissingFile(t *testing.T) {
	_, err := LoadProfile("/no/such/path/profile.json")
	if err == nil {
		t.Error("expected error for missing file, got nil")
	}
}

func TestLoadProfile_BadJSON(t *testing.T) {
	path := writeTempFile(t, `{bad json`)
	_, err := LoadProfile(path)
	if err == nil {
		t.Error("expected error for bad JSON, got nil")
	}
}

func TestSaveProfile_RoundTrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "profile.json")

	want := types.Profile{
		ID:          "round-trip",
		Version:     1,
		Name:        "RoundTrip",
		Description: "test",
		Input:       types.IOConfig{Connector: "csv", Path: "in.csv"},
		Output:      types.IOConfig{Connector: "csv", Path: "out.csv"},
		Mappings: []types.Mapping{
			{InputCol: "A", OutputCol: "B", Transforms: []types.Transform{}},
		},
	}

	if err := SaveProfile(path, want); err != nil {
		t.Fatalf("SaveProfile() error: %v", err)
	}

	got, err := LoadProfile(path)
	if err != nil {
		t.Fatalf("LoadProfile() error: %v", err)
	}
	if got.ID != want.ID || got.Name != want.Name || got.Input.Connector != want.Input.Connector {
		t.Errorf("round-trip mismatch: got %+v, want %+v", got, want)
	}
}

// ── LoadSettings / SaveSettings ──────────────────────────────────────────────

func TestLoadSettings_MissingFile(t *testing.T) {
	s, err := LoadSettings("/no/such/path/settings.json")
	if err != nil {
		t.Fatalf("LoadSettings() unexpected error: %v", err)
	}
	if s.LastConnector != "csv" {
		t.Errorf("LastConnector = %q, want %q", s.LastConnector, "csv")
	}
}

func TestLoadSettings_ExistingFile(t *testing.T) {
	path := writeTempFile(t, `{"lastConnector":"mssql","lastProfilePath":"foo.json"}`)

	s, err := LoadSettings(path)
	if err != nil {
		t.Fatalf("LoadSettings() error: %v", err)
	}
	if s.LastConnector != "mssql" {
		t.Errorf("LastConnector = %q, want %q", s.LastConnector, "mssql")
	}
	if s.LastProfilePath != "foo.json" {
		t.Errorf("LastProfilePath = %q, want %q", s.LastProfilePath, "foo.json")
	}
}

func TestSaveSettings_RoundTrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "settings.json")

	want := types.AppSettings{
		LastConnector:   "postgres",
		LastProfilePath: "my.json",
	}
	if err := SaveSettings(path, want); err != nil {
		t.Fatalf("SaveSettings() error: %v", err)
	}
	got, err := LoadSettings(path)
	if err != nil {
		t.Fatalf("LoadSettings() error: %v", err)
	}
	if got.LastConnector != want.LastConnector || got.LastProfilePath != want.LastProfilePath {
		t.Errorf("round-trip mismatch: got %+v, want %+v", got, want)
	}
}

// ── helpers ───────────────────────────────────────────────────────────────────

func writeTempFile(t *testing.T, content string) string {
	t.Helper()
	f, err := os.CreateTemp("", "config_test_*.json")
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
