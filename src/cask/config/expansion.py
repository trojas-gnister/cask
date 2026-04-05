"""Environment variable and tilde expansion for config values."""
from __future__ import annotations

import os
import re

_ENV_PATTERN = re.compile(r"\$\{([^}]+)\}")


def _expand_env_match(match: re.Match) -> str:
    expr = match.group(1)
    if ":-" in expr:
        name, default = expr.split(":-", 1)
        return os.environ.get(name, default)
    return os.environ.get(expr, "")


def expand_value(value: str) -> str:
    """Expand ~ and ${VAR} / ${VAR:-default} in a string."""
    if "~" in value:
        value = os.path.expanduser(value)
    if "${" in value:
        value = _ENV_PATTERN.sub(_expand_env_match, value)
    return value


def expand_config(data: dict | list | str | object) -> dict | list | str | object:
    """Recursively expand all string values in a config dict/list."""
    if isinstance(data, dict):
        return {k: expand_config(v) for k, v in data.items()}
    if isinstance(data, list):
        return [expand_config(item) for item in data]
    if isinstance(data, str):
        return expand_value(data)
    return data
