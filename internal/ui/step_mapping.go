package ui

import (
	"errors"
	"fmt"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"

	"github.com/mitchellbauer/data-coupler/internal/types"
)

// mappingRow tracks one output→input column pair and the widgets that represent it.
type mappingRow struct {
	outputEntry *widget.Entry  // free mode only
	outputLabel *widget.Label  // template mode only
	inputSelect *widget.Select
	container   fyne.CanvasObject // the grid row shown in the UI
}

// StepMapping is Step 4 of the wizard: the user defines which input columns map
// to which output column names.
type StepMapping struct {
	wiz     *Wizard
	content fyne.CanvasObject

	rows         []*mappingRow
	rowsBox      *fyne.Container // VBox — rows are appended/removed dynamically
	lastOutputKey string          // rebuild guard
}

func (s *StepMapping) Title() string { return "Map Columns" }

// outputKey returns a string that changes whenever the template or headers change,
// triggering a content rebuild.
func (s *StepMapping) outputKey() string {
	return s.wiz.State.OutputTemplate + "|" + strings.Join(s.wiz.State.InputHeaders, ",")
}

func (s *StepMapping) Content() fyne.CanvasObject {
	key := s.outputKey()
	if s.content != nil && s.lastOutputKey == key {
		return s.content
	}
	s.lastOutputKey = key
	s.rows = nil

	if s.wiz.State.OutputTemplate != "" {
		s.content = s.buildTemplateContent()
	} else {
		s.content = s.buildFreeContent()
	}
	return s.content
}

// ── Free mode ─────────────────────────────────────────────────────────────────

func (s *StepMapping) buildFreeContent() fyne.CanvasObject {
	outHeader := widget.NewLabel("Output Column")
	outHeader.TextStyle = fyne.TextStyle{Bold: true}

	inHeader := widget.NewLabel("From Input Column")
	inHeader.TextStyle = fyne.TextStyle{Bold: true}

	headerRow := container.New(
		layout.NewGridLayout(3),
		outHeader,
		inHeader,
		widget.NewLabel(""),
	)

	topSection := container.NewVBox(
		container.NewPadded(headerRow),
		widget.NewSeparator(),
	)

	s.rowsBox = container.NewVBox()
	s.addFreeRow()

	addBtn := widget.NewButton("＋  Add Mapping Row", s.addFreeRow)

	return container.NewBorder(
		topSection,
		container.NewPadded(addBtn),
		nil, nil,
		container.NewVScroll(s.rowsBox),
	)
}

func (s *StepMapping) addFreeRow() {
	row := &mappingRow{
		outputEntry: widget.NewEntry(),
		inputSelect: widget.NewSelect(s.wiz.State.InputHeaders, nil),
	}
	row.outputEntry.SetPlaceHolder("Output column name")
	row.inputSelect.PlaceHolder = "Select input column..."

	row.outputEntry.OnChanged = func(_ string) { s.syncMappings() }
	row.inputSelect.OnChanged = func(_ string) { s.syncMappings() }

	deleteBtn := widget.NewButton("✕", func() { s.removeFreeRow(row) })
	deleteBtn.Importance = widget.DangerImportance

	row.container = container.New(
		layout.NewGridLayout(3),
		row.outputEntry,
		row.inputSelect,
		deleteBtn,
	)

	s.rows = append(s.rows, row)
	s.rowsBox.Add(row.container)
}

func (s *StepMapping) removeFreeRow(row *mappingRow) {
	for i, r := range s.rows {
		if r == row {
			s.rows = append(s.rows[:i], s.rows[i+1:]...)
			break
		}
	}

	filtered := make([]fyne.CanvasObject, 0, len(s.rowsBox.Objects))
	for _, obj := range s.rowsBox.Objects {
		if obj != row.container {
			filtered = append(filtered, obj)
		}
	}
	s.rowsBox.Objects = filtered
	s.rowsBox.Refresh()
	s.syncMappings()
}

// ── Template mode ─────────────────────────────────────────────────────────────

func (s *StepMapping) buildTemplateContent() fyne.CanvasObject {
	outHeader := widget.NewLabel("Output Column")
	outHeader.TextStyle = fyne.TextStyle{Bold: true}

	inHeader := widget.NewLabel("From Input Column")
	inHeader.TextStyle = fyne.TextStyle{Bold: true}

	headerRow := container.New(
		layout.NewGridLayout(2),
		outHeader,
		inHeader,
	)

	topSection := container.NewVBox(
		container.NewPadded(headerRow),
		widget.NewSeparator(),
	)

	s.rowsBox = container.NewVBox()

	for _, col := range s.wiz.State.OutputTemplateColumns {
		col := col // capture

		// Preserve any existing mapping for this column (e.g. after back-navigation).
		existingInput := ""
		for _, m := range s.wiz.State.Mappings {
			if m.OutputCol == col.Name {
				existingInput = m.InputCol
				break
			}
		}

		lbl := widget.NewLabel(col.Name)
		if col.Required {
			lbl.SetText("★ " + col.Name)
			lbl.Importance = widget.DangerImportance
		}

		sel := widget.NewSelect(s.wiz.State.InputHeaders, nil)
		sel.PlaceHolder = "Select input column..."
		if existingInput != "" {
			sel.SetSelected(existingInput)
		}
		sel.OnChanged = func(_ string) { s.syncMappings() }

		row := &mappingRow{
			outputLabel: lbl,
			inputSelect: sel,
		}
		row.container = container.New(
			layout.NewGridLayout(2),
			lbl,
			sel,
		)
		s.rows = append(s.rows, row)
		s.rowsBox.Add(row.container)
	}

	s.syncMappings()

	return container.NewBorder(
		topSection,
		nil, nil, nil,
		container.NewVScroll(s.rowsBox),
	)
}

// ── Shared ────────────────────────────────────────────────────────────────────

// syncMappings rebuilds WizardState.Mappings from the current row widgets,
// preserving any Transforms that were previously set.
func (s *StepMapping) syncMappings() {
	// Build a lookup of existing transforms by output column name.
	existing := map[string][]types.Transform{}
	for _, m := range s.wiz.State.Mappings {
		if len(m.Transforms) > 0 {
			existing[m.OutputCol] = m.Transforms
		}
	}

	mappings := make([]types.Mapping, 0, len(s.rows))
	for _, r := range s.rows {
		var out, in string
		if r.outputEntry != nil {
			out = strings.TrimSpace(r.outputEntry.Text)
		} else if r.outputLabel != nil {
			// Template mode: strip the "★ " prefix to recover the real column name.
			out = strings.TrimPrefix(r.outputLabel.Text, "★ ")
		}
		in = r.inputSelect.Selected

		if out != "" || in != "" {
			transforms := existing[out]
			if transforms == nil {
				transforms = []types.Transform{}
			}
			mappings = append(mappings, types.Mapping{
				InputCol:   in,
				OutputCol:  out,
				Transforms: transforms,
			})
		}
	}
	s.wiz.State.Mappings = mappings
}

func (s *StepMapping) Validate() error {
	if s.wiz.State.OutputTemplate != "" {
		// Template mode: all required fields must have an input column selected.
		for _, col := range s.wiz.State.OutputTemplateColumns {
			if !col.Required {
				continue
			}
			found := false
			for _, m := range s.wiz.State.Mappings {
				if m.OutputCol == col.Name && m.InputCol != "" {
					found = true
					break
				}
			}
			if !found {
				return fmt.Errorf("required column %q has no input column selected", col.Name)
			}
		}
		return nil
	}

	// Free mode validation.
	if len(s.rows) == 0 {
		return errors.New("add at least one column mapping")
	}
	for i, r := range s.rows {
		out := strings.TrimSpace(r.outputEntry.Text)
		in := r.inputSelect.Selected
		switch {
		case out == "" && in == "":
			return fmt.Errorf("row %d is empty — fill it in or delete it", i+1)
		case out != "" && in == "":
			return fmt.Errorf("row %d: please select an input column", i+1)
		case out == "" && in != "":
			return fmt.Errorf("row %d: please enter an output column name", i+1)
		}
	}
	return nil
}
