package flatpak

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/iskry/cask/internal/executor"
)

// DiscoverHostOverrides returns the set of app IDs that have local overrides.
func DiscoverHostOverrides() ([]string, error) {
	dir := filepath.Join(homeDir(), ".local", "share", "flatpak", "overrides")
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

// ApplyOverrides resets and re-applies per-app Flatpak overrides from config.
func ApplyOverrides(ctx context.Context, exec executor.Executor, overrides map[string]map[string]any) error {
	for appID, settings := range overrides {
		// Reset existing overrides for this app
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
			r, err := exec.Execute(ctx, []string{"flatpak", "override", flag, appID})
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

func homeDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return os.Getenv("HOME")
	}
	return home
}
