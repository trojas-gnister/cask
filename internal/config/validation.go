package config

import "fmt"

// ValidationWarning is a non-fatal config issue.
type ValidationWarning struct {
	Field   string
	Message string
}

// ValidationResult holds warnings and errors from config validation.
type ValidationResult struct {
	Warnings []ValidationWarning
	Errors   []string
}

// IsValid returns true if there are no errors.
func (v *ValidationResult) IsValid() bool {
	return len(v.Errors) == 0
}

func (v *ValidationResult) addWarning(field, message string) {
	v.Warnings = append(v.Warnings, ValidationWarning{Field: field, Message: message})
}

func (v *ValidationResult) addError(message string) {
	v.Errors = append(v.Errors, message)
}

// ValidateConfig validates the entire configuration, returning warnings and errors.
func ValidateConfig(cfg *CaskConfig) *ValidationResult {
	result := &ValidationResult{}
	validatePodman(cfg, result)
	validateFlatpak(cfg, result)
	return result
}

func validatePodman(cfg *CaskConfig, result *ValidationResult) {
	if cfg.Podman == nil {
		return
	}
	for i, c := range cfg.Podman.Containers {
		prefix := fmt.Sprintf("podman.containers[%d]", i)
		if c.Name == "" {
			result.addError(fmt.Sprintf("%s.name cannot be empty", prefix))
		}
		if c.Image == "" && c.Build == nil {
			result.addError(fmt.Sprintf("%s.image cannot be empty (unless using build)", prefix))
		}
		if c.Autostart == nil || !*c.Autostart {
			result.addWarning(
				prefix+".autostart",
				fmt.Sprintf("Container '%s' has autostart disabled or not set", c.Name),
			)
		}
		if c.Security != nil {
			sec := c.Security
			if sec.SeccompProfile != "" && sec.SeccompProfile[0] != '/' {
				result.addError(fmt.Sprintf("%s.security.seccomp_profile: path must be absolute", prefix))
			}
			if sec.DropAllCaps && len(sec.AddCaps) == 0 {
				result.addWarning(
					prefix+".security",
					"drop_all_caps enabled with no add_caps — container may not function",
				)
			}
		}
	}
}

func validateFlatpak(cfg *CaskConfig, result *ValidationResult) {
	if cfg.Flatpak == nil {
		return
	}
	if cfg.Flatpak.Remotes != nil {
		for i, r := range cfg.Flatpak.Remotes {
			if r.Name == "" {
				result.addError(fmt.Sprintf("flatpak.remotes[%d].name cannot be empty", i))
			}
			if r.URL == "" {
				result.addError(fmt.Sprintf("flatpak.remotes[%d].url cannot be empty", i))
			}
		}
	}
}
