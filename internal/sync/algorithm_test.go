package sync

import (
	"context"
	"testing"

	"github.com/iskry/cask/internal/executor"
)

// mockSync is a test implementation of ResourceSync.
type mockSync struct {
	host   []Resource
	config []Resource
}

func (m *mockSync) ResourceType() string { return "test" }

func (m *mockSync) GetHostResources(_ context.Context, _ executor.Executor) ([]Resource, error) {
	return m.host, nil
}

func (m *mockSync) GetConfigResources() []Resource {
	return m.config
}

func (m *mockSync) ResourceID(r Resource) string {
	name, _ := r["name"].(string)
	return name
}

func (m *mockSync) Apply(_ context.Context, _ executor.Executor, _ Resource) bool {
	return true
}

func (m *mockSync) Remove(_ context.Context, _ executor.Executor, _ string) bool {
	return true
}

func (m *mockSync) NeedsUpdate(host, config Resource) bool {
	hv, _ := host["version"].(string)
	cv, _ := config["version"].(string)
	return hv != cv
}

func TestSyncApplyNew(t *testing.T) {
	mgr := &mockSync{
		host: nil,
		config: []Resource{
			{"name": "vim", "version": "1"},
			{"name": "git", "version": "2"},
		},
	}

	result, err := SyncResources(context.Background(), executor.NewMockExecutor(), mgr, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Stats.Applied != 2 {
		t.Errorf("expected 2 applied, got %d", result.Stats.Applied)
	}
}

func TestSyncNoChanges(t *testing.T) {
	resources := []Resource{
		{"name": "vim", "version": "1"},
	}
	mgr := &mockSync{host: resources, config: resources}

	result, err := SyncResources(context.Background(), executor.NewMockExecutor(), mgr, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Stats.Applied != 0 && result.Stats.Updated != 0 && result.Stats.Removed != 0 {
		t.Error("no changes expected")
	}
}

func TestSyncUpdate(t *testing.T) {
	mgr := &mockSync{
		host:   []Resource{{"name": "vim", "version": "1"}},
		config: []Resource{{"name": "vim", "version": "2"}},
	}

	result, err := SyncResources(context.Background(), executor.NewMockExecutor(), mgr, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Stats.Updated != 1 {
		t.Errorf("expected 1 updated, got %d", result.Stats.Updated)
	}
}

func TestSyncRemoveUndeclared(t *testing.T) {
	mgr := &mockSync{
		host: []Resource{
			{"name": "vim", "version": "1"},
			{"name": "emacs", "version": "1"},
		},
		config: []Resource{
			{"name": "vim", "version": "1"},
		},
	}

	// With --no: remove undeclared
	result, err := SyncResources(context.Background(), executor.NewMockExecutor(), mgr, &SyncOptions{No: true})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Stats.Removed != 1 {
		t.Errorf("expected 1 removed, got %d", result.Stats.Removed)
	}
}

func TestSyncKeepUndeclaredDefault(t *testing.T) {
	mgr := &mockSync{
		host:   []Resource{{"name": "emacs", "version": "1"}},
		config: nil,
	}

	// Default: keep undeclared
	result, err := SyncResources(context.Background(), executor.NewMockExecutor(), mgr, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Stats.Removed != 0 {
		t.Errorf("expected 0 removed (keep by default), got %d", result.Stats.Removed)
	}
}

func TestSyncKeepUndeclaredWithYes(t *testing.T) {
	mgr := &mockSync{
		host:   []Resource{{"name": "emacs", "version": "1"}},
		config: nil,
	}

	result, err := SyncResources(context.Background(), executor.NewMockExecutor(), mgr, &SyncOptions{Yes: true})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Stats.Removed != 0 {
		t.Errorf("--yes should keep undeclared, got %d removed", result.Stats.Removed)
	}
}

func TestSyncMixedActions(t *testing.T) {
	mgr := &mockSync{
		host: []Resource{
			{"name": "vim", "version": "1"},   // unchanged
			{"name": "git", "version": "1"},   // needs update
			{"name": "emacs", "version": "1"}, // undeclared
		},
		config: []Resource{
			{"name": "vim", "version": "1"}, // unchanged
			{"name": "git", "version": "2"}, // update
			{"name": "go", "version": "1"},  // new
		},
	}

	result, err := SyncResources(context.Background(), executor.NewMockExecutor(), mgr, &SyncOptions{No: true})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Stats.Applied != 1 {
		t.Errorf("expected 1 applied (go), got %d", result.Stats.Applied)
	}
	if result.Stats.Updated != 1 {
		t.Errorf("expected 1 updated (git), got %d", result.Stats.Updated)
	}
	if result.Stats.Removed != 1 {
		t.Errorf("expected 1 removed (emacs), got %d", result.Stats.Removed)
	}
	if !result.Success {
		t.Error("should succeed")
	}
}

func TestSyncEmpty(t *testing.T) {
	mgr := &mockSync{}
	result, err := SyncResources(context.Background(), executor.NewMockExecutor(), mgr, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.Success {
		t.Error("empty sync should succeed")
	}
}
