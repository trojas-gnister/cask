"""cask diff command."""
from __future__ import annotations

import asyncio
from typing import Optional

import typer
from rich.console import Console

console = Console()


def diff_cmd(sections: Optional[list[str]] = typer.Argument(None)):
    """Preview what sync would change."""
    from cask.cli.app import get_config, get_executor
    from cask.sync.flatpak import FlatpakSync
    from cask.sync.containers import ContainerSync
    from cask.sync.devbox import DevboxSync
    from cask.sync.tools import ToolSync
    from cask.sync.pacman import PacmanSync, AURSync

    cfg = get_config()
    executor = get_executor()

    async def _run():
        diffs = []
        if cfg.pacman and cfg.pacman.packages and (not sections or "pacman" in sections):
            diffs.append(("Pacman", PacmanSync(), cfg.pacman))
        if cfg.pacman and cfg.pacman.aur_packages and (not sections or "aur" in sections):
            diffs.append(("AUR", AURSync(), cfg.pacman))
        if cfg.flatpak and (not sections or "flatpak" in sections):
            diffs.append(("Flatpak", FlatpakSync(), cfg.flatpak))
        if cfg.podman and (not sections or "podman" in sections):
            diffs.append(("Podman", ContainerSync(), cfg.podman))
        if cfg.devbox and (not sections or "devbox" in sections):
            diffs.append(("Devbox", DevboxSync(), cfg.devbox))
        if cfg.tools and (not sections or "tools" in sections):
            diffs.append(("Tools", ToolSync(), cfg.tools))

        for name, sync, section_cfg in diffs:
            host = await sync.get_host_resources(executor)
            config = sync.get_config_resources(section_cfg)
            host_map = {sync.resource_id(r): r for r in host}
            config_map = {sync.resource_id(r): r for r in config}

            to_apply = set(config_map) - set(host_map)
            common = set(config_map) & set(host_map)
            undeclared = set(host_map) - set(config_map)
            to_update = {rid for rid in common if sync.needs_update(host_map[rid], config_map[rid])}

            if to_apply or to_update or undeclared:
                console.print(f"\n[bold]{name}:[/bold]")
                for rid in sorted(to_apply):
                    console.print(f"  [green]+[/green] {rid}")
                for rid in sorted(to_update):
                    console.print(f"  [yellow]~[/yellow] {rid}")
                for rid in sorted(undeclared):
                    console.print(f"  [dim]?[/dim] {rid}")

    asyncio.run(_run())
