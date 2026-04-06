"""cask sync command."""
from __future__ import annotations

import asyncio
import os
from typing import Optional

import typer
from rich.console import Console

console = Console()


def sync_cmd(sections: Optional[list[str]] = typer.Argument(None)):
    """Bidirectional sync of declared resources."""
    from cask.cli.app import get_config, get_executor, get_sync_options
    from cask.sync.algorithm import run_sync
    from cask.sync.flatpak import FlatpakSync
    from cask.sync.containers import ContainerSync
    from cask.sync.devbox import DevboxSync
    from cask.sync.tools import ToolSync
    from cask.sync.pacman import PacmanSync, AURSync
    from cask.state.manager import StateManager
    from cask.state.generations import GenerationManager
    from cask.state.hashing import hash_section
    from cask.config.paths import state_dir

    cfg = get_config()
    executor = get_executor()
    opts = get_sync_options()

    sdir = state_dir()
    state_mgr = StateManager(sdir)
    state_mgr.load()
    gen_mgr = GenerationManager(os.path.join(sdir, "generations"))

    async def _run():
        total_applied = total_updated = total_removed = total_kept = total_failed = 0

        syncs = []
        if cfg.pacman and cfg.pacman.packages and (not sections or "pacman" in sections):
            syncs.append(("pacman", "Pacman", PacmanSync(), cfg.pacman))
        if cfg.pacman and cfg.pacman.aur_packages and (not sections or "aur" in sections):
            syncs.append(("aur", "AUR", AURSync(), cfg.pacman))
        if cfg.flatpak and (not sections or "flatpak" in sections):
            syncs.append(("flatpak", "Flatpak", FlatpakSync(), cfg.flatpak))
        if cfg.podman and (not sections or "podman" in sections):
            syncs.append(("podman", "Podman", ContainerSync(), cfg.podman))
        if cfg.devbox and (not sections or "devbox" in sections):
            syncs.append(("devbox", "Devbox", DevboxSync(), cfg.devbox))
        if cfg.tools and (not sections or "tools" in sections):
            syncs.append(("tools", "Tools", ToolSync(), cfg.tools))

        any_synced = False
        for section_name, label, sync, section_cfg in syncs:
            section_hash = hash_section(section_cfg.model_dump() if hasattr(section_cfg, "model_dump") else section_cfg)
            if not state_mgr.has_changed(section_name, section_hash):
                console.print(f"\n[bold]{label}:[/bold] [dim]unchanged, skipping[/dim]")
                continue

            console.print(f"\n[bold]{label}:[/bold]")
            stats = await run_sync(sync, section_cfg, executor, opts)
            total_applied += stats.applied
            total_updated += stats.updated
            total_removed += stats.removed
            total_kept += stats.kept
            total_failed += stats.failed
            console.print(f"  {stats.applied} applied, {stats.updated} updated, "
                         f"{stats.removed} removed, {stats.kept} kept")

            if stats.failed == 0:
                state_mgr.mark_applied(section_name, section_hash)
                any_synced = True

        console.print(f"\n[bold]Total:[/bold] {total_applied} applied, {total_updated} updated, "
                     f"{total_removed} removed, {total_kept} kept, {total_failed} failed")

        if any_synced:
            state_mgr.save()
            state_data = {
                k: {"config_hash": v.config_hash, "applied": v.applied, "last_applied": v.last_applied}
                for k, v in state_mgr._state.sections.items()
            }
            gen_name = gen_mgr.create(state_data)
            console.print(f"[dim]State saved. Generation: {gen_name}[/dim]")

    asyncio.run(_run())
