"""Generation snapshot management."""
from __future__ import annotations

import json
import os
from datetime import datetime


class GenerationManager:
    """Manages config+state snapshots per successful sync."""

    def __init__(self, gen_dir: str) -> None:
        self._dir = gen_dir

    def create(self, state_data: dict) -> str:
        os.makedirs(self._dir, exist_ok=True)
        gens = self.list_generations()
        gen_num = len(gens) + 1
        name = f"gen-{gen_num:03d}.json"
        path = os.path.join(self._dir, name)
        snapshot = {
            "generation": gen_num,
            "timestamp": datetime.now().isoformat(),
            "state": state_data,
        }
        with open(path, "w") as f:
            json.dump(snapshot, f, indent=2)
        return name

    def list_generations(self) -> list[str]:
        if not os.path.exists(self._dir):
            return []
        files = sorted(f for f in os.listdir(self._dir) if f.startswith("gen-"))
        return files
