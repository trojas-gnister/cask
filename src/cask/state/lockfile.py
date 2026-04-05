"""Version pinning lockfile."""
from __future__ import annotations

import json
import os


class Lockfile:
    """Pin exact package versions for reproducibility."""

    def __init__(self, path: str) -> None:
        self._path = path
        self._pins: dict[str, str] = {}

    def load(self) -> None:
        if os.path.exists(self._path):
            with open(self._path) as f:
                self._pins = json.load(f)

    def save(self) -> None:
        os.makedirs(os.path.dirname(self._path), exist_ok=True)
        with open(self._path, "w") as f:
            json.dump(self._pins, f, indent=2, sort_keys=True)

    def pin(self, name: str, version: str) -> None:
        self._pins[name] = version

    def get(self, name: str) -> str | None:
        return self._pins.get(name)

    def verify(self, name: str, version: str) -> bool:
        pinned = self._pins.get(name)
        if pinned is None:
            return True  # Not pinned = OK
        return pinned == version

    @property
    def pins(self) -> dict[str, str]:
        return dict(self._pins)
