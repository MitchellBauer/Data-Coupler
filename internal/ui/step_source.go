package ui

import (
	"errors"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

// StepSource is Step 1 of the wizard: the user picks an input connector type.
type StepSource struct {
	wiz     *Wizard
	content fyne.CanvasObject

	// Retained button refs so we can highlight the selected one.
	csvBtn      *widget.Button
	mssqlBtn    *widget.Button
	sqliteBtn   *widget.Button
	mysqlBtn    *widget.Button
	postgresBtn *widget.Button
	odbcBtn     *widget.Button
}

func (s *StepSource) Title() string { return "Choose Input Source" }

func (s *StepSource) Content() fyne.CanvasObject {
	if s.content != nil {
		return s.content
	}

	// ── CSV card ──────────────────────────────────────────────────────────────
	s.csvBtn = widget.NewButton("Select", func() { s.onSelect("csv", s.csvBtn) })

	csvCard := widget.NewCard(
		"📄  CSV File",
		"Load data from a .csv file on your computer.",
		container.NewCenter(s.csvBtn),
	)

	// ── SQL Database card ─────────────────────────────────────────────────────
	s.mssqlBtn = widget.NewButton("SQL Server", func() { s.onSelect("mssql", s.mssqlBtn) })
	s.sqliteBtn = widget.NewButton("SQLite", func() { s.onSelect("sqlite", s.sqliteBtn) })
	s.mysqlBtn = widget.NewButton("MySQL", func() { s.onSelect("mysql", s.mysqlBtn) })
	s.postgresBtn = widget.NewButton("PostgreSQL", func() { s.onSelect("postgres", s.postgresBtn) })
	s.odbcBtn = widget.NewButton("ODBC", func() { s.onSelect("odbc", s.odbcBtn) })

	dbButtons := container.New(layout.NewGridLayout(2),
		s.mssqlBtn, s.sqliteBtn,
		s.mysqlBtn, s.postgresBtn,
		s.odbcBtn, widget.NewLabel(""),
	)

	sqlCard := widget.NewCard(
		"🗄️  SQL Database",
		"Connect to SQL Server, MySQL, SQLite, PostgreSQL, or a Windows ODBC data source.",
		container.NewPadded(dbButtons),
	)

	s.content = container.NewPadded(
		container.New(layout.NewGridLayout(2), csvCard, sqlCard),
	)
	return s.content
}

// onSelect highlights the chosen button and records the connector name in WizardState.
func (s *StepSource) onSelect(name string, chosen *widget.Button) {
	// Reset all buttons to default importance.
	for _, b := range []*widget.Button{s.csvBtn, s.mssqlBtn, s.sqliteBtn, s.mysqlBtn, s.postgresBtn, s.odbcBtn} {
		if b != nil {
			b.Importance = widget.MediumImportance
			b.Refresh()
		}
	}

	s.wiz.State.InputConnectorName = name

	// Highlight the chosen button.
	chosen.Importance = widget.HighImportance
	chosen.Refresh()
}

func (s *StepSource) Validate() error {
	if s.wiz.State.InputConnectorName == "" {
		return errors.New("please select an input source")
	}
	return nil
}
