package templates

import "testing"

func TestListTemplates_Count(t *testing.T) {
	tmpls, err := ListTemplates()
	if err != nil {
		t.Fatalf("ListTemplates() error: %v", err)
	}
	if len(tmpls) != 7 {
		t.Errorf("expected 7 templates, got %d", len(tmpls))
	}
}

func TestListTemplates_RequiredFieldsNonEmpty(t *testing.T) {
	tmpls, err := ListTemplates()
	if err != nil {
		t.Fatalf("ListTemplates() error: %v", err)
	}
	for _, tmpl := range tmpls {
		for _, col := range tmpl.Columns {
			if col.Required && col.Name == "" {
				t.Errorf("template %q has a required column with empty name", tmpl.ID)
			}
		}
	}
}

func TestLoadTemplate_NotFound(t *testing.T) {
	_, err := LoadTemplate("fishbowl/nonexistent")
	if err == nil {
		t.Error("expected error for nonexistent template, got nil")
	}
}
