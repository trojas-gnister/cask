package cli

import (
	"encoding/json"
	"fmt"

	"github.com/iskry/cask/internal/config"
	"github.com/iskry/cask/internal/state"
	"github.com/spf13/cobra"
)

func newDiffCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "diff",
		Short: "Show config sections that have changed since last apply",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, _, err := config.FindAndLoadConfig(flagConfigPath)
			if err != nil {
				return err
			}

			mgr := state.NewManager()

			// Build a map of section names to their data
			sections := make(map[string]any)
			cfgJSON, _ := json.Marshal(cfg)
			var cfgMap map[string]any
			json.Unmarshal(cfgJSON, &cfgMap)

			for name, data := range cfgMap {
				if data != nil {
					sections[name] = data
				}
			}

			changed, err := mgr.GetChangedSections(sections)
			if err != nil {
				return err
			}

			if len(changed) == 0 {
				fmt.Println("No changes detected")
				return nil
			}

			fmt.Printf("Changed sections (%d):\n", len(changed))
			for _, name := range changed {
				fmt.Printf("  - %s\n", name)
			}
			return nil
		},
	}
}
