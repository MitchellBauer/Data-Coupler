package ui

import (
	"errors"
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/widget"

	"github.com/mitchellbauer/data-coupler/internal/templates"
)

// StepOutput is Step 3 of the wizard: the user picks an output destination,
// sets the output file path, and chooses a column delimiter.
type StepOutput struct {
	wiz     *Wizard
	content fyne.CanvasObject

	// Retained widget refs for in-place updates.
	csvBtn    *widget.Button
	fishBtn   *widget.Button
	pathRow   *fyne.Container // revealed when CSV or Fishbowl is selected
	pathEntry *widget.Entry
}

func (s *StepOutput) Title() string { return "Choose Output" }

func (s *StepOutput) Content() fyne.CanvasObject {
	if s.content != nil {
		return s.content
	}

	// ── CSV output card ───────────────────────────────────────────────────────
	s.csvBtn = widget.NewButton("Select", s.onCSVSelect)

	csvCard := widget.NewCard(
		"📄  CSV File",
		"Save output to a .csv file.",
		container.NewCenter(s.csvBtn),
	)

	// ── Fishbowl card ─────────────────────────────────────────────────────────
	s.fishBtn = widget.NewButton("Select Template", s.onFishbowlSelect)

	fishCard := widget.NewCard(
		"🐟  Fishbowl Template",
		"Export using a Fishbowl import template.",
		container.NewCenter(s.fishBtn),
	)

	cards := container.New(layout.NewGridLayout(2), csvCard, fishCard)

	// ── Output path row (hidden until a destination is selected) ──────────────
	s.pathEntry = widget.NewEntry()
	s.pathEntry.SetPlaceHolder("Enter or browse to the output file path...")
	s.pathEntry.OnChanged = func(p string) {
		s.wiz.State.OutputConfig.FilePath = p
	}

	browseBtn := widget.NewButton("Browse...", s.onBrowse)
	pathBox := container.NewBorder(nil, nil, nil, browseBtn, s.pathEntry)

	// ── Delimiter radio ───────────────────────────────────────────────────────
	delimRadio := widget.NewRadioGroup(
		[]string{`Comma (,)`, `Tab (\t)`, `Pipe (|)`},
		func(sel string) {
			switch sel {
			case `Comma (,)`:
				s.wiz.State.OutputDelimiter = ','
			case `Tab (\t)`:
				s.wiz.State.OutputDelimiter = '\t'
			case `Pipe (|)`:
				s.wiz.State.OutputDelimiter = '|'
			}
		},
	)
	delimRadio.Horizontal = true
	delimRadio.SetSelected(`Comma (,)`)

	s.pathRow = container.NewVBox(
		widget.NewSeparator(),
		widget.NewLabel("Output File Path:"),
		pathBox,
		widget.NewLabel("Delimiter:"),
		delimRadio,
	)
	s.pathRow.Hide()

	s.content = container.NewPadded(
		container.NewVBox(cards, s.pathRow),
	)
	return s.content
}

func (s *StepOutput) onCSVSelect() {
	s.wiz.State.OutputConnectorName = "csv"
	s.wiz.State.OutputTemplate = ""
	s.wiz.State.OutputTemplateColumns = nil
	s.csvBtn.SetText("✓  Selected")
	s.csvBtn.Importance = widget.HighImportance
	s.csvBtn.Refresh()
	s.fishBtn.SetText("Select Template")
	s.fishBtn.Importance = widget.MediumImportance
	s.fishBtn.Refresh()
	s.pathRow.Show()
}

func (s *StepOutput) onFishbowlSelect() {
	tmpls, err := templates.ListTemplates()
	if err != nil {
		s.wiz.showError("Failed to load templates: " + err.Error())
		return
	}

	var d dialog.Dialog

	// Build a scrollable grid of template cards.
	var cards []fyne.CanvasObject
	for _, t := range tmpls {
		t := t // capture loop variable

		reqCount := 0
		for _, col := range t.Columns {
			if col.Required {
				reqCount++
			}
		}

		selectBtn := widget.NewButton("Select", func() {
			d.Hide()
			s.wiz.State.OutputConnectorName = "csv"
			s.wiz.State.OutputTemplate = t.ID
			s.wiz.State.OutputTemplateColumns = t.Columns

			s.fishBtn.SetText("✓  " + t.Name)
			s.fishBtn.Importance = widget.HighImportance
			s.fishBtn.Refresh()
			s.csvBtn.SetText("Select")
			s.csvBtn.Importance = widget.MediumImportance
			s.csvBtn.Refresh()
			s.pathRow.Show()
		})
		selectBtn.Importance = widget.HighImportance

		summary := fmt.Sprintf("%s\n%d required columns · %d total",
			t.Description, reqCount, len(t.Columns))

		card := widget.NewCard(t.Name, summary, container.NewCenter(selectBtn))
		cards = append(cards, card)
	}

	grid := container.New(layout.NewGridLayout(2), cards...)
	scroll := container.NewVScroll(grid)
	scroll.SetMinSize(fyne.NewSize(600, 400))

	d = dialog.NewCustom("Choose a Fishbowl Template", "Cancel", scroll, s.wiz.Window())
	d.Show()
}

func (s *StepOutput) onBrowse() {
	d := dialog.NewFileSave(func(w fyne.URIWriteCloser, err error) {
		if err != nil {
			fyne.Do(func() {
				s.wiz.showError(err.Error())
			})
			return
		}
		if w == nil {
			return // user cancelled
		}
		path := uriToFilePath(w.URI())
		w.Close() // we only need the path; the engine writes the file later
		fyne.Do(func() {
			s.pathEntry.SetText(path)
			s.wiz.State.OutputConfig.FilePath = path
		})
	}, s.wiz.Window())
	d.SetFilter(storage.NewExtensionFileFilter([]string{".csv"}))
	d.Show()
}

func (s *StepOutput) Validate() error {
	if s.wiz.State.OutputConnectorName == "" {
		return errors.New("please select an output destination")
	}
	if s.wiz.State.OutputConfig.FilePath == "" {
		return errors.New("please enter or browse to an output file path")
	}
	return nil
}
