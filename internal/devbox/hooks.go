// Package devbox generates shell hooks for auto-entering Distrobox environments.
package devbox

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/iskry/cask/internal/config"
)

// MatchProject finds the first project whose path pattern matches the given directory.
func MatchProject(cwd string, projects []config.DevboxProject) *config.DevboxProject {
	for i := range projects {
		p := &projects[i]
		// Support glob patterns
		if matched, _ := filepath.Match(p.Path, cwd); matched {
			return p
		}
		// Also match if cwd is under the project path
		expandedPath := p.Path
		if strings.HasPrefix(expandedPath, "~/") {
			// Already expanded by config loader, but handle just in case
		}
		rel, err := filepath.Rel(expandedPath, cwd)
		if err == nil && !strings.HasPrefix(rel, "..") {
			return p
		}
	}
	return nil
}

func buildEnterCommand(p *config.DevboxProject, shell string) string {
	if p.Hook != "" {
		return fmt.Sprintf(`distrobox enter %s -- sh -c "%s && exec %s"`, p.BoxName, p.Hook, shell)
	}
	return fmt.Sprintf("distrobox enter %s -- %s", p.BoxName, shell)
}

// GenerateBashHook generates a bash hook that auto-enters distrobox on cd.
func GenerateBashHook(projects []config.DevboxProject) string {
	var cases []string
	for i := range projects {
		p := &projects[i]
		enterCmd := buildEnterCommand(p, "bash")
		cases = append(cases, fmt.Sprintf("        %s*)\n            %s ;;", p.Path, enterCmd))
	}

	caseBlock := strings.Join(cases, "\n")
	return fmt.Sprintf(`# cask devbox auto-enter hook (bash)
_cask_devbox_hook() {
    case "$PWD" in
%s
    esac
}

_cask_cd() {
    builtin cd "$@" && _cask_devbox_hook
}
alias cd=_cask_cd

_cask_pushd() {
    builtin pushd "$@" && _cask_devbox_hook
}
alias pushd=_cask_pushd

_cask_popd() {
    builtin popd "$@" && _cask_devbox_hook
}
alias popd=_cask_popd
`, caseBlock)
}

// GenerateZshHook generates a zsh hook that auto-enters distrobox on directory change.
func GenerateZshHook(projects []config.DevboxProject) string {
	var cases []string
	for i := range projects {
		p := &projects[i]
		enterCmd := buildEnterCommand(p, "zsh")
		cases = append(cases, fmt.Sprintf("        %s*)\n            %s ;;", p.Path, enterCmd))
	}

	caseBlock := strings.Join(cases, "\n")
	return fmt.Sprintf(`# cask devbox auto-enter hook (zsh)
_cask_devbox_chpwd() {
    case "$PWD" in
%s
    esac
}

autoload -U add-zsh-hook
add-zsh-hook chpwd _cask_devbox_chpwd
`, caseBlock)
}

// GenerateFishHook generates a fish hook that auto-enters distrobox on directory change.
func GenerateFishHook(projects []config.DevboxProject) string {
	var conditions []string
	for i := range projects {
		p := &projects[i]
		enterCmd := buildEnterCommand(p, "fish")
		conditions = append(conditions, fmt.Sprintf(`    if string match -q "%s*" $PWD
        %s
    end`, p.Path, enterCmd))
	}

	condBlock := strings.Join(conditions, "\n    else\n")
	if len(conditions) == 0 {
		condBlock = "    # No projects configured"
	}
	return fmt.Sprintf(`# cask devbox auto-enter hook (fish)
function _cask_devbox_hook --on-variable PWD
%s
end
`, condBlock)
}

// GenerateHook generates a shell hook for the given shell type.
func GenerateHook(shell string, projects []config.DevboxProject) string {
	switch shell {
	case "bash":
		return GenerateBashHook(projects)
	case "zsh":
		return GenerateZshHook(projects)
	case "fish":
		return GenerateFishHook(projects)
	default:
		return fmt.Sprintf("# Unsupported shell: %s", shell)
	}
}
