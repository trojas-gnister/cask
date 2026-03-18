// Package sync implements bidirectional resource synchronization.
package sync

import (
	"context"

	"github.com/iskry/cask/internal/executor"
)

// SyncOptions controls sync behavior.
type SyncOptions struct {
	Yes     bool // Auto-keep all undeclared resources
	No      bool // Auto-remove all undeclared resources
	Verbose bool
}

// SyncStats holds statistics from a sync operation.
type SyncStats struct {
	Applied int
	Updated int
	Removed int
}

// SyncResult holds the full result of a sync operation.
type SyncResult struct {
	Success bool
	Stats   SyncStats
	Errors  []string
}

// ResourceSync is the interface for bidirectional resource synchronization.
// Implementations define how to discover, compare, apply, and remove
// a specific type of resource (packages, containers, tools, etc.).
type ResourceSync interface {
	// ResourceType returns a human-readable name (e.g., "containers").
	ResourceType() string

	// GetHostResources discovers resources currently on the host system.
	GetHostResources(ctx context.Context, exec executor.Executor) ([]Resource, error)

	// GetConfigResources returns resources declared in configuration.
	GetConfigResources() []Resource

	// ResourceID extracts a unique identifier from a resource.
	ResourceID(r Resource) string

	// Apply creates or updates a resource on the host. Returns success.
	Apply(ctx context.Context, exec executor.Executor, r Resource) bool

	// Remove removes a resource from the host. Returns success.
	Remove(ctx context.Context, exec executor.Executor, id string) bool

	// NeedsUpdate checks if a host resource differs from its config version.
	NeedsUpdate(host, config Resource) bool
}

// Resource is a generic map representing a syncable resource.
type Resource = map[string]any
