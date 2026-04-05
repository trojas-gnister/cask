"""cask config commands."""
from __future__ import annotations

import os
from typing import Optional

import typer
from rich.console import Console

console = Console()

_DEFAULT_CONFIG = """# Cask configuration
# See: https://github.com/trojas-gnister/cask

# [pacman]
# packages = ["firefox", "git"]
# aur_packages = []

# [flatpak]
# remotes = ["flathub"]
# packages = []

# [podman.containers.example]
# image = "nginx:latest"
# ports = ["8080:80"]

# [tools]
# node = "22.0.0"
"""


def config_init(
    config: Optional[str] = typer.Option(None, "-c", "--config", help="Config file path override"),
):
    """Generate default config file."""
    from cask.cli.app import _config_path
    from cask.config.paths import default_config_path

    # Prefer local --config arg, then global _config_path, then default
    if config:
        path = config
    elif _config_path:
        path = _config_path
    else:
        path = default_config_path()

    if os.path.exists(path):
        console.print(f"[yellow]Config already exists: {path}[/yellow]")
        return
    parent = os.path.dirname(path)
    if parent:
        os.makedirs(parent, exist_ok=True)
    with open(path, "w") as f:
        f.write(_DEFAULT_CONFIG)
    console.print(f"[green]Config written to {path}[/green]")
