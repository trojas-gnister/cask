package state

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewManagerLoadsEmptyState(t *testing.T) {
	dir := t.TempDir()
	m := NewManagerWithPath(filepath.Join(dir, "state.json"))

	s := m.State()
	if s == nil {
		t.Fatal("state should not be nil")
	}
	if len(s.Sections) != 0 {
		t.Error("initial state should have no sections")
	}
	if s.Version != "1" {
		t.Errorf("expected version '1', got '%s'", s.Version)
	}
}

func TestManagerSaveAndLoad(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "state.json")

	m := NewManagerWithPath(path)
	m.State().UpdateSection("flatpak", "abc123")

	if err := m.Save(); err != nil {
		t.Fatalf("save failed: %v", err)
	}

	m2 := NewManagerWithPath(path)
	s := m2.Load()
	sec := s.Sections["flatpak"]
	if sec == nil {
		t.Fatal("flatpak section should exist after load")
	}
	if sec.ConfigHash != "abc123" {
		t.Errorf("expected hash 'abc123', got '%s'", sec.ConfigHash)
	}
	if !sec.Applied {
		t.Error("section should be marked as applied")
	}
	if sec.LastApplied == "" {
		t.Error("last_applied should be set")
	}
}

func TestManagerHasChanged(t *testing.T) {
	dir := t.TempDir()
	m := NewManagerWithPath(filepath.Join(dir, "state.json"))

	// First check: no previous state, should always be changed
	changed, err := m.HasChanged("flatpak", map[string]string{"packages": "vim"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !changed {
		t.Error("new section should be changed")
	}

	// Mark as applied
	if err := m.MarkApplied("flatpak", map[string]string{"packages": "vim"}); err != nil {
		t.Fatalf("mark applied failed: %v", err)
	}

	// Same data: should not be changed
	changed, err = m.HasChanged("flatpak", map[string]string{"packages": "vim"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if changed {
		t.Error("same data should not be changed")
	}

	// Different data: should be changed
	changed, err = m.HasChanged("flatpak", map[string]string{"packages": "emacs"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !changed {
		t.Error("different data should be changed")
	}
}

func TestManagerGetChangedSections(t *testing.T) {
	dir := t.TempDir()
	m := NewManagerWithPath(filepath.Join(dir, "state.json"))

	m.MarkApplied("flatpak", map[string]string{"pkg": "vim"})

	sections := map[string]any{
		"flatpak": map[string]string{"pkg": "vim"},      // unchanged
		"podman":  map[string]string{"name": "postgres"}, // new
	}
	changed, err := m.GetChangedSections(sections)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(changed) != 1 || changed[0] != "podman" {
		t.Errorf("expected [podman], got %v", changed)
	}
}

func TestManagerCorruptStateFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "state.json")

	os.WriteFile(path, []byte("not json"), 0o644)

	m := NewManagerWithPath(path)
	s := m.State()
	if len(s.Sections) != 0 {
		t.Error("corrupt state should return empty state")
	}
}

func TestGlobalStateRunOnce(t *testing.T) {
	s := NewGlobalState()
	if s.HasRunOnce("abc") {
		t.Error("should not have run-once hash initially")
	}
	s.AddRunOnce("abc")
	if !s.HasRunOnce("abc") {
		t.Error("should have run-once hash after adding")
	}
	// Adding again should not duplicate
	s.AddRunOnce("abc")
	if len(s.RunOnceHashes) != 1 {
		t.Errorf("should not duplicate, got %d hashes", len(s.RunOnceHashes))
	}
}
