"""State data models."""
from __future__ import annotations

from dataclasses import dataclass, field
from datetime import datetime


@dataclass
class SectionState:
    """State of a single config section."""
    config_hash: str
    applied: bool = True
    last_applied: str = ""

    def __post_init__(self):
        if not self.last_applied:
            self.last_applied = datetime.now().isoformat()


@dataclass
class GlobalState:
    """Full application state."""
    sections: dict[str, SectionState] = field(default_factory=dict)
