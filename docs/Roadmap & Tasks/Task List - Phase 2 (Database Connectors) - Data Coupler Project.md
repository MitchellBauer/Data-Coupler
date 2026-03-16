---
date created: Sunday, March 15th 2026, 12:00:00 pm
date modified: Sunday, March 15th 2026, 12:00:00 pm
---

**Focus:** SQL Database Connectors + Encrypted Credential Storage
**Parent:** [[Roadmap - Data Coupler Project]]
**Prerequisite:** [[Task List - Phase 1B (GUI) - Data Coupler Project]]

> **Why database connectors now?**
> Phase 1B completed a fully functional CSV-to-CSV wizard. The architecture was deliberately built around a `Connector` interface so that adding new data sources requires no changes to the engine or GUI shell. Phase 2 activates that architecture. The first and most important connector is Microsoft SQL Server, which is the backend for Macola — the primary real-world use case driving this project.

---

# 🔐 Stage 1: Credential Store (`internal/credentials/`)

*SQL passwords must never appear in profile JSON files. This stage builds the encrypted local vault that the engine and GUI will use to store and retrieve credentials at runtime.*

## Step 1.1: Define the Store Interface

* [ ] Define the `Store` interface in `internal/credentials/credentials.go`:
  ```go
  type Store interface {
      Save(ref string, password string) error
      Load(ref string) (string, error)
      Delete(ref string) error
      List() ([]string, error)
  }
  ```
* [ ] Define `ErrNotFound` sentinel error for `Load()` when the ref doesn't exist.
* [ ] Define `FileStore struct` that implements `Store` using a local encrypted file.

## Step 1.2: Implement Encryption

* [ ] Use **AES-256-GCM** for encryption. Use Go's `crypto/aes` and `crypto/cipher` standard library — no external dependencies.
* [ ] Derive the encryption key from a machine-specific identifier:
  * Windows: use the machine's `MachineGuid` from the registry (`HKLM\SOFTWARE\Microsoft\Cryptography`).
  * macOS/Linux: use `/etc/machine-id` or a fallback UUID stored in the app directory.
  * ⚠️ **Constraint:** The goal is preventing casual exposure (e.g., someone finding a shared profile JSON). This is not enterprise-grade secrets management. A machine-derived key is sufficient.
* [ ] `Save()`: encrypt the password with AES-GCM, store the ciphertext + nonce in `credentials.bin` as a JSON map keyed by `ref`.
* [ ] `Load()`: read `credentials.bin`, decrypt the entry for the given `ref`.
* [ ] `Delete()`: remove the entry and rewrite the file.
* [ ] `List()`: return all stored `ref` keys (not passwords).
* [ ] ⚠️ **Constraint:** Never log, print, or return a plaintext password in an error message.

## Step 1.3: Unit Tests

* [ ] Save and Load round-trip: saved password is recovered exactly.
* [ ] Load with unknown ref returns `ErrNotFound`.
* [ ] Delete removes the entry; subsequent Load returns `ErrNotFound`.
* [ ] File is not human-readable plaintext after a Save.
* [ ] **Milestone:** `go test ./internal/credentials/...` passes.

---

# 🗄️ Stage 2: Database Connectors

*Each connector is one self-contained package implementing the `Connector` interface from `internal/connector/connector.go`. The engine and GUI never import database drivers directly.*

## Step 2A: Microsoft SQL Server (`internal/connector/mssql/`)

*Priority connector — this is the Macola backend.*

* [ ] Add dependency: `go get github.com/microsoft/go-mssqldb`
* [ ] Create `internal/connector/mssql/mssql.go` with `MSSQLConnector struct`.
* [ ] Implement `Name()`: returns `"mssql"`.
* [ ] Implement `Connect(cfg connector.ConnectionConfig) error`:
  * Build DSN: `sqlserver://username:password@host:port?database=dbname`
  * Password comes from `cfg.Password` (resolved from credential store before this call — the connector itself does not touch the store).
  * Call `sql.Open("sqlserver", dsn)` and `db.Ping()` to validate.
  * Store `*sql.DB` internally.
* [ ] Implement `Columns(query string) ([]string, error)`:
  * Run `query` with `db.QueryContext`, read column names from `rows.Columns()`, close rows immediately.
  * ⚠️ **Constraint:** Do not fetch actual data rows in `Columns()` — this is used for preview headers only and must be fast.
* [ ] Implement `Rows(query string) (<-chan []string, error)`:
  * Launch a goroutine that streams rows as `[]string` through the returned channel.
  * Close the channel when the query is exhausted or an error occurs.
  * All values converted to `string` via `fmt.Sprintf`.
* [ ] Implement `Disconnect() error`: call `db.Close()`.
* [ ] Register in `cmd/datacoupler/main.go`:
  ```go
  connector.Register(&mssqlconn.MSSQLConnector{})
  ```
* [ ] Unit tests (use `database/sql` with a mock driver — no live server required):
  * `Connect()` returns error on bad DSN.
  * `Columns()` returns correct column names.
  * `Rows()` streams all rows and closes the channel.
* [ ] Document in `docs/Technical Design and Setup/` how to run integration tests against a Docker SQL Server instance.
* [ ] **Milestone:** `go build ./...` succeeds. CLI run with `-profile` pointing to a valid MSSQL profile produces correct output CSV.

---

## Step 2B: SQLite (`internal/connector/sqlite/`)

* [ ] Add dependency (CGO-free preferred for single binary): `go get modernc.org/sqlite`
* [ ] Create `internal/connector/sqlite/sqlite.go` with `SQLiteConnector struct`.
* [ ] Implement `Name()`: returns `"sqlite"`.
* [ ] Implement `Connect(cfg connector.ConnectionConfig) error`:
  * Open `cfg.FilePath` using `sql.Open("sqlite", filePath)`.
  * No Host/Port/credentials needed.
* [ ] Implement `Columns()`, `Rows()`, `Disconnect()` following the same pattern as MSSQL.
* [ ] Register in `main.go`.
* [ ] Unit tests: open a temp SQLite file, insert test rows, assert `Columns()` and `Rows()` return expected data.
* [ ] **Milestone:** End-to-end test: SQLite `.db` file → CSV output via CLI.

---

## Step 2C: MySQL (`internal/connector/mysql/`)

* [ ] Add dependency: `go get github.com/go-sql-driver/mysql`
* [ ] Create `internal/connector/mysql/mysql.go` with `MySQLConnector struct`.
* [ ] Implement `Name()`: returns `"mysql"`.
* [ ] `Connect()` DSN format: `username:password@tcp(host:port)/database`
* [ ] Implement `Columns()`, `Rows()`, `Disconnect()` following the same pattern as MSSQL.
* [ ] Register in `main.go`.
* [ ] Unit tests with mock driver.

---

## Step 2D: PostgreSQL (`internal/connector/postgres/`)

* [ ] Add dependency: `go get github.com/lib/pq`
* [ ] Create `internal/connector/postgres/postgres.go` with `PostgreSQLConnector struct`.
* [ ] Implement `Name()`: returns `"postgres"`.
* [ ] `Connect()` DSN format: `host=h port=p dbname=db user=u password=pw sslmode=disable`
* [ ] Implement `Columns()`, `Rows()`, `Disconnect()` following the same pattern as MSSQL.
* [ ] Register in `main.go`.
* [ ] Unit tests with mock driver.

---

# 🖥️ Stage 3: GUI — SQL Input Step (`internal/ui/`)

*With working connectors behind us, activate the SQL Database option in the wizard that has been grayed out since Phase 1B.*

## Step 3.1: Enable SQL Source Card (`internal/ui/step_source.go`)

* [ ] Remove the "Coming in Phase 2" badge and disabled state from the SQL Database card.
* [ ] On SQL card click: set `WizardState.InputConnectorName` to `""` (sub-selection required) and navigate to a DB type sub-selection screen.
* [ ] Sub-selection: four cards — MSSQL, SQLite, MySQL, PostgreSQL. On selection, set `WizardState.InputConnectorName` to the connector's `Name()` string.

## Step 3.2: New SQL Input Step (`internal/ui/step_input_sql.go`)

* [ ] Create `step_input_sql.go` implementing the `Step` interface.
* [ ] **Server connection form** (shown for MSSQL, MySQL, PostgreSQL — hidden for SQLite):
  * Host entry
  * Port entry (numeric, with sensible defaults: MSSQL=1433, MySQL=3306, PostgreSQL=5432)
  * Database name entry
  * Username entry
  * Password entry (masked)
  * "Credential name" entry — the `credentialRef` used to save/load from the store.
  * "Save credentials" checkbox — if checked, saves to credential store on successful test.
* [ ] **SQLite variant**: file picker for `.db` / `.sqlite` file only. No server fields shown.
* [ ] **"Test Connection" button**:
  * Calls the appropriate connector's `Connect()` then `Disconnect()`.
  * On success: show inline green checkmark ✓ "Connection successful."
  * On failure: show inline red ✗ with the error message.
  * ⚠️ **Constraint:** Run the connection test in a goroutine — never block the UI thread.
* [ ] **SQL query text area**: multi-line entry widget.
* [ ] **"Preview Query" button**:
  * Calls `Columns()` and fetches first 10 rows via `Rows()`.
  * Stores headers in `WizardState.InputHeaders` and rows in `WizardState.InputPreviewRows`.
  * Renders results in a `widget.Table`.
  * On error: show inline error message.
* [ ] **Macola suggestion banner** (shown only when `InputConnectorName == "mssql"`):
  * Persistent (non-dismissible) info box: "Macola data often has trailing spaces. Consider adding TrimSpace to all columns in the Transform step."
* [ ] `Validate()`: returns error if connection test has not passed, or if query is empty.
* [ ] **Milestone:** Full GUI flow: select SQL Database → MSSQL → fill form → Test Connection → enter query → Preview → Next proceeds to mapping step.

---

# ✅ Stage 4: End-to-End Verification

* [ ] **CLI test:** Run `go run ./cmd/datacoupler -profile profiles/mssql-test.json -out result.csv` against a Docker SQL Server. Verify output matches expected CSV.
* [ ] **GUI test (manual):** Full wizard flow for each connector type (MSSQL, SQLite) from source selection to CSV output.
* [ ] **Credential test:** Save credentials, restart app, load a profile with `credentialRef`, verify the engine resolves the password without prompting.
* [ ] `go test ./...` passes with zero errors.
* [ ] `build.bat` produces a working single binary.
