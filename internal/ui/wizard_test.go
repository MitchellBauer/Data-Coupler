package ui

import (
	"errors"
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/test"
	"fyne.io/fyne/v2/widget"
)

// testStep is a minimal Step for use in wizard navigation tests.
type testStep struct {
	title       string
	validateErr error
}

func (s *testStep) Title() string              { return s.title }
func (s *testStep) Content() fyne.CanvasObject { return widget.NewLabel(s.title) }
func (s *testStep) Validate() error            { return s.validateErr }

// newTestWizard creates a headless wizard with the given steps.
func newTestWizard(steps []Step) *Wizard {
	a := test.NewApp()
	w := a.NewWindow("test")
	wiz := NewWizard(w, steps)
	return wiz
}

func TestWizardStartsOnFirstStep(t *testing.T) {
	wiz := newTestWizard([]Step{
		&testStep{"A", nil},
		&testStep{"B", nil},
	})
	wiz.Start()

	if wiz.currentStep != 0 {
		t.Errorf("after Start: expected step 0, got %d", wiz.currentStep)
	}
}

func TestWizardNextAdvancesStep(t *testing.T) {
	wiz := newTestWizard([]Step{
		&testStep{"A", nil},
		&testStep{"B", nil},
	})
	wiz.Start()
	wiz.Next()

	if wiz.currentStep != 1 {
		t.Errorf("after Next: expected step 1, got %d", wiz.currentStep)
	}
}

func TestWizardBackRetreats(t *testing.T) {
	wiz := newTestWizard([]Step{
		&testStep{"A", nil},
		&testStep{"B", nil},
	})
	wiz.Start()
	wiz.Next()
	wiz.Back()

	if wiz.currentStep != 0 {
		t.Errorf("after Next+Back: expected step 0, got %d", wiz.currentStep)
	}
}

func TestWizardBackOnFirstStepIsNoop(t *testing.T) {
	wiz := newTestWizard([]Step{
		&testStep{"A", nil},
	})
	wiz.Start()
	wiz.Back()

	if wiz.currentStep != 0 {
		t.Errorf("Back on first step should be noop, got step %d", wiz.currentStep)
	}
}

func TestWizardNextBlockedByValidate(t *testing.T) {
	wiz := newTestWizard([]Step{
		&testStep{"A", errors.New("not ready")},
		&testStep{"B", nil},
	})
	wiz.Start()
	wiz.Next()

	if wiz.currentStep != 0 {
		t.Errorf("Next with failing Validate should stay at step 0, got %d", wiz.currentStep)
	}
}

func TestWizardGoToJumpsDirectly(t *testing.T) {
	wiz := newTestWizard([]Step{
		&testStep{"A", nil},
		&testStep{"B", nil},
		&testStep{"C", nil},
	})
	wiz.Start()
	wiz.GoTo(2)

	if wiz.currentStep != 2 {
		t.Errorf("GoTo(2): expected step 2, got %d", wiz.currentStep)
	}
}
