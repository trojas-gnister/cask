package sync

import (
	"context"
	"strings"

	"github.com/iskry/cask/internal/config"
	"github.com/iskry/cask/internal/executor"
)

// DevboxSyncManager synchronizes distrobox instances between host and config.
type DevboxSyncManager struct {
	Config *config.DevboxConfig
}

func (m *DevboxSyncManager) ResourceType() string { return "devboxes" }

func (m *DevboxSyncManager) GetHostResources(ctx context.Context, exec executor.Executor) ([]Resource, error) {
	r, err := exec.Execute(ctx, []string{"distrobox", "list", "--no-color"})
	if err != nil {
		return nil, err
	}
	if !r.Success {
		return nil, nil
	}
	var resources []Resource
	lines := strings.Split(r.Stdout, "\n")
	for _, line := range lines[1:] { // Skip header
		parts := strings.Split(line, "|")
		if len(parts) >= 3 {
			resources = append(resources, Resource{
				"id":     strings.TrimSpace(parts[0]),
				"name":   strings.TrimSpace(parts[1]),
				"status": strings.TrimSpace(parts[2]),
			})
		}
	}
	return resources, nil
}

func (m *DevboxSyncManager) GetConfigResources() []Resource {
	if m.Config == nil {
		return nil
	}
	var resources []Resource
	for _, inst := range m.Config.Instances {
		resources = append(resources, Resource{
			"name":  inst.Name,
			"image": inst.Image,
		})
	}
	return resources
}

func (m *DevboxSyncManager) ResourceID(r Resource) string {
	name, _ := r["name"].(string)
	return name
}

func (m *DevboxSyncManager) Apply(ctx context.Context, exec executor.Executor, r Resource) bool {
	name, _ := r["name"].(string)
	image, _ := r["image"].(string)
	result, err := exec.Execute(ctx, []string{
		"distrobox", "create", "--name", name, "--image", image, "--yes",
	})
	return err == nil && result.Success
}

func (m *DevboxSyncManager) Remove(ctx context.Context, exec executor.Executor, id string) bool {
	r, err := exec.Execute(ctx, []string{"distrobox", "rm", "--force", id})
	return err == nil && r.Success
}

func (m *DevboxSyncManager) NeedsUpdate(host, cfg Resource) bool {
	hostImage, _ := host["image"].(string)
	cfgImage, _ := cfg["image"].(string)
	return hostImage != cfgImage
}
