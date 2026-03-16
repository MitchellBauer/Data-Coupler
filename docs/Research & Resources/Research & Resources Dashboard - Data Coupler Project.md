---
date created: Saturday, January 17th 2026, 9:11:23 am
date modified: Wednesday, February 4th 2026, 6:10:00 pm
---

**Parent:** [[Project - Data Coupler Summary]]
**Tags:** #learning #golang #fyne

# 1. Key Documentation (The "Truth")

* **Go Language:** [pkg.go.dev](https://pkg.go.dev/) (Standard Library)
* **Fyne Framework:**
* [Fyne Tour](https://tour.fyne.io/) (Interactive examples - **Start Here**)
* [Fyne Widget List](https://developer.fyne.io/explore/widgets) (Visual gallery of all components)
* [Fyne Layouts](https://www.google.com/search?q=https://developer.fyne.io/tutorial/layout) (How to arrange items)
* **CSV Handling:** [Go CSV Package Docs](https://pkg.go.dev/encoding/csv)

# 2. Implementation Snippets (Cookbook)

*Paste useful code blocks here so I don't have to search for them again.*

## Fyne: Basic Window Setup

```go
myApp := app.New()
myWindow := myApp.NewWindow("Data Coupler")
myWindow.Resize(fyne.NewSize(800, 600))

// content goes here

myWindow.ShowAndRun()

```

## Go: Reading a CSV File

```go
file, _ := os.Open("data.csv")
reader := csv.NewReader(file)
records, _ := reader.ReadAll() // returns [][]string

for _, row := range records {
    fmt.Println(row[0]) // Print first column
}

```

## Go: Marshaling JSON (Saving Profiles)

```go
import "encoding/json"

type Profile struct {
    Name string `json:"name"`
}

// Saving
data, _ := json.MarshalIndent(myProfile, "", "  ")
os.WriteFile("profile.json", data, 0644)

```

# 3. Solved Problems (The "Gotchas")

*Log specific errors you encountered and how you fixed them.*

* **Issue:** Console window appears when running the `.exe`.
* **Fix:** Use the command `go build -ldflags "-H=windowsgui"`.
* **Issue:** Fyne text looks blurry on high DPI.
* **Fix:** Check `FYNE_SCALE` environment variable.
