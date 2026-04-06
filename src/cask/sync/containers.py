"""Podman container sync implementation."""
from __future__ import annotations

from dataclasses import dataclass, field

from cask.config.models import PodmanConfig, ContainerConfig
from cask.executor.protocol import Executor
from cask.managers.podman import PodmanManager
from cask.result import Result


@dataclass
class ContainerResource:
    name: str
    config: ContainerConfig | None = None
    image: str = ""
    ports: list[str] = field(default_factory=list)
    volumes: list[str] = field(default_factory=list)
    read_only: bool = False


class ContainerSync:
    def __init__(self) -> None:
        self._mgr = PodmanManager()

    async def get_host_resources(self, exec: Executor) -> list[ContainerResource]:
        containers = await self._mgr.list_containers(exec)
        resources = []
        for name, info in containers.items():
            if isinstance(info, dict):
                image = info.get("image", "")
                ports = info.get("ports", [])
                volumes = info.get("volumes", [])
            else:
                # Legacy: info is just the image string
                image = info
                ports = []
                volumes = []
            resources.append(ContainerResource(name=name, image=image, ports=ports, volumes=volumes))
        return resources

    def get_config_resources(self, config: PodmanConfig) -> list[ContainerResource]:
        return [
            ContainerResource(
                name=n,
                config=c,
                image=c.image,
                ports=list(c.ports),
                volumes=list(c.volumes),
                read_only=c.security.read_only,
            )
            for n, c in config.containers.items()
        ]

    async def apply(self, resource: ContainerResource, exec: Executor) -> Result:
        if resource.config:
            return await self._mgr.create(resource.name, resource.config, exec)
        return Result(ok=False, message=f"No config for {resource.name}", actions=[])

    async def remove(self, resource_id: str, exec: Executor) -> Result:
        return await self._mgr.remove(resource_id, exec)

    def needs_update(self, host: ContainerResource, config: ContainerResource) -> bool:
        return (
            host.image != config.image
            or sorted(host.ports) != sorted(config.ports)
            or sorted(host.volumes) != sorted(config.volumes)
        )

    def resource_id(self, resource: ContainerResource) -> str:
        return resource.name
