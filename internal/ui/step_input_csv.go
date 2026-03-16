package ui

import (
	"errors"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/widget"

	"github.com/mitchellbauer/data-coupler/internal/connector"
	csvconn "github.com/mitchellbauer/data-coupler/internal/connector/csv"
)

// StepInputCSV is Step 2 of the wizard: the user picks the input CSV file and
// sees a preview of the first 10 data rows.
type StepInputCSV struct {
	wiz     *Wizard
	content fyne.CanvasObject

	// Retained widget refs for in-place updates.
	fileEntry    *widget.Entry
	errorLbl     *widget.Label
	previewStack *fyne.Container
	table        *widget.Table

	// Local data cache (mirrors WizardState after a successful parse).
	headers []string
	rows    [][]string
}

func (s *StepInputCSV) Title() string { return "Input File & Preview" }

func (s *StepInputCSV) Content() fyne.CanvasObject {
	if s.content != nil {
		return s.content
	}

	// ── File picker row ───────────────────────────────────────────────────────
	s.fileEntry = widget.NewEntry()
	s.fileEntry.SetPlaceHolder("No file selected...")
	s.fileEntry.Disable() // read-only; populated by the dialog

	browseBtn := widget.NewButton("Browse...", s.onBrowse)
	fileRow := container.NewBorder(nil, nil, nil, browseBtn, s.fileEntry)

	// ── Inline error label (hidden by default) ────────────────────────────────
	s.errorLbl = widget.NewLabel("")
	s.errorLbl.Importance = widget.DangerImportance
	s.errorLbl.Wrapping = fyne.TextWrapWord
	s.errorLbl.Hide()

	// ── Preview table ─────────────────────────────────────────────────────────
	// Row 0 is the header row (bold); rows 1‥N are the data preview rows.
	s.table = widget.NewTable(
		func() (int, int) {
			if len(s.headers) == 0 {
				return 0, 0
			}
			return 1 + len(s.rows), len(s.headers)
		},
		func() fyne.CanvasObject {
			return widget.NewLabel("")
		},
		func(id widget.TableCellID, obj fyne.CanvasObject) {
			lbl := obj.(*widget.Label)
			if id.Row == 0 {
				lbl.TextStyle = fyne.TextStyle{Bold: true}
				if id.Col < len(s.headers) {
					lbl.SetText(s.headers[id.Col])
				}
			} else {
				lbl.TextStyle = fyne.TextStyle{}
				if id.Row-1 < len(s.rows) {
					row := s.rows[id.Row-1]
					if id.Col < len(row) {
						lbl.SetText(row[id.Col])
					} else {
						lbl.SetText("")
					}
				}
			}
		},
	)

	// Placeholder shown before a file is loaded.
	noFileLbl := widget.NewLabel("Select a CSV file above to see a data preview.")
	noFileLbl.Alignment = fyne.TextAlignCenter

	s.previewStack = container.NewStack(container.NewCenter(noFileLbl))

	// ── Compose layout ────────────────────────────────────────────────────────
	// The top section is fixed height; the preview stack fills the rest.
	topSection := container.NewVBox(
		widget.NewLabel("Input CSV File"),
		fileRow,
		s.errorLbl,
		widget.NewSeparator(),
		widget.NewLabel("Preview (first 10 rows):"),
	)

	s.content = container.NewBorder(
		container.NewPadded(topSection),
		nil, nil, nil,
		container.NewPadded(s.previewStack),
	)
	return s.content
}

func (s *StepInputCSV) onBrowse() {
	d := dialog.NewFileOpen(func(r fyne.URIReadCloser, err error) {
		if err != nil {
			fyne.Do(func() { s.showError(err.Error()) })
			return
		}
		if r == nil {
			return // user cancelled
		}
		path := uriToFilePath(r.URI())
		r.Close()
		go s.loadFile(path)
	}, s.wiz.Window())
	d.SetFilter(storage.NewExtensionFileFilter([]string{".csv", ".CSV"}))
	d.Show()
}

func (s *StepInputCSV) loadFile(path string) {
	conn := &csvconn.CSVConnector{}
	if err := conn.Connect(connector.ConnectionConfig{FilePath: path}); err != nil {
		fyne.Do(func() { s.showError(err.Error()) })
		return
	}
	defer conn.Disconnect()

	headers, err := conn.Columns("")
	if err != nil {
		fyne.Do(func() { s.showError(err.Error()) })
		return
	}

	rowCh, err := conn.Rows("")
	if err != nil {
		fyne.Do(func() { s.showError(err.Error()) })
		return
	}

	// Collect up to 10 preview rows; drain the rest to avoid a goroutine leak.
	var rows [][]string
	for row := range rowCh {
		if len(rows) < 10 {
			rows = append(rows, row)
		}
	}

	// Persist into WizardState.
	s.headers = headers
	s.rows = rows
	s.wiz.State.InputHeaders = headers
	s.wiz.State.InputPreviewRows = rows
	s.wiz.State.InputConfig.FilePath = path

	fyne.Do(func() {
		s.fileEntry.Enable()
		s.fileEntry.SetText(path)
		s.fileEntry.Disable()

		s.errorLbl.Hide()
		s.table.Refresh()
		s.previewStack.Objects[0] = s.table
		s.previewStack.Refresh()
	})
}

func (s *StepInputCSV) showError(msg string) {
	// Reset any previously loaded data.
	s.headers = nil
	s.rows = nil
	s.wiz.State.InputHeaders = nil
	s.wiz.State.InputPreviewRows = nil
	s.wiz.State.InputConfig.FilePath = ""

	s.errorLbl.SetText("⚠  " + msg)
	s.errorLbl.Show()
}

func (s *StepInputCSV) Validate() error {
	if s.wiz.State.InputConfig.FilePath == "" {
		return errors.New("please select an input CSV file")
	}
	return nil
}
