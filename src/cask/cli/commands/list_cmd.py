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
    from cask.cli.app import get_config, get_executor

    cfg = get_config()
    executor = get_executor()

    async def _run():
        if cfg.pacman and (not sections or "pacman" in sections):
            table = Table(title="Pacman Packages")
            table.add_column("Package")
            table.add_column("Status")
            declared = set(cfg.pacman.packages)
            for p in cfg.pacman.packages:
                table.add_row(p, "")
            for p in cfg.pacman.aur_packages:
                table.add_row(f"{p} (AUR)", "")

            if all_resources:
                from cask.managers.pacman import PacmanManager
                installed = await PacmanManager().list_installed(executor)
                for pkg in sorted(installed):
                    if pkg not in declared:
                        table.add_row(pkg, "[yellow](undeclared)[/yellow]")

            console.print(table)

        if cfg.flatpak and (not sections or "flatpak" in sections):
            table = Table(title="Flatpak")
            table.add_column("App ID")
            table.add_column("Status")
            declared = set(cfg.flatpak.packages)
            for p in cfg.flatpak.packages:
                table.add_row(p, "")

            if all_resources:
                from cask.managers.flatpak import FlatpakManager
                installed = await FlatpakManager().list_installed(executor)
                for app_id in sorted(installed):
                    if app_id not in declared:
                        table.add_row(app_id, "[yellow](undeclared)[/yellow]")

            console.print(table)

        if cfg.podman and (not sections or "podman" in sections):
            table = Table(title="Containers")
            table.add_column("Name")
            table.add_column("Image")
            table.add_column("Status")
            declared = set(cfg.podman.containers)
            for name, c in cfg.podman.containers.items():
                table.add_row(name, c.image, "")

            if all_resources:
                from cask.managers.podman import PodmanManager
                running = await PodmanManager().list_containers(executor)
                for name, image in sorted(running.items()):
                    if name not in declared:
                        table.add_row(name, image, "[yellow](undeclared)[/yellow]")

            console.print(table)
        elif all_resources and (not sections or "podman" in sections):
            # podman section not in config but --all requested: still show undeclared
            from cask.managers.podman import PodmanManager
            running = await PodmanManager().list_containers(executor)
            if running:
                table = Table(title="Containers")
                table.add_column("Name")
                table.add_column("Image")
                table.add_column("Status")
                for name, image in sorted(running.items()):
                    table.add_row(name, image, "[yellow](undeclared)[/yellow]")
                console.print(table)

        if cfg.tools and (not sections or "tools" in sections):
            table = Table(title="Tools")
            table.add_column("Tool")
            table.add_column("Version")
            for tool, ver in cfg.tools.items():
                table.add_row(tool, ver)
            console.print(table)

    asyncio.run(_run())
