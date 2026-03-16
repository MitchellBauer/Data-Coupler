package transform

import "testing"

func TestTrimSpace(t *testing.T) {
	tr := &TrimSpace{}

	cases := []struct {
		input string
		want  string
	}{
		{"  hello", "hello"},
		{"hello  ", "hello"},
		{"  hello  ", "hello"},
		{"hello", "hello"},
	}

	for _, c := range cases {
		got, err := tr.Apply(c.input, nil)
		if err != nil {
			t.Errorf("TrimSpace.Apply(%q) unexpected error: %v", c.input, err)
		}
		if got != c.want {
			t.Errorf("TrimSpace.Apply(%q) = %q, want %q", c.input, got, c.want)
		}
	}
}

func TestDefault(t *testing.T) {
	tr := &Default{}
	params := map[string]string{"value": "N/A"}

	cases := []struct {
		input string
		want  string
	}{
		{"", "N/A"},
		{"hello", "hello"},
	}

	for _, c := range cases {
		got, err := tr.Apply(c.input, params)
		if err != nil {
			t.Errorf("Default.Apply(%q) unexpected error: %v", c.input, err)
		}
		if got != c.want {
			t.Errorf("Default.Apply(%q) = %q, want %q", c.input, got, c.want)
		}
	}
}
