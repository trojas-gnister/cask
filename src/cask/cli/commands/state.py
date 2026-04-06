"""cask state commands."""
from __future__ import annotations

import os
from typing import Optional

import typer
from rich.console import Console

from cask.config.paths import state_dir

console = Console()


def state_show():
    """Show current state."""
    from cask.state.manager import StateManager
    mgr = StateManager(state_dir())
    mgr.load()
    console.print("[bold]Current state:[/bold]")
    for section, s in mgr._state.sections.items():
        console.print(f"  {section}: hash={s.config_hash[:12]}... applied={s.applied} at={s.last_applied}")


def state_reset(sections: Optional[list[str]] = typer.Argument(None)):
    """Clear state to force re-sync."""
    from cask.state.manager import StateManager
    mgr = StateManager(state_dir())
    mgr.load()
    if sections:
        for s in sections:
            mgr.reset(s)
            console.print(f"  Reset {s}")
    else:
        mgr.reset()
        console.print("  Reset all state")
    mgr.save()


def state_generations():
    """List generation snapshots."""
    from cask.state.generations import GenerationManager
    gm = GenerationManager(os.path.join(state_dir(), "generations"))
    gens = gm.list_generations()
    if not gens:
        console.print("No generations yet")
    else:
        for g in gens:
            console.print(f"  {g}")
