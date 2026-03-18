package cli

import (
	"fmt"

	"github.com/iskry/cask/internal/state"
	"github.com/spf13/cobra"
)

func newStateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "state",
		Short: "Manage state and generations",
	}
	cmd.AddCommand(newStateGCCmd())
	cmd.AddCommand(newStateGenerationsCmd())
	cmd.AddCommand(newStateDiffGenerationsCmd())
	cmd.AddCommand(newStateRollbackCmd())
	return cmd
}

func newStateGCCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "gc",
		Short: "Garbage collect old state",
		RunE: func(cmd *cobra.Command, args []string) error {
			mgr := state.NewManager()
			s := mgr.State()
			fmt.Printf("State has %d sections, %d run-once hashes\n",
				len(s.Sections), len(s.RunOnceHashes))
			return nil
		},
	}
}

func newStateGenerationsCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "generations",
		Short: "List all generations",
		RunE: func(cmd *cobra.Command, args []string) error {
			gens, err := state.ListGenerations()
			if err != nil {
				return err
			}
			if len(gens) == 0 {
				fmt.Println("No generations found")
				return nil
			}
			for _, gen := range gens {
				fmt.Printf("gen-%04d  %s  hash=%s\n", gen.ID, gen.Timestamp, gen.ConfigHash[:12])
			}
			return nil
		},
	}
}

func newStateDiffGenerationsCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "diff-generations <id-a> <id-b>",
		Short: "Compare two generations",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			var aID, bID int
			fmt.Sscanf(args[0], "%d", &aID)
			fmt.Sscanf(args[1], "%d", &bID)

			diff, err := state.DiffGenerations(aID, bID)
			if err != nil {
				return err
			}
			printDiffList("Flatpaks added", diff.FlatpaksAdded)
			printDiffList("Flatpaks removed", diff.FlatpaksRemoved)
			printDiffList("Containers added", diff.ContainersAdded)
			printDiffList("Containers removed", diff.ContainersRemoved)
			printDiffList("Devboxes added", diff.DevboxesAdded)
			printDiffList("Devboxes removed", diff.DevboxesRemoved)
			printDiffList("Tools added", diff.ToolsAdded)
			printDiffList("Tools removed", diff.ToolsRemoved)
			return nil
		},
	}
}

func newStateRollbackCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "rollback <generation-id>",
		Short: "Show rollback target (preview only)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var genID int
			fmt.Sscanf(args[0], "%d", &genID)

			gen, err := state.LoadGeneration(genID)
			if err != nil {
				return err
			}
			fmt.Printf("Generation %d (from %s):\n", gen.ID, gen.Timestamp)
			fmt.Printf("  Config hash: %s\n", gen.ConfigHash)
			fmt.Printf("  Flatpaks: %d\n", len(gen.Flatpaks))
			fmt.Printf("  Containers: %d\n", len(gen.Containers))
			fmt.Printf("  Devboxes: %d\n", len(gen.Devboxes))
			fmt.Printf("  Tools: %d\n", len(gen.Tools))
			return nil
		},
	}
}

func printDiffList(label string, items []string) {
	if len(items) == 0 {
		return
	}
	fmt.Printf("%s:\n", label)
	for _, item := range items {
		fmt.Printf("  %s\n", item)
	}
}
