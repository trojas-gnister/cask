package cli

import (
	"context"
	"fmt"

	"github.com/iskry/cask/internal/config"
	"github.com/iskry/cask/internal/executor"
	"github.com/iskry/cask/internal/flatpak"
	"github.com/spf13/cobra"
)

func newAddCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add",
		Short: "Add packages, containers, or devbox instances",
	}
	cmd.AddCommand(newAddFlatpakCmd())
	cmd.AddCommand(newAddContainerCmd())
	cmd.AddCommand(newAddDevboxCmd())
	return cmd
}

func newAddFlatpakCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "flatpak [packages...]",
		Short: "Install Flatpak packages and add to config",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			exec := executor.NewSystemExecutor(flagDryRun)
			ctx := context.Background()

			r, err := flatpak.Install(ctx, exec, args)
			if err != nil {
				return err
			}
			if !r.Success {
				return fmt.Errorf("flatpak install failed: %s", r.Stderr)
			}

			// Update config
			cfgPath := flagConfigPath
			if cfgPath == "" {
				cfgPath = config.MainConfigPath()
			}
			for _, pkg := range args {
				if err := config.AddToConfigList(cfgPath, "flatpak", "packages", pkg); err != nil {
					return fmt.Errorf("updating config: %w", err)
				}
			}
			fmt.Printf("Added %d Flatpak package(s)\n", len(args))
			return nil
		},
	}
}

func newAddContainerCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "container <name> <image>",
		Short: "Create a Podman container and add to config",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			exec := executor.NewSystemExecutor(flagDryRun)
			ctx := context.Background()

			name, image := args[0], args[1]
			r, err := exec.Execute(ctx, []string{"podman", "create", "--name", name, image})
			if err != nil {
				return err
			}
			if !r.Success {
				return fmt.Errorf("container creation failed: %s", r.Stderr)
			}

			cfgPath := flagConfigPath
			if cfgPath == "" {
				cfgPath = config.MainConfigPath()
			}
			if err := config.UpdateConfigSection(cfgPath, "podman", map[string]any{
				"containers": []map[string]any{
					{"name": name, "image": image, "scope": "user"},
				},
			}); err != nil {
				return fmt.Errorf("updating config: %w", err)
			}
			fmt.Printf("Created container %s\n", name)
			return nil
		},
	}
}

func newAddDevboxCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "devbox <name> <image>",
		Short: "Create a Distrobox instance and add to config",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			exec := executor.NewSystemExecutor(flagDryRun)
			ctx := context.Background()

			name, image := args[0], args[1]
			r, err := exec.Execute(ctx, []string{
				"distrobox", "create", "--name", name, "--image", image, "--yes",
			})
			if err != nil {
				return err
			}
			if !r.Success {
				return fmt.Errorf("devbox creation failed: %s", r.Stderr)
			}

			fmt.Printf("Created devbox %s\n", name)
			return nil
		},
	}
}
