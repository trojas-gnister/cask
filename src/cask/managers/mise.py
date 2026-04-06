"""Mise tool version management."""
from __future__ import annotations

import json

from cask.executor.protocol import Executor
from cask.result import Result


class MiseManager:
    """Manages tool versions via mise."""

    async def install(self, tool: str, version: str, exec: Executor) -> Result:
        r = await exec.execute(["mise", "install", f"{tool}@{version}"])
        if r.exit_code == 0:
            await exec.execute(["mise", "use", "--global", f"{tool}@{version}"])
            return Result(ok=True, message=f"Installed {tool} {version}",
                         actions=[f"mise install {tool}@{version}"])
        return Result(ok=False, message=f"Failed to install {tool}: {r.stderr}", actions=[])

    async def remove(self, tool: str, exec: Executor) -> Result:
        r = await exec.execute(["mise", "uninstall", tool])
        return Result(ok=r.exit_code == 0, message=f"Removed {tool}", actions=[])

    async def list_tools(self, exec: Executor) -> dict[str, str]:
        """Return {tool: version} of installed tools."""
        r = await exec.execute(["mise", "list", "--json"])
        tools = {}
        if r.exit_code == 0 and r.stdout.strip():
            try:
                data = json.loads(r.stdout)
                for tool, info in data.items():
                    if isinstance(info, dict):
                        tools[tool] = info.get("version", "")
                    elif isinstance(info, list) and info:
                        tools[tool] = info[0].get("version", "") if isinstance(info[0], dict) else ""
            except json.JSONDecodeError:
                pass
        return tools
