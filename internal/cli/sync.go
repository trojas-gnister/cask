package cli

import (
	"context"
	"fmt"

	"github.com/iskry/cask/internal/config"
	"github.com/iskry/cask/internal/executor"
	csync "github.com/iskry/cask/internal/sync"
	"github.com/spf13/cobra"
)

func newSyncCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "sync",
		Short: "Bidirectional sync between host and config",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, _, err := config.FindAndLoadConfig(flagConfigPath)
			if err != nil {
				return err
			}

			exec := executor.NewSystemExecutor(flagDryRun)
			ctx := context.Background()
			opts := &csync.SyncOptions{Yes: flagYes, No: flagNo, Verbose: flagVerbose}

			var managers []csync.ResourceSync
			if cfg.Podman != nil {
				managers = append(managers, &csync.ContainerSyncManager{Config: cfg.Podman})
			}
			if cfg.Devbox != nil {
				managers = append(managers, &csync.DevboxSyncManager{Config: cfg.Devbox})
			}
			if cfg.Tools != nil {
				managers = append(managers, &csync.ToolsSyncManager{Config: cfg.Tools})
			}

			for _, mgr := range managers {
				fmt.Printf("Syncing %s...\n", mgr.ResourceType())
				result, err := csync.SyncResources(ctx, exec, mgr, opts)
				if err != nil {
					return fmt.Errorf("sync %s: %w", mgr.ResourceType(), err)
				}
				s := result.Stats
				fmt.Printf("  applied=%d updated=%d removed=%d\n", s.Applied, s.Updated, s.Removed)
				for _, e := range result.Errors {
					fmt.Printf("  error: %s\n", e)
				}
			}

			// Flatpak overrides (bespoke sync)
			if cfg.Flatpak != nil && cfg.Flatpak.ManageOverrides {
				fmt.Println("Syncing flatpak overrides...")
				fs := &csync.FlatpakOverrideSync{Config: cfg.Flatpak}
				if err := fs.Sync(ctx, exec); err != nil {
					return fmt.Errorf("flatpak overrides: %w", err)
				}
			}

			fmt.Println("Sync complete")
			return nil
		},
	}
}
