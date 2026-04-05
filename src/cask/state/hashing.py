"""SHA256 hashing for config change detection."""
from __future__ import annotations

import hashlib
import json


def hash_section(data: object) -> str:
    """Return SHA256 hex digest of a config section (deterministic)."""
    serialized = json.dumps(data, sort_keys=True, default=str)
    return hashlib.sha256(serialized.encode()).hexdigest()
