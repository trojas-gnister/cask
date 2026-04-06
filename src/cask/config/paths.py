"""XDG-compliant path resolution."""
from __future__ import annotations

import os


def config_dir() -> str:
    """Return the cask config directory (~/.config/cask)."""
    xdg = os.environ.get("XDG_CONFIG_HOME", os.path.expanduser("~/.config"))
    return os.path.join(xdg, "cask")


def state_dir() -> str:
    """Return the cask state directory (~/.config/cask/state)."""
    return os.path.join(config_dir(), "state")


def default_config_path() -> str:
    """Return the default config file path."""
    return os.path.join(config_dir(), "config.toml")
