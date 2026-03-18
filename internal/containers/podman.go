// Package containers handles Podman container and Distrobox environment management.
package containers

import (
	"context"
	"fmt"
	"strings"

	"github.com/iskry/cask/internal/config"
	"github.com/iskry/cask/internal/executor"
)

// BuildSecurityFlags constructs podman security flags from container security options.
func BuildSecurityFlags(sec *config.ContainerSecurityOptions) []string {
	if sec == nil {
		return nil
	}
	var flags []string

	if sec.ReadOnlyRootfs {
		flags = append(flags, "--read-only")
	}
	if sec.DropAllCaps {
		flags = append(flags, "--cap-drop=ALL")
	}
	for _, cap := range sec.AddCaps {
		flags = append(flags, fmt.Sprintf("--cap-add=%s", cap))
	}
	if sec.NoNewPrivileges {
		flags = append(flags, "--security-opt=no-new-privileges")
	}
	if sec.SeccompProfile != "" && sec.SeccompProfile != "default" {
		flags = append(flags, fmt.Sprintf("--security-opt=seccomp=%s", sec.SeccompProfile))
	}
	if sec.User != "" {
		flags = append(flags, "--user", sec.User)
	}
	if sec.AppArmorProfile != "" {
		flags = append(flags, fmt.Sprintf("--security-opt=apparmor=%s", sec.AppArmorProfile))
	}
	for _, ip := range sec.DNS {
		flags = append(flags, "--dns", ip)
	}
	for _, domain := range sec.DNSSearch {
		flags = append(flags, "--dns-search", domain)
	}
	for _, opt := range sec.DNSOptions {
		flags = append(flags, "--dns-option", opt)
	}
	for _, mount := range sec.Tmpfs {
		val := mount.Path
		if mount.Options != "" {
			val = fmt.Sprintf("%s:%s", mount.Path, mount.Options)
		}
		flags = append(flags, "--tmpfs", val)
	}
	return flags
}

// BuildImage runs podman build if the container has a build config.
func BuildImage(ctx context.Context, exec executor.Executor, c *config.Container) error {
	if c.Build == nil {
		return nil
	}
	useSudo := c.Scope == config.ScopeSystem

	var cmd []string
	if useSudo {
		cmd = []string{"sudo", "podman"}
	} else {
		cmd = []string{"podman"}
	}
	cmd = append(cmd, "build", "--tag", c.Image)

	if c.Build.Dockerfile != "" {
		cmd = append(cmd, "-f", c.Build.Dockerfile)
	}
	for k, v := range c.Build.BuildArgs {
		cmd = append(cmd, "--build-arg", fmt.Sprintf("%s=%s", k, v))
	}
	cmd = append(cmd, c.Build.ExtraFlags...)
	cmd = append(cmd, c.Build.Context)

	r, err := exec.Execute(ctx, cmd)
	if err != nil {
		return err
	}
	if !r.Success {
		return fmt.Errorf("building image for %s: %s", c.Name, strings.TrimSpace(r.Stderr))
	}
	return nil
}

// SetupContainers creates Podman containers from config.
func SetupContainers(ctx context.Context, exec executor.Executor, cfg *config.PodmanConfig) error {
	if cfg == nil {
		return nil
	}

	// Run pre-setup commands
	for _, setup := range cfg.PreContainerSetup {
		r, err := exec.ExecuteShell(ctx, setup.Command)
		if err != nil {
			return fmt.Errorf("pre-setup command %q: %w", setup.Description, err)
		}
		if !r.Success {
			return fmt.Errorf("pre-setup command %q failed: %s", setup.Description, strings.TrimSpace(r.Stderr))
		}
	}

	// Create containers
	for i := range cfg.Containers {
		c := &cfg.Containers[i]
		useSudo := c.Scope == config.ScopeSystem

		// Build image if needed
		if c.Build != nil {
			if err := BuildImage(ctx, exec, c); err != nil {
				return err
			}
		}

		// Create container
		cmd := []string{"podman", "create", "--name", c.Name}
		if c.RawFlags != "" {
			cmd = append(cmd, splitFlags(c.RawFlags)...)
		}
		cmd = append(cmd, BuildSecurityFlags(c.Security)...)
		cmd = append(cmd, c.Image)

		var r *executor.CommandResult
		var err error
		if useSudo {
			r, err = exec.ExecuteSudo(ctx, cmd)
		} else {
			r, err = exec.Execute(ctx, cmd)
		}
		if err != nil {
			return fmt.Errorf("creating container %s: %w", c.Name, err)
		}
		if !r.Success {
			return fmt.Errorf("creating container %s: %s", c.Name, strings.TrimSpace(r.Stderr))
		}

		// Write Quadlet for autostart
		if c.Autostart != nil && *c.Autostart {
			_, changed, err := WriteQuadlet(c)
			if err != nil {
				return fmt.Errorf("writing quadlet for %s: %w", c.Name, err)
			}
			if changed {
				if useSudo {
					exec.ExecuteSudo(ctx, []string{"systemctl", "daemon-reload"})
				} else {
					exec.Execute(ctx, []string{"systemctl", "--user", "daemon-reload"})
				}
			}
		}
	}
	return nil
}

// ListContainers returns a list of Podman containers.
func ListContainers(ctx context.Context, exec executor.Executor) ([]map[string]string, error) {
	r, err := exec.Execute(ctx, []string{
		"podman", "ps", "-a", "--format", "{{.Names}}\t{{.Image}}\t{{.Status}}",
	})
	if err != nil {
		return nil, err
	}
	if !r.Success {
		return nil, nil
	}
	var containers []map[string]string
	for _, line := range strings.Split(strings.TrimSpace(r.Stdout), "\n") {
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, "\t", 3)
		if len(parts) < 3 {
			continue
		}
		containers = append(containers, map[string]string{
			"name":   strings.TrimSpace(parts[0]),
			"image":  strings.TrimSpace(parts[1]),
			"status": strings.TrimSpace(parts[2]),
		})
	}
	return containers, nil
}

// splitFlags splits a flags string respecting quotes.
func splitFlags(flags string) []string {
	var result []string
	var current strings.Builder
	inQuote := false
	quoteChar := byte(0)

	for i := 0; i < len(flags); i++ {
		c := flags[i]
		if inQuote {
			if c == quoteChar {
				inQuote = false
			} else {
				current.WriteByte(c)
			}
		} else if c == '\'' || c == '"' {
			inQuote = true
			quoteChar = c
		} else if c == ' ' || c == '\t' {
			if current.Len() > 0 {
				result = append(result, current.String())
				current.Reset()
			}
		} else {
			current.WriteByte(c)
		}
	}
	if current.Len() > 0 {
		result = append(result, current.String())
	}
	return result
}
