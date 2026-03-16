---
date created: Sunday, March 15th 2026, 12:00:00 pm
date modified: Sunday, March 15th 2026, 12:00:00 pm
---

**Focus:** Built-in Data Transforms + Transform Configuration UI
**Parent:** [[Roadmap - Data Coupler Project]]
**Prerequisite:** [[Task List - Phase 2 (Database Connectors) - Data Coupler Project]]

> **Why a transform pipeline?**
> Raw data almost never matches the format a target system expects. Macola fields have trailing whitespace. Dates come out as `MM/DD/YYYY` but Fishbowl wants `YYYY-MM-DD`. Customer names are stored in two columns but the output needs one. The transform pipeline applies per-column, ordered, chainable operations between reading and writing — without the user ever needing to manipulate the data manually.
>
> `TrimSpace` and `Default` were implemented as stubs in Phase 1B. This phase completes the full built-in library and adds the GUI step to configure transforms interactively.

---

# 🔧 Stage 1: Remaining Built-in Transforms (`internal/transform/builtins.go`)

*All transforms implement the `Transformer` interface. They are stateless — a string value goes in, a string value comes out.*

```go
type Transformer interface {
    Name()  string
    Apply(value string, params map[string]string) (string, error)
}
```

## Step 1.1: Case Transforms

* [ ] Implement `ToUpper`:
  * No params.
  * `"acme corp"` → `"ACME CORP"`
* [ ] Implement `ToLower`:
  * No params.
  * `"ACME CORP"` → `"acme corp"`
* [ ] Register both in `main.go`.
* [ ] Unit tests (table-driven):
  * Normal mixed-case string.
  * Already upper/lower (no-op behavior).
  * Empty string returns empty string.

## Step 1.2: Date Format Transform

* [ ] Implement `DateFormat`:
  * Params: `from` (Go time layout string), `to` (Go time layout string).
  * Example: `from="01/02/2006"`, `to="2006-01-02"` converts `"03/15/2025"` → `"2025-03-15"`.
  * ⚠️ **Constraint:** Use Go's `time.Parse` / `time.Format` with the exact layout strings provided. Document in the UI that Go uses `01/02/2006` (not `MM/DD/YYYY`). Consider showing a human-readable hint in the param form.
  * On parse failure: return a descriptive error including the input value and the expected layout. The engine will log a warning and write a blank — it will not abort the run.
* [ ] Register in `main.go`.
* [ ] Unit tests:
  * Valid date converts correctly.
  * Input not matching `from` layout returns error.
  * Empty input returns empty string (no error).

## Step 1.3: String Manipulation Transforms

* [ ] Implement `Split`:
  * Params: `separator` (string), `index` (zero-based integer as string).
  * Example: separator=`"-"`, index=`"1"` on `"123-456-789"` → `"456"`.
  * If `index` is out of range, return empty string (no error).
* [ ] Implement `Prefix`:
  * Param: `value` (string to prepend).
  * `"123"` with value=`"PART-"` → `"PART-123"`.
* [ ] Implement `Suffix`:
  * Param: `value` (string to append).
  * `"123"` with value=`"-US"` → `"123-US"`.
* [ ] Register all three in `main.go`.
* [ ] Unit tests for each: normal input, empty input, edge cases (index out of range for `Split`).

## Step 1.4: Lookup Replace Transform

* [ ] Implement `LookupReplace`:
  * Param: `map` — a JSON-encoded object string (e.g., `{"01":"Category A","02":"Category B"}`).
  * On `Apply()`: parse the `map` param, look up the input value. If found, return the replacement. If not found, return the original value unchanged.
  * ⚠️ **Constraint:** Parse the `map` param on every call (it's a string). For large maps used repeatedly, this is acceptable given the stateless design; do not cache state in the transformer.
* [ ] Register in `main.go`.
* [ ] Unit tests:
  * Key present → returns replacement value.
  * Key absent → returns original value unchanged.
  * Invalid JSON in `map` param → returns error.
  * Empty input → returns empty string.

## Step 1.5: Concatenate Transform (Engine Change Required)

* [ ] ⚠️ **Constraint:** `Concatenate` reads from multiple input columns, not just the value of the current column. This requires a small change to how the engine calls transforms.
* [ ] **Engine change** in `internal/engine/engine.go`:
  * Add a second signature variant or pass the full input row + header map to `applyTransforms()` so that multi-column transforms can look up other columns by name.
  * Existing single-value transforms ignore the extra context (no behavior change).
* [ ] Implement `Concatenate`:
  * Params: `cols` (comma-separated list of input column names), `separator` (string).
  * Example: cols=`"FIRSTNAME,LASTNAME"`, separator=`" "` → `"John Doe"`.
  * The `value` parameter (current column's value) is ignored — `Concatenate` builds its output entirely from `cols`.
  * If a named column doesn't exist in the row, treat it as an empty string.
* [ ] Register in `main.go`.
* [ ] Unit tests:
  * Two columns concatenated with separator.
  * Missing column treated as empty.
  * Single column (degenerate case).

## Step 1.6: Engine Error Hardening

* [ ] In `applyTransforms()`, if a transform `Name()` is not found in the registry, return an error that includes the row number, column name, and the unknown transform type.
* [ ] Add test: profile referencing an unknown transform type `"Frobnicate"` returns a clear error message.
* [ ] **Milestone:** `go test ./internal/transform/...` and `go test ./internal/engine/...` pass with all new transforms covered.

---

# 🖥️ Stage 2: Transform Configuration Step (`internal/ui/step_transform.go`)

*A new wizard step, inserted between Map Columns (Step 4) and Review & Run (Step 5). It is always shown but never blocks progress — transforms are optional.*

## Step 2.1: Wizard Integration

* [ ] Add `step_transform.go` to the wizard step list in `wizard.go`, between `step_mapping` and `step_review`.
* [ ] Update step count indicator in the wizard chrome (e.g., "Step 5 of 6").
* [ ] `Title()` returns `"Configure Transforms (Optional)"`.
* [ ] `Validate()` always returns `nil` — this step is never blocking.

## Step 2.2: Per-Column Transform Editor

* [ ] Render a scrollable list of all configured mappings from `WizardState.Mappings`.
* [ ] Each mapping row shows: **Output Column Name** (label) ← **Input Column** (label) + an "Add Transform" button.
* [ ] "Add Transform" button: opens a dropdown of all registered transform names.
  * On selection: append a transform entry to the mapping's `Transforms` slice.
  * Show a param form appropriate to the selected transform type (see below).
* [ ] Each active transform shows:
  * Transform name label.
  * Param fields (dynamically rendered based on transform type).
  * A "Remove" button (×).
  * Up/Down arrow buttons to reorder within the column.
* [ ] All changes sync live to `WizardState.Mappings[i].Transforms`.

## Step 2.3: Dynamic Param Forms

*Each transform type has different parameters. Render the right fields automatically.*

| Transform | Params UI |
|---|---|
| `TrimSpace` | No params — just a label |
| `ToUpper` / `ToLower` | No params |
| `DateFormat` | Two text entries: "From format" and "To format" with hint text showing Go layout |
| `Split` | Text entry for separator + numeric entry for index |
| `Prefix` / `Suffix` | Single text entry labeled "Value" |
| `LookupReplace` | Multi-row key/value editor (Add Row / Remove Row) — serialized to JSON internally |
| `Concatenate` | Multi-select of input column names + separator entry |
| `Default` | Single text entry labeled "Default value" |

## Step 2.4: Live Preview

* [ ] Below each mapping row with at least one transform configured, show a before/after preview:
  * "Before: `ACME CORP   `" → "After: `ACME CORP`"
  * Source the "before" value from `WizardState.InputPreviewRows[0]` for the corresponding input column.
  * Run the transform chain live as params change.
  * Show `(no preview data)` if `InputPreviewRows` is empty.

## Step 2.5: Macola Auto-Suggest

* [ ] If `WizardState.InputConnectorName == "mssql"` and no mappings have `TrimSpace` configured:
  * Show a persistent info banner at the top of the step: "Macola data often has trailing spaces. Click here to add TrimSpace to all columns."
  * "Add to all" button: appends `TrimSpace` to every mapping that doesn't already have it.
  * Banner disappears once all mappings have `TrimSpace`.

* [ ] **Milestone:** Full manual test: load a Macola-style CSV with trailing spaces → configure TrimSpace via GUI → run → verify output has no trailing whitespace.

---

# ✅ Stage 3: End-to-End Verification

* [ ] Test each transform via the GUI: configure, preview, run, verify output.
* [ ] Test `DateFormat` with a malformed date value: verify the engine logs a warning for that cell, writes blank, and completes the run (does not abort).
* [ ] Test `Concatenate` in a profile with MSSQL source: two name columns merged into one output column.
* [ ] Test `LookupReplace` with a map of 20+ entries.
* [ ] `go test ./...` passes with zero errors.
* [ ] Update `test_profile.json` to include at least one mapping with `TrimSpace` and one with `DateFormat` as living documentation of the transform syntax.
