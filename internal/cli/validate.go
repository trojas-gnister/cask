package cli

import (
	"fmt"

	"github.com/iskry/cask/internal/config"
	"github.com/spf13/cobra"
)

func newValidateCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "validate",
		Short: "Validate configuration file",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, path, err := config.FindAndLoadConfig(flagConfigPath)
			if err != nil {
				return err
			}

			result := config.ValidateConfig(cfg)

			for _, w := range result.Warnings {
				fmt.Printf("warning: %s: %s\n", w.Field, w.Message)
			}
			for _, e := range result.Errors {
				fmt.Printf("error: %s\n", e)
			}

			if result.IsValid() {
				fmt.Printf("Config %s is valid\n", path)
				return nil
			}
			return fmt.Errorf("config has %d error(s)", len(result.Errors))
		},
	}
}
