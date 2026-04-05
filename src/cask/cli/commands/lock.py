"""cask lock commands."""
from __future__ import annotations

import asyncio

from rich.console import Console

from cask.config.paths import state_dir

console = Console()


def lock_create():
    """Pin current versions to lockfile."""
    import os
    from cask.cli.app import get_config, get_executor
    from cask.state.lockfile import Lockfile
    from cask.managers.pacman import PacmanManager
    from cask.managers.flatpak import FlatpakManager

    cfg = get_config()
    executor = get_executor()
    lf = Lockfile(os.path.join(state_dir(), "lock.json"))

    async def _run():
        if cfg.pacman:
            installed = await PacmanManager().list_installed(executor)
            for pkg in cfg.pacman.packages:
                if pkg in installed:
                    lf.pin(pkg, installed[pkg])
        if cfg.flatpak:
            installed = await FlatpakManager().list_installed(executor)
            for pkg in cfg.flatpak.packages:
                if pkg in installed:
                    lf.pin(pkg, installed[pkg])
        lf.save()
        console.print(f"Locked {len(lf.pins)} packages")

    asyncio.run(_run())


def lock_verify():
    """Verify installed versions match lockfile."""
    import os
    from cask.cli.app import get_executor
    from cask.state.lockfile import Lockfile
    from cask.managers.pacman import PacmanManager

    executor = get_executor()
    lf = Lockfile(os.path.join(state_dir(), "lock.json"))
    lf.load()

    async def _run():
        installed = await PacmanManager().list_installed(executor)
        mismatches = 0
        for name, pinned in lf.pins.items():
            actual = installed.get(name, "not installed")
            if actual != pinned:
                console.print(f"  [red]MISMATCH[/red] {name}: {pinned} -> {actual}")
                mismatches += 1
        if mismatches == 0:
            console.print("[green]All versions match lockfile[/green]")

    asyncio.run(_run())


def lock_apply():
    """Install exact locked versions."""
    console.print("[yellow]lock apply not yet implemented[/yellow]")
