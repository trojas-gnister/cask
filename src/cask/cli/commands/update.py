"""cask update command."""
from __future__ import annotations

import asyncio
from typing import Optional

import typer
from rich.console import Console

console = Console()


def update_cmd(sections: Optional[list[str]] = typer.Argument(None)):
    """Update all resources to latest versions."""
    from cask.cli.app import get_config, get_executor
    from cask.managers.pacman import PacmanManager
    from cask.managers.flatpak import FlatpakManager

    cfg = get_config()
    executor = get_executor()

    async def _run():
        if cfg.pacman and (not sections or "pacman" in sections):
            await PacmanManager().sync_db(executor)
            result = await PacmanManager().install(cfg.pacman.packages, executor)
            console.print(f"  Pacman: {result.message}")

        if cfg.flatpak and (not sections or "flatpak" in sections):
            r = await executor.execute_sudo(["flatpak", "update", "-y"])
            console.print(f"  Flatpak: {'updated' if r.exit_code == 0 else 'failed'}")

        if cfg.podman and (not sections or "podman" in sections):
            for name, container in cfg.podman.containers.items():
                r = await executor.execute(["podman", "pull", container.image])
                status = "pulled" if r.exit_code == 0 else "failed"
                console.print(f"  Podman [{name}]: {status}")

        if cfg.devbox and (not sections or "devbox" in sections):
            for name in cfg.devbox.instances:
                r = await executor.execute(["distrobox", "upgrade", name])
                status = "upgraded" if r.exit_code == 0 else "failed"
                console.print(f"  Devbox [{name}]: {status}")

        if cfg.tools and (not sections or "tools" in sections):
            for tool in cfg.tools:
                r = await executor.execute(["mise", "install", f"{tool}@latest"])
                if r.exit_code == 0:
                    await executor.execute(["mise", "use", "--global", f"{tool}@latest"])
                    console.print(f"  Tools [{tool}]: updated to latest")
                else:
                    console.print(f"  Tools [{tool}]: failed")

    asyncio.run(_run())
