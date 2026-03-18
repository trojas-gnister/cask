package sync

import (
	"context"
	"fmt"
	"sort"

	"github.com/iskry/cask/internal/executor"
)

// SyncResources synchronizes resources between host and config using an 8-step process:
//  1. Gather host and config state
//  2. Build lookup maps by resource ID
//  3. Categorize: to_apply (config-only), common, undeclared (host-only)
//  4. Identify resources needing update within common set
//  5. Handle undeclared resources based on options
//  6. Execute apply/update actions
//  7. Execute remove actions
//  8. Return stats
func SyncResources(ctx context.Context, exec executor.Executor, manager ResourceSync, opts *SyncOptions) (*SyncResult, error) {
	if opts == nil {
		opts = &SyncOptions{}
	}

	var errors []string
	stats := SyncStats{}

	// Step 1: Gather current state
	hostResources, err := manager.GetHostResources(ctx, exec)
	if err != nil {
		return nil, fmt.Errorf("getting host %s: %w", manager.ResourceType(), err)
	}
	configResources := manager.GetConfigResources()

	// Step 2: Build lookup maps
	hostByID := make(map[string]Resource, len(hostResources))
	for _, r := range hostResources {
		hostByID[manager.ResourceID(r)] = r
	}
	configByID := make(map[string]Resource, len(configResources))
	for _, r := range configResources {
		configByID[manager.ResourceID(r)] = r
	}

	// Step 3: Categorize
	hostIDs := mapKeys(hostByID)
	configIDs := mapKeys(configByID)

	toApply := setDiff(configIDs, hostIDs)
	common := setIntersect(configIDs, hostIDs)
	undeclared := setDiff(hostIDs, configIDs)

	// Step 4: Find resources needing update
	var toUpdate []string
	for _, id := range common {
		if manager.NeedsUpdate(hostByID[id], configByID[id]) {
			toUpdate = append(toUpdate, id)
		}
	}
	sort.Strings(toUpdate)

	// Step 5: Handle undeclared
	var toRemove []string
	if len(undeclared) > 0 && opts.No {
		toRemove = undeclared
	}
	// If opts.Yes: keep all (default behavior)
	// If neither: keep all (safe default)
	sort.Strings(toRemove)

	// Step 6: Apply new resources
	sort.Strings(toApply)
	for _, id := range toApply {
		if manager.Apply(ctx, exec, configByID[id]) {
			stats.Applied++
		} else {
			errors = append(errors, fmt.Sprintf("Failed to apply %s: %s", manager.ResourceType(), id))
		}
	}

	// Update changed resources
	for _, id := range toUpdate {
		if manager.Apply(ctx, exec, configByID[id]) {
			stats.Updated++
		} else {
			errors = append(errors, fmt.Sprintf("Failed to update %s: %s", manager.ResourceType(), id))
		}
	}

	// Step 7: Remove undeclared
	for _, id := range toRemove {
		if manager.Remove(ctx, exec, id) {
			stats.Removed++
		} else {
			errors = append(errors, fmt.Sprintf("Failed to remove %s: %s", manager.ResourceType(), id))
		}
	}

	return &SyncResult{
		Success: len(errors) == 0,
		Stats:   stats,
		Errors:  errors,
	}, nil
}

func mapKeys(m map[string]Resource) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func setDiff(a, b []string) []string {
	set := make(map[string]bool, len(b))
	for _, v := range b {
		set[v] = true
	}
	var result []string
	for _, v := range a {
		if !set[v] {
			result = append(result, v)
		}
	}
	return result
}

func setIntersect(a, b []string) []string {
	set := make(map[string]bool, len(b))
	for _, v := range b {
		set[v] = true
	}
	var result []string
	for _, v := range a {
		if set[v] {
			result = append(result, v)
		}
	}
	return result
}
