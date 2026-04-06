"""cask remove command."""
from __future__ import annotations

import asyncio

import typer
from rich.console import Console

console = Console()


def remove_cmd(
    section: str = typer.Argument(..., help="Resource type"),
    items: list[str] = typer.Argument(..., help="Items to remove"),
):
    """Remove resources from config and uninstall."""
    from cask.cli.app import get_executor, _config_path
    from cask.config.writer import remove_from_config
    from cask.managers.pacman import PacmanManager
    from cask.managers.flatpak import FlatpakManager

    executor = get_executor()

    async def _run():
        for item in items:
            if section == "pacman":
                remove_from_config(_config_path, section, item)
                result = await PacmanManager().remove(item, executor)
                console.print(f"  {'[green]OK[/green]' if result.ok else '[red]FAIL[/red]'} {result.message}")
            elif section == "flatpak":
                remove_from_config(_config_path, section, item)
                result = await FlatpakManager().remove(item, executor)
                console.print(f"  {'[green]OK[/green]' if result.ok else '[red]FAIL[/red]'} {result.message}")
            elif section == "aur":
                remove_from_config(_config_path, section, item)
                from cask.managers.aur import AURManager
                result = await AURManager().remove(item, executor)
                console.print(f"  {'[green]OK[/green]' if result.ok else '[red]FAIL[/red]'} {result.message}")
            elif section == "tool":
                from cask.managers.mise import MiseManager
                from cask.config.writer import _load_raw, _save
                # Remove from config tools table
                try:
                    data = _load_raw(_config_path)
                except Exception:
                    data = {}
                if "tools" in data and item in data["tools"]:
                    del data["tools"][item]
                    _save(_config_path, data)
                console.print(f"  Removed {item} from config")
                result = await MiseManager().remove(item, executor)
                console.print(f"  {'[green]OK[/green]' if result.ok else '[red]FAIL[/red]'} {result.message}")
            elif section == "container":
                from cask.managers.podman import PodmanManager
                from cask.config.writer import _load_raw, _save
                # Remove from config
                try:
                    data = _load_raw(_config_path)
                except Exception:
                    data = {}
                if "podman" in data and "containers" in data["podman"] and item in data["podman"]["containers"]:
                    del data["podman"]["containers"][item]
                    _save(_config_path, data)
                console.print(f"  Removed container {item} from config")
                result = await PodmanManager().remove(item, executor)
                console.print(f"  {'[green]OK[/green]' if result.ok else '[red]FAIL[/red]'} {result.message}")
            elif section == "devbox":
                from cask.managers.distrobox import DistroboxManager
                from cask.config.writer import _load_raw, _save
                # Remove from config
                try:
                    data = _load_raw(_config_path)
                except Exception:
                    data = {}
                if "devbox" in data and "instances" in data["devbox"] and item in data["devbox"]["instances"]:
                    del data["devbox"]["instances"][item]
                    _save(_config_path, data)
                console.print(f"  Removed devbox instance {item} from config")
                result = await DistroboxManager().remove(item, executor)
                console.print(f"  {'[green]OK[/green]' if result.ok else '[red]FAIL[/red]'} {result.message}")
            else:
                console.print(f"[yellow]Unknown section: {section}[/yellow]")

    asyncio.run(_run())
