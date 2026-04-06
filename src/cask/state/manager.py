"""State persistence manager."""
from __future__ import annotations

import json
import os

from cask.state.models import SectionState, GlobalState


class StateManager:
    """Manages state persistence to JSON."""

    def __init__(self, state_dir: str) -> None:
        self._path = os.path.join(state_dir, "global.json")
        self._state = GlobalState()

    def load(self) -> None:
        if os.path.exists(self._path):
            with open(self._path) as f:
                data = json.load(f)
            self._state = GlobalState(
                sections={
                    k: SectionState(**v) for k, v in data.get("sections", {}).items()
                }
            )

    def save(self) -> None:
        os.makedirs(os.path.dirname(self._path), exist_ok=True)
        data = {
            "sections": {
                k: {"config_hash": v.config_hash, "applied": v.applied, "last_applied": v.last_applied}
                for k, v in self._state.sections.items()
            }
        }
        with open(self._path, "w") as f:
            json.dump(data, f, indent=2)

    def has_changed(self, section: str, config_hash: str) -> bool:
        s = self._state.sections.get(section)
        if not s:
            return True
        return s.config_hash != config_hash

    def mark_applied(self, section: str, config_hash: str) -> None:
        self._state.sections[section] = SectionState(config_hash=config_hash)

    def reset(self, section: str | None = None) -> None:
        if section:
            self._state.sections.pop(section, None)
        else:
            self._state.sections.clear()
