"""Flatpak operations."""
from __future__ import annotations

from cask.executor.protocol import Executor
from cask.result import Result


class FlatpakManager:
    """Manages Flatpak packages and remotes."""

    async def install(self, app_id: str, remote: str, exec: Executor) -> Result:
        r = await exec.execute_sudo(["flatpak", "install", "-y", remote, app_id])
        if r.exit_code == 0:
            return Result(ok=True, message=f"Installed {app_id}", actions=[f"flatpak install {app_id}"])
        return Result(ok=False, message=f"Failed to install {app_id}: {r.stderr}", actions=[])

    async def remove(self, app_id: str, exec: Executor) -> Result:
        r = await exec.execute_sudo(["flatpak", "uninstall", "-y", app_id])
        if r.exit_code == 0:
            return Result(ok=True, message=f"Removed {app_id}", actions=[f"flatpak uninstall {app_id}"])
        return Result(ok=False, message=f"Failed to remove {app_id}: {r.stderr}", actions=[])

    async def list_installed(self, exec: Executor) -> dict[str, str]:
        """Return {app_id: version} of installed apps."""
        r = await exec.execute(["flatpak", "list", "--app", "--columns=application,version"])
        apps = {}
        if r.exit_code == 0:
            for line in r.stdout.strip().splitlines():
                parts = line.split("\t")
                if len(parts) >= 2:
                    apps[parts[0]] = parts[1]
                elif len(parts) == 1 and parts[0]:
                    apps[parts[0]] = ""
        return apps

    async def add_remote(self, name: str, url: str, exec: Executor) -> Result:
        r = await exec.execute_sudo(["flatpak", "remote-add", "--if-not-exists", name, url])
        return Result(ok=r.exit_code == 0, message=f"Added remote {name}", actions=[])

    async def set_override(self, app_id: str, flags: list[str], exec: Executor) -> Result:
        r = await exec.execute_sudo(["flatpak", "override", "--system", app_id, *flags])
        return Result(ok=r.exit_code == 0, message=f"Set overrides for {app_id}", actions=flags)
