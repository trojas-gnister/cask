"""AUR helper (yay) operations."""
from __future__ import annotations

from cask.executor.protocol import Executor
from cask.result import Result


class AURManager:
    """Manages AUR packages via yay."""

    async def install(self, packages: list[str], exec: Executor) -> Result:
        r = await exec.execute(["yay", "-S", "--noconfirm", "--needed", *packages])
        if r.exit_code == 0:
            return Result(ok=True, message=f"Installed {len(packages)} AUR packages",
                         actions=[f"yay -S {p}" for p in packages])
        return Result(ok=False, message=f"AUR install failed: {r.stderr}", actions=[])

    async def remove(self, package: str, exec: Executor) -> Result:
        r = await exec.execute(["yay", "-Rs", "--noconfirm", package])
        if r.exit_code == 0:
            return Result(ok=True, message=f"Removed AUR package {package}", actions=[])
        return Result(ok=False, message=f"Failed to remove {package}: {r.stderr}", actions=[])

    async def list_installed(self, exec: Executor) -> dict[str, str]:
        """Return {name: version} of foreign (AUR) packages."""
        r = await exec.execute(["yay", "-Qm"])
        packages = {}
        if r.exit_code == 0:
            for line in r.stdout.strip().splitlines():
                parts = line.split()
                if len(parts) >= 2:
                    packages[parts[0]] = parts[1]
        return packages
