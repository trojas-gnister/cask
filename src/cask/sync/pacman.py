"""Pacman and AUR sync implementations."""
from __future__ import annotations

from dataclasses import dataclass

from cask.config.models import PacmanConfig
from cask.executor.protocol import Executor
from cask.managers.pacman import PacmanManager
from cask.managers.aur import AURManager
from cask.result import Result


@dataclass
class PacmanResource:
    name: str
    version: str = ""


class PacmanSync:
    """Syncs pacman (official repo) packages."""

    def __init__(self) -> None:
        self._mgr = PacmanManager()

    async def get_host_resources(self, exec: Executor) -> list[PacmanResource]:
        installed = await self._mgr.list_installed(exec)
        return [PacmanResource(name=n, version=v) for n, v in installed.items()]

    def get_config_resources(self, config: PacmanConfig) -> list[PacmanResource]:
        return [PacmanResource(name=p) for p in config.packages]

    async def apply(self, resource: PacmanResource, exec: Executor) -> Result:
        return await self._mgr.install([resource.name], exec)

    async def remove(self, resource_id: str, exec: Executor) -> Result:
        return await self._mgr.remove(resource_id, exec)

    def needs_update(self, host: PacmanResource, config: PacmanResource) -> bool:
        # If config specifies a version and it differs from host, flag for update
        return bool(config.version) and host.version != config.version

    def resource_id(self, resource: PacmanResource) -> str:
        return resource.name


class AURSync:
    """Syncs AUR packages via yay."""

    def __init__(self) -> None:
        self._mgr = AURManager()

    async def get_host_resources(self, exec: Executor) -> list[PacmanResource]:
        installed = await self._mgr.list_installed(exec)
        return [PacmanResource(name=n, version=v) for n, v in installed.items()]

    def get_config_resources(self, config: PacmanConfig) -> list[PacmanResource]:
        return [PacmanResource(name=p) for p in config.aur_packages]

    async def apply(self, resource: PacmanResource, exec: Executor) -> Result:
        return await self._mgr.install([resource.name], exec)

    async def remove(self, resource_id: str, exec: Executor) -> Result:
        return await self._mgr.remove(resource_id, exec)

    def needs_update(self, host: PacmanResource, config: PacmanResource) -> bool:
        return bool(config.version) and host.version != config.version

    def resource_id(self, resource: PacmanResource) -> str:
        return resource.name
