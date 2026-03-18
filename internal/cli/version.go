package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

// Version is set at build time via -ldflags.
var Version = "dev"

func newVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print cask version",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("cask %s\n", Version)
		},
	}
}
