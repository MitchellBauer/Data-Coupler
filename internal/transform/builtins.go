package transform

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"
)

// ── No-param transforms ───────────────────────────────────────────────────────

// TrimSpace strips leading and trailing whitespace from a value.
type TrimSpace struct{}

func (t *TrimSpace) Name() string { return "TrimSpace" }
func (t *TrimSpace) Apply(value string, _ map[string]string) (string, error) {
	return strings.TrimSpace(value), nil
}

// ToUpper converts a value to uppercase.
type ToUpper struct{}

func (t *ToUpper) Name() string { return "ToUpper" }
func (t *ToUpper) Apply(value string, _ map[string]string) (string, error) {
	return strings.ToUpper(value), nil
}

// ToLower converts a value to lowercase.
type ToLower struct{}

func (t *ToLower) Name() string { return "ToLower" }
func (t *ToLower) Apply(value string, _ map[string]string) (string, error) {
	return strings.ToLower(value), nil
}

// ── Single-param transforms ───────────────────────────────────────────────────

// Default returns a fallback value (params["value"]) when the input is empty.
type Default struct{}

func (d *Default) Name() string { return "Default" }
func (d *Default) Apply(value string, params map[string]string) (string, error) {
	if value == "" {
		return params["value"], nil
	}
	return value, nil
}

// Prefix prepends params["value"] to the input.
type Prefix struct{}

func (t *Prefix) Name() string { return "Prefix" }
func (t *Prefix) Apply(value string, params map[string]string) (string, error) {
	return params["value"] + value, nil
}

// Suffix appends params["value"] to the input.
type Suffix struct{}

func (t *Suffix) Name() string { return "Suffix" }
func (t *Suffix) Apply(value string, params map[string]string) (string, error) {
	return value + params["value"], nil
}

// ── Multi-param transforms ────────────────────────────────────────────────────

// DateFormat reformats a date string.
// Params: "from" — Go time layout of the input, "to" — Go time layout for the output.
// Empty input passes through as empty. Parse failure returns an error (engine writes blank, continues).
type DateFormat struct{}

func (t *DateFormat) Name() string { return "DateFormat" }
func (t *DateFormat) Apply(value string, params map[string]string) (string, error) {
	if value == "" {
		return "", nil
	}
	parsed, err := time.Parse(params["from"], value)
	if err != nil {
		return "", fmt.Errorf("DateFormat: cannot parse %q with layout %q", value, params["from"])
	}
	return parsed.Format(params["to"]), nil
}

// Split extracts one segment of a delimited string.
// Params: "separator" — delimiter string, "index" — zero-based segment index.
// Out-of-range index returns empty string without error.
type Split struct{}

func (t *Split) Name() string { return "Split" }
func (t *Split) Apply(value string, params map[string]string) (string, error) {
	parts := strings.Split(value, params["separator"])
	idx, _ := strconv.Atoi(params["index"])
	if idx < 0 || idx >= len(parts) {
		return "", nil
	}
	return parts[idx], nil
}

// LookupReplace swaps values via a lookup table.
// Params: "map" — JSON object string, e.g. {"01":"Category A","02":"Category B"}.
// Keys absent from the map pass through unchanged. Invalid JSON returns an error.
// The parsed map is cached per unique JSON string to avoid repeated JSON parsing in hot loops.
type LookupReplace struct{}

var lookupCache sync.Map // key: JSON string → value: map[string]string

func (t *LookupReplace) Name() string { return "LookupReplace" }
func (t *LookupReplace) Apply(value string, params map[string]string) (string, error) {
	raw := params["map"]
	var m map[string]string
	if cached, ok := lookupCache.Load(raw); ok {
		m = cached.(map[string]string)
	} else {
		if err := json.Unmarshal([]byte(raw), &m); err != nil {
			return "", fmt.Errorf("LookupReplace: invalid map JSON: %w", err)
		}
		lookupCache.Store(raw, m)
	}
	if replacement, ok := m[value]; ok {
		return replacement, nil
	}
	return value, nil
}

func init() {
	Register(&TrimSpace{})
	Register(&Default{})
	Register(&ToUpper{})
	Register(&ToLower{})
	Register(&DateFormat{})
	Register(&Split{})
	Register(&Prefix{})
	Register(&Suffix{})
	Register(&LookupReplace{})
	Register(&Concatenate{})
}

// ── Multi-column transform ────────────────────────────────────────────────────

// Concatenate merges values from multiple input columns into one output value.
// Params: "cols" — comma-separated list of input column names, "separator" — join string.
// Implements RowTransformer so the engine passes the full row context.
// Missing columns are treated as empty strings.
type Concatenate struct{}

func (t *Concatenate) Name() string { return "Concatenate" }

// Apply satisfies the base Transformer interface; unused when engine detects RowTransformer.
func (t *Concatenate) Apply(value string, _ map[string]string) (string, error) {
	return value, nil
}

// ApplyRow is the real implementation — uses the full input row.
func (t *Concatenate) ApplyRow(inputRow []string, headerMap map[string]int, params map[string]string) (string, error) {
	cols := strings.Split(params["cols"], ",")
	sep := params["separator"]
	parts := make([]string, 0, len(cols))
	for _, col := range cols {
		col = strings.TrimSpace(col)
		if idx, ok := headerMap[col]; ok && idx < len(inputRow) {
			parts = append(parts, inputRow[idx])
		} else {
			parts = append(parts, "")
		}
	}
	return strings.Join(parts, sep), nil
}
