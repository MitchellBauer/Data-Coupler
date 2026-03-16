package updater

import "testing"

func TestIsNewer(t *testing.T) {
	tests := []struct {
		current string
		latest  string
		want    bool
	}{
		{"0.3.0", "0.4.0", true},
		{"0.4.0", "0.4.0", false},
		{"0.5.0", "0.4.0", false},
		{"0.3.0", "v0.4.0", true}, // leading "v" stripped
		{"1.0.0", "2.0.0", true},
		{"2.0.0", "1.9.9", false},
		{"0.4.1", "0.4.2", true},
	}

	for _, tc := range tests {
		got := IsNewer(tc.current, tc.latest)
		if got != tc.want {
			t.Errorf("IsNewer(%q, %q) = %v, want %v", tc.current, tc.latest, got, tc.want)
		}
	}
}
