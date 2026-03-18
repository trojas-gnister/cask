package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestConfigDir(t *testing.T) {
	dir := ConfigDir()
	if !strings.HasSuffix(dir, filepath.Join(".config", "cask")) {
		t.Errorf("ConfigDir should end with .config/cask, got %s", dir)
	}
}

func TestConfigDirXDG(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", "/tmp/xdg-test")
	dir := ConfigDir()
	if dir != "/tmp/xdg-test/cask" {
		t.Errorf("expected /tmp/xdg-test/cask, got %s", dir)
	}
}

func TestStateDir(t *testing.T) {
	dir := StateDir()
	if !strings.HasSuffix(dir, filepath.Join("cask", "state")) {
		t.Errorf("StateDir should end with cask/state, got %s", dir)
	}
}

func TestMainConfigPath(t *testing.T) {
	p := MainConfigPath()
	if !strings.HasSuffix(p, filepath.Join("cask", "config.toml")) {
		t.Errorf("MainConfigPath should end with cask/config.toml, got %s", p)
	}
}

func TestStatePath(t *testing.T) {
	p := StatePath("global.json")
	if !strings.HasSuffix(p, filepath.Join("state", "global.json")) {
		t.Errorf("StatePath should end with state/global.json, got %s", p)
	}
}

func TestEnsureDir(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "a", "b", "c")
	if err := EnsureDir(dir); err != nil {
		t.Fatalf("EnsureDir failed: %v", err)
	}
	info, err := os.Stat(dir)
	if err != nil {
		t.Fatalf("directory should exist: %v", err)
	}
	if !info.IsDir() {
		t.Error("should be a directory")
	}
}

func TestResolveConfigPathAbsolute(t *testing.T) {
	p := ResolveConfigPath("/etc/cask/config.toml")
	if p != "/etc/cask/config.toml" {
		t.Errorf("absolute paths should be returned as-is, got %s", p)
	}
}

func TestResolveConfigPathRelative(t *testing.T) {
	// For a relative path that doesn't exist anywhere, should fall back to ConfigDir
	p := ResolveConfigPath("nonexistent.toml")
	expected := filepath.Join(ConfigDir(), "nonexistent.toml")
	if p != expected {
		t.Errorf("expected %s, got %s", expected, p)
	}
}
