"""Pacman package manager operations."""
from __future__ import annotations

from cask.executor.protocol import Executor
from cask.result import Result


class PacmanManager:
    """Manages pacman packages."""

    async def install(self, packages: list[str], exec: Executor) -> Result:
        r = await exec.execute_sudo(["pacman", "-S", "--noconfirm", "--needed", *packages])
        if r.exit_code == 0:
            return Result(ok=True, message=f"Installed {len(packages)} packages",
                         actions=[f"pacman -S {p}" for p in packages])
        return Result(ok=False, message=f"Failed to install: {r.stderr}", actions=[])

    async def remove(self, package: str, exec: Executor) -> Result:
        r = await exec.execute_sudo(["pacman", "-Rs", "--noconfirm", package])
        if r.exit_code == 0:
            return Result(ok=True, message=f"Removed {package}", actions=[f"pacman -Rs {package}"])
        return Result(ok=False, message=f"Failed to remove {package}: {r.stderr}", actions=[])

    async def list_installed(self, exec: Executor) -> dict[str, str]:
        """Return {name: version} of explicitly installed packages."""
        r = await exec.execute(["pacman", "-Qe"])
        packages = {}
        if r.exit_code == 0:
            for line in r.stdout.strip().splitlines():
                parts = line.split()
                if len(parts) >= 2:
                    packages[parts[0]] = parts[1]
        return packages

    async def sync_db(self, exec: Executor) -> Result:
        r = await exec.execute_sudo(["pacman", "-Sy"])
        return Result(ok=r.exit_code == 0, message="Synced package database", actions=[])
