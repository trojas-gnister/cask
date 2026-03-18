package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadConfigBasic(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.toml")

	content := `
[flatpak]
packages = ["org.mozilla.Firefox", "org.gimp.GIMP"]
manage_overrides = true

[[flatpak.remotes]]
name = "flathub"
url = "https://flathub.org/repo/flathub.flatpakrepo"

[[podman.containers]]
name = "postgres"
image = "docker.io/library/postgres:16"
scope = "user"

[podman.containers.security]
read_only_rootfs = true
drop_all_caps = true
add_caps = ["NET_BIND_SERVICE"]
no_new_privileges = true
`
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	cfg, err := LoadConfig(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.Flatpak == nil {
		t.Fatal("flatpak config should not be nil")
	}
	if len(cfg.Flatpak.Packages) != 2 {
		t.Errorf("expected 2 flatpak packages, got %d", len(cfg.Flatpak.Packages))
	}
	if !cfg.Flatpak.ManageOverrides {
		t.Error("manage_overrides should be true")
	}
	if len(cfg.Flatpak.Remotes) != 1 {
		t.Fatalf("expected 1 remote, got %d", len(cfg.Flatpak.Remotes))
	}
	if cfg.Flatpak.Remotes[0].Name != "flathub" {
		t.Errorf("expected remote name 'flathub', got '%s'", cfg.Flatpak.Remotes[0].Name)
	}

	if cfg.Podman == nil {
		t.Fatal("podman config should not be nil")
	}
	if len(cfg.Podman.Containers) != 1 {
		t.Fatalf("expected 1 container, got %d", len(cfg.Podman.Containers))
	}
	c := cfg.Podman.Containers[0]
	if c.Name != "postgres" {
		t.Errorf("expected container name 'postgres', got '%s'", c.Name)
	}
	if c.Security == nil {
		t.Fatal("container security should not be nil")
	}
	if !c.Security.ReadOnlyRootfs {
		t.Error("read_only_rootfs should be true")
	}
	if !c.Security.NoNewPrivileges {
		t.Error("no_new_privileges should be true")
	}
}

func TestLoadConfigWithEnvExpansion(t *testing.T) {
	t.Setenv("MY_IMAGE", "docker.io/library/redis:7")

	dir := t.TempDir()
	path := filepath.Join(dir, "config.toml")

	content := `
[[podman.containers]]
name = "redis"
image = "${MY_IMAGE}"
scope = "user"
`
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	cfg, err := LoadConfig(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.Podman.Containers[0].Image != "docker.io/library/redis:7" {
		t.Errorf("expected expanded image, got '%s'", cfg.Podman.Containers[0].Image)
	}
}

func TestLoadConfigNotFound(t *testing.T) {
	_, err := LoadConfig("/nonexistent/config.toml")
	if err == nil {
		t.Error("expected error for missing config")
	}
}

func TestLoadConfigInvalidTOML(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "bad.toml")
	os.WriteFile(path, []byte("{{invalid toml"), 0o644)

	_, err := LoadConfig(path)
	if err == nil {
		t.Error("expected error for invalid TOML")
	}
}

func TestLoadConfigWithIncludes(t *testing.T) {
	dir := t.TempDir()

	// Base file
	base := `
[tools]
shell_integration = true

[[tools.tools]]
name = "node"
version = "20"
`
	os.WriteFile(filepath.Join(dir, "base.toml"), []byte(base), 0o644)

	// Main file includes base
	main := `
include = ["base.toml"]

[flatpak]
packages = ["org.mozilla.Firefox"]
`
	mainPath := filepath.Join(dir, "config.toml")
	os.WriteFile(mainPath, []byte(main), 0o644)

	cfg, err := LoadConfig(mainPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.Flatpak == nil || len(cfg.Flatpak.Packages) != 1 {
		t.Error("main config flatpak should be loaded")
	}
	if cfg.Tools == nil {
		t.Error("included tools config should be merged")
	}
}

func TestLoadConfigEmptyFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "empty.toml")
	os.WriteFile(path, []byte(""), 0o644)

	cfg, err := LoadConfig(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Flatpak != nil {
		t.Error("empty config should have nil sections")
	}
}
