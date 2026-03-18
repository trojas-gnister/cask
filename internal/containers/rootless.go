package containers

import (
	"context"
	"fmt"

	"github.com/iskry/cask/internal/config"
	"github.com/iskry/cask/internal/executor"
)

// SetupPodmanRootless configures sysctl for rootless Podman.
func SetupPodmanRootless(ctx context.Context, exec executor.Executor, cfg *config.PodmanRootlessConfig) error {
	if cfg == nil || !cfg.Enabled {
		return nil
	}

	maxNS := cfg.MaxUserNamespaces
	if maxNS == 0 {
		maxNS = 65536
	}

	content := fmt.Sprintf("kernel.unprivileged_userns_clone = 1\nuser.max_user_namespaces = %d\n", maxNS)
	r, err := exec.WriteFile("/etc/sysctl.d/99-userns.conf", content, true)
	if err != nil {
		return err
	}
	if !r.Success {
		return fmt.Errorf("writing sysctl config: %s", r.Stderr)
	}

	// Apply immediately
	r, err = exec.ExecuteSudo(ctx, []string{"sysctl", "--system"})
	if err != nil {
		return err
	}
	if !r.Success {
		return fmt.Errorf("applying sysctl: %s", r.Stderr)
	}
	return nil
}
