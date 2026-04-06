"""cask add command."""
from __future__ import annotations

import asyncio
from typing import Optional

import typer
from rich.console import Console

console = Console()


def add_cmd(
    section: str = typer.Argument(..., help="Resource type (pacman, aur, flatpak, tool, container, devbox)"),
    items: list[str] = typer.Argument(..., help="Items to add"),
    image: Optional[str] = typer.Option(None, "--image", help="Image for container/devbox"),
):
    """Add resources to config and install."""
    from cask.cli.app import get_executor, _config_path
    from cask.config.writer import add_to_config
    from cask.managers.pacman import PacmanManager
    from cask.managers.aur import AURManager
    from cask.managers.flatpak import FlatpakManager

    executor = get_executor()

    async def _run():
        if section in ("pacman", "aur", "flatpak"):
            if section == "pacman":
                result = await PacmanManager().install(items, executor)
                console.print(f"  {'[green]OK[/green]' if result.ok else '[red]FAIL[/red]'} {result.message}")
                if result.ok:
                    for item in items:
                        add_to_config(_config_path, section, item)
                        console.print(f"  Added {item} to config")
            elif section == "aur":
                result = await AURManager().install(items, executor)
                console.print(f"  {'[green]OK[/green]' if result.ok else '[red]FAIL[/red]'} {result.message}")
                if result.ok:
                    for item in items:
                        add_to_config(_config_path, section, item)
                        console.print(f"  Added {item} to config")
            elif section == "flatpak":
                mgr = FlatpakManager()
                for item in items:
                    result = await mgr.install(item, "flathub", executor)
                    console.print(f"  {'[green]OK[/green]' if result.ok else '[red]FAIL[/red]'} {result.message}")
                    if result.ok:
                        add_to_config(_config_path, section, item)
                        console.print(f"  Added {item} to config")
        elif section == "tool":
            if len(items) >= 2:
                from cask.managers.mise import MiseManager
                from cask.config.writer import _load_raw, _save
                name, version = items[0], items[1]
                result = await MiseManager().install(name, version, executor)
                console.print(f"  {'[green]OK[/green]' if result.ok else '[red]FAIL[/red]'} {result.message}")
                if result.ok:
                    try:
                        data = _load_raw(_config_path)
                    except Exception:
                        data = {}
                    if "tools" not in data:
                        data["tools"] = {}
                    data["tools"][name] = version
                    _save(_config_path, data)
                    console.print(f"  Added {name}@{version} to config")
            else:
                console.print("[red]tool requires <name> <version>[/red]")
        elif section == "container":
            if not image:
                console.print("[red]container requires --image <image>[/red]")
                return
            if not items:
                console.print("[red]container requires a name[/red]")
                return
            from cask.managers.podman import PodmanManager
            from cask.config.models import ContainerConfig
            from cask.config.writer import _load_raw, _save
            name = items[0]
            cfg = ContainerConfig(image=image)
            result = await PodmanManager().create(name, cfg, executor)
            console.print(f"  {'[green]OK[/green]' if result.ok else '[red]FAIL[/red]'} {result.message}")
            if result.ok:
                try:
                    data = _load_raw(_config_path)
                except Exception:
                    data = {}
                if "podman" not in data:
                    data["podman"] = {}
                if "containers" not in data["podman"]:
                    data["podman"]["containers"] = {}
                data["podman"]["containers"][name] = {"image": image}
                _save(_config_path, data)
                console.print(f"  Added container {name} ({image}) to config")
        elif section == "devbox":
            if not image:
                console.print("[red]devbox requires --image <image>[/red]")
                return
            if not items:
                console.print("[red]devbox requires a name[/red]")
                return
            from cask.managers.distrobox import DistroboxManager
            from cask.config.models import DevboxInstanceConfig
            from cask.config.writer import _load_raw, _save
            name = items[0]
            cfg = DevboxInstanceConfig(image=image)
            result = await DistroboxManager().create(name, cfg, executor)
            console.print(f"  {'[green]OK[/green]' if result.ok else '[red]FAIL[/red]'} {result.message}")
            if result.ok:
                try:
                    data = _load_raw(_config_path)
                except Exception:
                    data = {}
                if "devbox" not in data:
                    data["devbox"] = {}
                if "instances" not in data["devbox"]:
                    data["devbox"]["instances"] = {}
                data["devbox"]["instances"][name] = {"image": image}
                _save(_config_path, data)
                console.print(f"  Added devbox instance {name} ({image}) to config")
        else:
            console.print(f"[red]Unknown section: {section}[/red]")

    asyncio.run(_run())
