package sync

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/iskry/cask/internal/config"
	"github.com/iskry/cask/internal/executor"
)

// ToolsSyncManager synchronizes mise-managed tools between host and config.
type ToolsSyncManager struct {
	Config *config.ToolsConfig
}

func (m *ToolsSyncManager) ResourceType() string { return "tools" }

func (m *ToolsSyncManager) GetHostResources(ctx context.Context, exec executor.Executor) ([]Resource, error) {
	r, err := exec.Execute(ctx, []string{"mise", "list", "--json"})
	if err != nil || !r.Success || strings.TrimSpace(r.Stdout) == "" {
		return nil, nil
	}

	var data map[string]any
	if err := json.Unmarshal([]byte(r.Stdout), &data); err != nil {
		return nil, nil
	}

	var resources []Resource
	for toolName, versions := range data {
		switch v := versions.(type) {
		case []any:
			for _, ver := range v {
				if verMap, ok := ver.(map[string]any); ok {
					version, _ := verMap["version"].(string)
					resources = append(resources, Resource{
						"name":    toolName,
						"version": version,
					})
				}
			}
		case map[string]any:
			version, _ := v["version"].(string)
			resources = append(resources, Resource{
				"name":    toolName,
				"version": version,
			})
		}
	}
	return resources, nil
}

func (m *ToolsSyncManager) GetConfigResources() []Resource {
	if m.Config == nil {
		return nil
	}
	var resources []Resource
	for _, t := range m.Config.Tools {
		resources = append(resources, Resource{
			"name":           t.Name,
			"version":        t.Version,
			"global_install": t.GlobalInstall,
		})
	}
	return resources
}

func (m *ToolsSyncManager) ResourceID(r Resource) string {
	name, _ := r["name"].(string)
	return name
}

func (m *ToolsSyncManager) Apply(ctx context.Context, exec executor.Executor, r Resource) bool {
	name, _ := r["name"].(string)
	version, _ := r["version"].(string)
	if version == "" {
		version = "latest"
	}
	spec := fmt.Sprintf("%s@%s", name, version)

	result, err := exec.Execute(ctx, []string{"mise", "install", spec})
	if err != nil || !result.Success {
		return false
	}

	if globalInstall, _ := r["global_install"].(bool); globalInstall {
		result, err = exec.Execute(ctx, []string{"mise", "use", "--global", spec})
		return err == nil && result.Success
	}
	return true
}

func (m *ToolsSyncManager) Remove(ctx context.Context, exec executor.Executor, id string) bool {
	r, err := exec.Execute(ctx, []string{"mise", "uninstall", "--all", id})
	return err == nil && r.Success
}

func (m *ToolsSyncManager) NeedsUpdate(host, cfg Resource) bool {
	cfgVersion, _ := cfg["version"].(string)
	if cfgVersion == "latest" || cfgVersion == "stable" {
		return false
	}
	hostVersion, _ := host["version"].(string)
	return !strings.HasPrefix(hostVersion, cfgVersion)
}
