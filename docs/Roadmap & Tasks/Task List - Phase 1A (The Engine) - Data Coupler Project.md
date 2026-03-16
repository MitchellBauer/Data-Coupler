---
date created: Saturday, January 17th 2026, 9:28:54 am
date modified: Wednesday, February 4th 2026, 6:10:00 pm
---

**Focus:** Core Logic, Data Structures, & CLI
**Parent:** [[Roadmap - Data Coupler Project]]

# 📂 Step 0: Project Initialization

*Setting up the workspace foundation.*

* [x] **Init Module:** Run `go mod init data-coupler`
- [x] Create Directory Structure:
	* /cmd/datacoupler (Main entry point)
	* /internal/engine (Core logic)
	* /internal/config (Profile I/O)
	* /internal/types (Data structs)
	* /internal/ui (Future GUI code)

# 🛠️ Step 1: The Core Logic (`internal/`)

*Goal: Write the "Brains" that don't care about the UI.*

## 1.1 Data Structures (`internal/types`)

* [x] **Define Structs:** Create `Profile`, `Mapping`, and `IOConfig` structs in a dedicated `types` package.
* [x] **Create Mock Data:** Manually write a `test_profile.json` and `test_data.csv` in the root folder.

## 1.2 The Engine (`internal/mapper`)

* [x] **CSV Reader (BOM Safe):** Write `ReadCSV` function.
* ⚠️ **Constraint:** Implement logic to detect and strip the UTF-8 Byte Order Mark (BOM) from the first byte, or headers will fail to match.
* [x] **Header Validator:** Write logic to read the first row and map `ColumnName -> Index`.
* ⚠️ **Constraint:** Return an error immediately if a `SourceCol` in the profile does not exist in the CSV headers.
* [x] **The Mapper Logic:** Write `MapRow(inputRow []string, headerMap map[string]int, profile types.Profile)`.
* *Logic:* Create a slice of size `len(profile.Mappings)`, look up indices, and fill.
* [x] **The Coordinator:** Create `Run(input, output, profile)` to tie it all together.

## 1.3 Quality Assurance (Tests)

* [x] **Unit Test - Mapping:** Test `MapRow` with a hardcoded map.
* [x] **Unit Test - BOM Handling:** Create a test case with a BOM-prefixed string to ensure it parses correctly.
* [x] **Unit Test - Missing Columns:** Ensure the validator returns the correct error message when a column is missing.

# 💻 Step 2: The CLI Wrapper (`cmd/cli`)

*Goal: The interface for testing the engine.*

* [x] **Main Entry Point:** Create `main.go`.
* [x] **Flags:** Implement `flag` package:
* `-in` (input csv)
* `-out` (output csv)
* `-profile` (path or ID)
* `-dry-run` (print to console instead of file - *Added for easier testing*)
* [x] **Profile Loader:** Simple function to read the JSON file from the `-profile` path.
* [x] **Milestone Check:** Run `./data-coupler -in test.csv -out result.csv -profile profiles/test.json` and verify the output.
