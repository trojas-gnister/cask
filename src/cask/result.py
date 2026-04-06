"""Core result types."""
from dataclasses import dataclass, field


@dataclass
class ExecResult:
    """Result of a subprocess execution."""
    exit_code: int
    stdout: str
    stderr: str


@dataclass
class Result:
    """Result of a cask operation."""
    ok: bool
    message: str
    actions: list[str] = field(default_factory=list)
