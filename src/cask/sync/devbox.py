"""Distrobox sync implementation."""
from __future__ import annotations

from dataclasses import dataclass, field

from cask.config.models import DevboxConfig, DevboxInstanceConfig
from cask.executor.protocol import Executor
from cask.managers.distrobox import DistroboxManager
from cask.result import Result


@dataclass
class DevboxResource:
    name: str
    config: DevboxInstanceConfig | None = None
    image: str = ""
    packages: list[str] = field(default_factory=list)


class DevboxSync:
    def __init__(self) -> None:
        self._mgr = DistroboxManager()

    async def get_host_resources(self, exec: Executor) -> list[DevboxResource]:
        instances = await self._mgr.list_instances(exec)
        return [DevboxResource(name=n, image=img) for n, img in instances.items()]

    def get_config_resources(self, config: DevboxConfig) -> list[DevboxResource]:
        return [
            DevboxResource(name=n, config=c, image=c.image, packages=list(c.packages))
            for n, c in config.instances.items()
        ]

    async def apply(self, resource: DevboxResource, exec: Executor) -> Result:
        if resource.config:
            return await self._mgr.create(resource.name, resource.config, exec)
        return Result(ok=False, message=f"No config for {resource.name}", actions=[])

    async def remove(self, resource_id: str, exec: Executor) -> Result:
        return await self._mgr.remove(resource_id, exec)

    def needs_update(self, host: DevboxResource, config: DevboxResource) -> bool:
        return host.image != config.image

    def resource_id(self, resource: DevboxResource) -> str:
        return resource.name
