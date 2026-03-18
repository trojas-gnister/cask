package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	toml "github.com/pelletier/go-toml/v2"
)

// LoadConfig loads a CaskConfig from a TOML file.
// If configPath is empty, uses the default location.
func LoadConfig(configPath string) (*CaskConfig, error) {
	path := configPath
	if path == "" {
		path = MainConfigPath()
	} else if !filepath.IsAbs(path) {
		path = ResolveConfigPath(path)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("config file not found: %s", path)
	}

	// Parse TOML into generic map for expansion pipeline
	var raw map[string]any
	if err := toml.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("invalid TOML in %s: %w", path, err)
	}

	// Expansion pipeline: includes → env vars/tilde
	raw = ExpandIncludes(raw, filepath.Dir(path))
	expanded := ExpandStringsRecursive(raw).(map[string]any)

	// Re-serialize to JSON then unmarshal into typed struct.
	// This lets us use json tags for flexible deserialization
	// while TOML tags handle the initial parse.
	jsonBytes, err := json.Marshal(expanded)
	if err != nil {
		return nil, fmt.Errorf("config serialization failed: %w", err)
	}

	var cfg CaskConfig
	if err := json.Unmarshal(jsonBytes, &cfg); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return &cfg, nil
}

// FindAndLoadConfig finds and loads config, returning both the config and resolved path.
func FindAndLoadConfig(specifiedPath string) (*CaskConfig, string, error) {
	path := specifiedPath
	if path == "" {
		path = MainConfigPath()
	} else {
		path = ResolveConfigPath(path)
	}
	cfg, err := LoadConfig(path)
	return cfg, path, err
}
