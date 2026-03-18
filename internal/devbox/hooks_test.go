package devbox

import (
	"strings"
	"testing"

	"github.com/iskry/cask/internal/config"
)

func TestGenerateBashHook(t *testing.T) {
	projects := []config.DevboxProject{
		{Path: "/home/user/projects/web", BoxName: "web-dev", AutoEnter: true},
	}
	hook := GenerateBashHook(projects)
	if !strings.Contains(hook, "distrobox enter web-dev -- bash") {
		t.Error("should contain enter command for bash")
	}
	if !strings.Contains(hook, "alias cd=_cask_cd") {
		t.Error("should alias cd")
	}
}

func TestGenerateZshHook(t *testing.T) {
	projects := []config.DevboxProject{
		{Path: "/home/user/projects/api", BoxName: "api-dev"},
	}
	hook := GenerateZshHook(projects)
	if !strings.Contains(hook, "distrobox enter api-dev -- zsh") {
		t.Error("should contain enter command for zsh")
	}
	if !strings.Contains(hook, "add-zsh-hook chpwd") {
		t.Error("should use zsh chpwd hook")
	}
}

func TestGenerateFishHook(t *testing.T) {
	projects := []config.DevboxProject{
		{Path: "/home/user/projects/ml", BoxName: "ml-dev"},
	}
	hook := GenerateFishHook(projects)
	if !strings.Contains(hook, "distrobox enter ml-dev -- fish") {
		t.Error("should contain enter command for fish")
	}
	if !strings.Contains(hook, "--on-variable PWD") {
		t.Error("should use fish PWD variable event")
	}
}

func TestGenerateHookWithCustomHook(t *testing.T) {
	projects := []config.DevboxProject{
		{Path: "/home/user/projects/web", BoxName: "web-dev", Hook: "source .env"},
	}
	hook := GenerateBashHook(projects)
	if !strings.Contains(hook, "source .env && exec bash") {
		t.Error("should include custom hook command")
	}
}

func TestGenerateHookRouter(t *testing.T) {
	projects := []config.DevboxProject{
		{Path: "/test", BoxName: "dev"},
	}
	if !strings.Contains(GenerateHook("bash", projects), "_cask_cd") {
		t.Error("bash should route to bash hook")
	}
	if !strings.Contains(GenerateHook("zsh", projects), "chpwd") {
		t.Error("zsh should route to zsh hook")
	}
	if !strings.Contains(GenerateHook("fish", projects), "--on-variable") {
		t.Error("fish should route to fish hook")
	}
	if !strings.Contains(GenerateHook("tcsh", projects), "Unsupported") {
		t.Error("unknown shell should say unsupported")
	}
}

func TestGenerateFishHookEmpty(t *testing.T) {
	hook := GenerateFishHook(nil)
	if !strings.Contains(hook, "No projects configured") {
		t.Error("empty projects should produce placeholder comment")
	}
}

func TestMatchProject(t *testing.T) {
	projects := []config.DevboxProject{
		{Path: "/home/user/projects/web", BoxName: "web-dev"},
		{Path: "/home/user/projects/api", BoxName: "api-dev"},
	}

	// Direct match
	p := MatchProject("/home/user/projects/web", projects)
	if p == nil || p.BoxName != "web-dev" {
		t.Error("should match web project directly")
	}

	// Subdirectory match
	p = MatchProject("/home/user/projects/web/src/components", projects)
	if p == nil || p.BoxName != "web-dev" {
		t.Error("should match web project from subdirectory")
	}

	// No match
	p = MatchProject("/home/user/documents", projects)
	if p != nil {
		t.Error("should not match unrelated directory")
	}
}
