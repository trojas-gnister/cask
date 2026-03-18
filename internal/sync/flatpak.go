package sync

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/iskry/cask/internal/config"
	"github.com/iskry/cask/internal/executor"
)

// FlatpakOverrideSync handles per-app Flatpak permission override sync.
// This is a bespoke sync (not via ResourceSync) since overrides are INI-based.
type FlatpakOverrideSync struct {
	Config *config.FlatpakConfig
}

// Sync synchronizes Flatpak overrides between host and config.
func (s *FlatpakOverrideSync) Sync(ctx context.Context, exec executor.Executor) error {
	if s.Config == nil || !s.Config.ManageOverrides || len(s.Config.Overrides) == 0 {
		return nil
	}

	for appID, settings := range s.Config.Overrides {
		// Reset existing overrides
		r, err := exec.Execute(ctx, []string{"flatpak", "override", "--reset", appID})
		if err != nil {
			return fmt.Errorf("resetting overrides for %s: %w", appID, err)
		}
		if !r.Success {
			return fmt.Errorf("resetting overrides for %s: %s", appID, strings.TrimSpace(r.Stderr))
		}

		// Apply each override
		for key, val := range settings {
			flag := fmt.Sprintf("--%s=%v", key, val)
			r, err = exec.Execute(ctx, []string{"flatpak", "override", flag, appID})
			if err != nil {
				return fmt.Errorf("applying override %s for %s: %w", key, appID, err)
			}
			if !r.Success {
				return fmt.Errorf("applying override %s for %s: %s", key, appID, strings.TrimSpace(r.Stderr))
			}
		}
	}
	return nil
}

// DiscoverHostOverrides returns app IDs with local override files.
func (s *FlatpakOverrideSync) DiscoverHostOverrides() ([]string, error) {
	home, _ := os.UserHomeDir()
	dir := filepath.Join(home, ".local", "share", "flatpak", "overrides")
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var apps []string
	for _, e := range entries {
		if !e.IsDir() && e.Name() != "global" {
			apps = append(apps, e.Name())
		}
	}
	return apps, nil
}
