---
date created: Saturday, January 17th 2026, 8:56:22 am
date modified: Wednesday, February 4th 2026, 6:10:00 pm
---

**Parent:** [[Project - Data Coupler Summary]]

# 1. Environment Setup (The "Getting Started" Guide)

*For when I come back to this project in 12 months.*

* **Go Version:** 1.22+
* **Fyne Installation:** `go install fyne.io/fyne/v2/cmd/fyne@latest`
* **Build Command:** `go build -ldflags "-H=windowsgui" -o data-coupler.exe ./cmd/datacoupler`

# 2. Project Structure

* `/cmd/datacoupler` - **Main entry point.** Contains `main.go` which decides between CLI and GUI modes.
* `/internal/engine` - **Core Logic.** Pure CSV mapping logic (no UI dependencies).
* `/internal/ui` - **Fyne GUI.** Windows, widgets, and event handling.
* `/internal/config` - **I/O.** Loading and saving JSON profiles.
* `/internal/types` - **Data Models.** Shared structs like `Profile` and `Mapping`.
* `profiles/` - User-generated JSON profile files.

# 3. Data Structures

## Profile JSON Schema

*Drafting how the save file looks.*

```json
{
  "id": "payclock-to-paychex",
  "name": "PayClock to PayChex Export",
  "description": "Weekly hours export for payroll.",
  "settings": {
    "skipHeader": true,
    "delimiter": ","
  },
  "mappings": [
    {
      "inputCol": "EmployeeID",
      "outputCol": "Worker_ID",
      "transform": "none"
    }
  ]
}

```

# 4. Known Technical Constraints

* **Fyne & Windows Legacy:** Text rendering can be blurry on older Windows versions or high-DPI scaling without proper manifest settings.
* **CSV Complexity:** Standard Go CSV parser handles quoted commas, but BOM (Byte Order Mark) from Excel files must be stripped manually before parsing headers.
* **Single Binary Limits:** The CLI and GUI share one executable. This means the file size is larger (~20MB+) even if you only use the CLI features.

## Why split it?

If you keep the "Project Setup" in the main dashboard, you have to scroll past 50 lines of configuration details just to remember what the goal of the project was. By separating them, you keep your thinking clear: **One mode for planning (Dashboard), one mode for engineering (Tech Doc).**
