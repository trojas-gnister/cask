// Package config handles TOML configuration loading, validation, and persistence.
package config

import (
	"os"
	"path/filepath"
)

const (
	AppName    = "cask"
	MainConfig = "config.toml"
)

// HomeDir returns the user's home directory.
func HomeDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return os.Getenv("HOME")
	}
	return home
}

// ConfigDir returns ~/.config/cask/
func ConfigDir() string {
	if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
		return filepath.Join(xdg, AppName)
	}
	return filepath.Join(HomeDir(), ".config", AppName)
}

// StateDir returns ~/.config/cask/state/
func StateDir() string {
	return filepath.Join(ConfigDir(), "state")
}

// MainConfigPath returns the default config file path.
func MainConfigPath() string {
	return filepath.Join(ConfigDir(), MainConfig)
}

// StatePath returns the path for a named state file.
func StatePath(name string) string {
	return filepath.Join(StateDir(), name)
}

// GenerationsDir returns the path to the generations directory.
func GenerationsDir() string {
	return filepath.Join(StateDir(), "generations")
}

// LockfilePath returns the path to the lockfile.
func LockfilePath() string {
	return filepath.Join(ConfigDir(), "cask.lock")
}

// EnsureDir creates a directory and all parents if they don't exist.
func EnsureDir(path string) error {
	return os.MkdirAll(path, 0o755)
}

// ResolveConfigPath resolves a config file path.
// Absolute paths are returned as-is. Relative paths are resolved
// first against ConfigDir, then the current directory.
func ResolveConfigPath(path string) string {
	if filepath.IsAbs(path) {
		return path
	}
	candidate := filepath.Join(ConfigDir(), path)
	if _, err := os.Stat(candidate); err == nil {
		return candidate
	}
	abs, err := filepath.Abs(path)
	if err == nil {
		if _, err := os.Stat(abs); err == nil {
			return abs
		}
	}
	return candidate
}
