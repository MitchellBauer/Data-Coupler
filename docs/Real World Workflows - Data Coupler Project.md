---
date created: Sunday, March 15th 2026, 12:00:00 pm
date modified: Sunday, March 15th 2026, 12:00:00 pm
---

**Parent:** [[Project - Data Coupler Summary]]

# Real World Workflows — Data Coupler

This document captures concrete, real-world use cases for Data Coupler. Each workflow describes a complete end-to-end scenario: where the data starts, what problems exist with it, what the mapping and transformation looks like, and what the end result is. These serve as both product specification examples and functional test scenarios.

---

# Workflow 1: Macola Parts → Fishbowl Parts Import

**The Problem:**
The business runs Macola as its ERP. Fishbowl is replacing it. Before going live on Fishbowl, the entire parts catalog needs to be imported. Doing this manually is hundreds of rows of copy-paste. Macola's data is also dirty — every string field has trailing spaces baked into it (a quirk of how Macola stores fixed-width data in SQL), so anything copied over raw will look wrong in Fishbowl and break exact-match lookups.

**The Workflow:**

**Step 1 — Input: Microsoft SQL Server (Macola)**

Connect to the Macola database. The connection details:
* Host: `[internal server address]`
* Database: `MACOLA_PROD`
* Credentials stored under ref: `macola-prod`

Query to run:
```sql
SELECT
    ITEMNO,
    DESCRIP,
    CATEGORY,
    UOFMEAS,
    SELLPRICE,
    STDCOST,
    ITEMSTATUS
FROM IM_ITEM
WHERE ITEMSTATUS = 'A'
ORDER BY ITEMNO
```

*Preview shows columns with visible trailing spaces on DESCRIP and CATEGORY.*

**Step 2 — Output: Fishbowl Parts Template**

Select output type: Fishbowl Template → Parts / Products.

The output columns are pre-defined by the template:

| Column | Required | Notes |
|---|---|---|
| Part Number | ✅ | |
| Description | ✅ | |
| Part Type | ✅ | Must be "Inventory", "Service", or "Labor" |
| UOM | ✅ | Unit of measure |
| Price | ☐ | |
| Cost | ☐ | |
| Active | ☐ | "true" or "false" |

**Step 3 — Mapping**

| Input Column (Macola) | Output Column (Fishbowl) | Notes |
|---|---|---|
| ITEMNO | Part Number | |
| DESCRIP | Description | |
| *(static value)* | Part Type | "Inventory" — all parts in this export are inventory items |
| UOFMEAS | UOM | |
| SELLPRICE | Price | |
| STDCOST | Cost | |
| *(derived)* | Active | Based on ITEMSTATUS = 'A' — but since we filtered the query to active only, static "true" |

**Step 4 — Transforms**

| Column | Transform(s) | Reason |
|---|---|---|
| Part Number | `TrimSpace` | Macola trailing spaces |
| Description | `TrimSpace`, `ToUpper` | Trailing spaces + Fishbowl convention |
| UOM | `TrimSpace`, `ToUpper` | Trailing spaces + must match Fishbowl UOM codes |
| Part Type | `Default("Inventory")` | Not sourced from Macola — static value injected |
| Active | `Default("true")` | Not sourced from Macola — static value injected |

**Step 5 — Result**

A clean `fishbowl-parts.csv` with all required columns, no trailing spaces, ready to drag into Fishbowl's import tool. Zero manual reformatting.

**Save as Profile:** `macola-parts-to-fishbowl`

---

# Workflow 2: Macola Customers → Fishbowl Customers Import

**The Problem:**
Same migration scenario. Customer records need to move from Macola to Fishbowl. Macola stores customer name, billing address, and contact info split across several columns. Fishbowl wants a single "Name" field and separate address fields.

**Query:**
```sql
SELECT
    CUSTNO,
    CUSTNAME,
    ADDR1,
    ADDR2,
    CITY,
    STATE,
    ZIPCODE,
    COUNTRY,
    PHONE1,
    EMAIL
FROM AR_CUST
WHERE CUSTSTATUS = 'A'
```

**Mapping & Transforms:**

| Input | Output | Transform |
|---|---|---|
| CUSTNO | Customer ID | `TrimSpace` |
| CUSTNAME | Name | `TrimSpace` |
| ADDR1 | Address | `TrimSpace` |
| ADDR2 | Address 2 | `TrimSpace` |
| CITY | City | `TrimSpace` |
| STATE | State | `TrimSpace` |
| ZIPCODE | Zip | `TrimSpace` |
| COUNTRY | Country | `TrimSpace`, `Default("US")` |
| PHONE1 | Main Phone | `TrimSpace` |
| EMAIL | Email | `TrimSpace`, `ToLower` |

*Note: `Default("US")` on Country means any blank country field gets filled in as "US" — appropriate since the entire customer base is domestic.*

**Save as Profile:** `macola-customers-to-fishbowl`

---

# Workflow 3: CSV Re-format (PayClock → Paychex Payroll)

**The Problem:**
The time clock system exports a weekly hours report as a CSV. The payroll processor (Paychex) requires a CSV in a completely different column order with different header names. An accountant does this manually every week by opening both files and copy-pasting. It takes 20 minutes and introduces errors.

**Input CSV (PayClock export):**
```
EmployeeID, LastName, FirstName, Hours, OvertimeHours, Department
10042, Smith, John, 40.00, 2.50, Warehouse
```

**Output CSV (Paychex format):**
```
Worker_ID, Worker_Name, Reg_Hours, OT_Hours, Dept_Code
10042, John Smith, 40.00, 2.50, Warehouse
```

**Mapping & Transforms:**

| Input | Output | Transform |
|---|---|---|
| EmployeeID | Worker_ID | *(none)* |
| FirstName + LastName | Worker_Name | `Concatenate(cols: "FirstName,LastName", separator: " ")` |
| Hours | Reg_Hours | *(none)* |
| OvertimeHours | OT_Hours | *(none)* |
| Department | Dept_Code | *(none)* |

*The `Concatenate` transform here takes two input columns and merges them. This is the one case where a single output column is fed from multiple input columns.*

**Save as Profile:** `payclock-to-paychex-weekly`

**Workflow in Practice:** The accountant opens Data Coupler, it auto-loads the last used profile (`payclock-to-paychex-weekly`), they click the file picker to select this week's PayClock export, hit Run. Done in 30 seconds.

---

# Workflow 4: Exporting a SQLite Hobby Database to CSV

**The Problem:**
A personal project uses a SQLite database to track book reading history. The goal is to export it to a CSV to share with a friend who uses a spreadsheet.

**Input: SQLite**
File: `~/Documents/reading.db`

Query:
```sql
SELECT title, author, date_finished, rating, notes
FROM books
WHERE date_finished IS NOT NULL
ORDER BY date_finished DESC
```

**Output: CSV**
File: `~/Desktop/my-books.csv`

**Mapping:** Direct 1:1 — column names are already clean. No transforms needed.

**Transforms:**
* `date_finished` → `DateFormat(from: "2006-01-02", to: "01/02/2006")` — reformat from SQLite ISO date to MM/DD/YYYY for readability in the spreadsheet.

**Save as Profile:** `reading-db-export`

---

# Workflow 5: CSV Cleanup (Importing a Messy Vendor Price List)

**The Problem:**
A vendor sends a monthly price list CSV. The headers are inconsistent, some rows have blank Part Numbers, prices include dollar signs and commas (`$1,234.56`), and the category codes are numeric but the internal system expects text labels.

**Input CSV (Vendor):**
```
Part #, Desc., Category, List Price
ABC-001, Widget A, 3, "$12.50"
ABC-002, Widget B, 3, "$1,234.00"
ABC-003, , 5, "$0.99"
```

**Mapping & Transforms:**

| Input | Output | Transforms |
|---|---|---|
| Part # | Part Number | `TrimSpace` |
| Desc. | Description | `TrimSpace`, `Default("No Description")` |
| Category | Category | `LookupReplace(map: {"3": "Hardware", "5": "Fasteners"})` |
| List Price | Price | `TrimSpace`, *(custom: strip `$` and `,`)* |

*The `LookupReplace` transform is perfect for category code translation — the profile JSON stores the full lookup table.*

---

# Workflow Patterns & Lessons Learned

From the above scenarios, several patterns emerge that should inform the design:

**The TrimSpace-first rule:** Any data coming from Macola (or any MSSQL system with fixed-width legacy columns) needs `TrimSpace` on every single string column. The GUI should suggest this automatically when MSSQL is selected as the input source.

**Static / Injected values:** Sometimes an output column has no corresponding input column — it needs a static value (like `Part Type = "Inventory"`). The `Default` transform with an empty input column handles this. The GUI should make "no input column — use a static value" a first-class option in the mapping step, not a workaround.

**Multi-column → single column:** The `Concatenate` transform (First + Last → Full Name) comes up constantly. It should be prominently available in the transform picker.

**Query as the filter:** For database sources, the SQL query itself handles filtering (e.g., `WHERE ITEMSTATUS = 'A'`). Users should be encouraged to filter in the query rather than relying on Data Coupler to filter rows — it keeps the tool simpler and the query more readable.

**Profiles are the product:** The real value isn't the one-time conversion — it's the saved profile. Once `macola-parts-to-fishbowl` is built and tested, it can be run again in 10 seconds any time the parts catalog is updated before go-live. This should be emphasized in the UI.
