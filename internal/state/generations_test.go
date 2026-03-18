package state

import (
	"testing"
)

func TestCreateAndLoadGeneration(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())

	gen, err := CreateGeneration("hash123",
		[]string{"org.mozilla.Firefox"},
		[]string{"postgres"},
		[]string{"dev"},
		[]string{"node@20"},
	)
	if err != nil {
		t.Fatalf("create failed: %v", err)
	}
	if gen.ID != 1 {
		t.Errorf("first generation should be ID 1, got %d", gen.ID)
	}
	if gen.ConfigHash != "hash123" {
		t.Errorf("expected hash 'hash123', got '%s'", gen.ConfigHash)
	}
	if gen.Timestamp == "" {
		t.Error("timestamp should be set")
	}

	loaded, err := LoadGeneration(1)
	if err != nil {
		t.Fatalf("load failed: %v", err)
	}
	if loaded.ConfigHash != "hash123" {
		t.Error("loaded generation should match")
	}
	if len(loaded.Flatpaks) != 1 || loaded.Flatpaks[0] != "org.mozilla.Firefox" {
		t.Error("flatpaks should be preserved")
	}
}

func TestListGenerations(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())

	CreateGeneration("h1", nil, nil, nil, nil)
	CreateGeneration("h2", nil, nil, nil, nil)
	CreateGeneration("h3", nil, nil, nil, nil)

	gens, err := ListGenerations()
	if err != nil {
		t.Fatalf("list failed: %v", err)
	}
	if len(gens) != 3 {
		t.Errorf("expected 3 generations, got %d", len(gens))
	}
}

func TestGetCurrentGeneration(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())

	CreateGeneration("h1", nil, nil, nil, nil)
	CreateGeneration("h2", nil, nil, nil, nil)

	gen, err := GetCurrentGeneration()
	if err != nil {
		t.Fatalf("get current failed: %v", err)
	}
	if gen.ConfigHash != "h2" {
		t.Error("current should be the most recent generation")
	}
}

func TestGenerationAutoPrune(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())

	// Create more than maxGenerations
	for i := 0; i < 15; i++ {
		_, err := CreateGeneration("hash", nil, nil, nil, nil)
		if err != nil {
			t.Fatalf("create failed at %d: %v", i, err)
		}
	}

	gens, _ := ListGenerations()
	if len(gens) > maxGenerations {
		t.Errorf("should prune to %d, got %d", maxGenerations, len(gens))
	}

	// First generation should have been pruned
	_, err := LoadGeneration(1)
	if err == nil {
		t.Error("generation 1 should have been pruned")
	}
}

func TestDiffGenerations(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())

	CreateGeneration("h1",
		[]string{"org.mozilla.Firefox"},
		[]string{"postgres"},
		nil, nil,
	)
	CreateGeneration("h2",
		[]string{"org.mozilla.Firefox", "org.gimp.GIMP"},
		nil,
		[]string{"dev"},
		nil,
	)

	diff, err := DiffGenerations(1, 2)
	if err != nil {
		t.Fatalf("diff failed: %v", err)
	}
	if len(diff.FlatpaksAdded) != 1 || diff.FlatpaksAdded[0] != "org.gimp.GIMP" {
		t.Errorf("expected GIMP added, got %v", diff.FlatpaksAdded)
	}
	if len(diff.ContainersRemoved) != 1 || diff.ContainersRemoved[0] != "postgres" {
		t.Errorf("expected postgres removed, got %v", diff.ContainersRemoved)
	}
	if len(diff.DevboxesAdded) != 1 || diff.DevboxesAdded[0] != "dev" {
		t.Errorf("expected dev added, got %v", diff.DevboxesAdded)
	}
}

func TestDiffGenerationsNotFound(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())

	_, err := DiffGenerations(999, 1000)
	if err == nil {
		t.Error("diff of nonexistent generations should fail")
	}
}

func TestGetCurrentGenerationEmpty(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())

	_, err := GetCurrentGeneration()
	if err == nil {
		t.Error("should fail when no generations exist")
	}
}

func TestGenerationNilSlicesBecomEmpty(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())

	gen, _ := CreateGeneration("h", nil, nil, nil, nil)
	if gen.Flatpaks == nil {
		t.Error("nil slice should become empty slice")
	}
	if gen.Containers == nil {
		t.Error("nil slice should become empty slice")
	}
}
