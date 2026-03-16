---
date created: Saturday, January 17th 2026, 10:09:52 am
date modified: Sunday, March 15th 2026, 12:00:00 pm
---

**Focus:** Code Refactor for Modularity + Fyne Wizard GUI
**Parent:** [[Roadmap - Data Coupler Project]]
**Prerequisite:** [[Task List - Phase 1A (The Engine) - Data Coupler Project]]

> **Why a refactor first?**
> Phase 1A was intentionally minimal — get a working CSV-to-CSV engine as fast as possible. Now that the larger vision is clear (database connectors, transform pipeline, Fishbowl templates), the code needs to be restructured *before* the GUI is built on top of it. Building the wizard on top of the old layout would mean ripping it apart again in Phase 2. One clean pass now saves three messy ones later.
>
> The refactor does not change any behavior. The CLI still works identically at the end of this step. It only reorganizes the internals.

---

# 🔧 Stage 1: Refactor — Lay the Foundation

## Step 1.1: Restructure the Directory Layout

*Align the actual project folders with the architecture defined in the Technical Design Document.*

* [ ] Rename `internal/mapper` → `internal/engine`
  * Update all import paths that reference the old package name.
  * Verify `go build ./...` still compiles cleanly.
* [ ] Create new empty packages (files can be stubs for now):
  * `internal/connector/` — will hold the Connector interface and registry
  * `internal/connector/csv/` — will hold the CSV connector
  * `internal/transform/` — will hold the Transformer interface and built-ins
  * `internal/credentials/` — will hold encrypted credential storage (stub only in 1B)
* [ ] Move `internal/config` profile loading logic — verify it still satisfies the `internal/types` structs.
* [ ] **Milestone:** `go build ./...` and `go test ./...` pass with zero errors after restructure.

---

## Step 1.2: Upgrade the Types (`internal/types`)

*The existing `Profile` struct was built for CSV-only. It needs to support the full connector model.*

* [ ] **Update `Profile` struct:**
  * Replace the flat `Settings` block with a dedicated `Input IOConfig` and `Output IOConfig`.
  * Add `Version int` field for future schema migrations.
* [ ] **Update `IOConfig` struct:**
  * Fields: `Connector string`, `CredentialRef string`, `Query string`, `Path string`, `Template string`
  * For now, only `Connector: "csv"` and `Path` are used. The other fields are wired up but dormant.
* [ ] **Update `Mapping` struct:**
  * Add `Transforms []Transform` field.
  * Define `Transform struct` with `Type string` and `Params map[string]string`.
* [ ] **Write a profile migration helper** in `internal/config`:
  * Detect old-format profiles (no `input`/`output` blocks) and upgrade them on load.
  * This ensures existing profiles from Phase 1A don't break.
* [ ] **Update test fixtures** (`test_profile.json`) to use the new schema.
* [ ] **Milestone:** All existing unit tests pass against the updated structs.

---

## Step 1.3: Extract the CSV Connector (`internal/connector`)

*Move the CSV read/write logic out of the engine and behind the Connector interface. The engine should never call `csv.NewReader` directly.*

* [ ] **Define the `Connector` interface** in `internal/connector/connector.go`:
  ```go
  type Connector interface {
      Name()    string
      Connect(cfg ConnectionConfig) error
      Disconnect() error
      Columns(query string) ([]string, error)
      Rows(query string) (<-chan []string, error)
  }
  ```
* [ ] **Define `ConnectionConfig` struct** in the same file (fields: `Host`, `Port`, `Database`, `Username`, `Password`, `FilePath`, `Extra map[string]string`).
* [ ] **Implement `CSVConnector`** in `internal/connector/csv/`:
  * `Name()` returns `"csv"`.
  * `Connect()` opens the file at `cfg.FilePath`, stores the reader internally.
  * `Columns()` reads and returns the header row (BOM-safe, reusing existing logic).
  * `Rows()` streams remaining rows as `[]string` through a channel.
  * `Disconnect()` closes the file handle.
* [ ] **Create the Connector Registry** in `internal/connector/registry.go`:
  * `Register(c Connector)` — adds to an internal map keyed by `c.Name()`.
  * `Get(name string) (Connector, bool)` — looks up by name.
* [ ] **Register `CSVConnector` in `main.go`** at startup.
* [ ] **Refactor `internal/engine`** to resolve connectors through the registry instead of calling CSV functions directly.
* [ ] **Write unit tests** for `CSVConnector`:
  * `Columns()` returns correct headers from a test CSV string.
  * `Rows()` streams all data rows correctly.
  * BOM stripping still works.
* [ ] **Milestone:** CLI end-to-end test still passes. The engine no longer imports `encoding/csv` directly.

---

## Step 1.4: Wire the Transform Pipeline (`internal/transform`)

*Set up the infrastructure for transforms even though only one transform is implemented right now. The engine already has a slot for it (`mapping.Transforms`). Now we make it real.*

* [ ] **Define the `Transformer` interface** in `internal/transform/transform.go`:
  ```go
  type Transformer interface {
      Name()  string
      Apply(value string, params map[string]string) (string, error)
  }
  ```
* [ ] **Create the Transform Registry** (same pattern as connector registry):
  * `Register(t Transformer)`
  * `Get(name string) (Transformer, bool)`
* [ ] **Implement `TrimSpace`** in `internal/transform/builtins.go` — the first and most critical transform. Strips leading and trailing whitespace.
* [ ] **Implement `Default`** — returns a fallback value if the input is empty. (Needed for static-value output columns.)
* [ ] **Register both transforms in `main.go`** at startup.
* [ ] **Update `internal/engine`** to run `applyTransforms()` on every value during row processing. If `Transforms` is empty, the value passes through unchanged (zero behavior change for existing profiles).
* [ ] **Write unit tests** for each transform:
  * `TrimSpace`: leading spaces, trailing spaces, both, no spaces (no-op).
  * `Default`: empty string returns default, non-empty returns original value.
* [ ] **Milestone:** Run existing end-to-end test with a profile that includes `"transforms": [{"type": "TrimSpace"}]` and verify output is trimmed.

---

## Step 1.5: Update the CLI & Settings

*Minor cleanup to align the CLI flags and settings file with the new schema.*

* [ ] Update the profile loader in the CLI to use the migration helper for old-format profiles.
* [ ] Confirm `-dry-run` flag still works correctly.
* [ ] Update `settings.json` struct to add `LastConnector string` (defaults to `"csv"` — no behavior change yet, prep for Phase 2).
* [ ] **Milestone:** All existing Phase 1A tests pass. `build_and_test.ps1` runs green.

---

# 🖥️ Stage 2: Fyne Wizard GUI

> The refactor is complete. Now build the GUI on the clean foundation. The wizard is step-by-step — one screen, one decision. The user can go Back at any point.

## Step 2.1: The Wizard Shell (`internal/ui/wizard.go`)

*Goal: A navigable multi-step container. No real content yet — just the skeleton that all steps plug into.*

* [ ] Define `WizardState` struct to carry user choices between steps:
  ```go
  type WizardState struct {
      InputConnectorName  string
      InputConfig         connector.ConnectionConfig
      InputHeaders        []string
      InputPreviewRows    [][]string
      OutputConnectorName string
      OutputConfig        connector.ConnectionConfig
      Mappings            []types.Mapping
      Settings            types.Settings
      ProfileName         string
  }
  ```
* [ ] Create `Wizard` struct with:
  * A `fyne.Window` reference.
  * A `*WizardState` shared across all steps.
  * A `steps []Step` slice and `currentStep int` index.
  * `Next()`, `Back()`, and `GoTo(n int)` navigation methods.
* [ ] Build the outer chrome:
  * Top: Application title "Data Coupler" + current step indicator (e.g., "Step 2 of 4").
  * Middle: `container.NewMax()` — swappable content area where each step renders.
  * Bottom: `Back` button (left), `Next` / `Run` button (right). Back is hidden on step 1.
* [ ] Define the `Step` interface:
  ```go
  type Step interface {
      Title()    string
      Content()  fyne.CanvasObject
      Validate() error  // Returns nil if Next is allowed, error message if not.
  }
  ```
* [ ] **Milestone:** Window opens with the chrome visible. Next/Back cycle through placeholder step screens.

---

## Step 2.2: Step 1 — Choose Input Source (`internal/ui/step_source.go`)

* [ ] Render two large clickable cards:
  * 📄 **CSV File** — "Load data from a .csv file on your computer."
  * 🗄️ **SQL Database** — "Connect to Microsoft SQL Server, MySQL, SQLite, or PostgreSQL." *(grayed out with "Coming in Phase 2" badge — wired up but not yet active)*
* [ ] On CSV card click: set `WizardState.InputConnectorName = "csv"`, enable Next button.
* [ ] `Validate()`: returns error if no source is selected.

---

## Step 2.3: Step 2 — CSV File Input + Preview (`internal/ui/step_input_csv.go`)

* [ ] File picker button (filters to `.csv` files).
* [ ] On file selection: call `CSVConnector.Columns()` and `CSVConnector.Rows()` to fetch a preview.
* [ ] Store headers in `WizardState.InputHeaders` and first 10 rows in `WizardState.InputPreviewRows`.
* [ ] Render preview table: `widget.Table` showing the first 10 rows with header labels.
* [ ] Error state: if the file can't be opened or has no headers, show an inline error message (not a popup).
* [ ] `Validate()`: returns error if no file is selected or if the file failed to parse.

---

## Step 2.4: Step 3 — Choose Output (`internal/ui/step_output.go`)

* [ ] Render two large clickable cards:
  * 📄 **CSV File** — saves output to a .csv file.
  * 🐟 **Fishbowl Template** — *(grayed out with "Coming in Phase 4" badge)*
* [ ] On CSV card click: show a save-file dialog or path entry widget. Let user name the output file.
* [ ] Set `WizardState.OutputConnectorName = "csv"` and `WizardState.OutputConfig.Path`.
* [ ] Delimiter selection: small radio group — Comma, Tab, Pipe. Default: Comma.
* [ ] `Validate()`: returns error if no output path is set.

---

## Step 2.5: Step 4 — Map Columns (`internal/ui/step_mapping.go`)

*This is the most complex screen. Keep it simple for 1B — full drag-and-drop can come later.*

* [ ] Two-column layout inside a scrollable container:
  * Left column header: "Output Column" — editable text entry for each desired output column name.
  * Right column header: "From Input Column" — a `widget.Select` dropdown populated from `WizardState.InputHeaders` for each row.
* [ ] "Add Row" button at the bottom to append a new mapping pair.
* [ ] Delete icon on each row to remove it.
* [ ] On changes, keep `WizardState.Mappings` in sync.
* [ ] `Validate()`: returns error if any row has an output column name but no input column selected, or vice versa.

---

## Step 2.6: Step 5 — Review & Run (`internal/ui/step_review.go`)

* [ ] Summary card displaying:
  * Input: filename
  * Output: filename
  * Mappings: count of column mappings defined
* [ ] **Run button** — calls `engine.Run()` in a goroutine (never block the UI thread).
* [ ] Progress bar (infinite / indeterminate) visible during run.
* [ ] On success: show row count processed + an "Open Output File" button.
* [ ] On error: show the error message clearly with a "Go Back" button.
* [ ] After a successful run, show: "Save this as a profile?" with a name entry field and Save button.
  * Save action: marshal the current `WizardState` into a `types.Profile` and write to `/profiles/[name].json`.

---

## Step 2.7: Persistence & App State

* [ ] On startup: read `settings.json`. If a `lastProfilePath` is present, offer "Resume Last Job" on a home screen or auto-load into the wizard.
* [ ] On app close: save current wizard state to `settings.json` (last input folder, last output folder, last profile used).
* [ ] Home screen (shown before the wizard starts): two buttons — "New Conversion" and "Load a Saved Profile."
  * Load Profile: opens a file picker scoped to the `/profiles` folder. Populates `WizardState` from the JSON and jumps the wizard to the Review step.

---

## Step 2.8: Final Polish & Testing

* [ ] Write a Fyne test for wizard navigation (Next/Back state transitions).
* [ ] Test the full end-to-end GUI flow manually: pick a file → map columns → run → verify output CSV.
* [ ] Test "Load Profile" path: save a profile, restart the app, load it, run it, verify identical output.
* [ ] Test the Phase 1A CLI still works after all changes: run `build_and_test.ps1`, verify it passes.
* [ ] **Final Milestone:** A single binary that runs in both CLI mode (flags) and GUI wizard mode (no flags), with modular internals ready for Phase 2 database connectors.
