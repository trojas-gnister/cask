"""cask devbox commands."""
from __future__ import annotations

import os

from rich.console import Console

console = Console()

HOOK_MARKER = "# Cask auto-enter hooks"


def hooks_install():
    """Install shell auto-enter hooks."""
    from cask.cli.app import get_config
    from cask.devbox.hooks import generate_hook

    cfg = get_config()
    if not cfg.devbox or not cfg.devbox.hooks:
        console.print("[yellow]No devbox hooks configured[/yellow]")
        return

    shell = os.environ.get("SHELL", "").split("/")[-1]
    if shell not in ("bash", "zsh", "fish"):
        console.print(f"[red]Unsupported shell: {shell}[/red]")
        return

    hook_code = generate_hook(shell, cfg.devbox.hooks)
    rc_file = {"bash": "~/.bashrc", "zsh": "~/.zshrc", "fish": "~/.config/fish/config.fish"}[shell]
    rc_path = os.path.expanduser(rc_file)

    # Check for existing hooks
    existing = ""
    if os.path.exists(rc_path):
        with open(rc_path) as f:
            existing = f.read()
        if HOOK_MARKER in existing:
            console.print("[yellow]Hooks already installed — remove first with 'cask devbox remove'[/yellow]")
            return

    with open(rc_path, "a") as f:
        f.write(f"\n{hook_code}")
    console.print(f"[green]Hooks installed to {rc_file}[/green]")


def hooks_remove():
    """Remove shell auto-enter hooks."""
    console.print("[yellow]Manual removal required — remove the cask_auto_enter block from your shell rc file[/yellow]")
