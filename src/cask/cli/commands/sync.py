"""cask sync command."""
from __future__ import annotations

import asyncio
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

    cfg = get_config()
    executor = get_executor()
    opts = get_sync_options()

    async def _run():
        total_applied = total_updated = total_removed = total_kept = total_failed = 0

        syncs = []
        if cfg.flatpak and (not sections or "flatpak" in sections):
            syncs.append(("Flatpak", FlatpakSync(), cfg.flatpak))
        if cfg.podman and (not sections or "podman" in sections):
            syncs.append(("Podman", ContainerSync(), cfg.podman))
        if cfg.devbox and (not sections or "devbox" in sections):
            syncs.append(("Devbox", DevboxSync(), cfg.devbox))
        if cfg.tools and (not sections or "tools" in sections):
            syncs.append(("Tools", ToolSync(), cfg.tools))

        for name, sync, section_cfg in syncs:
            console.print(f"\n[bold]{name}:[/bold]")
            stats = await run_sync(sync, section_cfg, executor, opts)
            total_applied += stats.applied
            total_updated += stats.updated
            total_removed += stats.removed
            total_kept += stats.kept
            total_failed += stats.failed
            console.print(f"  {stats.applied} applied, {stats.updated} updated, "
                         f"{stats.removed} removed, {stats.kept} kept")

        console.print(f"\n[bold]Total:[/bold] {total_applied} applied, {total_updated} updated, "
                     f"{total_removed} removed, {total_kept} kept, {total_failed} failed")

    asyncio.run(_run())
