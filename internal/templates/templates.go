package templates

import (
	"embed"
	"encoding/json"
	"fmt"
	"strings"
)

//go:embed fishbowl/*.json
var fishbowlFS embed.FS

// TemplateColumn describes one column in a Fishbowl import template.
type TemplateColumn struct {
	Name        string `json:"name"`
	Required    bool   `json:"required"`
	Description string `json:"description"`
}

// Template describes a Fishbowl import template.
type Template struct {
	ID          string           `json:"id"`
	Name        string           `json:"name"`
	Description string           `json:"description"`
	Columns     []TemplateColumn `json:"columns"`
}

// LoadTemplate loads a single template by ID (e.g. "fishbowl/parts").
func LoadTemplate(id string) (Template, error) {
	path := id + ".json"
	data, err := fishbowlFS.ReadFile(path)
	if err != nil {
		return Template{}, fmt.Errorf("template %q not found", id)
	}
	var t Template
	if err := json.Unmarshal(data, &t); err != nil {
		return Template{}, fmt.Errorf("template %q: %w", id, err)
	}
	return t, nil
}

// ListTemplates returns all embedded templates, sorted by ID.
func ListTemplates() ([]Template, error) {
	entries, err := fishbowlFS.ReadDir("fishbowl")
	if err != nil {
		return nil, err
	}
	var out []Template
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".json") {
			continue
		}
		id := "fishbowl/" + strings.TrimSuffix(e.Name(), ".json")
		t, err := LoadTemplate(id)
		if err != nil {
			return nil, err
		}
		out = append(out, t)
	}
	return out, nil
}
