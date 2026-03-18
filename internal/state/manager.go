package state

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/iskry/cask/internal/config"
)

const globalStateFile = "global.json"

// Manager handles state persistence and change detection.
type Manager struct {
	path  string
	state *GlobalState
}

// NewManager creates a Manager with the default state file path.
func NewManager() *Manager {
	return &Manager{
		path: config.StatePath(globalStateFile),
	}
}

// NewManagerWithPath creates a Manager with a custom state file path.
func NewManagerWithPath(path string) *Manager {
	return &Manager{path: path}
}

// State returns the current state, loading from disk if necessary.
func (m *Manager) State() *GlobalState {
	if m.state == nil {
		m.state = m.Load()
	}
	return m.state
}

// Load reads state from disk. Returns empty state if file doesn't exist or is corrupt.
func (m *Manager) Load() *GlobalState {
	data, err := os.ReadFile(m.path)
	if err != nil {
		return NewGlobalState()
	}
	var s GlobalState
	if err := json.Unmarshal(data, &s); err != nil {
		return NewGlobalState()
	}
	if s.Sections == nil {
		s.Sections = make(map[string]*SectionState)
	}
	if s.RunOnceHashes == nil {
		s.RunOnceHashes = []string{}
	}
	return &s
}

// Save persists state to disk.
func (m *Manager) Save() error {
	if m.state == nil {
		return nil
	}
	if err := os.MkdirAll(filepath.Dir(m.path), 0o755); err != nil {
		return fmt.Errorf("creating state directory: %w", err)
	}
	data, err := json.MarshalIndent(m.state, "", "  ")
	if err != nil {
		return fmt.Errorf("serializing state: %w", err)
	}
	return os.WriteFile(m.path, data, 0o644)
}

// HasChanged checks if a config section has changed since last apply.
func (m *Manager) HasChanged(section string, configData any) (bool, error) {
	newHash, err := HashData(configData)
	if err != nil {
		return false, err
	}
	return m.State().HasChanged(section, newHash), nil
}

// MarkApplied records that a config section was successfully applied.
func (m *Manager) MarkApplied(section string, configData any) error {
	hash, err := HashData(configData)
	if err != nil {
		return err
	}
	m.State().UpdateSection(section, hash)
	return nil
}

// GetChangedSections returns section names that have changed.
func (m *Manager) GetChangedSections(configSections map[string]any) ([]string, error) {
	var changed []string
	for name, data := range configSections {
		c, err := m.HasChanged(name, data)
		if err != nil {
			return nil, err
		}
		if c {
			changed = append(changed, name)
		}
	}
	return changed, nil
}
