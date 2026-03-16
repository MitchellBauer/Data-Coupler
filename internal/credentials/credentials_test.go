package credentials

import (
	"errors"
	"sort"
	"testing"
)

func newStore(t *testing.T) *FileStore {
	t.Helper()
	return NewFileStore(t.TempDir())
}

func TestSaveLoad(t *testing.T) {
	s := newStore(t)
	if err := s.Save("db-prod", "s3cr3t"); err != nil {
		t.Fatalf("Save() error: %v", err)
	}
	got, err := s.Load("db-prod")
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}
	if got != "s3cr3t" {
		t.Errorf("Load() = %q, want %q", got, "s3cr3t")
	}
}

func TestLoad_NotFound(t *testing.T) {
	s := newStore(t)
	_, err := s.Load("nonexistent-ref")
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("Load() error = %v, want ErrNotFound", err)
	}
}

func TestDelete(t *testing.T) {
	s := newStore(t)
	if err := s.Save("ref1", "pass1"); err != nil {
		t.Fatalf("Save() error: %v", err)
	}
	if err := s.Delete("ref1"); err != nil {
		t.Fatalf("Delete() error: %v", err)
	}
	_, err := s.Load("ref1")
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("Load() after Delete() error = %v, want ErrNotFound", err)
	}
}

func TestList(t *testing.T) {
	s := newStore(t)
	if err := s.Save("ref-a", "pass-a"); err != nil {
		t.Fatalf("Save(ref-a) error: %v", err)
	}
	if err := s.Save("ref-b", "pass-b"); err != nil {
		t.Fatalf("Save(ref-b) error: %v", err)
	}
	refs, err := s.List()
	if err != nil {
		t.Fatalf("List() error: %v", err)
	}
	sort.Strings(refs)
	if len(refs) != 2 || refs[0] != "ref-a" || refs[1] != "ref-b" {
		t.Errorf("List() = %v, want [ref-a ref-b]", refs)
	}
}

func TestSaveLoad_Unicode(t *testing.T) {
	s := newStore(t)
	password := "pässwörd-日本語-🔑"
	if err := s.Save("unicode-ref", password); err != nil {
		t.Fatalf("Save() error: %v", err)
	}
	got, err := s.Load("unicode-ref")
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}
	if got != password {
		t.Errorf("Load() = %q, want %q", got, password)
	}
}

func TestSaveLoad_Overwrite(t *testing.T) {
	s := newStore(t)
	if err := s.Save("ref", "first"); err != nil {
		t.Fatalf("Save(first) error: %v", err)
	}
	if err := s.Save("ref", "second"); err != nil {
		t.Fatalf("Save(second) error: %v", err)
	}
	got, err := s.Load("ref")
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}
	if got != "second" {
		t.Errorf("Load() = %q, want %q", got, "second")
	}
}
