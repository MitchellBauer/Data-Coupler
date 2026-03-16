package ui

import (
	"errors"
	"strconv"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/widget"

	"github.com/mitchellbauer/data-coupler/internal/connector"
	"github.com/mitchellbauer/data-coupler/internal/credentials"
)

// StepInputSQL is the SQL input configuration step. It renders differently
// depending on wiz.State.InputConnectorName:
//   - "sqlite": shows a file picker only
//   - "mssql" / "mysql" / "postgres": shows a server connection form + SQL query area
type StepInputSQL struct {
	wiz        *Wizard
	credStore  credentials.Store
	content    fyne.CanvasObject
	lastConnector string // tracks when to rebuild content

	// Server form widgets.
	hostEntry     *widget.Entry
	portEntry     *widget.Entry
	dbEntry       *widget.Entry
	userEntry     *widget.Entry
	passEntry     *widget.Entry
	credRefEntry  *widget.Entry
	saveCredsCheck *widget.Check
	testBtn       *widget.Button
	testStatusLbl *widget.Label

	// SQLite file picker widget.
	fileEntry *widget.Entry

	// Query widgets (server connectors only).
	queryEntry    *widget.Entry
	previewBtn    *widget.Button
	previewStack  *fyne.Container
	table         *widget.Table

	// State.
	connectionTested bool
	headers          []string
	rows             [][]string
	errorLbl         *widget.Label
}

func (s *StepInputSQL) Title() string { return "Configure SQL Input" }

func (s *StepInputSQL) Content() fyne.CanvasObject {
	// Rebuild if the connector type changed since last render.
	if s.content != nil && s.lastConnector == s.wiz.State.InputConnectorName {
		return s.content
	}
	s.lastConnector = s.wiz.State.InputConnectorName
	s.connectionTested = false

	s.errorLbl = widget.NewLabel("")
	s.errorLbl.Importance = widget.DangerImportance
	s.errorLbl.Wrapping = fyne.TextWrapWord
	s.errorLbl.Hide()

	if s.wiz.State.InputConnectorName == "sqlite" {
		s.content = s.buildSQLiteContent()
	} else {
		s.content = s.buildServerContent()
	}
	return s.content
}

// ── SQLite layout ─────────────────────────────────────────────────────────────

func (s *StepInputSQL) buildSQLiteContent() fyne.CanvasObject {
	s.fileEntry = widget.NewEntry()
	s.fileEntry.SetPlaceHolder("No file selected...")
	s.fileEntry.Disable()

	browseBtn := widget.NewButton("Browse...", s.onSQLiteBrowse)
	fileRow := container.NewBorder(nil, nil, nil, browseBtn, s.fileEntry)

	top := container.NewVBox(
		widget.NewLabel("SQLite Database File"),
		fileRow,
		s.errorLbl,
	)
	return container.NewBorder(container.NewPadded(top), nil, nil, nil, nil)
}

func (s *StepInputSQL) onSQLiteBrowse() {
	d := dialog.NewFileOpen(func(r fyne.URIReadCloser, err error) {
		if err != nil {
			fyne.Do(func() { s.showError(err.Error()) })
			return
		}
		if r == nil {
			return
		}
		path := uriToFilePath(r.URI())
		r.Close()
		go s.testSQLiteFile(path)
	}, s.wiz.Window())
	d.SetFilter(storage.NewExtensionFileFilter([]string{".db", ".sqlite", ".sqlite3"}))
	d.Show()
}

func (s *StepInputSQL) testSQLiteFile(path string) {
	conn, ok := connector.Get("sqlite")
	if !ok {
		fyne.Do(func() { s.showError("sqlite connector not registered") })
		return
	}
	if err := conn.Connect(connector.ConnectionConfig{FilePath: path}); err != nil {
		fyne.Do(func() { s.showError(err.Error()) })
		return
	}
	conn.Disconnect()

	s.wiz.State.InputConfig.FilePath = path
	s.wiz.State.InputQuery = "SELECT * FROM " + "sqlite_master" // placeholder; user can change
	s.connectionTested = true

	fyne.Do(func() {
		s.fileEntry.Enable()
		s.fileEntry.SetText(path)
		s.fileEntry.Disable()
		s.errorLbl.Hide()
	})
}

// ── Server connector layout (MSSQL / MySQL / PostgreSQL) ──────────────────────

func (s *StepInputSQL) buildServerContent() fyne.CanvasObject {
	defaultPort := defaultPortFor(s.wiz.State.InputConnectorName)

	// Connection form.
	s.hostEntry = widget.NewEntry()
	s.hostEntry.SetPlaceHolder("e.g. 192.168.1.10")

	s.portEntry = widget.NewEntry()
	s.portEntry.SetText(strconv.Itoa(defaultPort))

	s.dbEntry = widget.NewEntry()
	s.dbEntry.SetPlaceHolder("database name")

	s.userEntry = widget.NewEntry()
	s.userEntry.SetPlaceHolder("username")

	s.passEntry = widget.NewPasswordEntry()
	s.passEntry.SetPlaceHolder("password")

	s.credRefEntry = widget.NewEntry()
	s.credRefEntry.SetPlaceHolder("e.g. macola-prod")

	s.saveCredsCheck = widget.NewCheck("Save credentials", nil)

	s.testStatusLbl = widget.NewLabel("")
	s.testStatusLbl.Hide()

	s.testBtn = widget.NewButton("Test Connection", s.onTestConnection)
	s.testBtn.Importance = widget.MediumImportance

	formGrid := container.New(newTwoColFormLayout(),
		widget.NewLabel("Host"), s.hostEntry,
		widget.NewLabel("Port"), s.portEntry,
		widget.NewLabel("Database"), s.dbEntry,
		widget.NewLabel("Username"), s.userEntry,
		widget.NewLabel("Password"), s.passEntry,
		widget.NewLabel("Credential name"), s.credRefEntry,
		widget.NewLabel(""), s.saveCredsCheck,
	)

	testRow := container.NewVBox(
		container.NewBorder(nil, nil, nil, s.testBtn, s.testStatusLbl),
		widget.NewSeparator(),
	)

	// Query area.
	s.queryEntry = widget.NewMultiLineEntry()
	s.queryEntry.SetPlaceHolder("SELECT column1, column2 FROM table WHERE ...")
	s.queryEntry.SetMinRowsVisible(4)
	s.queryEntry.OnChanged = func(q string) {
		s.wiz.State.InputQuery = q
	}

	// Preview table.
	s.table = widget.NewTable(
		func() (int, int) {
			if len(s.headers) == 0 {
				return 0, 0
			}
			return 1 + len(s.rows), len(s.headers)
		},
		func() fyne.CanvasObject { return widget.NewLabel("") },
		func(id widget.TableCellID, obj fyne.CanvasObject) {
			lbl := obj.(*widget.Label)
			if id.Row == 0 {
				lbl.TextStyle = fyne.TextStyle{Bold: true}
				if id.Col < len(s.headers) {
					lbl.SetText(s.headers[id.Col])
				}
			} else {
				lbl.TextStyle = fyne.TextStyle{}
				dataRow := id.Row - 1
				if dataRow < len(s.rows) && id.Col < len(s.rows[dataRow]) {
					lbl.SetText(s.rows[dataRow][id.Col])
				} else {
					lbl.SetText("")
				}
			}
		},
	)

	noPreviewLbl := widget.NewLabel("Enter a query above and click Preview.")
	noPreviewLbl.Alignment = fyne.TextAlignCenter
	s.previewStack = container.NewStack(container.NewCenter(noPreviewLbl))

	s.previewBtn = widget.NewButton("Preview Query", s.onPreview)

	// Macola banner (only for MSSQL).
	var macolaBanner fyne.CanvasObject
	if s.wiz.State.InputConnectorName == "mssql" {
		banner := widget.NewLabel("ℹ  Macola data often has trailing spaces. Consider adding TrimSpace to all columns in the Transform step.")
		banner.Importance = widget.LowImportance
		banner.Wrapping = fyne.TextWrapWord
		macolaBanner = widget.NewCard("", "", banner)
	}

	querySection := container.NewVBox(
		widget.NewLabel("SQL Query"),
		s.queryEntry,
		container.NewBorder(nil, nil, nil, s.previewBtn, widget.NewLabel("")),
		widget.NewLabel("Preview (first 10 rows):"),
	)

	topParts := []fyne.CanvasObject{formGrid, testRow, s.errorLbl, querySection}
	if macolaBanner != nil {
		topParts = append([]fyne.CanvasObject{macolaBanner}, topParts...)
	}

	top := container.NewPadded(container.NewVBox(topParts...))
	return container.NewBorder(top, nil, nil, nil, container.NewPadded(s.previewStack))
}

func (s *StepInputSQL) onTestConnection() {
	s.connectionTested = false
	s.testStatusLbl.SetText("Testing...")
	s.testStatusLbl.Importance = widget.MediumImportance
	s.testStatusLbl.Show()
	s.testBtn.Disable()

	cfg := s.buildConnectionConfig()

	go func() {
		conn, ok := connector.Get(s.wiz.State.InputConnectorName)
		if !ok {
			fyne.Do(func() {
				s.setTestStatus(false, "connector not registered: "+s.wiz.State.InputConnectorName)
				s.testBtn.Enable()
			})
			return
		}

		err := conn.Connect(cfg)
		if err != nil {
			fyne.Do(func() {
				s.setTestStatus(false, err.Error())
				s.testBtn.Enable()
			})
			return
		}
		conn.Disconnect()

		// Optionally save credentials; always record the ref so the profile can
		// reference credentials saved in a previous session.
		fyne.Do(func() {
			ref := s.credRefEntry.Text
			if ref == "" {
				ref = s.wiz.State.InputConnectorName + "-default"
			}
			if s.saveCredsCheck.Checked && s.credStore != nil {
				_ = s.credStore.Save(ref, s.passEntry.Text)
			}
			s.wiz.State.InputCredentialRef = ref
			s.wiz.State.InputConfig = cfg
			s.connectionTested = true
			s.setTestStatus(true, "Connection successful.")
			s.testBtn.Enable()
		})
	}()
}

func (s *StepInputSQL) onPreview() {
	query := s.queryEntry.Text
	if query == "" {
		s.showError("Please enter a SQL query before previewing.")
		return
	}

	s.previewBtn.Disable()
	cfg := s.buildConnectionConfig()

	go func() {
		defer fyne.Do(func() { s.previewBtn.Enable() })

		conn, ok := connector.Get(s.wiz.State.InputConnectorName)
		if !ok {
			fyne.Do(func() { s.showError("connector not registered") })
			return
		}
		if err := conn.Connect(cfg); err != nil {
			fyne.Do(func() { s.showError(err.Error()) })
			return
		}
		defer conn.Disconnect()

		headers, err := conn.Columns(query)
		if err != nil {
			fyne.Do(func() { s.showError(err.Error()) })
			return
		}

		rowCh, err := conn.Rows(query)
		if err != nil {
			fyne.Do(func() { s.showError(err.Error()) })
			return
		}

		var rows [][]string
		for row := range rowCh {
			if len(rows) < 10 {
				rows = append(rows, row)
			}
		}

		fyne.Do(func() {
			s.headers = headers
			s.rows = rows
			s.wiz.State.InputHeaders = headers
			s.wiz.State.InputPreviewRows = rows
			s.errorLbl.Hide()
			s.table.Refresh()
			s.previewStack.Objects[0] = s.table
			s.previewStack.Refresh()
		})
	}()
}

func (s *StepInputSQL) buildConnectionConfig() connector.ConnectionConfig {
	port, _ := strconv.Atoi(s.portEntry.Text)
	return connector.ConnectionConfig{
		Host:     s.hostEntry.Text,
		Port:     port,
		Database: s.dbEntry.Text,
		Username: s.userEntry.Text,
		Password: s.passEntry.Text,
	}
}

func (s *StepInputSQL) setTestStatus(ok bool, msg string) {
	if ok {
		s.testStatusLbl.SetText("✓  " + msg)
		s.testStatusLbl.Importance = widget.SuccessImportance
	} else {
		s.testStatusLbl.SetText("✗  " + msg)
		s.testStatusLbl.Importance = widget.DangerImportance
	}
	s.testStatusLbl.Show()
}

func (s *StepInputSQL) showError(msg string) {
	s.errorLbl.SetText("⚠  " + msg)
	s.errorLbl.Show()
}

func (s *StepInputSQL) Validate() error {
	if s.wiz.State.InputConnectorName == "sqlite" {
		if s.wiz.State.InputConfig.FilePath == "" {
			return errors.New("please select a SQLite database file")
		}
		return nil
	}
	if !s.connectionTested {
		return errors.New("please test the connection before continuing")
	}
	if s.wiz.State.InputQuery == "" {
		return errors.New("please enter a SQL query")
	}
	return nil
}

// defaultPortFor returns the conventional port for a given connector name.
func defaultPortFor(name string) int {
	switch name {
	case "mssql":
		return 1433
	case "mysql":
		return 3306
	case "postgres":
		return 5432
	default:
		return 0
	}
}

// newTwoColFormLayout returns a grid layout that gives the label column a
// fixed narrow width and the input column the remaining space.
func newTwoColFormLayout() fyne.Layout {
	return layout.NewFormLayout()
}
