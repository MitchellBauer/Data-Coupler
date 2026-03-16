package updater

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCheckLatestRelease_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"tag_name":"v1.2.3"}`))
	}))
	defer srv.Close()

	got, err := checkRelease(srv.URL)
	if err != nil {
		t.Fatalf("checkRelease() error: %v", err)
	}
	if got != "v1.2.3" {
		t.Errorf("checkRelease() = %q, want %q", got, "v1.2.3")
	}
}

func TestCheckLatestRelease_Non200(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	_, err := checkRelease(srv.URL)
	if err == nil {
		t.Error("checkRelease() expected error for non-200 response, got nil")
	}
}

func TestCheckLatestRelease_BadJSON(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{bad json`))
	}))
	defer srv.Close()

	_, err := checkRelease(srv.URL)
	if err == nil {
		t.Error("checkRelease() expected error for bad JSON, got nil")
	}
}

func TestCheckLatestRelease_EmptyTagName(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"tag_name":""}`))
	}))
	defer srv.Close()

	_, err := checkRelease(srv.URL)
	if err == nil {
		t.Error("checkRelease() expected error for empty tag_name, got nil")
	}
}

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
