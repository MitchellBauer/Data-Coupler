package transform

import (
	"strings"
	"testing"
)

func TestTrimSpace(t *testing.T) {
	tr := &TrimSpace{}

	cases := []struct {
		input string
		want  string
	}{
		{"  hello", "hello"},
		{"hello  ", "hello"},
		{"  hello  ", "hello"},
		{"hello", "hello"},
	}

	for _, c := range cases {
		got, err := tr.Apply(c.input, nil)
		if err != nil {
			t.Errorf("TrimSpace.Apply(%q) unexpected error: %v", c.input, err)
		}
		if got != c.want {
			t.Errorf("TrimSpace.Apply(%q) = %q, want %q", c.input, got, c.want)
		}
	}
}

func TestDefault(t *testing.T) {
	tr := &Default{}
	params := map[string]string{"value": "N/A"}

	cases := []struct {
		input string
		want  string
	}{
		{"", "N/A"},
		{"hello", "hello"},
	}

	for _, c := range cases {
		got, err := tr.Apply(c.input, params)
		if err != nil {
			t.Errorf("Default.Apply(%q) unexpected error: %v", c.input, err)
		}
		if got != c.want {
			t.Errorf("Default.Apply(%q) = %q, want %q", c.input, got, c.want)
		}
	}
}

func TestToUpper(t *testing.T) {
	tr := &ToUpper{}
	got, err := tr.Apply("hello world", nil)
	if err != nil || got != "HELLO WORLD" {
		t.Errorf("ToUpper.Apply() = %q, %v, want %q, nil", got, err, "HELLO WORLD")
	}
}

func TestToLower(t *testing.T) {
	tr := &ToLower{}
	got, err := tr.Apply("HELLO WORLD", nil)
	if err != nil || got != "hello world" {
		t.Errorf("ToLower.Apply() = %q, %v, want %q, nil", got, err, "hello world")
	}
}

func TestPrefix(t *testing.T) {
	tr := &Prefix{}
	params := map[string]string{"value": "PRE-"}
	got, err := tr.Apply("value", params)
	if err != nil || got != "PRE-value" {
		t.Errorf("Prefix.Apply() = %q, %v, want %q, nil", got, err, "PRE-value")
	}
}

func TestSuffix(t *testing.T) {
	tr := &Suffix{}
	params := map[string]string{"value": "-SFX"}
	got, err := tr.Apply("value", params)
	if err != nil || got != "value-SFX" {
		t.Errorf("Suffix.Apply() = %q, %v, want %q, nil", got, err, "value-SFX")
	}
}

func TestDateFormat_Valid(t *testing.T) {
	tr := &DateFormat{}
	params := map[string]string{"from": "2006-01-02", "to": "01/02/2006"}
	got, err := tr.Apply("2024-03-15", params)
	if err != nil || got != "03/15/2024" {
		t.Errorf("DateFormat.Apply() = %q, %v, want %q, nil", got, err, "03/15/2024")
	}
}

func TestDateFormat_Invalid(t *testing.T) {
	tr := &DateFormat{}
	params := map[string]string{"from": "2006-01-02", "to": "01/02/2006"}
	_, err := tr.Apply("not-a-date", params)
	if err == nil {
		t.Error("DateFormat.Apply() expected error for invalid date, got nil")
	}
}

func TestDateFormat_EmptyInput(t *testing.T) {
	tr := &DateFormat{}
	params := map[string]string{"from": "2006-01-02", "to": "01/02/2006"}
	got, err := tr.Apply("", params)
	if err != nil || got != "" {
		t.Errorf("DateFormat.Apply(%q) = %q, %v, want %q, nil", "", got, err, "")
	}
}

func TestSplit_ValidIndex(t *testing.T) {
	tr := &Split{}
	params := map[string]string{"separator": "-", "index": "1"}
	got, err := tr.Apply("A-B-C", params)
	if err != nil || got != "B" {
		t.Errorf("Split.Apply() = %q, %v, want %q, nil", got, err, "B")
	}
}

func TestSplit_OutOfRange(t *testing.T) {
	tr := &Split{}
	params := map[string]string{"separator": "-", "index": "9"}
	got, err := tr.Apply("A-B", params)
	if err != nil || got != "" {
		t.Errorf("Split.Apply() out-of-range = %q, %v, want %q, nil", got, err, "")
	}
}

func TestLookupReplace_Hit(t *testing.T) {
	tr := &LookupReplace{}
	params := map[string]string{"map": `{"01":"Category A","02":"Category B"}`}
	got, err := tr.Apply("01", params)
	if err != nil || got != "Category A" {
		t.Errorf("LookupReplace.Apply() = %q, %v, want %q, nil", got, err, "Category A")
	}
}

func TestLookupReplace_Miss(t *testing.T) {
	tr := &LookupReplace{}
	params := map[string]string{"map": `{"01":"Category A"}`}
	got, err := tr.Apply("99", params)
	if err != nil || got != "99" {
		t.Errorf("LookupReplace.Apply() miss = %q, %v, want %q, nil", got, err, "99")
	}
}

func TestLookupReplace_BadJSON(t *testing.T) {
	tr := &LookupReplace{}
	params := map[string]string{"map": `{bad json`}
	_, err := tr.Apply("anything", params)
	if err == nil {
		t.Error("LookupReplace.Apply() expected error for bad JSON, got nil")
	}
}

func TestConcatenate(t *testing.T) {
	tr := &Concatenate{}
	headerMap := map[string]int{"First": 0, "Last": 1}
	inputRow := []string{"John", "Doe"}
	params := map[string]string{"cols": "First, Last", "separator": " "}

	got, err := tr.ApplyRow(inputRow, headerMap, params)
	if err != nil || got != "John Doe" {
		t.Errorf("Concatenate.ApplyRow() = %q, %v, want %q, nil", got, err, "John Doe")
	}
}

func TestRegistry_RegisterGet(t *testing.T) {
	// TrimSpace is registered by init(); verify Get returns it.
	tr, ok := Get("TrimSpace")
	if !ok {
		t.Fatal("Get(TrimSpace) returned false, want true")
	}
	if tr.Name() != "TrimSpace" {
		t.Errorf("Get(TrimSpace).Name() = %q, want %q", tr.Name(), "TrimSpace")
	}
}

func TestRegistry_NotFound(t *testing.T) {
	_, ok := Get("__nonexistent_transform__")
	if ok {
		t.Error("Get() returned true for unknown transform, want false")
	}
}

func TestRegistry_ListContainsAll(t *testing.T) {
	// All 10 built-in transforms must be registered via init().
	expected := []string{
		"Concatenate", "DateFormat", "Default", "LookupReplace",
		"Prefix", "Split", "Suffix", "ToLower", "ToUpper", "TrimSpace",
	}
	names := List()
	nameSet := make(map[string]bool, len(names))
	for _, n := range names {
		nameSet[n] = true
	}
	var missing []string
	for _, e := range expected {
		if !nameSet[e] {
			missing = append(missing, e)
		}
	}
	if len(missing) > 0 {
		t.Errorf("List() missing transforms: %s", strings.Join(missing, ", "))
	}
}
