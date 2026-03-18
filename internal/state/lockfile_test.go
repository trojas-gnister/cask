package state

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestLockfileSaveAndLoad(t *testing.T) {
	// Use a temp dir and override the lockfile path via env
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)

	lf := &Lockfile{
		Flatpaks: []FlatpakLock{
			{AppID: "org.mozilla.Firefox", Version: "120.0", Commit: "abc123"},
		},
		Containers: []ContainerLock{
			{Name: "postgres", Image: "postgres:16", ImageID: "sha256:def456"},
		},
		Tools: []ToolLock{
			{Name: "node", Version: "20.10.0"},
		},
		Devboxes: []DevboxLock{
			{Name: "dev", Image: "ubuntu:22.04"},
		},
	}

	if err := SaveLockfile(lf); err != nil {
		t.Fatalf("save failed: %v", err)
	}

	loaded, err := LoadLockfile()
	if err != nil {
		t.Fatalf("load failed: %v", err)
	}
	if loaded == nil {
		t.Fatal("loaded lockfile should not be nil")
	}
	if len(loaded.Flatpaks) != 1 {
		t.Errorf("expected 1 flatpak, got %d", len(loaded.Flatpaks))
	}
	if loaded.Flatpaks[0].AppID != "org.mozilla.Firefox" {
		t.Errorf("expected Firefox, got %s", loaded.Flatpaks[0].AppID)
	}
	if len(loaded.Containers) != 1 {
		t.Errorf("expected 1 container, got %d", len(loaded.Containers))
	}
	if len(loaded.Tools) != 1 {
		t.Errorf("expected 1 tool, got %d", len(loaded.Tools))
	}
}

func TestLoadLockfileNotFound(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	lf, err := LoadLockfile()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if lf != nil {
		t.Error("missing lockfile should return nil")
	}
}

func TestLockfileRoundTrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.lock")

	lf := &Lockfile{
		Tools: []ToolLock{{Name: "go", Version: "1.22"}},
	}
	data, _ := json.MarshalIndent(lf, "", "  ")
	os.WriteFile(path, data, 0o644)

	loaded, _ := os.ReadFile(path)
	var lf2 Lockfile
	json.Unmarshal(loaded, &lf2)

	if len(lf2.Tools) != 1 || lf2.Tools[0].Name != "go" {
		t.Error("round-trip failed")
	}
}
