"""cask add command."""
from __future__ import annotations

import asyncio
from typing import Optional

import typer
from rich.console import Console

console = Console()


def add_cmd(
    section: str = typer.Argument(..., help="Resource type (pacman, aur, flatpak, tool)"),
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
            for item in items:
                add_to_config(_config_path, section, item)
                console.print(f"  Added {item} to config")

            if section == "pacman":
                result = await PacmanManager().install(items, executor)
                console.print(f"  {'[green]OK[/green]' if result.ok else '[red]FAIL[/red]'} {result.message}")
            elif section == "aur":
                result = await AURManager().install(items, executor)
                console.print(f"  {'[green]OK[/green]' if result.ok else '[red]FAIL[/red]'} {result.message}")
            elif section == "flatpak":
                mgr = FlatpakManager()
                for item in items:
                    result = await mgr.install(item, "flathub", executor)
                    console.print(f"  {'[green]OK[/green]' if result.ok else '[red]FAIL[/red]'} {result.message}")
        elif section == "tool":
            if len(items) >= 2:
                from cask.managers.mise import MiseManager
                result = await MiseManager().install(items[0], items[1], executor)
                console.print(f"  {'[green]OK[/green]' if result.ok else '[red]FAIL[/red]'} {result.message}")
            else:
                console.print("[red]tool requires <name> <version>[/red]")
        else:
            console.print(f"[red]Unknown section: {section}[/red]")

    asyncio.run(_run())
