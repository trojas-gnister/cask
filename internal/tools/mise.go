// Package tools handles mise tool version management.
package tools

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/iskry/cask/internal/config"
	"github.com/iskry/cask/internal/executor"
)

// SetupMiseTools installs tools via mise, optionally setting global versions.
func SetupMiseTools(ctx context.Context, exec executor.Executor, cfg *config.ToolsConfig) error {
	if cfg == nil || len(cfg.Tools) == 0 {
		return nil
	}

	for _, tool := range cfg.Tools {
		spec := fmt.Sprintf("%s@%s", tool.Name, tool.Version)
		r, err := exec.Execute(ctx, []string{"mise", "install", spec})
		if err != nil {
			return fmt.Errorf("installing %s: %w", spec, err)
		}
		if !r.Success {
			return fmt.Errorf("installing %s: %s", spec, strings.TrimSpace(r.Stderr))
		}

		if tool.GlobalInstall {
			r, err = exec.Execute(ctx, []string{"mise", "use", "-g", spec})
			if err != nil {
				return fmt.Errorf("setting global %s: %w", spec, err)
			}
			if !r.Success {
				return fmt.Errorf("setting global %s: %s", spec, strings.TrimSpace(r.Stderr))
			}
		}
	}

	if cfg.ShellIntegration {
		log.Println("Mise shell integration: add 'eval \"$(mise activate bash)\"' (or zsh/fish) to your shell config")
	}

	return nil
}
