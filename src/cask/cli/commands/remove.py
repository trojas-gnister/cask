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
            remove_from_config(_config_path, section, item)

            if section == "pacman":
                result = await PacmanManager().remove(item, executor)
                console.print(f"  {'[green]OK[/green]' if result.ok else '[red]FAIL[/red]'} {result.message}")
            elif section == "flatpak":
                result = await FlatpakManager().remove(item, executor)
                console.print(f"  {'[green]OK[/green]' if result.ok else '[red]FAIL[/red]'} {result.message}")
            else:
                console.print(f"[yellow]Removed {item} from config[/yellow]")

    asyncio.run(_run())
