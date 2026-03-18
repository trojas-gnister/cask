package flatpak

import (
	"context"
	"fmt"
	"strings"

	"github.com/iskry/cask/internal/config"
	"github.com/iskry/cask/internal/executor"
)

// SetupHardening applies global Flatpak permission restrictions.
func SetupHardening(ctx context.Context, exec executor.Executor, cfg *config.FlatpakHardeningConfig) error {
	if cfg == nil || !cfg.Enabled {
		return nil
	}

	var overrides []string

	if cfg.RestrictFilesystem {
		overrides = append(overrides, "--nofilesystem=host")
	}
	if cfg.NetworkPolicy == config.NetworkDeny {
		overrides = append(overrides, "--unshare=network")
	}
	for _, denial := range cfg.DefaultDenials {
		overrides = append(overrides, fmt.Sprintf("--%s", denial))
	}

	if len(overrides) == 0 {
		return nil
	}

	cmd := append([]string{"flatpak", "override"}, overrides...)
	r, err := exec.ExecuteSudo(ctx, cmd)
	if err != nil {
		return err
	}
	if !r.Success {
		return fmt.Errorf("flatpak hardening failed: %s", strings.TrimSpace(r.Stderr))
	}
	return nil
}
