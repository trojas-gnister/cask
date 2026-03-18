package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestWriteAndLoadRoundTrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.toml")

	cfg := &CaskConfig{
		Flatpak: &FlatpakConfig{
			Packages:        []string{"org.mozilla.Firefox"},
			ManageOverrides: true,
		},
	}

	if err := WriteConfig(cfg, path); err != nil {
		t.Fatalf("write failed: %v", err)
	}

	loaded, err := LoadConfig(path)
	if err != nil {
		t.Fatalf("load failed: %v", err)
	}

	if len(loaded.Flatpak.Packages) != 1 || loaded.Flatpak.Packages[0] != "org.mozilla.Firefox" {
		t.Errorf("round-trip failed: %v", loaded.Flatpak.Packages)
	}
}

func TestAddToConfigList(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.toml")

	// Start with a config that has one package
	os.WriteFile(path, []byte(`[flatpak]
packages = ["org.mozilla.Firefox"]
`), 0o644)

	if err := AddToConfigList(path, "flatpak", "packages", "org.gimp.GIMP"); err != nil {
		t.Fatalf("add failed: %v", err)
	}

	cfg, err := LoadConfig(path)
	if err != nil {
		t.Fatalf("load failed: %v", err)
	}
	if len(cfg.Flatpak.Packages) != 2 {
		t.Errorf("expected 2 packages, got %d", len(cfg.Flatpak.Packages))
	}
}

func TestAddToConfigListNoDuplicates(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.toml")

	os.WriteFile(path, []byte(`[flatpak]
packages = ["org.mozilla.Firefox"]
`), 0o644)

	// Add the same package twice
	AddToConfigList(path, "flatpak", "packages", "org.mozilla.Firefox")
	AddToConfigList(path, "flatpak", "packages", "org.mozilla.Firefox")

	cfg, _ := LoadConfig(path)
	if len(cfg.Flatpak.Packages) != 1 {
		t.Errorf("should not duplicate, got %d packages", len(cfg.Flatpak.Packages))
	}
}

func TestAddToConfigListCreatesSection(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.toml")

	// Empty file
	os.WriteFile(path, []byte(""), 0o644)

	if err := AddToConfigList(path, "flatpak", "packages", "org.mozilla.Firefox"); err != nil {
		t.Fatalf("add failed: %v", err)
	}

	cfg, err := LoadConfig(path)
	if err != nil {
		t.Fatalf("load failed: %v", err)
	}
	if cfg.Flatpak == nil || len(cfg.Flatpak.Packages) != 1 {
		t.Error("should create section and list")
	}
}

func TestRemoveFromConfigList(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.toml")

	os.WriteFile(path, []byte(`[flatpak]
packages = ["org.mozilla.Firefox", "org.gimp.GIMP"]
`), 0o644)

	removed, err := RemoveFromConfigList(path, "flatpak", "packages", "org.mozilla.Firefox")
	if err != nil {
		t.Fatalf("remove failed: %v", err)
	}
	if !removed {
		t.Error("should have removed the package")
	}

	cfg, _ := LoadConfig(path)
	if len(cfg.Flatpak.Packages) != 1 || cfg.Flatpak.Packages[0] != "org.gimp.GIMP" {
		t.Errorf("unexpected packages after removal: %v", cfg.Flatpak.Packages)
	}
}

func TestRemoveFromConfigListNotFound(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.toml")

	os.WriteFile(path, []byte(`[flatpak]
packages = ["org.mozilla.Firefox"]
`), 0o644)

	removed, err := RemoveFromConfigList(path, "flatpak", "packages", "nonexistent")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if removed {
		t.Error("should not have found the package")
	}
}

func TestRemoveFromConfigListNoFile(t *testing.T) {
	removed, err := RemoveFromConfigList("/nonexistent/config.toml", "flatpak", "packages", "x")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if removed {
		t.Error("missing file should return false")
	}
}

func TestUpdateConfigSection(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.toml")

	os.WriteFile(path, []byte(`[flatpak]
manage_overrides = false
`), 0o644)

	err := UpdateConfigSection(path, "flatpak", map[string]any{
		"manage_overrides": true,
	})
	if err != nil {
		t.Fatalf("update failed: %v", err)
	}

	cfg, _ := LoadConfig(path)
	if !cfg.Flatpak.ManageOverrides {
		t.Error("manage_overrides should be updated to true")
	}
}
