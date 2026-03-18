package containers

import (
	"context"
	"fmt"
	"strings"

	"github.com/iskry/cask/internal/config"
	"github.com/iskry/cask/internal/executor"
)

// DetectPackageManager returns the package manager name and install command
// based on the container image name.
func DetectPackageManager(image string) (string, []string) {
	lower := strings.ToLower(image)
	for _, d := range []string{"fedora", "centos", "rhel", "rocky", "alma"} {
		if strings.Contains(lower, d) {
			return "dnf", []string{"dnf", "install", "-y"}
		}
	}
	for _, d := range []string{"debian", "ubuntu", "mint"} {
		if strings.Contains(lower, d) {
			return "apt", []string{"apt-get", "install", "-y"}
		}
	}
	if strings.Contains(lower, "arch") {
		return "pacman", []string{"pacman", "-S", "--noconfirm"}
	}
	if strings.Contains(lower, "alpine") {
		return "apk", []string{"apk", "add"}
	}
	for _, d := range []string{"suse", "sles", "tumbleweed"} {
		if strings.Contains(lower, d) {
			return "zypper", []string{"zypper", "install", "-y"}
		}
	}
	return "dnf", []string{"dnf", "install", "-y"}
}

// BuildHomeFlags builds distrobox create flags from the home field.
func BuildHomeFlags(instance *config.DevboxInstance) []string {
	if instance.Home == "" || instance.Home == "host" {
		return nil
	}
	if instance.Home == "isolated" {
		return []string{
			"--no-entry", "--home",
			fmt.Sprintf("~/.local/share/distrobox/%s", instance.Name),
		}
	}
	return []string{"--home", instance.Home}
}

// SetupDevboxes creates Distrobox instances from config.
func SetupDevboxes(ctx context.Context, exec executor.Executor, cfg *config.DevboxConfig) error {
	if cfg == nil {
		return nil
	}

	for i := range cfg.Instances {
		inst := &cfg.Instances[i]

		cmd := []string{"distrobox", "create", "--name", inst.Name, "--image", inst.Image}
		cmd = append(cmd, BuildHomeFlags(inst)...)

		// Environment variables
		for key, value := range inst.Environment {
			cmd = append(cmd, "--additional-flags", fmt.Sprintf("--env %s=%s", key, value))
		}

		// Extra flags
		cmd = append(cmd, inst.Flags...)
		cmd = append(cmd, "--yes")

		r, err := exec.Execute(ctx, cmd)
		if err != nil {
			return fmt.Errorf("creating devbox %s: %w", inst.Name, err)
		}
		if !r.Success {
			return fmt.Errorf("creating devbox %s: %s", inst.Name, strings.TrimSpace(r.Stderr))
		}

		// Init hooks
		for _, hook := range inst.InitHooks {
			r, err = exec.Execute(ctx, []string{
				"distrobox", "enter", inst.Name, "--", "sh", "-c", hook,
			})
			if err != nil {
				return fmt.Errorf("init hook for %s: %w", inst.Name, err)
			}
		}

		// Install packages
		if len(inst.Packages) > 0 {
			_, installCmd := DetectPackageManager(inst.Image)
			pkgCmd := strings.Join(append(append([]string{"sudo"}, installCmd...), inst.Packages...), " ")
			exec.Execute(ctx, []string{
				"distrobox", "enter", inst.Name, "--", "sh", "-c", pkgCmd,
			})
		}

		// Post-create commands
		for _, postCmd := range inst.PostCreate {
			exec.Execute(ctx, []string{
				"distrobox", "enter", inst.Name, "--", "sh", "-c", postCmd,
			})
		}

		// Export apps
		for _, app := range inst.ExportApps {
			exec.Execute(ctx, []string{
				"distrobox", "enter", inst.Name, "--", "distrobox-export", "--app", app,
			})
		}
	}
	return nil
}

// ListDevboxes lists Distrobox instances.
func ListDevboxes(ctx context.Context, exec executor.Executor) ([]map[string]string, error) {
	r, err := exec.Execute(ctx, []string{"distrobox", "list", "--no-color"})
	if err != nil {
		return nil, err
	}
	if !r.Success {
		return nil, nil
	}
	var instances []map[string]string
	lines := strings.Split(r.Stdout, "\n")
	for _, line := range lines[1:] { // Skip header
		parts := strings.Split(line, "|")
		if len(parts) >= 3 {
			instances = append(instances, map[string]string{
				"id":     strings.TrimSpace(parts[0]),
				"name":   strings.TrimSpace(parts[1]),
				"status": strings.TrimSpace(parts[2]),
			})
		}
	}
	return instances, nil
}
