package cli

import (
	"context"
	"fmt"
	"os"

	"github.com/iskry/cask/internal/config"
	devboxpkg "github.com/iskry/cask/internal/devbox"
	"github.com/iskry/cask/internal/executor"
	"github.com/spf13/cobra"
)

func newDevboxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "devbox",
		Short: "Manage Distrobox development environments",
	}
	cmd.AddCommand(newDevboxEnterCmd())
	cmd.AddCommand(newDevboxRunCmd())
	cmd.AddCommand(newDevboxBoxesCmd())
	cmd.AddCommand(newDevboxHookCmd())
	cmd.AddCommand(newDevboxCheckCmd())
	return cmd
}

func newDevboxEnterCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "enter <name>",
		Short: "Enter a Distrobox instance",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			exec := executor.NewSystemExecutor(flagDryRun)
			r, err := exec.Execute(context.Background(), []string{"distrobox", "enter", args[0]})
			if err != nil {
				return err
			}
			if !r.Success {
				return fmt.Errorf("enter failed: %s", r.Stderr)
			}
			return nil
		},
	}
}

func newDevboxRunCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "run <name> -- <command>",
		Short: "Run a command in a Distrobox instance",
		Args:  cobra.MinimumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			exec := executor.NewSystemExecutor(flagDryRun)
			execArgs := append([]string{"distrobox", "enter", args[0], "--"}, args[1:]...)
			r, err := exec.Execute(context.Background(), execArgs)
			if err != nil {
				return err
			}
			fmt.Print(r.Stdout)
			if !r.Success {
				return fmt.Errorf("command failed: %s", r.Stderr)
			}
			return nil
		},
	}
}

func newDevboxBoxesCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "boxes",
		Short: "List Distrobox instances",
		RunE: func(cmd *cobra.Command, args []string) error {
			exec := executor.NewSystemExecutor(flagDryRun)
			r, err := exec.Execute(context.Background(), []string{"distrobox", "list"})
			if err != nil {
				return err
			}
			fmt.Print(r.Stdout)
			return nil
		},
	}
}

func newDevboxHookCmd() *cobra.Command {
	var shell string
	cmd := &cobra.Command{
		Use:   "hook",
		Short: "Generate shell hook for auto-entering devbox on cd",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, _, err := config.FindAndLoadConfig(flagConfigPath)
			if err != nil {
				return err
			}
			if cfg.Devbox == nil || len(cfg.Devbox.Projects) == 0 {
				return fmt.Errorf("no devbox projects configured")
			}
			fmt.Print(devboxpkg.GenerateHook(shell, cfg.Devbox.Projects))
			return nil
		},
	}
	cmd.Flags().StringVar(&shell, "shell", "bash", "shell type (bash/zsh/fish)")
	return cmd
}

func newDevboxCheckCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "check",
		Short: "Check if current directory matches a devbox project",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, _, err := config.FindAndLoadConfig(flagConfigPath)
			if err != nil {
				return err
			}
			if cfg.Devbox == nil || len(cfg.Devbox.Projects) == 0 {
				fmt.Println("No devbox projects configured")
				return nil
			}

			cwd, err := os.Getwd()
			if err != nil {
				return err
			}
			p := devboxpkg.MatchProject(cwd, cfg.Devbox.Projects)
			if p == nil {
				fmt.Println("No matching project for current directory")
				return nil
			}
			fmt.Printf("Matched project: %s -> %s\n", p.Path, p.BoxName)
			return nil
		},
	}
}
