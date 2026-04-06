"""Podman container lifecycle management."""
from __future__ import annotations

import json

from cask.config.models import ContainerConfig
from cask.executor.protocol import Executor
from cask.result import Result


class PodmanManager:
    """Manages Podman containers."""

    def _build_run_args(self, name: str, cfg: ContainerConfig) -> list[str]:
        args = ["podman", "run", "-d", "--name", name]
        for port in cfg.ports:
            args.extend(["-p", port])
        for vol in cfg.volumes:
            args.extend(["-v", vol])
        for key, val in cfg.environment.items():
            args.extend(["-e", f"{key}={val}"])
        if cfg.security.read_only:
            args.append("--read-only")
        if cfg.security.no_new_privileges:
            args.append("--security-opt=no-new-privileges")
        for cap in cfg.security.drop_capabilities:
            args.extend(["--cap-drop", cap])
        for cap in cfg.security.add_capabilities:
            args.extend(["--cap-add", cap])
        args.append(cfg.image)
        return args

    async def create(self, name: str, cfg: ContainerConfig, exec: Executor) -> Result:
        args = self._build_run_args(name, cfg)
        r = await exec.execute(args)
        if r.exit_code == 0:
            return Result(ok=True, message=f"Created container {name}", actions=[" ".join(args)])
        return Result(ok=False, message=f"Failed to create {name}: {r.stderr}", actions=[])

    async def remove(self, name: str, exec: Executor) -> Result:
        await exec.execute(["podman", "stop", name])
        r = await exec.execute(["podman", "rm", "-f", name])
        if r.exit_code == 0:
            return Result(ok=True, message=f"Removed container {name}", actions=[])
        return Result(ok=False, message=f"Failed to remove {name}: {r.stderr}", actions=[])

    async def list_containers(self, exec: Executor) -> dict[str, dict]:
        """Return {name: {image, ports, volumes}} of all containers."""
        r = await exec.execute(["podman", "ps", "-a", "--format", "json"])
        containers: dict[str, dict] = {}
        if r.exit_code == 0 and r.stdout.strip():
            try:
                data = json.loads(r.stdout)
                for c in data:
                    names = c.get("Names", [])
                    name = names[0] if names else ""
                    image = c.get("Image", "")
                    # Ports: list of dicts like {"host_port": 8080, ...} or raw strings
                    raw_ports = c.get("Ports", []) or []
                    ports: list[str] = []
                    for p in raw_ports:
                        if isinstance(p, str):
                            ports.append(p)
                        elif isinstance(p, dict):
                            host = p.get("host_port") or p.get("hostPort", "")
                            container = p.get("container_port") or p.get("containerPort", "")
                            if host and container:
                                ports.append(f"{host}:{container}")
                    # Mounts: list of dicts with Source/Destination
                    raw_mounts = c.get("Mounts", []) or []
                    volumes: list[str] = []
                    for m in raw_mounts:
                        if isinstance(m, str):
                            volumes.append(m)
                        elif isinstance(m, dict):
                            src = m.get("Source", "")
                            dst = m.get("Destination", "")
                            if src and dst:
                                volumes.append(f"{src}:{dst}")
                    if name:
                        containers[name] = {"image": image, "ports": ports, "volumes": volumes}
            except json.JSONDecodeError:
                pass
        return containers
