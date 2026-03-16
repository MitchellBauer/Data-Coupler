package ui

import (
	"encoding/json"
	"fmt"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"

	"github.com/mitchellbauer/data-coupler/internal/transform"
	"github.com/mitchellbauer/data-coupler/internal/types"
)

// StepTransform is the optional transform configuration step.
// It sits between Map Columns and Review & Run.
// Users can add zero or more ordered transforms per output column.
type StepTransform struct {
	wiz *Wizard

	content       fyne.CanvasObject
	lastMappingKey string // detects when mappings changed so we can rebuild
}

func (s *StepTransform) Title() string { return "Configure Transforms (Optional)" }

// Validate always returns nil — transforms are never required.
func (s *StepTransform) Validate() error { return nil }

func (s *StepTransform) Content() fyne.CanvasObject {
	key := mappingKey(s.wiz.State.Mappings)
	if s.content != nil && s.lastMappingKey == key {
		return s.content
	}
	s.lastMappingKey = key
	s.content = s.build()
	return s.content
}

// build constructs the full step UI from the current WizardState.Mappings.
func (s *StepTransform) build() fyne.CanvasObject {
	sections := make([]fyne.CanvasObject, 0, len(s.wiz.State.Mappings)+1)

	// Macola auto-suggest banner.
	if s.wiz.State.InputConnectorName == "mssql" && s.anyMissingTrimSpace() {
		banner := widget.NewLabel("ℹ  Macola data often has trailing spaces. Add TrimSpace to all columns?")
		banner.Wrapping = fyne.TextWrapWord
		banner.Importance = widget.LowImportance

		addAllBtn := widget.NewButton("Add TrimSpace to All", func() {
			for i := range s.wiz.State.Mappings {
				if !s.hasTrimSpace(i) {
					s.wiz.State.Mappings[i].Transforms = append(
						[]types.Transform{{Type: "TrimSpace"}},
						s.wiz.State.Mappings[i].Transforms...,
					)
				}
			}
			// Invalidate cache and rebuild.
			s.lastMappingKey = ""
			s.content = s.build()
		})

		sections = append(sections, widget.NewCard("", "", container.NewBorder(nil, nil, nil, addAllBtn, banner)))
	}

	// Per-mapping sections.
	for i := range s.wiz.State.Mappings {
		sections = append(sections, s.buildMappingSection(i))
		sections = append(sections, widget.NewSeparator())
	}

	if len(s.wiz.State.Mappings) == 0 {
		lbl := widget.NewLabel("No mappings defined. Go back and add column mappings first.")
		lbl.Alignment = fyne.TextAlignCenter
		return container.NewCenter(lbl)
	}

	return container.NewBorder(nil, nil, nil, nil,
		container.NewVScroll(container.NewPadded(container.NewVBox(sections...))),
	)
}

// buildMappingSection renders the transform editor for one mapping row.
func (s *StepTransform) buildMappingSection(idx int) fyne.CanvasObject {
	m := &s.wiz.State.Mappings[idx]

	header := widget.NewLabel(fmt.Sprintf("  %s  ←  %s", m.OutputCol, m.InputCol))
	header.TextStyle = fyne.TextStyle{Bold: true}

	// Container that holds the transform rows; rebuilt when transforms change.
	transformsBox := container.NewVBox()

	// rebuildTransforms redraws the transform list and preview for this mapping.
	var rebuildTransforms func()
	rebuildTransforms = func() {
		transformsBox.Objects = nil

		for ti := range m.Transforms {
			ti := ti // capture
			t := &m.Transforms[ti]

			nameLabel := widget.NewLabel(t.Type)
			nameLabel.TextStyle = fyne.TextStyle{Bold: true}

			removeBtn := widget.NewButton("×", func() {
				m.Transforms = append(m.Transforms[:ti], m.Transforms[ti+1:]...)
				rebuildTransforms()
			})
			removeBtn.Importance = widget.DangerImportance

			upBtn := widget.NewButton("↑", func() {
				if ti > 0 {
					m.Transforms[ti], m.Transforms[ti-1] = m.Transforms[ti-1], m.Transforms[ti]
					rebuildTransforms()
				}
			})
			downBtn := widget.NewButton("↓", func() {
				if ti < len(m.Transforms)-1 {
					m.Transforms[ti], m.Transforms[ti+1] = m.Transforms[ti+1], m.Transforms[ti]
					rebuildTransforms()
				}
			})

			titleRow := container.NewBorder(nil, nil, nil,
				container.NewHBox(upBtn, downBtn, removeBtn),
				nameLabel,
			)

			paramForm := s.buildParamForm(t)
			preview := s.buildPreview(idx)

			transformsBox.Add(container.NewPadded(container.NewVBox(titleRow, paramForm, preview)))
		}

		transformsBox.Refresh()
	}

	// "Add Transform" select widget.
	addSelect := widget.NewSelect(transform.List(), func(name string) {
		if name == "" {
			return
		}
		m.Transforms = append(m.Transforms, types.Transform{
			Type:   name,
			Params: map[string]string{},
		})
		rebuildTransforms()
	})
	addSelect.PlaceHolder = "+ Add Transform"

	rebuildTransforms()

	return container.NewVBox(
		container.NewPadded(header),
		transformsBox,
		container.NewPadded(container.NewBorder(nil, nil, nil, nil, addSelect)),
	)
}

// buildParamForm returns a form widget appropriate for the given transform type.
func (s *StepTransform) buildParamForm(t *types.Transform) fyne.CanvasObject {
	if t.Params == nil {
		t.Params = map[string]string{}
	}

	makeEntry := func(key, placeholder string) *widget.Entry {
		e := widget.NewEntry()
		e.SetPlaceHolder(placeholder)
		e.SetText(t.Params[key])
		e.OnChanged = func(v string) { t.Params[key] = v }
		return e
	}

	switch t.Type {
	case "TrimSpace", "ToUpper", "ToLower":
		lbl := widget.NewLabel("No parameters")
		lbl.Importance = widget.LowImportance
		return container.NewPadded(lbl)

	case "DateFormat":
		return container.NewPadded(container.New(layout.NewFormLayout(),
			widget.NewLabel("From format"), makeEntry("from", "e.g. 01/02/2006"),
			widget.NewLabel("To format"), makeEntry("to", "e.g. 2006-01-02"),
		))

	case "Split":
		return container.NewPadded(container.New(layout.NewFormLayout(),
			widget.NewLabel("Separator"), makeEntry("separator", `e.g. -`),
			widget.NewLabel("Index (0-based)"), makeEntry("index", "0"),
		))

	case "Prefix", "Suffix":
		return container.NewPadded(container.New(layout.NewFormLayout(),
			widget.NewLabel("Value"), makeEntry("value", "text to add"),
		))

	case "Default":
		return container.NewPadded(container.New(layout.NewFormLayout(),
			widget.NewLabel("Default value"), makeEntry("value", "e.g. N/A"),
		))

	case "Concatenate":
		colsEntry := makeEntry("cols", "e.g. FirstName, LastName")
		sepEntry := makeEntry("separator", `e.g. (space)`)
		return container.NewPadded(container.New(layout.NewFormLayout(),
			widget.NewLabel("Input columns"), colsEntry,
			widget.NewLabel("Separator"), sepEntry,
		))

	case "LookupReplace":
		return container.NewPadded(s.buildLookupEditor(t))

	default:
		return container.NewPadded(widget.NewLabel("(no param editor for this transform)"))
	}
}

// buildLookupEditor renders a multi-row key→value editor for LookupReplace.
func (s *StepTransform) buildLookupEditor(t *types.Transform) fyne.CanvasObject {
	if t.Params == nil {
		t.Params = map[string]string{}
	}

	// Parse existing map param.
	pairs := make([][2]string, 0)
	if raw, ok := t.Params["map"]; ok && raw != "" {
		var m map[string]string
		if json.Unmarshal([]byte(raw), &m) == nil {
			for k, v := range m {
				pairs = append(pairs, [2]string{k, v})
			}
		}
	}

	pairsBox := container.NewVBox()

	syncMapParam := func() {
		m := make(map[string]string, len(pairs))
		for _, p := range pairs {
			if p[0] != "" {
				m[p[0]] = p[1]
			}
		}
		b, _ := json.Marshal(m)
		t.Params["map"] = string(b)
	}

	var rebuildPairs func()
	rebuildPairs = func() {
		pairsBox.Objects = nil
		for pi := range pairs {
			pi := pi
			keyEntry := widget.NewEntry()
			keyEntry.SetPlaceHolder("key")
			keyEntry.SetText(pairs[pi][0])
			keyEntry.OnChanged = func(v string) { pairs[pi][0] = v; syncMapParam() }

			valEntry := widget.NewEntry()
			valEntry.SetPlaceHolder("replacement")
			valEntry.SetText(pairs[pi][1])
			valEntry.OnChanged = func(v string) { pairs[pi][1] = v; syncMapParam() }

			removeBtn := widget.NewButton("×", func() {
				pairs = append(pairs[:pi], pairs[pi+1:]...)
				syncMapParam()
				rebuildPairs()
			})
			removeBtn.Importance = widget.DangerImportance

			pairsBox.Add(container.New(layout.NewGridLayout(3), keyEntry, valEntry, removeBtn))
		}
		pairsBox.Refresh()
	}

	addPairBtn := widget.NewButton("+ Add pair", func() {
		pairs = append(pairs, [2]string{"", ""})
		rebuildPairs()
	})

	rebuildPairs()

	return container.NewVBox(
		container.NewHBox(widget.NewLabel("Key → Replacement"), addPairBtn),
		pairsBox,
	)
}

// buildPreview shows a before/after sample value for the mapping's transform chain.
func (s *StepTransform) buildPreview(mappingIdx int) fyne.CanvasObject {
	m := s.wiz.State.Mappings[mappingIdx]
	if len(m.Transforms) == 0 {
		return widget.NewLabel("")
	}

	before := ""
	if len(s.wiz.State.InputPreviewRows) > 0 {
		row := s.wiz.State.InputPreviewRows[0]
		for ci, h := range s.wiz.State.InputHeaders {
			if h == m.InputCol && ci < len(row) {
				before = row[ci]
				break
			}
		}
	}

	if before == "" && len(s.wiz.State.InputPreviewRows) == 0 {
		lbl := widget.NewLabel("(no preview data available)")
		lbl.Importance = widget.LowImportance
		return lbl
	}

	after := runPreviewChain(before, m.Transforms, s.wiz.State.InputPreviewRows, s.wiz.State.InputHeaders, m.InputCol)

	preview := widget.NewLabel(fmt.Sprintf("Preview:  %q  →  %q", before, after))
	preview.Importance = widget.LowImportance
	return preview
}

// runPreviewChain applies the transform chain locally for live preview.
// It mirrors the engine logic without importing it.
func runPreviewChain(value string, transforms []types.Transform, previewRows [][]string, headers []string, inputCol string) string {
	// Build a minimal headerMap for RowTransformer support.
	headerMap := make(map[string]int, len(headers))
	for i, h := range headers {
		headerMap[h] = i
	}
	var inputRow []string
	if len(previewRows) > 0 {
		inputRow = previewRows[0]
	}

	for _, t := range transforms {
		tr, ok := transform.Get(t.Type)
		if !ok {
			continue
		}
		var err error
		if rt, ok := tr.(transform.RowTransformer); ok {
			value, err = rt.ApplyRow(inputRow, headerMap, t.Params)
		} else {
			value, err = tr.Apply(value, t.Params)
		}
		if err != nil {
			return fmt.Sprintf("(error: %v)", err)
		}
	}
	return value
}

// ── helpers ───────────────────────────────────────────────────────────────────

// mappingKey produces a string that changes when the mapping list changes
// (different length or different column names), used to decide when to rebuild.
func mappingKey(mappings []types.Mapping) string {
	var sb strings.Builder
	for _, m := range mappings {
		sb.WriteString(m.InputCol)
		sb.WriteByte('|')
		sb.WriteString(m.OutputCol)
		sb.WriteByte(';')
	}
	return sb.String()
}

// anyMissingTrimSpace returns true if at least one mapping lacks a TrimSpace transform.
func (s *StepTransform) anyMissingTrimSpace() bool {
	for _, m := range s.wiz.State.Mappings {
		if !hasTrimSpaceInSlice(m.Transforms) {
			return true
		}
	}
	return false
}

// hasTrimSpace returns true if mapping[idx] already has TrimSpace.
func (s *StepTransform) hasTrimSpace(idx int) bool {
	return hasTrimSpaceInSlice(s.wiz.State.Mappings[idx].Transforms)
}

func hasTrimSpaceInSlice(transforms []types.Transform) bool {
	for _, t := range transforms {
		if t.Type == "TrimSpace" {
			return true
		}
	}
	return false
}
