"""Quadlet systemd unit generation for containers."""
from __future__ import annotations

from cask.config.models import ContainerConfig
from cask.executor.protocol import Executor
from cask.result import Result

_QUADLET_DIR = "/etc/containers/systemd"


def generate_quadlet(name: str, cfg: ContainerConfig) -> str:
    """Generate a .container quadlet file content."""
    lines = [
        "[Unit]",
        f"Description={name} container",
        "",
        "[Container]",
        f"Image={cfg.image}",
    ]
    for port in cfg.ports:
        lines.append(f"PublishPort={port}")
    for vol in cfg.volumes:
        lines.append(f"Volume={vol}")
    for key, val in cfg.environment.items():
        lines.append(f"Environment={key}={val}")
    if cfg.security.read_only:
        lines.append("ReadOnly=true")
    if cfg.security.no_new_privileges:
        lines.append("SecurityLabelDisable=true")
    lines.extend(["", "[Install]", "WantedBy=default.target", ""])
    return "\n".join(lines)


async def install_quadlet(name: str, cfg: ContainerConfig, exec: Executor) -> Result:
    """Write a quadlet file and reload systemd."""
    content = generate_quadlet(name, cfg)
    path = f"{_QUADLET_DIR}/{name}.container"
    exec.write_file(path, content, sudo=True)
    await exec.execute_sudo(["systemctl", "daemon-reload"])
    await exec.execute_sudo(["systemctl", "start", f"{name}.service"])
    return Result(ok=True, message=f"Installed quadlet for {name}", actions=[f"Wrote {path}"])
