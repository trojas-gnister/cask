"""TOML config loading with include support and env expansion."""
from __future__ import annotations

import os
import tomllib
from pathlib import Path

from cask.config.expansion import expand_config
from cask.config.models import CaskConfig


def _deep_merge(base: dict, overlay: dict) -> dict:
    """Recursively merge overlay into base. overlay wins on conflict."""
    result = dict(base)
    for key, value in overlay.items():
        if key in result and isinstance(result[key], dict) and isinstance(value, dict):
            result[key] = _deep_merge(result[key], value)
        else:
            result[key] = value
    return result


def _load_toml(path: str) -> dict:
    """Load a single TOML file."""
    with open(path, "rb") as f:
        return tomllib.load(f)


def load_config(path: str) -> CaskConfig:
    """Load config from a TOML file, processing includes and env expansion."""
    if not os.path.exists(path):
        raise FileNotFoundError(f"Config file not found: {path}")

    data = _load_toml(path)
    base_dir = os.path.dirname(os.path.abspath(path))

    # Process includes
    includes = data.pop("include", [])
    for inc_path in includes:
        if not os.path.isabs(inc_path):
            inc_path = os.path.join(base_dir, inc_path)
        if os.path.exists(inc_path):
            inc_data = _load_toml(inc_path)
            data = _deep_merge(data, inc_data)

    # Expand environment variables and tildes
    data = expand_config(data)

    return CaskConfig.model_validate(data)
