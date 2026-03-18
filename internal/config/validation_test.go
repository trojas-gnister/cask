package config

import (
	"testing"
)

func boolPtr(b bool) *bool { return &b }

func TestValidateConfigEmptyIsValid(t *testing.T) {
	cfg := &CaskConfig{}
	result := ValidateConfig(cfg)
	if !result.IsValid() {
		t.Errorf("empty config should be valid, got errors: %v", result.Errors)
	}
}

func TestValidateContainerNoName(t *testing.T) {
	cfg := &CaskConfig{
		Podman: &PodmanConfig{
			Containers: []Container{{Image: "alpine"}},
		},
	}
	result := ValidateConfig(cfg)
	if result.IsValid() {
		t.Error("container without name should be invalid")
	}
}

func TestValidateContainerNoImageNoBuild(t *testing.T) {
	cfg := &CaskConfig{
		Podman: &PodmanConfig{
			Containers: []Container{{Name: "test"}},
		},
	}
	result := ValidateConfig(cfg)
	if result.IsValid() {
		t.Error("container without image or build should be invalid")
	}
}

func TestValidateContainerWithBuildNoImage(t *testing.T) {
	cfg := &CaskConfig{
		Podman: &PodmanConfig{
			Containers: []Container{{
				Name:  "test",
				Build: &ContainerBuildConfig{Context: "."},
			}},
		},
	}
	result := ValidateConfig(cfg)
	if !result.IsValid() {
		t.Errorf("container with build should be valid, got: %v", result.Errors)
	}
}

func TestValidateContainerAutostartWarning(t *testing.T) {
	cfg := &CaskConfig{
		Podman: &PodmanConfig{
			Containers: []Container{{
				Name:  "test",
				Image: "alpine",
			}},
		},
	}
	result := ValidateConfig(cfg)
	if len(result.Warnings) == 0 {
		t.Error("container without autostart should produce warning")
	}
}

func TestValidateContainerAutostartNoWarning(t *testing.T) {
	cfg := &CaskConfig{
		Podman: &PodmanConfig{
			Containers: []Container{{
				Name:      "test",
				Image:     "alpine",
				Autostart: boolPtr(true),
			}},
		},
	}
	result := ValidateConfig(cfg)
	if len(result.Warnings) != 0 {
		t.Errorf("container with autostart=true should not warn, got: %v", result.Warnings)
	}
}

func TestValidateContainerSeccompRelativePath(t *testing.T) {
	cfg := &CaskConfig{
		Podman: &PodmanConfig{
			Containers: []Container{{
				Name:  "test",
				Image: "alpine",
				Autostart: boolPtr(true),
				Security: &ContainerSecurityOptions{
					SeccompProfile: "profiles/custom.json",
				},
			}},
		},
	}
	result := ValidateConfig(cfg)
	if result.IsValid() {
		t.Error("relative seccomp profile path should be invalid")
	}
}

func TestValidateContainerDropCapsNoAdd(t *testing.T) {
	cfg := &CaskConfig{
		Podman: &PodmanConfig{
			Containers: []Container{{
				Name:      "test",
				Image:     "alpine",
				Autostart: boolPtr(true),
				Security: &ContainerSecurityOptions{
					DropAllCaps: true,
				},
			}},
		},
	}
	result := ValidateConfig(cfg)
	if len(result.Warnings) == 0 {
		t.Error("drop_all_caps without add_caps should warn")
	}
}

func TestValidateFlatpakRemoteNoName(t *testing.T) {
	cfg := &CaskConfig{
		Flatpak: &FlatpakConfig{
			Remotes: []FlatpakRemote{{URL: "https://example.com"}},
		},
	}
	result := ValidateConfig(cfg)
	if result.IsValid() {
		t.Error("remote without name should be invalid")
	}
}

func TestValidateFlatpakRemoteNoURL(t *testing.T) {
	cfg := &CaskConfig{
		Flatpak: &FlatpakConfig{
			Remotes: []FlatpakRemote{{Name: "test"}},
		},
	}
	result := ValidateConfig(cfg)
	if result.IsValid() {
		t.Error("remote without URL should be invalid")
	}
}
