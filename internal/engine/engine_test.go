package engine

import (
	"encoding/csv"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/mitchellbauer/data-coupler/internal/connector"
	"github.com/mitchellbauer/data-coupler/internal/transform"
	"github.com/mitchellbauer/data-coupler/internal/types"

	// Trigger init() registration for csv connector and all built-in transforms.
	_ "github.com/mitchellbauer/data-coupler/internal/connector/csv"
	_ "github.com/mitchellbauer/data-coupler/internal/transform"
)

// ── MapRow ────────────────────────────────────────────────────────────────────

// Test 1: Does the Mapper actually move data to the right slots?
func TestMapRow(t *testing.T) {
	// Setup: A fake profile that wants "ID" -> "TargetID" and "Name" -> "TargetName"
	profile := types.Profile{
		Mappings: []types.Mapping{
			{InputCol: "ID", OutputCol: "TargetID"},
			{InputCol: "Name", OutputCol: "TargetName"},
		},
	}

	// Setup: A fake header map telling us where "ID" and "Name" are in the source
	// source_csv: [ "Name", "Irrelevant", "ID" ]
	// indices:       0         1          2
	headerMap := map[string]int{
		"Name": 0,
		"ID":   2,
	}

	// Input Row corresponding to indices above: Name="Alice", Irrelevant="X", ID="101"
	inputRow := []string{"Alice", "X", "101"}

	// Execution
	result, err := MapRow(inputRow, headerMap, profile)
	if err != nil {
		t.Fatalf("MapRow() unexpected error: %v", err)
	}

	// Expectation: [ "101", "Alice" ] (Order based on Mappings list)
	expected := []string{"101", "Alice"}

	// Assertion
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("MapRow() = %v, want %v", result, expected)
	}
}

// ── ValidateHeaders ───────────────────────────────────────────────────────────

// Test 2: Does the Validator catch missing columns?
func TestValidateHeaders_MissingColumn(t *testing.T) {
	// Profile asks for "Email", but CSV headers won't have it
	profile := types.Profile{
		Mappings: []types.Mapping{
			{InputCol: "Email", OutputCol: "Contact"},
		},
	}

	csvHeaders := []string{"ID", "Name"} // No Email here!

	_, err := ValidateHeaders(csvHeaders, profile)

	if err == nil {
		t.Error("Expected an error for missing column 'Email', but got nil")
	}
}

// Test 3: Does it safely ignore irrelevant columns?
func TestValidateHeaders_Success(t *testing.T) {
	profile := types.Profile{
		Mappings: []types.Mapping{
			{InputCol: "ID", OutputCol: "OutputID"},
		},
	}

	csvHeaders := []string{"ID", "ExtraColumn1", "ExtraColumn2"}

	_, err := ValidateHeaders(csvHeaders, profile)

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}

// ── applyTransforms ───────────────────────────────────────────────────────────

func TestApplyTransforms_Single(t *testing.T) {
	// TrimSpace is registered via the blank import of the transform package above.
	transforms := []types.Transform{{Type: "TrimSpace"}}
	got, err := applyTransforms("  hello  ", transforms, nil, nil)
	if err != nil || got != "hello" {
		t.Errorf("applyTransforms(TrimSpace) = %q, %v, want %q, nil", got, err, "hello")
	}
}

func TestApplyTransforms_Chain(t *testing.T) {
	transforms := []types.Transform{
		{Type: "TrimSpace"},
		{Type: "ToUpper"},
	}
	got, err := applyTransforms("  hello  ", transforms, nil, nil)
	if err != nil || got != "HELLO" {
		t.Errorf("applyTransforms(TrimSpace+ToUpper) = %q, %v, want %q, nil", got, err, "HELLO")
	}
}

func TestApplyTransforms_Unknown(t *testing.T) {
	transforms := []types.Transform{{Type: "__no_such_transform__"}}
	_, err := applyTransforms("value", transforms, nil, nil)
	if err == nil {
		t.Error("applyTransforms() expected error for unknown transform, got nil")
	}
}

func TestApplyTransforms_RowTransformer(t *testing.T) {
	// Concatenate implements RowTransformer.
	headerMap := map[string]int{"First": 0, "Last": 1}
	inputRow := []string{"Jane", "Smith"}
	transforms := []types.Transform{{
		Type:   "Concatenate",
		Params: map[string]string{"cols": "First, Last", "separator": " "},
	}}
	got, err := applyTransforms("", transforms, inputRow, headerMap)
	if err != nil || got != "Jane Smith" {
		t.Errorf("applyTransforms(Concatenate) = %q, %v, want %q, nil", got, err, "Jane Smith")
	}
}

// ── Run (integration) ─────────────────────────────────────────────────────────

// writeTempCSVFile creates a CSV temp file and returns its path.
func writeTempCSVFile(t *testing.T, content string) string {
	t.Helper()
	f, err := os.CreateTemp("", "engine_test_*.csv")
	if err != nil {
		t.Fatalf("CreateTemp: %v", err)
	}
	if _, err := f.WriteString(content); err != nil {
		t.Fatalf("WriteString: %v", err)
	}
	f.Close()
	t.Cleanup(func() { os.Remove(f.Name()) })
	return f.Name()
}

func TestRun_CSV(t *testing.T) {
	inPath := writeTempCSVFile(t, "Name,Age\nAlice,30\nBob,25\n")
	outPath := filepath.Join(t.TempDir(), "out.csv")

	profile := types.Profile{
		Input:  types.IOConfig{Connector: "csv", Path: inPath},
		Output: types.IOConfig{Connector: "csv", Path: outPath},
		Mappings: []types.Mapping{
			{InputCol: "Name", OutputCol: "FullName", Transforms: []types.Transform{}},
			{InputCol: "Age", OutputCol: "Years", Transforms: []types.Transform{}},
		},
	}

	count, err := Run(profile, nil)
	if err != nil {
		t.Fatalf("Run() error: %v", err)
	}
	if count != 2 {
		t.Errorf("Run() row count = %d, want 2", count)
	}

	// Read and verify output file.
	f, err := os.Open(outPath)
	if err != nil {
		t.Fatalf("open output: %v", err)
	}
	defer f.Close()

	r := csv.NewReader(f)
	records, err := r.ReadAll()
	if err != nil {
		t.Fatalf("ReadAll: %v", err)
	}
	if len(records) != 3 { // header + 2 data rows
		t.Fatalf("output rows = %d, want 3", len(records))
	}
	if !reflect.DeepEqual(records[0], []string{"FullName", "Years"}) {
		t.Errorf("header row = %v, want [FullName Years]", records[0])
	}
	if !reflect.DeepEqual(records[1], []string{"Alice", "30"}) {
		t.Errorf("row 1 = %v, want [Alice 30]", records[1])
	}
}

func TestRun_UnknownInputConnector(t *testing.T) {
	profile := types.Profile{
		Input:  types.IOConfig{Connector: "__no_such_connector__"},
		Output: types.IOConfig{Connector: "csv"},
	}
	_, err := Run(profile, nil)
	if err == nil {
		t.Error("Run() expected error for unknown input connector, got nil")
	}
}

func TestRun_UnknownOutputConnector(t *testing.T) {
	// Use a header-only CSV so no rows are streamed (prevents goroutine leak).
	inPath := writeTempCSVFile(t, "Name,Age\n")

	profile := types.Profile{
		Input:  types.IOConfig{Connector: "csv", Path: inPath},
		Output: types.IOConfig{Connector: "__no_such_output__"},
		Mappings: []types.Mapping{
			{InputCol: "Name", OutputCol: "FullName", Transforms: []types.Transform{}},
		},
	}
	_, err := Run(profile, nil)
	if err == nil {
		t.Error("Run() expected error for unknown output connector, got nil")
	}
}

func TestRun_OutputNotWriter(t *testing.T) {
	// Register a read-only connector that does not implement connector.Writer.
	connector.Register(&readOnlyConn{})

	inPath := writeTempCSVFile(t, "Name,Age\n")

	profile := types.Profile{
		Input:  types.IOConfig{Connector: "csv", Path: inPath},
		Output: types.IOConfig{Connector: "read-only-test"},
		Mappings: []types.Mapping{
			{InputCol: "Name", OutputCol: "FullName", Transforms: []types.Transform{}},
		},
	}
	_, err := Run(profile, nil)
	if err == nil {
		t.Error("Run() expected error when output connector does not support writing, got nil")
	}
}

// readOnlyConn satisfies connector.Connector but not connector.Writer.
type readOnlyConn struct{}

func (c *readOnlyConn) Name() string                                         { return "read-only-test" }
func (c *readOnlyConn) Connect(cfg connector.ConnectionConfig) error         { return nil }
func (c *readOnlyConn) Disconnect() error                                    { return nil }
func (c *readOnlyConn) Columns(_ string) ([]string, error)                   { return nil, nil }
func (c *readOnlyConn) Rows(_ string) (<-chan []string, error)                { return nil, nil }

// Ensure the blank imports are used (suppress "imported and not used" in some editors).
var _ = transform.List
