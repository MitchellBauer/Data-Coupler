package ui

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"github.com/mitchellbauer/data-coupler/internal/connector"
	"github.com/mitchellbauer/data-coupler/internal/templates"
	"github.com/mitchellbauer/data-coupler/internal/types"
)

// WizardState carries user choices across all wizard steps.
type WizardState struct {
	InputConnectorName string
	InputConfig        connector.ConnectionConfig
	InputQuery         string // SQL query (database connectors only)
	InputCredentialRef string // credential ref used to save/load the password
	InputHeaders       []string
	InputPreviewRows   [][]string
	OutputConnectorName    string
	OutputConfig           connector.ConnectionConfig
	OutputDelimiter        rune
	OutputTemplate         string                    // e.g. "fishbowl/parts"; empty = plain CSV
	OutputTemplateColumns  []templates.TemplateColumn // loaded columns; nil = plain CSV
	Mappings               []types.Mapping
	ProfileName            string
}

// Step is the interface every wizard screen must implement.
type Step interface {
	Title() string
	Content() fyne.CanvasObject
	Validate() error // nil means Next is allowed; non-nil blocks navigation.
}

// Wizard manages the multi-step navigation container.
type Wizard struct {
	window      fyne.Window
	State       *WizardState
	steps       []Step
	currentStep int

	// chrome is the full wizard UI (top bar + content area + bottom bar).
	// Callers swap it into the window via w.SetContent(wiz.Chrome()).
	chrome fyne.CanvasObject

	// Retained references for in-place updates.
	stepLabel *widget.Label
	content   *fyne.Container // stack — Objects[0] is swapped on navigation
	backBtn   *widget.Button
	nextBtn   *widget.Button
	errorLbl  *widget.Label
}

// NewWizard builds the wizard chrome and returns the Wizard.
// The caller is responsible for calling w.SetContent(wiz.Chrome()) when ready
// to show the wizard, and wiz.Start() after all step.wiz fields have been set.
func NewWizard(w fyne.Window, steps []Step) *Wizard {
	wiz := &Wizard{
		window: w,
		State:  &WizardState{OutputDelimiter: ','},
		steps:  steps,
	}

	// ── Top bar ──────────────────────────────────────────────────────────────
	title := widget.NewLabel("Data Coupler")
	title.TextStyle = fyne.TextStyle{Bold: true}
	title.Alignment = fyne.TextAlignCenter

	wiz.stepLabel = widget.NewLabel("")
	wiz.stepLabel.Alignment = fyne.TextAlignCenter

	topBar := container.NewVBox(
		title,
		wiz.stepLabel,
		widget.NewSeparator(),
	)

	// ── Content area (swappable) ──────────────────────────────────────────────
	wiz.content = container.NewStack(widget.NewLabel(""))

	// ── Inline error label (hidden until a Validate() call fails) ────────────
	wiz.errorLbl = widget.NewLabel("")
	wiz.errorLbl.Importance = widget.DangerImportance
	wiz.errorLbl.Alignment = fyne.TextAlignCenter
	wiz.errorLbl.Hide()

	// ── Bottom bar ───────────────────────────────────────────────────────────
	wiz.backBtn = widget.NewButton("← Back", wiz.Back)
	wiz.nextBtn = widget.NewButton("Next →", wiz.Next)
	wiz.nextBtn.Importance = widget.HighImportance

	bottomBar := container.NewBorder(
		container.NewVBox(wiz.errorLbl, widget.NewSeparator()),
		nil,
		wiz.backBtn,
		wiz.nextBtn,
	)

	// ── Assemble chrome ──────────────────────────────────────────────────────
	wiz.chrome = container.NewBorder(topBar, bottomBar, nil, nil, wiz.content)

	return wiz
}

// Chrome returns the wizard's full UI container. Pass it to w.SetContent to
// make the wizard visible in the window.
func (wiz *Wizard) Chrome() fyne.CanvasObject {
	return wiz.chrome
}

// Start renders the first step. Call this after all step.wiz fields have been set.
func (wiz *Wizard) Start() {
	wiz.refresh()
}

// Window returns the window this wizard is attached to.
func (wiz *Wizard) Window() fyne.Window {
	return wiz.window
}

// Next validates the current step and advances if valid.
func (wiz *Wizard) Next() {
	if wiz.currentStep >= len(wiz.steps)-1 {
		return
	}
	if err := wiz.steps[wiz.currentStep].Validate(); err != nil {
		wiz.showError(err.Error())
		return
	}
	wiz.clearError()
	wiz.currentStep++
	wiz.refresh()
}

// Back moves to the previous step without validation.
func (wiz *Wizard) Back() {
	if wiz.currentStep <= 0 {
		return
	}
	wiz.clearError()
	wiz.currentStep--
	wiz.refresh()
}

// GoTo jumps directly to a step by index (used by Load Profile).
func (wiz *Wizard) GoTo(n int) {
	if n < 0 || n >= len(wiz.steps) {
		return
	}
	wiz.clearError()
	wiz.currentStep = n
	wiz.refresh()
}

// refresh redraws the step indicator, swaps content, and syncs button states.
func (wiz *Wizard) refresh() {
	total := len(wiz.steps)
	step := wiz.steps[wiz.currentStep]

	wiz.stepLabel.SetText(
		fmt.Sprintf("%s  —  Step %d of %d", step.Title(), wiz.currentStep+1, total),
	)

	// Swap the content object in-place so the surrounding chrome is not rebuilt.
	wiz.content.Objects[0] = step.Content()
	wiz.content.Refresh()

	// Back button: hidden on the first step.
	if wiz.currentStep == 0 {
		wiz.backBtn.Hide()
	} else {
		wiz.backBtn.Show()
	}

	// Next button: relabelled on the last step (Review has its own Run button).
	if wiz.currentStep == total-1 {
		wiz.nextBtn.Hide()
	} else {
		wiz.nextBtn.Show()
		wiz.nextBtn.SetText("Next →")
	}
}

func (wiz *Wizard) showError(msg string) {
	wiz.errorLbl.SetText("⚠  " + msg)
	wiz.errorLbl.Show()
}

func (wiz *Wizard) clearError() {
	wiz.errorLbl.Hide()
	wiz.errorLbl.SetText("")
}
