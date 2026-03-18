package cli

import (
	"context"
	"fmt"

	"github.com/iskry/cask/internal/executor"
	"github.com/iskry/cask/internal/state"
	"github.com/spf13/cobra"
)

func newLockCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "lock",
		Short: "Manage version lockfile",
	}
	cmd.AddCommand(newLockCreateCmd())
	cmd.AddCommand(newLockVerifyCmd())
	return cmd
}

func newLockCreateCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "create",
		Short: "Generate lockfile from current system state",
		RunE: func(cmd *cobra.Command, args []string) error {
			exec := executor.NewSystemExecutor(flagDryRun)
			ctx := context.Background()

			lf, err := state.GenerateLockfile(ctx, exec)
			if err != nil {
				return err
			}
			if err := state.SaveLockfile(lf); err != nil {
				return err
			}

			fmt.Printf("Lockfile created: %d flatpaks, %d containers, %d tools\n",
				len(lf.Flatpaks), len(lf.Containers), len(lf.Tools))
			return nil
		},
	}
}

func newLockVerifyCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "verify",
		Short: "Verify system matches lockfile",
		RunE: func(cmd *cobra.Command, args []string) error {
			exec := executor.NewSystemExecutor(flagDryRun)
			ctx := context.Background()

			mismatches, err := state.VerifyLockfile(ctx, exec)
			if err != nil {
				return err
			}
			if len(mismatches) == 0 {
				fmt.Println("System matches lockfile")
				return nil
			}

			fmt.Printf("Found %d mismatch(es):\n", len(mismatches))
			for _, m := range mismatches {
				fmt.Printf("  - %s\n", m)
			}
			return fmt.Errorf("lockfile verification failed")
		},
	}
}
