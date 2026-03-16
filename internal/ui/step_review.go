package ui

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"github.com/mitchellbauer/data-coupler/internal/audit"
	"github.com/mitchellbauer/data-coupler/internal/config"
	"github.com/mitchellbauer/data-coupler/internal/credentials"
	"github.com/mitchellbauer/data-coupler/internal/engine"
	"github.com/mitchellbauer/data-coupler/internal/types"
)

// StepReview is Step 5 of the wizard: shows a summary, runs the conversion,
// and offers to save the configuration as a named profile.
type StepReview struct {
	wiz       *Wizard
	credStore credentials.Store
	content   fyne.CanvasObject

	// Summary labels — refreshed every time the step is re-entered.
	inputFileLbl    *widget.Label
	outputFileLbl   *widget.Label
	mappingCountLbl *widget.Label

	// Template validation label (shown when required fields are unmapped).
	templateValidationLbl *widget.Label

	// Run-state widgets.
	runBtn   *widget.Button
	progress *widget.ProgressBarInfinite

	// Success / error sections (shown after a run attempt).
	successSection *fyne.Container
	rowCountLbl    *widget.Label

	errorSection *fyne.Container
	errorMsgLbl  *widget.Label

	// Save section (shown after a successful run).
	saveSection      *fyne.Container
	profileNameEntry *widget.Entry
	saveStatusLbl    *widget.Label
}

func (s *StepReview) Title() string { return "Review & Run" }

func (s *StepReview) Content() fyne.CanvasObject {
	if s.content != nil {
		s.updateSummary()
		return s.content
	}

	// ── Summary card ──────────────────────────────────────────────────────────
	s.inputFileLbl = widget.NewLabel("")
	s.outputFileLbl = widget.NewLabel("")
	s.mappingCountLbl = widget.NewLabel("")

	summaryBody := container.NewVBox(
		container.NewHBox(widget.NewLabel("Input:"), s.inputFileLbl),
		container.NewHBox(widget.NewLabel("Output:"), s.outputFileLbl),
		container.NewHBox(widget.NewLabel("Mappings:"), s.mappingCountLbl),
	)
	summaryCard := widget.NewCard("Conversion Summary", "", summaryBody)

	// ── Progress bar (hidden while idle) ─────────────────────────────────────
	s.progress = widget.NewProgressBarInfinite()
	s.progress.Hide()

	// ── Success section (hidden until run succeeds) ───────────────────────────
	s.rowCountLbl = widget.NewLabel("")
	s.rowCountLbl.TextStyle = fyne.TextStyle{Bold: true}

	openBtn := widget.NewButton("Open Output File", func() {
		openInFileBrowser(s.wiz.State.OutputConfig.FilePath)
	})

	s.successSection = container.NewVBox(s.rowCountLbl, openBtn)
	s.successSection.Hide()

	// ── Error section (hidden until run fails) ────────────────────────────────
	s.errorMsgLbl = widget.NewLabel("")
	s.errorMsgLbl.Importance = widget.DangerImportance
	s.errorMsgLbl.Wrapping = fyne.TextWrapWord

	s.errorSection = container.NewVBox(s.errorMsgLbl)
	s.errorSection.Hide()

	// ── Template validation label (hidden when no issues) ────────────────────
	s.templateValidationLbl = widget.NewLabel("")
	s.templateValidationLbl.Importance = widget.DangerImportance
	s.templateValidationLbl.Wrapping = fyne.TextWrapWord
	s.templateValidationLbl.Hide()

	// ── Run button ────────────────────────────────────────────────────────────
	s.runBtn = widget.NewButton("▶  Run Conversion", s.onRun)
	s.runBtn.Importance = widget.HighImportance

	runArea := container.NewVBox(
		widget.NewSeparator(),
		s.templateValidationLbl,
		container.NewCenter(s.runBtn),
		container.NewCenter(s.progress),
		s.successSection,
		s.errorSection,
	)

	// ── Save section (shown after success) ────────────────────────────────────
	s.profileNameEntry = widget.NewEntry()
	s.profileNameEntry.SetPlaceHolder("Profile name, e.g. Monthly Export")

	s.saveStatusLbl = widget.NewLabel("")
	s.saveStatusLbl.Alignment = fyne.TextAlignCenter

	saveBtn := widget.NewButton("Save Profile", s.onSaveProfile)

	s.saveSection = container.NewVBox(
		widget.NewSeparator(),
		widget.NewLabel("Save this configuration as a profile for future use:"),
		container.NewBorder(nil, nil, nil, saveBtn, s.profileNameEntry),
		s.saveStatusLbl,
	)
	s.saveSection.Hide()

	s.content = container.NewPadded(
		container.NewVBox(summaryCard, runArea, s.saveSection),
	)

	s.updateSummary()
	return s.content
}

// updateSummary refreshes the summary labels and resets the run-state UI.
// Called every time the step is re-entered so stale results don't linger.
func (s *StepReview) updateSummary() {
	// Show a meaningful input label for both CSV and SQL connectors.
	inputLabel := s.wiz.State.InputConfig.FilePath
	if inputLabel == "" {
		inputLabel = s.wiz.State.InputConnectorName
		if s.wiz.State.InputConfig.Host != "" {
			inputLabel += " — " + s.wiz.State.InputConfig.Host
		}
	} else {
		inputLabel = filepath.Base(inputLabel)
	}
	s.inputFileLbl.SetText(inputLabel)

	outputLabel := filepath.Base(s.wiz.State.OutputConfig.FilePath)
	if s.wiz.State.OutputTemplate != "" {
		// Find the human-readable template name from the loaded columns context.
		// We use the template ID as a fallback; the name is stored in OutputTemplate.
		outputLabel += "  (" + s.wiz.State.OutputTemplate + ")"
	}
	s.outputFileLbl.SetText(outputLabel)
	s.mappingCountLbl.SetText(fmt.Sprintf("%d column mapping(s)", len(s.wiz.State.Mappings)))

	// Reset run state so re-entering the step shows a clean slate.
	s.progress.Stop()
	s.progress.Hide()
	s.successSection.Hide()
	s.errorSection.Hide()
	s.saveSection.Hide()

	// Validate required template fields and gate the Run button accordingly.
	if missing := s.validateTemplateMappings(); len(missing) > 0 {
		s.templateValidationLbl.SetText(
			"⚠  Required columns not mapped: " + strings.Join(missing, ", "),
		)
		s.templateValidationLbl.Show()
		s.runBtn.Disable()
	} else {
		s.templateValidationLbl.Hide()
		s.runBtn.Enable()
	}
	s.runBtn.Show()
}

func (s *StepReview) onRun() {
	s.runBtn.Disable()
	s.successSection.Hide()
	s.errorSection.Hide()
	s.saveSection.Hide()
	s.progress.Show()
	s.progress.Start()

	profile := s.buildProfile()
	startTime := time.Now()

	go func() {
		count, err := engine.Run(profile, s.credStore)
		fyne.Do(func() {
			s.progress.Stop()
			s.progress.Hide()
			if err != nil {
				s.errorMsgLbl.SetText("Error: " + err.Error())
				s.errorSection.Show()
				s.runBtn.Enable() // allow retry
			} else {
				s.rowCountLbl.SetText(fmt.Sprintf("✓  %d rows processed successfully.", count))
				s.successSection.Show()
				s.saveSection.Show()
				s.runBtn.Enable()
			}

			// Build a readable input source label for the audit entry.
			inputSource := s.wiz.State.InputConnectorName
			if s.wiz.State.InputConfig.Host != "" {
				inputSource += ": " + s.wiz.State.InputConfig.Host
			} else if s.wiz.State.InputConfig.FilePath != "" {
				inputSource += ": " + filepath.Base(s.wiz.State.InputConfig.FilePath)
			}
			errStr := ""
			if err != nil {
				errStr = err.Error()
			}
			_ = audit.AppendEntry(audit.AuditEntry{
				Timestamp:   startTime,
				ProfileName: profile.Name,
				ProfileID:   profile.ID,
				InputSource: inputSource,
				OutputPath:  profile.Output.Path,
				RowsOut:     count,
				DurationMs:  time.Since(startTime).Milliseconds(),
				Error:       errStr,
			})
		})
	}()
}

func (s *StepReview) onSaveProfile() {
	name := strings.TrimSpace(s.profileNameEntry.Text)
	if name == "" {
		s.saveStatusLbl.SetText("⚠  Please enter a profile name.")
		return
	}
	if strings.ContainsAny(name, `/\:*?"<>|`) {
		s.saveStatusLbl.SetText("⚠  Name contains invalid characters.")
		return
	}

	profile := s.buildProfile()
	profile.Name = name
	profile.ID = strings.ToLower(strings.ReplaceAll(name, " ", "-"))
	profile.Version = 1

	profileDir := filepath.Join(filepath.Dir(os.Args[0]), "profiles")
	savePath := filepath.Join(profileDir, profile.ID+".json")

	if err := config.SaveProfile(savePath, profile); err != nil {
		s.saveStatusLbl.SetText("⚠  " + err.Error())
		return
	}

	s.saveStatusLbl.SetText("✓  Saved to " + savePath)
	s.wiz.State.ProfileName = name
}

func (s *StepReview) buildProfile() types.Profile {
	return types.Profile{
		Version: 1,
		Input: types.IOConfig{
			Connector:     s.wiz.State.InputConnectorName,
			Path:          s.wiz.State.InputConfig.FilePath,
			Host:          s.wiz.State.InputConfig.Host,
			Port:          s.wiz.State.InputConfig.Port,
			Database:      s.wiz.State.InputConfig.Database,
			Username:      s.wiz.State.InputConfig.Username,
			CredentialRef: s.wiz.State.InputCredentialRef,
			Query:         s.wiz.State.InputQuery,
		},
		Output: types.IOConfig{
			Connector: s.wiz.State.OutputConnectorName,
			Path:      s.wiz.State.OutputConfig.FilePath,
			Template:  s.wiz.State.OutputTemplate,
		},
		Mappings: s.wiz.State.Mappings,
	}
}

// validateTemplateMappings returns the names of required template columns that
// have no InputCol assigned. Returns nil when there is no active template.
func (s *StepReview) validateTemplateMappings() []string {
	if len(s.wiz.State.OutputTemplateColumns) == 0 {
		return nil
	}
	mapped := map[string]bool{}
	for _, m := range s.wiz.State.Mappings {
		if m.InputCol != "" {
			mapped[m.OutputCol] = true
		}
	}
	var missing []string
	for _, col := range s.wiz.State.OutputTemplateColumns {
		if col.Required && !mapped[col.Name] {
			missing = append(missing, col.Name)
		}
	}
	return missing
}

func (s *StepReview) Validate() error { return nil }
