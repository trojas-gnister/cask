package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestExpandEnvVarsSimple(t *testing.T) {
	t.Setenv("TEST_VAR", "hello")
	result := ExpandEnvVars("${TEST_VAR}")
	if result != "hello" {
		t.Errorf("expected 'hello', got '%s'", result)
	}
}

func TestExpandEnvVarsWithDefault(t *testing.T) {
	os.Unsetenv("NONEXISTENT_VAR")
	result := ExpandEnvVars("${NONEXISTENT_VAR:-fallback}")
	if result != "fallback" {
		t.Errorf("expected 'fallback', got '%s'", result)
	}
}

func TestExpandEnvVarsSetOverridesDefault(t *testing.T) {
	t.Setenv("MY_VAR", "actual")
	result := ExpandEnvVars("${MY_VAR:-fallback}")
	if result != "actual" {
		t.Errorf("expected 'actual', got '%s'", result)
	}
}

func TestExpandEnvVarsNoMatch(t *testing.T) {
	os.Unsetenv("MISSING")
	result := ExpandEnvVars("${MISSING}")
	if result != "${MISSING}" {
		t.Errorf("unset var without default should be left unchanged, got '%s'", result)
	}
}

func TestExpandEnvVarsInContext(t *testing.T) {
	t.Setenv("HOME_DIR", "/home/user")
	result := ExpandEnvVars("${HOME_DIR}/projects")
	if result != "/home/user/projects" {
		t.Errorf("expected '/home/user/projects', got '%s'", result)
	}
}

func TestExpandStringsRecursiveMap(t *testing.T) {
	t.Setenv("PROJ", "myproject")
	data := map[string]any{
		"path": "${PROJ}/src",
		"nested": map[string]any{
			"inner": "${PROJ}/lib",
		},
	}
	result := ExpandStringsRecursive(data).(map[string]any)
	if result["path"] != "myproject/src" {
		t.Errorf("expected 'myproject/src', got '%v'", result["path"])
	}
	nested := result["nested"].(map[string]any)
	if nested["inner"] != "myproject/lib" {
		t.Errorf("expected 'myproject/lib', got '%v'", nested["inner"])
	}
}

func TestExpandStringsRecursiveSlice(t *testing.T) {
	t.Setenv("PKG", "vim")
	data := []any{"${PKG}", "git"}
	result := ExpandStringsRecursive(data).([]any)
	if result[0] != "vim" {
		t.Errorf("expected 'vim', got '%v'", result[0])
	}
}

func TestExpandStringsRecursiveNonString(t *testing.T) {
	result := ExpandStringsRecursive(42)
	if result != 42 {
		t.Errorf("non-string values should pass through unchanged")
	}
}

func TestDeepMerge(t *testing.T) {
	base := map[string]any{
		"a": "base_a",
		"b": map[string]any{
			"x": "base_x",
			"y": "base_y",
		},
	}
	overlay := map[string]any{
		"a": "overlay_a",
		"b": map[string]any{
			"y": "overlay_y",
			"z": "overlay_z",
		},
		"c": "new",
	}

	result := DeepMerge(base, overlay)
	if result["a"] != "overlay_a" {
		t.Error("overlay should win on scalar conflict")
	}
	if result["c"] != "new" {
		t.Error("overlay should add new keys")
	}
	nested := result["b"].(map[string]any)
	if nested["x"] != "base_x" {
		t.Error("base keys not in overlay should be preserved")
	}
	if nested["y"] != "overlay_y" {
		t.Error("overlay should win on nested conflict")
	}
	if nested["z"] != "overlay_z" {
		t.Error("overlay should add new nested keys")
	}
}

func TestExpandIncludes(t *testing.T) {
	dir := t.TempDir()

	// Write an included file
	includedContent := `
[flatpak]
packages = ["org.mozilla.Firefox"]
`
	if err := os.WriteFile(filepath.Join(dir, "base.toml"), []byte(includedContent), 0o644); err != nil {
		t.Fatal(err)
	}

	// Main config with include
	raw := map[string]any{
		"include": []any{"base.toml"},
		"flatpak": map[string]any{
			"packages": []any{"org.gimp.GIMP"},
		},
	}

	result := ExpandIncludes(raw, dir)
	flatpak := result["flatpak"].(map[string]any)
	pkgs := flatpak["packages"].([]any)

	// Main config should overlay: GIMP replaces Firefox
	if len(pkgs) != 1 || pkgs[0] != "org.gimp.GIMP" {
		t.Errorf("main config should overlay includes, got %v", pkgs)
	}
}

func TestExpandIncludesNoDirective(t *testing.T) {
	raw := map[string]any{"key": "value"}
	result := ExpandIncludes(raw, "/tmp")
	if result["key"] != "value" {
		t.Error("no-include config should pass through unchanged")
	}
}

func TestExpandIncludesMissingFile(t *testing.T) {
	raw := map[string]any{
		"include": []any{"nonexistent.toml"},
		"key":     "value",
	}
	result := ExpandIncludes(raw, t.TempDir())
	if result["key"] != "value" {
		t.Error("missing include should be silently skipped")
	}
}
