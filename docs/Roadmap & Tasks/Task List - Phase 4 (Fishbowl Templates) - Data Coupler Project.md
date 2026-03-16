---
date created: Sunday, March 15th 2026, 12:00:00 pm
date modified: Sunday, March 15th 2026, 12:00:00 pm
---

**Focus:** Fishbowl Import Template Library + Guided Mapping UI
**Parent:** [[Roadmap - Data Coupler Project]]
**Prerequisite:** [[Task List - Phase 3 (Transform Pipeline) - Data Coupler Project]]

> **Why Fishbowl templates?**
> Every Fishbowl import module expects a very specific set of column names. Users currently have to look these up in Fishbowl documentation, type them exactly (case-sensitive), and remember which ones are required. One typo means the import silently skips the field.
>
> Phase 4 embeds the official column definitions for each Fishbowl module directly into the binary. When the user picks "Fishbowl Parts" as their output, the output column names are already filled in — required fields are marked, optional fields are pre-populated, and the Run button is blocked until every required field has a mapping. The user's only job is choosing which input column feeds each output column.

---

# 📦 Stage 1: Template Definitions (`internal/templates/`)

*Templates are embedded into the binary at compile time. They are never external files the user can accidentally delete or corrupt.*

## Step 1.1: Package Setup

* [ ] Create `internal/templates/templates.go`:
  ```go
  //go:embed fishbowl/*.json
  var fishbowlFS embed.FS
  ```
* [ ] Define `Template` and `TemplateColumn` structs:
  ```go
  type Template struct {
      ID          string           `json:"id"`
      Name        string           `json:"name"`
      Description string           `json:"description"`
      Columns     []TemplateColumn `json:"columns"`
  }

  type TemplateColumn struct {
      Name        string `json:"name"`
      Required    bool   `json:"required"`
      Description string `json:"description"`
  }
  ```
* [ ] Implement `LoadTemplate(id string) (Template, error)` — reads `fishbowlFS`, unmarshals JSON.
* [ ] Implement `ListTemplates() ([]Template, error)` — reads all `.json` files from `fishbowlFS`, returns slice.

## Step 1.2: Author the Template JSON Files

*Each file lives at `internal/templates/fishbowl/<name>.json`. Column names must exactly match what Fishbowl's import utility expects.*

* [ ] `parts.json`:
  * Required: Part Number, Description, Part Type, UOM
  * Optional: Price, Cost, Active, Weight, Weight UOM, Length, Width, Height, Size UOM, Default Location, Custom Field 1–5
* [ ] `bom.json` (Bill of Materials):
  * Required: BOM Number, BOM Type, Quantity
  * Optional: Part Number, Description, Instructions, Stage, Note
* [ ] `customers.json`:
  * Required: Customer Name
  * Optional: Customer Number, Account Number, Default Terms, Tax Rate, Active, Address Name, Address, City, State, Zip, Country, Main Phone, Email, Sales Rep
* [ ] `vendors.json`:
  * Required: Vendor Name
  * Optional: Vendor Number, Account Number, Default Terms, Active, Address Name, Address, City, State, Zip, Country, Main Phone, Email, Currency
* [ ] `purchase_orders.json`:
  * Required: PO Number, Vendor Name, Part Number, Quantity
  * Optional: Date, Description, Unit Cost, Location, Note
* [ ] `sales_orders.json`:
  * Required: SO Number, Customer Name, Part Number, Quantity
  * Optional: Date, Description, Unit Price, Ship To Name, Ship To Address, Ship To City, Ship To State, Ship To Zip, Note
* [ ] `inventory_adjustments.json`:
  * Required: Part Number, Quantity
  * Optional: Location, Lot Number, Expiration Date, Note, Unit Cost
* [ ] ⚠️ **Constraint:** Column names are case-sensitive and must match Fishbowl exactly. Cross-reference against the Fishbowl CSV import documentation before finalizing. When in doubt, run a test import in Fishbowl with a one-row file.

## Step 1.3: Unit Tests

* [ ] `ListTemplates()` returns all 7 templates with no error.
* [ ] `LoadTemplate("fishbowl/parts")` returns the correct struct with required fields identified.
* [ ] `LoadTemplate("fishbowl/nonexistent")` returns a clear `ErrNotFound`-style error.
* [ ] Every template has at least one required column (sanity check).
* [ ] **Milestone:** `go test ./internal/templates/...` passes. `go build ./...` succeeds with all templates embedded.

---

# 🖥️ Stage 2: GUI — Output Template Picker

*Activate the Fishbowl Template card in the output step that has been grayed out since Phase 1B.*

## Step 2.1: Enable Fishbowl Card (`internal/ui/step_output.go`)

* [ ] Remove the "Coming in Phase 4" badge and disabled state from the Fishbowl Template card.
* [ ] On Fishbowl card click: navigate to the template picker sub-screen (built in Step 2.2).
* [ ] The CSV card path remains unchanged.

## Step 2.2: Template Picker Sub-Screen

* [ ] Create a scrollable grid of module cards — one card per template returned by `templates.ListTemplates()`.
* [ ] Each card shows:
  * Module name (e.g., "Fishbowl Parts Import")
  * Short description
  * Required column count (e.g., "4 required fields")
* [ ] On card click:
  * Set `WizardState.OutputConnectorName = "csv"`.
  * Set `WizardState.OutputTemplate` to the template ID.
  * Load the template's columns and store in `WizardState` so the mapping step can use them.
  * Show a file-path entry for the output CSV (same as the plain CSV path).
* [ ] Back button returns to the two-card output selection.
* [ ] `Validate()`: returns error if no template is selected or no output file path is set.

---

# 🗺️ Stage 3: Mapping Screen — Template Mode (`internal/ui/step_mapping.go`)

*When a Fishbowl template is active, the mapping screen changes from a blank canvas to a guided form.*

## Step 3.1: Detect Template Mode

* [ ] At the start of `step_mapping.go`'s `Content()`, check if `WizardState.OutputTemplate` is set.
* [ ] If set: render in **Template Mode**. If empty: render in the existing **Free Mode**.

## Step 3.2: Template Mode Layout

* [ ] Pre-populate the output column list from the loaded template columns (do not use the editable text entries from Free Mode).
* [ ] Output column names are **read-only labels**, not entry widgets.
* [ ] Required columns are marked with a red asterisk (**★**) next to the column name.
* [ ] Optional columns are visually de-emphasized (lighter text or italics).
* [ ] Each row has a `widget.Select` dropdown on the right, populated with `WizardState.InputHeaders` plus a `"— not mapped —"` option.
* [ ] The "Add Row" and "Delete" buttons are **hidden** — the template defines all output columns.
* [ ] On dropdown change: update `WizardState.Mappings` synchronously.
* [ ] Unmapped required columns show an inline warning icon (⚠️) next to the asterisk. This is informational — it doesn't block the Next button here (validation is enforced at the Review step).

---

# ✅ Stage 4: Validation & Review Step Updates (`internal/ui/step_review.go`)

*The Review step is the final gate. Required fields must be mapped before the run can start.*

## Step 4.1: Template Validation

* [ ] Before enabling the **Run** button, check all required template columns:
  * For each required column in the template, verify that a non-empty `InputCol` is mapped.
* [ ] If any required columns are unmapped:
  * **Disable the Run button.**
  * Show a validation summary card (red background):
    > ⚠️ **2 required fields are not mapped:**
    > • Part Type
    > • UOM
  * The summary lists each unmapped required column by name.
* [ ] If all required columns are mapped:
  * Run button is enabled.
  * Show a green confirmation line: "✓ All required fields are mapped."

## Step 4.2: Review Summary Update

* [ ] When a template is active, the review summary card shows:
  * Input: source name/connector type
  * Output: output file path + template name (e.g., "Fishbowl Parts Import")
  * Mappings: "X of Y fields mapped (Z required)"

* [ ] **Milestone:** Full manual test: MSSQL source → Fishbowl Parts template → map all required fields → run → import the output CSV into a Fishbowl test environment and verify the import succeeds.

---

# ✅ Stage 5: End-to-End Verification

* [ ] Test all 7 templates load and render correctly in the template picker.
* [ ] Test Free Mode mapping is unaffected when no template is selected.
* [ ] Test: attempt to run with a required field unmapped — Run button must be blocked.
* [ ] Test: map all required fields — Run button enables.
* [ ] Test each of the 5 real-world workflows from `Real World Workflows - Data Coupler Project.md` that involve Fishbowl output.
* [ ] `go test ./...` passes with zero errors.
* [ ] `go build ./...` produces a binary — verify all template JSON files are embedded (binary contains the column names as strings).
