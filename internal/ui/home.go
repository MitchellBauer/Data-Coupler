package ui

import (
	"net/url"
	"os"
	"path/filepath"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/widget"

	"github.com/mitchellbauer/data-coupler/internal/config"
	"github.com/mitchellbauer/data-coupler/internal/types"
)

// Home is the landing screen shown before the wizard starts.
// It offers New Conversion, Load a Saved Profile, and (if applicable) Resume Last Job.
type Home struct {
	window          fyne.Window
	wiz             *Wizard
	hasLastProfile  bool
	lastProfilePath string
	errorLbl        *widget.Label

	// Update banner — hidden until an update is detected.
	updateBanner *fyne.Container
	updateLink   *widget.Hyperlink

	// Callbacks set by the caller (LaunchApp) so Home can trigger transitions.
	onNewConversion func()
	onLoadProfile   func(path string) // called with the loaded profile file path
}

// Content builds and returns the home screen UI.
func (h *Home) Content() fyne.CanvasObject {
	// ── Update banner (hidden until ShowUpdateBanner is called) ───────────────
	h.updateLink = widget.NewHyperlink("", nil)
	dismissBtn := widget.NewButton("×", func() { h.updateBanner.Hide() })
	h.updateBanner = container.NewHBox(
		widget.NewLabel("🔔 Update available:"),
		h.updateLink,
		layout.NewSpacer(),
		dismissBtn,
	)
	h.updateBanner.Hide()

	// ── Title ──────────────────────────────────────────────────────────────────
	title := widget.NewLabel("Data Coupler")
	title.TextStyle = fyne.TextStyle{Bold: true}
	title.Alignment = fyne.TextAlignCenter

	subtitle := widget.NewLabel("The modular data conversion tool.")
	subtitle.Alignment = fyne.TextAlignCenter

	// ── Action buttons ────────────────────────────────────────────────────────
	newBtn := widget.NewButton("▶  New Conversion", func() {
		h.errorLbl.Hide()
		h.onNewConversion()
	})
	newBtn.Importance = widget.HighImportance

	loadBtn := widget.NewButton("📂  Load a Saved Profile", h.onBrowseProfile)

	buttons := container.NewVBox(newBtn, loadBtn)

	// ── Resume Last Job (conditional) ────────────────────────────────────────
	if h.hasLastProfile {
		resumeBtn := widget.NewButton(
			"↩  Resume Last Job  ("+filepath.Base(h.lastProfilePath)+")",
			func() {
				h.loadProfileFrom(h.lastProfilePath)
			},
		)
		buttons.Add(resumeBtn)
	}

	// ── Inline error label (hidden by default) ────────────────────────────────
	h.errorLbl = widget.NewLabel("")
	h.errorLbl.Importance = widget.DangerImportance
	h.errorLbl.Alignment = fyne.TextAlignCenter
	h.errorLbl.Wrapping = fyne.TextWrapWord
	h.errorLbl.Hide()

	body := container.NewVBox(
		title,
		subtitle,
		widget.NewSeparator(),
		buttons,
		h.errorLbl,
	)

	return container.NewBorder(
		h.updateBanner, nil, nil, nil,
		container.NewCenter(container.NewPadded(body)),
	)
}

// ShowUpdateBanner reveals the update banner with the given version and download URL.
// Safe to call from any goroutine via fyne.Do.
func (h *Home) ShowUpdateBanner(ver, rawURL string) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return
	}
	h.updateLink.SetText("Version " + ver + " available — click to download")
	h.updateLink.SetURL(u)
	h.updateBanner.Show()
}

// onBrowseProfile opens a file-open dialog scoped to the profiles folder.
func (h *Home) onBrowseProfile() {
	d := dialog.NewFileOpen(func(r fyne.URIReadCloser, err error) {
		if err != nil {
			fyne.Do(func() { h.showError("Could not open dialog: " + err.Error()) })
			return
		}
		if r == nil {
			return // user cancelled
		}
		path := uriToFilePath(r.URI())
		r.Close()
		h.loadProfileFrom(path)
	}, h.window)

	d.SetFilter(storage.NewExtensionFileFilter([]string{".json"}))

	// Scope the dialog to the profiles folder if it exists.
	profileDir := filepath.Join(filepath.Dir(os.Args[0]), "profiles")
	if profileURI := storage.NewFileURI(profileDir); profileURI != nil {
		if listable, err := storage.ListerForURI(profileURI); err == nil {
			d.SetLocation(listable)
		}
	}

	d.Show()
}

// loadProfileFrom reads the JSON profile at path, populates WizardState, and
// transitions to the Review step.
func (h *Home) loadProfileFrom(path string) {
	profile, err := config.LoadProfile(path)
	if err != nil {
		fyne.Do(func() { h.showError("Could not load profile: " + err.Error()) })
		return
	}

	populateWizardState(h.wiz.State, profile)

	fyne.Do(func() {
		h.errorLbl.Hide()
		h.onLoadProfile(path)
	})
}

// populateWizardState writes a loaded Profile's fields into WizardState.
func populateWizardState(state *WizardState, p types.Profile) {
	state.InputConnectorName = p.Input.Connector
	state.InputConfig.FilePath = p.Input.Path
	state.OutputConnectorName = p.Output.Connector
	state.OutputConfig.FilePath = p.Output.Path
	state.Mappings = p.Mappings
	state.ProfileName = p.Name
}

func (h *Home) showError(msg string) {
	h.errorLbl.SetText("⚠  " + msg)
	h.errorLbl.Show()
}
