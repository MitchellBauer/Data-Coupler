---
date created: Saturday, January 17th 2026, 9:27:41 am
date modified: Sunday, March 15th 2026, 12:00:00 pm
---

**Parent:** [[Project - Data Coupler Summary]]
**Status:** Active — [[Task List - Phase 1B (GUI) - Data Coupler Project]]

---

# ✅ Phase 1A: The Engine (CLI) — COMPLETE

**Goal:** A working CSV-to-CSV data mapper with a command-line interface.

* [x] Build the core mapping logic (Go code)
* [x] Create the JSON profile structure
* [x] Build a Command Line Interface to run conversions manually
* [x] Write unit tests for the engine

---

# 📍 Phase 1B: The Wizard GUI (CSV → CSV)

**Goal:** A Fyne-based graphical interface using a wizard / step-by-step layout. This replaces the original single-window GUI plan. Even for simple CSV work, the wizard pattern trains the user on the workflow they'll use in every later phase.

**Wizard Steps (CSV → CSV):**

1. **Welcome / Home** — "Start New Conversion" or "Load a Saved Profile"
2. **Step 1: Choose Input** — File picker for the input CSV. Preview the first 10 rows.
3. **Step 2: Choose Output** — File path for the output CSV. Choose delimiter.
4. **Step 3: Map Columns** — Two-column layout. Left = Input headers. Right = Output column name (editable). Drag or dropdown to connect them.
5. **Step 4: Review & Run** — Summary card. Row count estimate. Big "Run" button.
6. **Step 5: Save Profile (Optional)** — Prompt to name and save this workflow for reuse.

**Tasks:**
* [ ] Build Fyne wizard shell (multi-step navigation with Back/Next/Run buttons)
* [ ] Implement file picker + CSV header preview widget
* [ ] Implement column mapping screen
* [ ] Wire "Run" button to existing engine
* [ ] Implement profile save/load from the GUI
* [ ] Persistence: remember last used profile and last input folder on startup

---

# 📍 Phase 2: Database Connectors

**Goal:** Pull data directly from a SQL database instead of a pre-exported CSV. The first and most important connector is Microsoft SQL Server (for Macola).

**Architecture Note:** A Go `Connector` interface is defined once. Each database driver implements that interface. The engine only ever talks to the interface — it never knows or cares which database is on the other end. Adding a new database = adding one new file.

**Connector Interface (defined in `internal/connector`):**

```go
type Connector interface {
    Name()    string
    Connect(cfg ConnectionConfig) error
    Disconnect() error
    Columns(query string) ([]string, error)
    Rows(query string) (<-chan []string, error)
}
```

**Phase 2A — Microsoft SQL Server (Macola)**
* [ ] Implement `MSSQLConnector` using `go-mssqldb`
* [ ] GUI: "SQL Database" option on the Input Source screen
* [ ] Connection config form: Host, Port, Database, Username, Password
* [ ] "Test Connection" button with visual feedback
* [ ] SQL Query entry field with syntax highlighting (basic)
* [ ] "Preview Query" button — runs query, shows first 10 rows in a table widget
* [ ] Store connection credentials encrypted locally (never in profile JSON)

**Phase 2B — SQLite**
* [ ] Implement `SQLiteConnector` using `go-sqlite3`
* [ ] File picker for the `.db` / `.sqlite` file (no credentials needed)

**Phase 2C — MySQL**
* [ ] Implement `MySQLConnector` using `go-sql-driver/mysql`
* [ ] Reuses same connection config form as MSSQL

**Phase 2D — PostgreSQL**
* [ ] Implement `PostgreSQLConnector` using `lib/pq`
* [ ] Reuses same connection config form as MSSQL

---

# 📍 Phase 3: The Transform Pipeline

**Goal:** Data coming out of a source is rarely clean. Add a per-column transformation step between reading and writing.

**Architecture:** Each `Mapping` in the profile gets a `transforms` array. When the engine processes a row, it runs the value through each transform in order before writing it to the output.

**Profile JSON (updated):**
```json
{
  "inputCol": "CUSTNAME",
  "outputCol": "Customer Name",
  "transforms": [
    { "type": "TrimSpace" },
    { "type": "ToUpper" }
  ]
}
```

**Built-in Transforms (Priority Order):**
* [ ] `TrimSpace` — strip leading/trailing whitespace. **Critical for Macola.**
* [ ] `ToUpper` / `ToLower` — case normalization
* [ ] `DateFormat` — reformat date strings (params: `from`, `to` using Go time format)
* [ ] `Concatenate` — merge multiple input columns into one output column (params: `cols`, `separator`)
* [ ] `Split` — extract one segment of a delimited field (params: `separator`, `index`)
* [ ] `LookupReplace` — swap values via a lookup table defined in the profile (params: `map` object)
* [ ] `Default` — fill empty/blank cells with a fallback value (params: `value`)
* [ ] `Prefix` / `Suffix` — prepend or append a static string

**GUI Tasks:**
* [ ] Add "Transform" column to the mapping screen
* [ ] Transform editor: pick a transform type from a dropdown, fill in params
* [ ] Live before/after preview on a sample of rows when a transform is configured

---

# 📍 Phase 4: Fishbowl Template Library

**Goal:** When the output destination is Fishbowl, the user should never have to look up what columns Fishbowl expects. Select the module, and the output side of the mapping wizard is pre-populated with the official Fishbowl column names, marked required or optional.

**Bundled Templates:**
* [ ] Parts / Products
* [ ] Bill of Materials (BOM)
* [ ] Customers
* [ ] Vendors
* [ ] Purchase Orders
* [ ] Sales Orders
* [ ] Inventory Adjustments

**Tasks:**
* [ ] Store Fishbowl templates as embedded JSON in the binary (`embed.FS`)
* [ ] Add "Fishbowl Template" as an Output type option in the wizard
* [ ] Template picker screen: grid of module cards with descriptions
* [ ] Required columns marked with visual indicator; missing required fields block the Run step with a clear warning
* [ ] Validation summary before running: "3 required fields are empty"

---

# 📍 Phase 5: Polish & Distribution

* [ ] **Audit Log** — timestamped log file of every conversion run: profile name, input source, row count, any errors, duration
* [ ] **Auto-Update Check** — ping a GitHub releases endpoint on startup, notify user if a new version is available
* [ ] **Code Signing** — Windows (Authenticode) and macOS (Developer ID) to prevent OS security warnings
* [ ] **Installer / Packaging** — MSI for Windows, DMG for macOS, AppImage for Linux

---

# 🗂 Future Backlog (Not Scheduled)

* Smart Auto-Mapping (fuzzy column name matching)
* Query Builder (visual SQL helper for non-SQL users)
* Delta / Changed-Rows Mode (incremental exports)
* Lookup Table Editor (GUI for `LookupReplace` maps)
* Profile Sharing (`.dcp` export/import)
* Scheduled Runs (timer-based automatic exports)
* Output connector: direct database write (not just CSV export)
