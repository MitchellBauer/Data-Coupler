---
date created: Saturday, January 17th 2026, 8:27:28 am
date modified: Sunday, March 15th 2026, 12:00:00 pm
---

**Status:** 🟡 In Progress | **Last Updated:** 2026-03-15
**Repository:** [Link to Git] | **Current Version:** v0.2 (Phase 1A Complete)
**Current Task List**: [[Task List - Phase 1B (GUI) - Data Coupler Project]]

# 1. Quick Navigation

*Manage the project from here.*

* [[Roadmap - Data Coupler Project]] (Kanban or Todo list)
* [[Technical Design Document - Data Coupler Project]] (Detailed architecture)
* [[Real World Workflows - Data Coupler Project]] (Step-by-step examples for real use cases)
* [[Research & Resources Dashboard - Data Coupler Project]] (Links to Fyne docs, CSV specs, etc.)
* [[Setup Environment - Data Coupler Project]]
* [[Maintenance and Build Procedure - Data Coupler Project]]
* [[Daily Log / Dev Journal]]

---

# 2. The Pitch (Summary)

**Vision:** A desktop ETL wizard — a guided, step-by-step tool that pulls data from where it lives (a SQL database or a CSV file), cleans and transforms it, maps it to the format the destination needs, and exports it — ready to import. No manual exports, no reformatting spreadsheets by hand, no guesswork.

**Goal:** A streamlined, free, personal tool robust enough for real professional use. Targeted squarely at the small business data migration problem: moving data between legacy systems (like Macola) and modern platforms (like Fishbowl) without expensive middleware or consultants.

**Design Philosophy:**
* **Plain language everywhere.** "Input" and "Output." Not ETL jargon. Not "Source/Target/Ingest."
* **Wizard-driven, not form-driven.** Guide the user through the workflow step by step. Never dump everything on one screen.
* **Profiles are first-class citizens.** Every conversion can be saved as a reusable profile. Repeat jobs are one click.
* **Single binary.** No installers, no runtimes, no dependency hell. Drop it in a folder and run it.
* **Built to be extended, not rewritten.** A clean connector interface means adding a new database is adding one file — not rebuilding the app.

---

# 3. Core Feature Requirements

## Phase 1: Simple Data Coupler Tool (CSV → CSV)

* [x] Input .csv → Output .csv
* [x] Command Line tool (Phase 1A complete)
* [ ] **GUI: Wizard Interface** — step-by-step workflow replacing the planned single-window GUI
* [ ] **Profile System:** Save/Load mappings as JSON files
* *Workflow:* Load Profile → Select Input → Convert
* *Auto-load:* Application remembers the last used profile on startup
* [ ] **Smart Templating:** JSON structure simple enough for an LLM to generate from a text description
* [ ] Profile management: rename, duplicate, delete, switch
* [ ] Tests

## Phase 2: Database Connectors

* [ ] **Connector Interface:** A clean Go interface any new connector must implement. Register at startup. No hardcoding.
* [ ] **Built-in Connectors (Priority Order):**
  1. CSV (file) — already implemented in engine
  2. Microsoft SQL Server — primary use case (Macola)
  3. SQLite
  4. MySQL — future Fishbowl database
  5. PostgreSQL — personal projects
* [ ] **SQL Input Workflow:** Enter connection details → Write or paste a query → Preview results → Proceed to mapping
* [ ] **Credential storage:** Encrypted local storage. Never plaintext passwords in profile JSON.

## Phase 3: The Transform Pipeline

* [ ] **Built-in Transforms (Priority Order):**
  * `TrimSpace` — strips leading/trailing whitespace (critical for Macola data)
  * `ToUpper` / `ToLower` — case normalization
  * `DateFormat(from, to)` — reformat date strings (e.g., `MM/DD/YYYY` → `YYYY-MM-DD`)
  * `Concatenate(cols, separator)` — merge multiple input columns into one
  * `Split(index, separator)` — extract one part of a delimited field
  * `LookupReplace(map)` — swap codes for values (e.g., Macola category code → Fishbowl category name)
  * `Default(value)` — fill empty cells with a fallback value
  * `Prefix / Suffix` — prepend or append a static string
* [ ] Transforms are applied per-column in the profile JSON, executed in a pipeline
* [ ] Transform preview in the GUI: show before/after for a sample of rows

## Phase 4: Fishbowl Template Library

* [ ] **Bundled Fishbowl Import Templates** for all major Fishbowl modules:
  * Parts / Products
  * Bill of Materials (BOM)
  * Customers
  * Vendors
  * Purchase Orders
  * Sales Orders
  * Inventory Adjustments
* [ ] Templates define required vs. optional columns with validation
* [ ] Selecting a Fishbowl template pre-populates the Output side of the mapping wizard
* [ ] Validation step warns on missing required fields before running

## Phase 5: Polish & Distribution

* [ ] Audit log: timestamped record of every conversion run (profile used, row count, errors)
* [ ] Auto-update check
* [ ] Code signing (Windows & macOS)

---

# 4. Technical Architecture

* **Language:** Go (Golang)
* **GUI Framework:** Fyne
* **Persistence:** JSON (profiles) + encrypted local store (credentials)
* **Why this stack?**
  * **Single binary deployment.** No installers, no runtime, one executable.
  * **Go's backwards compatibility** means this tool will still build in 5 years with minimal maintenance.
  * **Fyne** is cross-platform (Windows, macOS, Linux) and keeps the UI code clean.
  * **Connector interface** keeps database-specific code isolated — adding MySQL doesn't touch the engine.

---

# 5. Future Ideas (Backlog)

These are not on the roadmap yet, but worth capturing:

* **Smart Auto-Mapping** — fuzzy-match input column names to output column names and suggest mappings automatically.
* **Data Preview Panel** — show a live preview of the first N rows before and after transforms, so you can spot issues before committing.
* **Query Builder** — a simple visual helper (table picker, column checkboxes, basic WHERE clause builder) for users who aren't fluent in SQL.
* **Delta / Changed-Rows Mode** — compare against a key column and only export rows that are new or changed since the last run. Useful for incremental syncs.
* **Lookup Table Editor** — a GUI for building and editing the `LookupReplace` transform maps without touching JSON.
* **Profile Sharing** — export a profile as a single `.dcp` file that can be emailed to a colleague and imported.
* **Scheduled Runs** — set a profile to run on a timer and auto-save output to a folder (useful for recurring exports).
* **Validation Rules** — per-column rules beyond required/optional: min length, numeric only, regex match, etc.
* **Reverse Mapping** — given an output template, auto-generate the starter profile JSON with empty input column fields.
