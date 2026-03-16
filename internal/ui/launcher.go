package ui

import (
	"os"
	"path/filepath"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"

	"github.com/mitchellbauer/data-coupler/internal/audit"
	"github.com/mitchellbauer/data-coupler/internal/config"
	"github.com/mitchellbauer/data-coupler/internal/credentials"
	"github.com/mitchellbauer/data-coupler/internal/updater"
	"github.com/mitchellbauer/data-coupler/internal/version"
)

// StepInputRouter sits at position 1 in the wizard step list.
// It delegates to StepInputCSV or StepInputSQL based on the chosen connector.
type StepInputRouter struct {
	wiz *Wizard
	csv *StepInputCSV
	sql *StepInputSQL
}

func (r *StepInputRouter) Title() string              { return r.active().Title() }
func (r *StepInputRouter) Content() fyne.CanvasObject { return r.active().Content() }
func (r *StepInputRouter) Validate() error            { return r.active().Validate() }

func (r *StepInputRouter) active() Step {
	if r.wiz.State.InputConnectorName == "csv" {
		return r.csv
	}
	return r.sql
}

// LaunchApp creates the Fyne application, shows the home screen, then launches
// the wizard when the user starts or loads a conversion.
func LaunchApp() {
	myApp := app.New()
	w := myApp.NewWindow("Data Coupler")
	w.Resize(fyne.NewSize(720, 520))
	w.CenterOnScreen()

	// Load persisted settings (non-fatal if missing).
	appDir := filepath.Dir(os.Args[0])
	settingsPath := filepath.Join(appDir, "settings.json")
	settings, _ := config.LoadSettings(settingsPath)

	// ── Audit log setup ───────────────────────────────────────────────────────
	audit.SetLogPath(appDir)
	_ = audit.TrimLog(1000)

	// Credential store (shared with engine via step_review).
	credStore := credentials.NewFileStore(appDir)

	// ── Build wizard steps ────────────────────────────────────────────────────
	source := &StepSource{}
	inputCSV := &StepInputCSV{}
	inputSQL := &StepInputSQL{credStore: credStore}
	inputRouter := &StepInputRouter{csv: inputCSV, sql: inputSQL}
	output := &StepOutput{}
	mapping := &StepMapping{}
	transformStep := &StepTransform{}
	review := &StepReview{credStore: credStore}

	steps := []Step{source, inputRouter, output, mapping, transformStep, review}
	wiz := NewWizard(w, steps)

	source.wiz = wiz
	inputCSV.wiz = wiz
	inputSQL.wiz = wiz
	inputRouter.wiz = wiz
	output.wiz = wiz
	mapping.wiz = wiz
	transformStep.wiz = wiz
	review.wiz = wiz

	// ── Build home screen ─────────────────────────────────────────────────────
	home := &Home{
		window:          w,
		wiz:             wiz,
		hasLastProfile:  settings.LastProfilePath != "",
		lastProfilePath: settings.LastProfilePath,

		onNewConversion: func() {
			w.SetContent(wiz.Chrome())
			wiz.Start()
		},
		onLoadProfile: func(loadedPath string) {
			settings.LastProfilePath = loadedPath
			w.SetContent(wiz.Chrome())
			wiz.GoTo(len(steps) - 1) // jump to Review & Run (last step)
		},
	}

	w.SetContent(home.Content())

	// ── On close: persist last-used folders and profile path ──────────────────
	w.SetCloseIntercept(func() {
		if wiz.State.InputConfig.FilePath != "" {
			settings.LastInputFolder = filepath.Dir(wiz.State.InputConfig.FilePath)
		}
		if wiz.State.OutputConfig.FilePath != "" {
			settings.LastOutputFolder = filepath.Dir(wiz.State.OutputConfig.FilePath)
		}
		_ = config.SaveSettings(settingsPath, settings)
		w.Close()
	})

	// ── Auto-update check (non-blocking) ─────────────────────────────────────
	go func() {
		latest, err := updater.CheckLatestRelease("mitchellbauer", "data-coupler")
		if err != nil || !updater.IsNewer(version.AppVersion, latest) {
			return
		}
		fyne.Do(func() {
			home.ShowUpdateBanner(latest, "https://github.com/mitchellbauer/data-coupler/releases/latest")
		})
	}()

	w.ShowAndRun()
}
