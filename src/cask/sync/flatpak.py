"""Flatpak sync implementation."""
from __future__ import annotations

from dataclasses import dataclass

from cask.config.models import FlatpakConfig
from cask.executor.protocol import Executor
from cask.managers.flatpak import FlatpakManager
from cask.result import Result


@dataclass
class FlatpakResource:
    app_id: str
    remote: str = "flathub"


class FlatpakSync:
    def __init__(self) -> None:
        self._mgr = FlatpakManager()

    async def get_host_resources(self, exec: Executor) -> list[FlatpakResource]:
        apps = await self._mgr.list_installed(exec)
        return [FlatpakResource(app_id=aid) for aid in apps]

    def get_config_resources(self, config: FlatpakConfig) -> list[FlatpakResource]:
        remote = config.remotes[0] if config.remotes else "flathub"
        return [FlatpakResource(app_id=p, remote=remote) for p in config.packages]

    async def apply(self, resource: FlatpakResource, exec: Executor) -> Result:
        return await self._mgr.install(resource.app_id, resource.remote, exec)

    async def remove(self, resource_id: str, exec: Executor) -> Result:
        return await self._mgr.remove(resource_id, exec)

    def needs_update(self, host: FlatpakResource, config: FlatpakResource) -> bool:
        return False  # Flatpak auto-updates; override detection is separate

    def resource_id(self, resource: FlatpakResource) -> str:
        return resource.app_id
