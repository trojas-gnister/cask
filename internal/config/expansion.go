package config

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"

	toml "github.com/pelletier/go-toml/v2"
)

// envPattern matches ${VAR} and ${VAR:-default}.
var envPattern = regexp.MustCompile(`\$\{([A-Za-z_][A-Za-z0-9_]*)(?::-([^}]*))?\}`)

// ExpandEnvVars expands ${VAR} and ${VAR:-default} in a string.
func ExpandEnvVars(value string) string {
	return envPattern.ReplaceAllStringFunc(value, func(match string) string {
		groups := envPattern.FindStringSubmatch(match)
		if groups == nil {
			return match
		}
		varName := groups[1]
		defaultVal := groups[2]

		if envVal, ok := os.LookupEnv(varName); ok {
			return envVal
		}
		if defaultVal != "" {
			return defaultVal
		}
		return match
	})
}

// expandTilde expands a leading ~ to the user's home directory.
func expandTilde(value string) string {
	if !strings.HasPrefix(value, "~") {
		return value
	}
	home := HomeDir()
	if value == "~" {
		return home
	}
	if strings.HasPrefix(value, "~/") {
		return filepath.Join(home, value[2:])
	}
	return value
}

// ExpandStringsRecursive recursively expands env vars and tilde in all string values.
func ExpandStringsRecursive(data any) any {
	switch v := data.(type) {
	case string:
		return expandTilde(ExpandEnvVars(v))
	case map[string]any:
		result := make(map[string]any, len(v))
		for k, val := range v {
			result[k] = ExpandStringsRecursive(val)
		}
		return result
	case []any:
		result := make([]any, len(v))
		for i, val := range v {
			result[i] = ExpandStringsRecursive(val)
		}
		return result
	default:
		return data
	}
}

// DeepMerge merges overlay into base (overlay wins on conflict).
func DeepMerge(base, overlay map[string]any) map[string]any {
	result := make(map[string]any, len(base))
	for k, v := range base {
		result[k] = v
	}
	for k, v := range overlay {
		if baseMap, ok := result[k].(map[string]any); ok {
			if overlayMap, ok := v.(map[string]any); ok {
				result[k] = DeepMerge(baseMap, overlayMap)
				continue
			}
		}
		result[k] = v
	}
	return result
}

// ExpandIncludes processes the include = [...] directive, loading and merging
// included TOML files. Included files are loaded first (depth-first), then the
// main config overlays on top.
func ExpandIncludes(raw map[string]any, baseDir string) map[string]any {
	includesRaw, ok := raw["include"]
	if !ok {
		return raw
	}
	delete(raw, "include")

	includes, ok := includesRaw.([]any)
	if !ok {
		return raw
	}

	merged := make(map[string]any)
	for _, inc := range includes {
		includePath, ok := inc.(string)
		if !ok {
			continue
		}
		resolved := filepath.Join(baseDir, includePath)
		data, err := os.ReadFile(resolved)
		if err != nil {
			continue
		}
		var includedRaw map[string]any
		if err := toml.Unmarshal(data, &includedRaw); err != nil {
			continue
		}
		// Recursively process includes in the included file
		includedRaw = ExpandIncludes(includedRaw, filepath.Dir(resolved))
		merged = DeepMerge(merged, includedRaw)
	}

	// Main config overlays on top of includes
	return DeepMerge(merged, raw)
}
