"""Typer CLI application."""
from __future__ import annotations

import typer
from rich.console import Console

import cask
from cask.config.paths import default_config_path
from cask.executor.system import SystemExecutor

app = typer.Typer(name="cask", help="Declarative package and container management.")
console = Console()

# Global state
_config_path: str = ""
_dry_run: bool = False
_verbose: bool = False
_yes: bool = False
_no: bool = False


@app.callback()
def main(
    config: str = typer.Option("", "-c", "--config", help="Config file path"),
    dry_run: bool = typer.Option(False, "--dry-run", help="Preview without applying"),
    verbose: bool = typer.Option(False, "-v", "--verbose", help="Verbose output"),
    yes: bool = typer.Option(False, "-y", "--yes", help="Auto-keep undeclared resources"),
    no: bool = typer.Option(False, "-n", "--no", help="Auto-remove undeclared resources"),
):
    global _config_path, _dry_run, _verbose, _yes, _no
    _config_path = config if config else default_config_path()
    _dry_run = dry_run
    _verbose = verbose
    _yes = yes
    _no = no


def get_config():
    from cask.config.loader import load_config
    return load_config(_config_path)


def get_executor():
    return SystemExecutor(dry_run=_dry_run)


def get_sync_options():
    from cask.sync.protocol import SyncOptions
    return SyncOptions(yes=_yes, no=_no, interactive=not (_yes or _no))


# Import command modules (must happen after app is defined)
from cask.cli.commands import sync as _sync_mod
from cask.cli.commands import diff as _diff_mod
from cask.cli.commands import add as _add_mod
from cask.cli.commands import remove as _remove_mod
from cask.cli.commands import list_cmd as _list_mod
from cask.cli.commands import update as _update_mod
from cask.cli.commands import lock as _lock_mod
from cask.cli.commands import validate as _validate_mod
from cask.cli.commands import state as _state_mod
from cask.cli.commands import devbox as _devbox_mod
from cask.cli.commands import config_cmd as _config_mod


# Register flat commands
app.command("sync")(_sync_mod.sync_cmd)
app.command("diff")(_diff_mod.diff_cmd)
app.command("add")(_add_mod.add_cmd)
app.command("remove")(_remove_mod.remove_cmd)
app.command("list")(_list_mod.list_cmd)
app.command("update")(_update_mod.update_cmd)
app.command("validate")(_validate_mod.validate_cmd)


@app.command("version")
def version_cmd():
    """Show cask version."""
    console.print(f"cask {cask.__version__}")


# lock sub-app
lock_app = typer.Typer(help="Version pinning")
lock_app.command("create")(_lock_mod.lock_create)
lock_app.command("verify")(_lock_mod.lock_verify)
lock_app.command("apply")(_lock_mod.lock_apply)
app.add_typer(lock_app, name="lock")

# state sub-app
state_app = typer.Typer(help="State management")
state_app.command("show")(_state_mod.state_show)
state_app.command("reset")(_state_mod.state_reset)
state_app.command("generations")(_state_mod.state_generations)
app.add_typer(state_app, name="state")

# devbox sub-app
devbox_app = typer.Typer(help="Devbox hook management")
devbox_app.command("install")(_devbox_mod.hooks_install)
devbox_app.command("remove")(_devbox_mod.hooks_remove)
app.add_typer(devbox_app, name="devbox")

# config sub-app
config_app = typer.Typer(help="Config management")
config_app.command("init")(_config_mod.config_init)
app.add_typer(config_app, name="config")


def run():
    app()
