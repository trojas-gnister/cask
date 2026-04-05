"""Shell auto-enter hooks for distrobox instances."""
from __future__ import annotations

import os


def generate_hook(shell: str, hooks: dict[str, str]) -> str:
    """Generate shell hook code for auto-entering distrobox on cd."""
    if shell == "zsh":
        return _zsh_hook(hooks)
    elif shell == "bash":
        return _bash_hook(hooks)
    elif shell == "fish":
        return _fish_hook(hooks)
    return ""


def _expand_path(path: str) -> str:
    return os.path.expanduser(path)


def _zsh_hook(hooks: dict[str, str]) -> str:
    lines = ["# Cask auto-enter hooks", "cask_auto_enter() {"]
    for path, instance in hooks.items():
        expanded = _expand_path(path)
        lines.append(f'  if [[ "$PWD" == {expanded}* ]]; then')
        lines.append(f'    [ -z "$CONTAINER_ID" ] && distrobox enter {instance}')
        lines.append("  fi")
    lines.extend(["}", "autoload -U add-zsh-hook", "add-zsh-hook chpwd cask_auto_enter"])
    return "\n".join(lines) + "\n"


def _bash_hook(hooks: dict[str, str]) -> str:
    lines = ["# Cask auto-enter hooks", "cask_auto_enter() {"]
    for path, instance in hooks.items():
        expanded = _expand_path(path)
        lines.append(f'  case "$PWD" in {expanded}*)')
        lines.append(f'    [ -z "$CONTAINER_ID" ] && distrobox enter {instance} ;;')
        lines.append("  esac")
    lines.extend(["}", 'PROMPT_COMMAND="cask_auto_enter;$PROMPT_COMMAND"'])
    return "\n".join(lines) + "\n"


def _fish_hook(hooks: dict[str, str]) -> str:
    lines = ["# Cask auto-enter hooks", "function cask_auto_enter --on-variable PWD"]
    for path, instance in hooks.items():
        expanded = _expand_path(path)
        lines.append(f'  if string match -q "{expanded}*" $PWD')
        lines.append(f"    if not set -q CONTAINER_ID")
        lines.append(f"      distrobox enter {instance}")
        lines.append("    end")
        lines.append("  end")
    lines.append("end")
    return "\n".join(lines) + "\n"
