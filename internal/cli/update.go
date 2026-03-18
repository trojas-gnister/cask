package cli

import (
	"context"
	"fmt"

	"github.com/iskry/cask/internal/executor"
	"github.com/spf13/cobra"
)

func newUpdateCmd() *cobra.Command {
	var (
		flagFlatpak    bool
		flagContainers bool
		flagAll        bool
	)

	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update packages and containers",
		RunE: func(cmd *cobra.Command, args []string) error {
			exec := executor.NewSystemExecutor(flagDryRun)
			ctx := context.Background()

			if !flagFlatpak && !flagContainers {
				flagAll = true
			}

			if flagAll || flagFlatpak {
				fmt.Println("Updating Flatpak packages...")
				r, err := exec.ExecuteSudo(ctx, []string{"flatpak", "update", "-y", "--noninteractive"})
				if err != nil {
					return err
				}
				if !r.Success {
					fmt.Printf("Flatpak update failed: %s\n", r.Stderr)
				} else {
					fmt.Println("Flatpak packages updated")
				}
			}

			if flagAll || flagContainers {
				fmt.Println("Pulling container image updates...")
				r, err := exec.Execute(ctx, []string{"podman", "auto-update"})
				if err != nil {
					return err
				}
				if !r.Success {
					fmt.Printf("Container update failed: %s\n", r.Stderr)
				} else {
					fmt.Println("Container images updated")
				}
			}

			return nil
		},
	}

	cmd.Flags().BoolVar(&flagFlatpak, "flatpak", false, "update Flatpak packages only")
	cmd.Flags().BoolVar(&flagContainers, "containers", false, "update container images only")
	cmd.Flags().BoolVar(&flagAll, "all", false, "update everything")

	return cmd
}
