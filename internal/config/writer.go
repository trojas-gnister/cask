package config

import (
	"fmt"
	"os"
	"path/filepath"

	toml "github.com/pelletier/go-toml/v2"
)

// WriteConfig serializes a CaskConfig to a TOML file.
func WriteConfig(cfg *CaskConfig, path string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("creating config directory: %w", err)
	}
	data, err := toml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("serializing config: %w", err)
	}
	return os.WriteFile(path, data, 0o644)
}

// UpdateConfigSection updates a single top-level section in an existing TOML config file.
func UpdateConfigSection(path string, section string, sectionData map[string]any) error {
	raw := make(map[string]any)
	if data, err := os.ReadFile(path); err == nil {
		if err := toml.Unmarshal(data, &raw); err != nil {
			return fmt.Errorf("parsing existing config: %w", err)
		}
	}

	if existing, ok := raw[section].(map[string]any); ok {
		for k, v := range sectionData {
			existing[k] = v
		}
	} else {
		raw[section] = sectionData
	}

	data, err := toml.Marshal(raw)
	if err != nil {
		return fmt.Errorf("serializing config: %w", err)
	}
	return os.WriteFile(path, data, 0o644)
}

// AddToConfigList adds a value to a list field in a config section (no duplicates).
func AddToConfigList(path string, section string, key string, value string) error {
	raw := make(map[string]any)
	if data, err := os.ReadFile(path); err == nil {
		if err := toml.Unmarshal(data, &raw); err != nil {
			return fmt.Errorf("parsing config: %w", err)
		}
	}

	sec, ok := raw[section].(map[string]any)
	if !ok {
		sec = make(map[string]any)
		raw[section] = sec
	}

	list, _ := sec[key].([]any)
	for _, item := range list {
		if s, ok := item.(string); ok && s == value {
			return nil // Already present
		}
	}
	sec[key] = append(list, value)

	data, err := toml.Marshal(raw)
	if err != nil {
		return fmt.Errorf("serializing config: %w", err)
	}
	return os.WriteFile(path, data, 0o644)
}

// RemoveFromConfigList removes a value from a list field in a config section.
// Returns true if the value was found and removed.
func RemoveFromConfigList(path string, section string, key string, value string) (bool, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return false, nil
	}

	var raw map[string]any
	if err := toml.Unmarshal(data, &raw); err != nil {
		return false, fmt.Errorf("parsing config: %w", err)
	}

	sec, ok := raw[section].(map[string]any)
	if !ok {
		return false, nil
	}

	list, ok := sec[key].([]any)
	if !ok {
		return false, nil
	}

	found := false
	filtered := make([]any, 0, len(list))
	for _, item := range list {
		if s, ok := item.(string); ok && s == value {
			found = true
			continue
		}
		filtered = append(filtered, item)
	}
	if !found {
		return false, nil
	}

	sec[key] = filtered
	out, err := toml.Marshal(raw)
	if err != nil {
		return false, fmt.Errorf("serializing config: %w", err)
	}
	return true, os.WriteFile(path, out, 0o644)
}
