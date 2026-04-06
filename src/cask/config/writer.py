"""Config file modification for add/remove commands."""
from __future__ import annotations

import tomllib
import tomli_w


def _load_raw(path: str) -> dict:
    with open(path, "rb") as f:
        return tomllib.load(f)


def _save(path: str, data: dict) -> None:
    with open(path, "wb") as f:
        tomli_w.dump(data, f)


def add_to_config(path: str, section: str, item: str) -> None:
    """Add an item to a config section's package list."""
    try:
        data = _load_raw(path)
    except Exception:
        data = {}

    # Map section to the right list key
    list_key = _section_list_key(section)

    # AUR packages live under [pacman] in the TOML, not under a separate [aur] table
    toml_section = "pacman" if section == "aur" else section

    if toml_section not in data:
        data[toml_section] = {}
    if list_key not in data[toml_section]:
        data[toml_section][list_key] = []

    if item not in data[toml_section][list_key]:
        data[toml_section][list_key].append(item)

    _save(path, data)


def remove_from_config(path: str, section: str, item: str) -> None:
    """Remove an item from a config section's package list."""
    try:
        data = _load_raw(path)
    except Exception:
        return

    list_key = _section_list_key(section)
    toml_section = "pacman" if section == "aur" else section

    if toml_section in data and list_key in data[toml_section]:
        data[toml_section][list_key] = [
            i for i in data[toml_section][list_key] if i != item
        ]
        _save(path, data)


def _section_list_key(section: str) -> str:
    """Map section name to its list key in the TOML."""
    if section in ("pacman", "flatpak"):
        return "packages"
    if section == "aur":
        return "aur_packages"
    return "packages"
