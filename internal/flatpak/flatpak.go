// Package flatpak handles Flatpak package management operations.
package flatpak

import (
	"context"
	"strings"

	"github.com/iskry/cask/internal/executor"
)

// Install installs Flatpak packages.
func Install(ctx context.Context, exec executor.Executor, packages []string) (*executor.CommandResult, error) {
	if len(packages) == 0 {
		return &executor.CommandResult{Success: true, Stdout: "No Flatpak packages to install"}, nil
	}
	cmd := append([]string{"flatpak", "install", "-y", "--noninteractive"}, packages...)
	return exec.ExecuteSudo(ctx, cmd)
}

// Remove removes Flatpak packages.
func Remove(ctx context.Context, exec executor.Executor, packages []string) (*executor.CommandResult, error) {
	if len(packages) == 0 {
		return &executor.CommandResult{Success: true, Stdout: "No Flatpak packages to remove"}, nil
	}
	cmd := append([]string{"flatpak", "uninstall", "-y", "--noninteractive"}, packages...)
	return exec.ExecuteSudo(ctx, cmd)
}

// List returns installed Flatpak application IDs.
func List(ctx context.Context, exec executor.Executor) ([]string, error) {
	r, err := exec.Execute(ctx, []string{"flatpak", "list", "--app", "--columns=application"})
	if err != nil {
		return nil, err
	}
	if !r.Success {
		return nil, nil
	}
	var apps []string
	for _, line := range strings.Split(r.Stdout, "\n") {
		line = strings.TrimSpace(line)
		if line != "" {
			apps = append(apps, line)
		}
	}
	return apps, nil
}

// AddRemote adds a Flatpak remote repository.
func AddRemote(ctx context.Context, exec executor.Executor, name, url string) (*executor.CommandResult, error) {
	return exec.ExecuteSudo(ctx, []string{
		"flatpak", "remote-add", "--if-not-exists", name, url,
	})
}
