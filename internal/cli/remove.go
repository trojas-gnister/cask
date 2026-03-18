package cli

import (
	"context"
	"fmt"

	"github.com/iskry/cask/internal/config"
	"github.com/iskry/cask/internal/executor"
	"github.com/iskry/cask/internal/flatpak"
	"github.com/spf13/cobra"
)

func newRemoveCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "remove",
		Aliases: []string{"rm"},
		Short:   "Remove packages, containers, or devbox instances",
	}
	cmd.AddCommand(newRemoveFlatpakCmd())
	cmd.AddCommand(newRemoveContainerCmd())
	cmd.AddCommand(newRemoveDevboxCmd())
	return cmd
}

func newRemoveFlatpakCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "flatpak [packages...]",
		Short: "Remove Flatpak packages and update config",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			exec := executor.NewSystemExecutor(flagDryRun)
			ctx := context.Background()

			r, err := flatpak.Remove(ctx, exec, args)
			if err != nil {
				return err
			}
			if !r.Success {
				return fmt.Errorf("flatpak remove failed: %s", r.Stderr)
			}

			cfgPath := flagConfigPath
			if cfgPath == "" {
				cfgPath = config.MainConfigPath()
			}
			for _, pkg := range args {
				config.RemoveFromConfigList(cfgPath, "flatpak", "packages", pkg)
			}
			fmt.Printf("Removed %d Flatpak package(s)\n", len(args))
			return nil
		},
	}
}

func newRemoveContainerCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "container <name>",
		Short: "Stop and remove a Podman container",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			exec := executor.NewSystemExecutor(flagDryRun)
			ctx := context.Background()
			name := args[0]

			exec.Execute(ctx, []string{"podman", "stop", name})
			r, err := exec.Execute(ctx, []string{"podman", "rm", "-f", name})
			if err != nil {
				return err
			}
			if !r.Success {
				return fmt.Errorf("container removal failed: %s", r.Stderr)
			}
			fmt.Printf("Removed container %s\n", name)
			return nil
		},
	}
}

func newRemoveDevboxCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "devbox <name>",
		Short: "Remove a Distrobox instance",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			exec := executor.NewSystemExecutor(flagDryRun)
			ctx := context.Background()

			r, err := exec.Execute(ctx, []string{"distrobox", "rm", "--force", args[0]})
			if err != nil {
				return err
			}
			if !r.Success {
				return fmt.Errorf("devbox removal failed: %s", r.Stderr)
			}
			fmt.Printf("Removed devbox %s\n", args[0])
			return nil
		},
	}
}
