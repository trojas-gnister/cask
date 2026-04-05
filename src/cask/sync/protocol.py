"""Sync protocol and data types."""
from __future__ import annotations

from dataclasses import dataclass, field
from typing import Protocol, TypeVar, Generic

from cask.executor.protocol import Executor
from cask.result import Result

T = TypeVar("T")


@dataclass
class SyncOptions:
    """Options for sync behavior."""
    yes: bool = False       # Keep all undeclared
    no: bool = False        # Remove all undeclared
    interactive: bool = True


@dataclass
class SyncStats:
    """Statistics from a sync run."""
    applied: int = 0
    updated: int = 0
    removed: int = 0
    kept: int = 0
    failed: int = 0


class ResourceSync(Protocol[T]):
    """Protocol for bidirectional resource synchronization."""
    async def get_host_resources(self, exec: Executor) -> list[T]: ...
    def get_config_resources(self, config: object) -> list[T]: ...
    async def apply(self, resource: T, exec: Executor) -> Result: ...
    async def remove(self, resource_id: str, exec: Executor) -> Result: ...
    def needs_update(self, host: T, config: T) -> bool: ...
    def resource_id(self, resource: T) -> str: ...
