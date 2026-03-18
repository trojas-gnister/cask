package state

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/iskry/cask/internal/config"
)

const maxGenerations = 10

// Generation is a snapshot of the system state at a point in time.
type Generation struct {
	ID         int            `json:"id"`
	Timestamp  string         `json:"timestamp"`
	ConfigHash string         `json:"config_hash"`
	Flatpaks   []string       `json:"flatpaks"`
	Containers []string       `json:"containers"`
	Devboxes   []string       `json:"devboxes"`
	Tools      []string       `json:"tools"`
	Metadata   map[string]any `json:"metadata,omitempty"`
}

// GenerationsIndex tracks all generations and the current one.
type GenerationsIndex struct {
	Current     int   `json:"current"`
	Generations []int `json:"generations"`
}

func generationsDir() string {
	return config.GenerationsDir()
}

func indexPath() string {
	return filepath.Join(generationsDir(), "generations.json")
}

func genFile(id int) string {
	return filepath.Join(generationsDir(), fmt.Sprintf("gen-%04d.json", id))
}

func loadIndex() *GenerationsIndex {
	data, err := os.ReadFile(indexPath())
	if err != nil {
		return &GenerationsIndex{}
	}
	var idx GenerationsIndex
	if err := json.Unmarshal(data, &idx); err != nil {
		return &GenerationsIndex{}
	}
	return &idx
}

func saveIndex(idx *GenerationsIndex) error {
	if err := config.EnsureDir(generationsDir()); err != nil {
		return err
	}
	data, err := json.MarshalIndent(idx, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(indexPath(), data, 0o644)
}

// CreateGeneration creates a new generation snapshot and persists it.
func CreateGeneration(configHash string, flatpaks, containers, devboxes, tools []string) (*Generation, error) {
	if err := config.EnsureDir(generationsDir()); err != nil {
		return nil, err
	}

	idx := loadIndex()
	newID := 1
	if len(idx.Generations) > 0 {
		newID = idx.Generations[len(idx.Generations)-1] + 1
	}

	gen := &Generation{
		ID:         newID,
		Timestamp:  time.Now().UTC().Format(time.RFC3339),
		ConfigHash: configHash,
		Flatpaks:   orEmpty(flatpaks),
		Containers: orEmpty(containers),
		Devboxes:   orEmpty(devboxes),
		Tools:      orEmpty(tools),
	}

	data, err := json.MarshalIndent(gen, "", "  ")
	if err != nil {
		return nil, err
	}
	if err := os.WriteFile(genFile(newID), data, 0o644); err != nil {
		return nil, err
	}

	idx.Generations = append(idx.Generations, newID)
	idx.Current = newID

	// Auto-prune old generations
	for len(idx.Generations) > maxGenerations {
		oldID := idx.Generations[0]
		idx.Generations = idx.Generations[1:]
		os.Remove(genFile(oldID))
	}

	if err := saveIndex(idx); err != nil {
		return nil, err
	}
	return gen, nil
}

// LoadGeneration loads a specific generation by ID.
func LoadGeneration(id int) (*Generation, error) {
	data, err := os.ReadFile(genFile(id))
	if err != nil {
		return nil, fmt.Errorf("generation %d not found", id)
	}
	var gen Generation
	if err := json.Unmarshal(data, &gen); err != nil {
		return nil, fmt.Errorf("parsing generation %d: %w", id, err)
	}
	return &gen, nil
}

// ListGenerations lists all stored generations.
func ListGenerations() ([]*Generation, error) {
	idx := loadIndex()
	var gens []*Generation
	for _, id := range idx.Generations {
		gen, err := LoadGeneration(id)
		if err != nil {
			continue
		}
		gens = append(gens, gen)
	}
	return gens, nil
}

// GetCurrentGeneration returns the most recent generation.
func GetCurrentGeneration() (*Generation, error) {
	idx := loadIndex()
	if idx.Current == 0 {
		return nil, fmt.Errorf("no generations exist")
	}
	return LoadGeneration(idx.Current)
}

// DiffGenerations compares two generations and returns differences.
func DiffGenerations(aID, bID int) (*GenerationDiff, error) {
	a, err := LoadGeneration(aID)
	if err != nil {
		return nil, err
	}
	b, err := LoadGeneration(bID)
	if err != nil {
		return nil, err
	}

	return &GenerationDiff{
		FlatpaksAdded:    diffSets(a.Flatpaks, b.Flatpaks),
		FlatpaksRemoved:  diffSets(b.Flatpaks, a.Flatpaks),
		ContainersAdded:  diffSets(a.Containers, b.Containers),
		ContainersRemoved: diffSets(b.Containers, a.Containers),
		DevboxesAdded:    diffSets(a.Devboxes, b.Devboxes),
		DevboxesRemoved:  diffSets(b.Devboxes, a.Devboxes),
		ToolsAdded:       diffSets(a.Tools, b.Tools),
		ToolsRemoved:     diffSets(b.Tools, a.Tools),
	}, nil
}

// GenerationDiff holds the differences between two generations.
type GenerationDiff struct {
	FlatpaksAdded     []string `json:"flatpaks_added"`
	FlatpaksRemoved   []string `json:"flatpaks_removed"`
	ContainersAdded   []string `json:"containers_added"`
	ContainersRemoved []string `json:"containers_removed"`
	DevboxesAdded     []string `json:"devboxes_added"`
	DevboxesRemoved   []string `json:"devboxes_removed"`
	ToolsAdded        []string `json:"tools_added"`
	ToolsRemoved      []string `json:"tools_removed"`
}

// diffSets returns elements in b that are not in a (sorted).
func diffSets(a, b []string) []string {
	set := make(map[string]bool, len(a))
	for _, v := range a {
		set[v] = true
	}
	var diff []string
	for _, v := range b {
		if !set[v] {
			diff = append(diff, v)
		}
	}
	sort.Strings(diff)
	return diff
}

func orEmpty(s []string) []string {
	if s == nil {
		return []string{}
	}
	return s
}
