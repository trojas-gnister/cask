package sync

import (
	"context"
	"encoding/json"

	"github.com/iskry/cask/internal/config"
	"github.com/iskry/cask/internal/executor"
)

// ContainerSyncManager synchronizes Podman containers between host and config.
type ContainerSyncManager struct {
	Config *config.PodmanConfig
}

func (m *ContainerSyncManager) ResourceType() string { return "containers" }

func (m *ContainerSyncManager) GetHostResources(ctx context.Context, exec executor.Executor) ([]Resource, error) {
	var containers []Resource

	// User-scoped containers
	r, err := exec.Execute(ctx, []string{"podman", "ps", "-a", "--format", "json"})
	if err == nil && r.Success && r.Stdout != "" {
		containers = append(containers, parseContainerJSON(r.Stdout, "user")...)
	}

	// System-scoped containers
	r, err = exec.ExecuteSudo(ctx, []string{"podman", "ps", "-a", "--format", "json"})
	if err == nil && r.Success && r.Stdout != "" {
		containers = append(containers, parseContainerJSON(r.Stdout, "system")...)
	}

	return containers, nil
}

func (m *ContainerSyncManager) GetConfigResources() []Resource {
	if m.Config == nil {
		return nil
	}
	var resources []Resource
	for _, c := range m.Config.Containers {
		resources = append(resources, Resource{
			"name":  c.Name,
			"image": c.Image,
			"scope": string(c.Scope),
		})
	}
	return resources
}

func (m *ContainerSyncManager) ResourceID(r Resource) string {
	name, _ := r["name"].(string)
	return name
}

func (m *ContainerSyncManager) Apply(ctx context.Context, exec executor.Executor, r Resource) bool {
	name, _ := r["name"].(string)
	image, _ := r["image"].(string)
	scope, _ := r["scope"].(string)

	cmd := []string{"podman", "create", "--name", name, image}
	var result *executor.CommandResult
	var err error
	if scope == "system" {
		result, err = exec.ExecuteSudo(ctx, cmd)
	} else {
		result, err = exec.Execute(ctx, cmd)
	}
	return err == nil && result.Success
}

func (m *ContainerSyncManager) Remove(ctx context.Context, exec executor.Executor, id string) bool {
	exec.Execute(ctx, []string{"podman", "stop", id})
	r, err := exec.Execute(ctx, []string{"podman", "rm", "-f", id})
	return err == nil && r.Success
}

func (m *ContainerSyncManager) NeedsUpdate(host, cfg Resource) bool {
	hostImage, _ := host["image"].(string)
	cfgImage, _ := cfg["image"].(string)
	return hostImage != cfgImage
}

func parseContainerJSON(jsonStr string, scope string) []Resource {
	var raw []map[string]any
	if err := json.Unmarshal([]byte(jsonStr), &raw); err != nil {
		return nil
	}
	var resources []Resource
	for _, c := range raw {
		name := ""
		if names, ok := c["Names"].([]any); ok && len(names) > 0 {
			name, _ = names[0].(string)
		} else if n, ok := c["Name"].(string); ok {
			name = n
		}
		image, _ := c["Image"].(string)
		resources = append(resources, Resource{
			"name":  name,
			"image": image,
			"scope": scope,
		})
	}
	return resources
}
