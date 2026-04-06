"""cask list command."""
from __future__ import annotations

import asyncio
from typing import Optional

import typer
from rich.console import Console
from rich.table import Table

console = Console()


def list_cmd(
    sections: Optional[list[str]] = typer.Argument(None),
    all_resources: bool = typer.Option(False, "--all", help="Include undeclared"),
):
    """List managed resources."""
    from cask.cli.app import get_config

    cfg = get_config()

    async def _run():
        if cfg.pacman and (not sections or "pacman" in sections):
            table = Table(title="Pacman Packages")
            table.add_column("Package")
            for p in cfg.pacman.packages:
                table.add_row(p)
            for p in cfg.pacman.aur_packages:
                table.add_row(f"{p} (AUR)")
            console.print(table)

        if cfg.flatpak and (not sections or "flatpak" in sections):
            table = Table(title="Flatpak")
            table.add_column("App ID")
            for p in cfg.flatpak.packages:
                table.add_row(p)
            console.print(table)

        if cfg.podman and (not sections or "podman" in sections):
            table = Table(title="Containers")
            table.add_column("Name")
            table.add_column("Image")
            for name, c in cfg.podman.containers.items():
                table.add_row(name, c.image)
            console.print(table)

        if cfg.tools and (not sections or "tools" in sections):
            table = Table(title="Tools")
            table.add_column("Tool")
            table.add_column("Version")
            for tool, ver in cfg.tools.items():
                table.add_row(tool, ver)
            console.print(table)

    asyncio.run(_run())
