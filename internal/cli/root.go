// Package cli defines all cobra commands for the cask CLI.
package cli

import (
	"github.com/spf13/cobra"
)

var (
	flagVerbose    bool
	flagYes        bool
	flagNo         bool
	flagDryRun     bool
	flagConfigPath string
)

func newRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cask",
		Short: "Distro-independent package and container management",
		Long:  "cask manages Flatpak packages, Podman containers, Distrobox environments, and mise tool versions with bidirectional sync and state tracking.",
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	cmd.PersistentFlags().BoolVarP(&flagVerbose, "verbose", "v", false, "enable verbose output")
	cmd.PersistentFlags().BoolVarP(&flagYes, "yes", "y", false, "keep undeclared resources during sync")
	cmd.PersistentFlags().BoolVarP(&flagNo, "no", "n", false, "remove undeclared resources during sync")
	cmd.PersistentFlags().BoolVar(&flagDryRun, "dry-run", false, "preview changes without applying")
	cmd.PersistentFlags().StringVarP(&flagConfigPath, "config", "c", "", "path to config file")

	cmd.AddCommand(newVersionCmd())
	cmd.AddCommand(newAddCmd())
	cmd.AddCommand(newRemoveCmd())
	cmd.AddCommand(newListCmd())
	cmd.AddCommand(newUpdateCmd())
	cmd.AddCommand(newSyncCmd())
	cmd.AddCommand(newValidateCmd())
	cmd.AddCommand(newDiffCmd())
	cmd.AddCommand(newDevboxCmd())
	cmd.AddCommand(newLockCmd())
	cmd.AddCommand(newStateCmd())

	return cmd
}

// Execute runs the root command.
func Execute() error {
	return newRootCmd().Execute()
}
