package state

import "time"

// SectionState tracks the state of a single config section.
type SectionState struct {
	ConfigHash  string         `json:"config_hash"`
	Applied     bool           `json:"applied"`
	LastApplied string         `json:"last_applied,omitempty"`
	Metadata    map[string]any `json:"metadata,omitempty"`
}

// GlobalState tracks all config section hashes and run-once deduplication.
type GlobalState struct {
	Sections      map[string]*SectionState `json:"sections"`
	RunOnceHashes []string                 `json:"run_once_hashes"`
	Version       string                   `json:"version"`
}

// NewGlobalState creates an empty GlobalState.
func NewGlobalState() *GlobalState {
	return &GlobalState{
		Sections:      make(map[string]*SectionState),
		RunOnceHashes: []string{},
		Version:       "1",
	}
}

// GetSection returns the SectionState for a named section, creating it if absent.
func (g *GlobalState) GetSection(name string) *SectionState {
	sec, ok := g.Sections[name]
	if !ok {
		sec = &SectionState{}
		g.Sections[name] = sec
	}
	return sec
}

// HasChanged returns true if the section's hash differs from newHash.
func (g *GlobalState) HasChanged(name string, newHash string) bool {
	sec, ok := g.Sections[name]
	if !ok {
		return true
	}
	return sec.ConfigHash != newHash
}

// UpdateSection updates a section's hash and marks it as applied.
func (g *GlobalState) UpdateSection(name string, configHash string) {
	sec := g.GetSection(name)
	sec.ConfigHash = configHash
	sec.Applied = true
	sec.LastApplied = time.Now().UTC().Format(time.RFC3339)
}

// HasRunOnce checks if a run-once hash has been recorded.
func (g *GlobalState) HasRunOnce(hash string) bool {
	for _, h := range g.RunOnceHashes {
		if h == hash {
			return true
		}
	}
	return false
}

// AddRunOnce records a run-once hash for deduplication.
func (g *GlobalState) AddRunOnce(hash string) {
	if !g.HasRunOnce(hash) {
		g.RunOnceHashes = append(g.RunOnceHashes, hash)
	}
}
