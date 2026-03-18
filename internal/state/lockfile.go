package state

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/iskry/cask/internal/config"
	"github.com/iskry/cask/internal/executor"
)

// FlatpakLock pins an exact Flatpak version.
type FlatpakLock struct {
	AppID   string `json:"app_id"`
	Commit  string `json:"commit"`
	Version string `json:"version"`
}

// ContainerLock pins an exact container image.
type ContainerLock struct {
	Name    string `json:"name"`
	ImageID string `json:"image_id"`
	Image   string `json:"image"`
}

// ToolLock pins an exact tool version.
type ToolLock struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

// DevboxLock pins a Distrobox instance image.
type DevboxLock struct {
	Name  string `json:"name"`
	Image string `json:"image"`
}

// Lockfile contains pinned versions of all managed resources.
type Lockfile struct {
	Flatpaks   []FlatpakLock   `json:"flatpaks"`
	Containers []ContainerLock `json:"containers"`
	Tools      []ToolLock      `json:"tools"`
	Devboxes   []DevboxLock    `json:"devboxes"`
}

// LoadLockfile loads the lockfile from the default path.
func LoadLockfile() (*Lockfile, error) {
	path := config.LockfilePath()
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, nil // No lockfile is not an error
	}
	var lf Lockfile
	if err := json.Unmarshal(data, &lf); err != nil {
		return nil, fmt.Errorf("parsing lockfile: %w", err)
	}
	return &lf, nil
}

// SaveLockfile saves the lockfile to the default path.
func SaveLockfile(lf *Lockfile) error {
	path := config.LockfilePath()
	if err := config.EnsureDir(config.ConfigDir()); err != nil {
		return err
	}
	data, err := json.MarshalIndent(lf, "", "  ")
	if err != nil {
		return fmt.Errorf("serializing lockfile: %w", err)
	}
	return os.WriteFile(path, data, 0o644)
}

// GenerateLockfile queries the system to generate a lockfile with exact versions.
func GenerateLockfile(ctx context.Context, exec executor.Executor) (*Lockfile, error) {
	lf := &Lockfile{}

	// Pin flatpak commits
	r, err := exec.Execute(ctx, []string{"flatpak", "list", "--app", "--columns=application,version"})
	if err == nil && r.Success {
		for _, line := range strings.Split(strings.TrimSpace(r.Stdout), "\n") {
			parts := strings.SplitN(line, "\t", 2)
			if len(parts) < 2 {
				continue
			}
			appID := strings.TrimSpace(parts[0])
			version := strings.TrimSpace(parts[1])
			commit := ""
			cr, cerr := exec.Execute(ctx, []string{"flatpak", "info", "--show-commit", appID})
			if cerr == nil && cr.Success {
				commit = strings.TrimSpace(cr.Stdout)
			}
			lf.Flatpaks = append(lf.Flatpaks, FlatpakLock{
				AppID: appID, Commit: commit, Version: version,
			})
		}
	}

	// Pin container images
	r, err = exec.Execute(ctx, []string{
		"podman", "ps", "-a", "--format", "{{.Names}}\t{{.Image}}\t{{.ImageID}}",
	})
	if err == nil && r.Success {
		for _, line := range strings.Split(strings.TrimSpace(r.Stdout), "\n") {
			parts := strings.SplitN(line, "\t", 3)
			if len(parts) < 3 {
				continue
			}
			lf.Containers = append(lf.Containers, ContainerLock{
				Name:    strings.TrimSpace(parts[0]),
				Image:   strings.TrimSpace(parts[1]),
				ImageID: strings.TrimSpace(parts[2]),
			})
		}
	}

	// Pin tool versions
	r, err = exec.Execute(ctx, []string{"mise", "list", "--json"})
	if err == nil && r.Success && strings.TrimSpace(r.Stdout) != "" {
		var toolData map[string][]map[string]any
		if jerr := json.Unmarshal([]byte(r.Stdout), &toolData); jerr == nil {
			for toolName, versions := range toolData {
				if len(versions) > 0 {
					version, _ := versions[0]["version"].(string)
					lf.Tools = append(lf.Tools, ToolLock{
						Name: toolName, Version: version,
					})
				}
			}
		}
	}

	return lf, nil
}

// VerifyLockfile checks that the current system matches the lockfile.
func VerifyLockfile(ctx context.Context, exec executor.Executor) ([]string, error) {
	lf, err := LoadLockfile()
	if err != nil {
		return nil, err
	}
	if lf == nil {
		return nil, fmt.Errorf("no lockfile found")
	}

	var mismatches []string
	for _, tool := range lf.Tools {
		r, err := exec.Execute(ctx, []string{"mise", "list", tool.Name, "--json"})
		if err != nil || !r.Success {
			mismatches = append(mismatches, fmt.Sprintf("%s: unable to check version", tool.Name))
			continue
		}
		var data map[string][]map[string]any
		if jerr := json.Unmarshal([]byte(r.Stdout), &data); jerr != nil {
			mismatches = append(mismatches, fmt.Sprintf("%s: unable to parse version info", tool.Name))
			continue
		}
		versions := data[tool.Name]
		found := false
		for _, v := range versions {
			if ver, _ := v["version"].(string); ver == tool.Version {
				found = true
				break
			}
		}
		if !found {
			mismatches = append(mismatches, fmt.Sprintf("%s: expected %s, not installed", tool.Name, tool.Version))
		}
	}

	return mismatches, nil
}
