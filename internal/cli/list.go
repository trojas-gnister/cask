package cli

import (
	"context"
	"fmt"

	"github.com/iskry/cask/internal/containers"
	"github.com/iskry/cask/internal/executor"
	"github.com/iskry/cask/internal/flatpak"
	"github.com/spf13/cobra"
)

func newListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List managed resources",
	}
	cmd.AddCommand(newListFlatpakCmd())
	cmd.AddCommand(newListContainersCmd())
	cmd.AddCommand(newListDevboxesCmd())
	cmd.AddCommand(newListAllCmd())
	return cmd
}

func newListFlatpakCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "flatpak",
		Short: "List installed Flatpak packages",
		RunE: func(cmd *cobra.Command, args []string) error {
			exec := executor.NewSystemExecutor(flagDryRun)
			apps, err := flatpak.List(context.Background(), exec)
			if err != nil {
				return err
			}
			if len(apps) == 0 {
				fmt.Println("No Flatpak packages installed")
				return nil
			}
			for _, app := range apps {
				fmt.Println(app)
			}
			return nil
		},
	}
}

func newListContainersCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "containers",
		Short: "List Podman containers",
		RunE: func(cmd *cobra.Command, args []string) error {
			exec := executor.NewSystemExecutor(flagDryRun)
			ctrs, err := containers.ListContainers(context.Background(), exec)
			if err != nil {
				return err
			}
			if len(ctrs) == 0 {
				fmt.Println("No containers found")
				return nil
			}
			for _, c := range ctrs {
				fmt.Printf("%-20s %-40s %s\n", c["name"], c["image"], c["status"])
			}
			return nil
		},
	}
}

func newListDevboxesCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "devboxes",
		Short: "List Distrobox instances",
		RunE: func(cmd *cobra.Command, args []string) error {
			exec := executor.NewSystemExecutor(flagDryRun)
			instances, err := containers.ListDevboxes(context.Background(), exec)
			if err != nil {
				return err
			}
			if len(instances) == 0 {
				fmt.Println("No devbox instances found")
				return nil
			}
			for _, inst := range instances {
				fmt.Printf("%-20s %s\n", inst["name"], inst["status"])
			}
			return nil
		},
	}
}

func newListAllCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "all",
		Short: "List all managed resources",
		RunE: func(cmd *cobra.Command, args []string) error {
			exec := executor.NewSystemExecutor(flagDryRun)
			ctx := context.Background()

			fmt.Println("=== Flatpak Packages ===")
			apps, _ := flatpak.List(ctx, exec)
			for _, app := range apps {
				fmt.Printf("  %s\n", app)
			}
			if len(apps) == 0 {
				fmt.Println("  (none)")
			}

			fmt.Println("\n=== Containers ===")
			ctrs, _ := containers.ListContainers(ctx, exec)
			for _, c := range ctrs {
				fmt.Printf("  %-20s %s\n", c["name"], c["image"])
			}
			if len(ctrs) == 0 {
				fmt.Println("  (none)")
			}

			fmt.Println("\n=== Devbox Instances ===")
			instances, _ := containers.ListDevboxes(ctx, exec)
			for _, inst := range instances {
				fmt.Printf("  %-20s %s\n", inst["name"], inst["status"])
			}
			if len(instances) == 0 {
				fmt.Println("  (none)")
			}

			return nil
		},
	}
}
