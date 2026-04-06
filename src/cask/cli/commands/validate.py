"""cask validate command."""
from __future__ import annotations

from typing import Optional

import typer
from rich.console import Console

console = Console()


def validate_cmd(
    config: Optional[str] = typer.Option(None, "-c", "--config", help="Config file path override"),
):
    """Validate config file."""
    from cask.cli.app import get_config, _config_path
    from cask.config.validation import validate_config
    from cask.config.loader import load_config

    try:
        path = config if config else _config_path
        cfg = load_config(path)
        errors = validate_config(cfg)
        if errors:
            for e in errors:
                console.print(f"  [red]ERROR[/red] {e}")
        else:
            console.print(f"[green]Config OK:[/green] {path}")
    except Exception as e:
        console.print(f"[red]Config error:[/red] {e}")
