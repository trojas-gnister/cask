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
    from cask.cli.app import get_config, get_executor
    from cask.state.lockfile import Lockfile
    from cask.managers.pacman import PacmanManager
    from cask.managers.flatpak import FlatpakManager

    executor = get_executor()
    lf = Lockfile(os.path.join(state_dir(), "lock.json"))
    lf.load()

    if not lf.pins:
        console.print("[yellow]No lockfile found or empty[/yellow]")
        return

    cfg = get_config()
    flatpak_apps = set(cfg.flatpak.packages) if cfg.flatpak else set()

    async def _run():
        pacman_installed = await PacmanManager().list_installed(executor)
        flatpak_installed = await FlatpakManager().list_installed(executor)

        mismatches = 0
        for name, pinned in lf.pins.items():
            if name in flatpak_apps:
                actual = flatpak_installed.get(name, "not installed")
            else:
                actual = pacman_installed.get(name, "not installed")

            if actual != pinned:
                console.print(f"  [red]MISMATCH[/red] {name}: locked={pinned} actual={actual}")
                mismatches += 1
            else:
                console.print(f"  [green]OK[/green] {name}: {pinned}")

        if mismatches == 0:
            console.print("\n[green]All versions match lockfile[/green]")
        else:
            console.print(f"\n[red]{mismatches} mismatch(es)[/red]")

    asyncio.run(_run())


def lock_apply():
    """Install exact locked versions."""
    import os
    from cask.cli.app import get_config, get_executor
    from cask.state.lockfile import Lockfile
    from cask.managers.pacman import PacmanManager
    from cask.managers.flatpak import FlatpakManager

    cfg = get_config()
    executor = get_executor()
    lf = Lockfile(os.path.join(state_dir(), "lock.json"))
    lf.load()

    if not lf.pins:
        console.print("[yellow]No lockfile found or lockfile is empty. Run 'cask lock create' first.[/yellow]")
        return

    # Determine which pinned packages belong to flatpak
    flatpak_pkgs: set[str] = set()
    if cfg.flatpak:
        flatpak_pkgs = set(cfg.flatpak.packages)

    async def _run():
        mgr = PacmanManager()
        installed_pacman = await mgr.list_installed(executor)
        installed_flatpak = await FlatpakManager().list_installed(executor)

        to_install: list[tuple[str, str]] = []
        for name, pinned_version in lf.pins.items():
            if name in flatpak_pkgs:
                # Flatpak: report mismatch as a warning — exact version pinning not supported
                actual = installed_flatpak.get(name, "not installed")
                if actual != pinned_version:
                    console.print(
                        f"  [yellow]WARNING[/yellow] flatpak {name}: "
                        f"pinned={pinned_version}, installed={actual} "
                        f"(flatpak does not support exact version installs)"
                    )
                else:
                    console.print(f"  {name}: [green]{actual}[/green] (ok)")
            else:
                actual = installed_pacman.get(name)
                if actual != pinned_version:
                    to_install.append((name, pinned_version))
                    status = f"[yellow]{actual}[/yellow]" if actual else "[yellow]not installed[/yellow]"
                    console.print(f"  {name}: {status} -> [cyan]{pinned_version}[/cyan]")
                else:
                    console.print(f"  {name}: [green]{actual}[/green] (ok)")

        if not to_install:
            console.print("[green]All pinned pacman versions already installed[/green]")
            return

        console.print(f"\nInstalling {len(to_install)} pinned package(s)...")
        for name, version in to_install:
            pkg_spec = f"{name}={version}"
            r = await executor.execute_sudo(["pacman", "-S", "--noconfirm", pkg_spec])
            if r.exit_code == 0:
                console.print(f"  [green]OK[/green] installed {name}={version}")
            else:
                console.print(f"  [red]FAIL[/red] {name}={version}: {r.stderr}")

    asyncio.run(_run())
