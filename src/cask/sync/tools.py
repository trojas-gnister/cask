"""Mise tool sync implementation."""
from __future__ import annotations

from dataclasses import dataclass

from cask.executor.protocol import Executor
from cask.managers.mise import MiseManager
from cask.result import Result


@dataclass
class ToolResource:
    name: str
    version: str


class ToolSync:
    def __init__(self) -> None:
        self._mgr = MiseManager()

    async def get_host_resources(self, exec: Executor) -> list[ToolResource]:
        tools = await self._mgr.list_tools(exec)
        return [ToolResource(name=n, version=v) for n, v in tools.items()]

    def get_config_resources(self, config: dict[str, str]) -> list[ToolResource]:
        return [ToolResource(name=n, version=v) for n, v in config.items()]

    async def apply(self, resource: ToolResource, exec: Executor) -> Result:
        return await self._mgr.install(resource.name, resource.version, exec)

    async def remove(self, resource_id: str, exec: Executor) -> Result:
        return await self._mgr.remove(resource_id, exec)

    def needs_update(self, host: ToolResource, config: ToolResource) -> bool:
        return host.version != config.version

    def resource_id(self, resource: ToolResource) -> str:
        return resource.name
