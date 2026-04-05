"""Real system executor using asyncio subprocess."""
from __future__ import annotations

import asyncio
import os

from cask.result import ExecResult

_TIMEOUT = 120


class SystemExecutor:
    """Executes real system commands."""

    def __init__(self, dry_run: bool = False) -> None:
        self.dry_run = dry_run

    async def execute(self, cmd: list[str]) -> ExecResult:
        if self.dry_run:
            print(f"[dry-run] {' '.join(cmd)}")
            return ExecResult(0, "", "")
        return await self._run(cmd)

    async def execute_sudo(self, cmd: list[str]) -> ExecResult:
        if self.dry_run:
            print(f"[dry-run] sudo {' '.join(cmd)}")
            return ExecResult(0, "", "")
        return await self._run(["sudo", *cmd])

    async def execute_shell(self, cmd: str) -> ExecResult:
        if self.dry_run:
            print(f"[dry-run] sh -c '{cmd}'")
            return ExecResult(0, "", "")
        proc = await asyncio.create_subprocess_shell(
            cmd, stdout=asyncio.subprocess.PIPE, stderr=asyncio.subprocess.PIPE,
        )
        stdout, stderr = await asyncio.wait_for(proc.communicate(), timeout=_TIMEOUT)
        return ExecResult(proc.returncode or 0, stdout.decode(), stderr.decode())

    def file_exists(self, path: str) -> bool:
        return os.path.exists(path)

    def read_file(self, path: str) -> str:
        with open(path) as f:
            return f.read()

    def write_file(self, path: str, content: str, sudo: bool = False) -> None:
        if self.dry_run:
            print(f"[dry-run] write {path} ({len(content)} bytes)")
            return
        if sudo:
            import subprocess
            subprocess.run(["sudo", "tee", path], input=content.encode(),
                          stdout=subprocess.DEVNULL, check=True)
        else:
            parent = os.path.dirname(path)
            if parent:
                os.makedirs(parent, exist_ok=True)
            with open(path, "w") as f:
                f.write(content)

    async def _run(self, cmd: list[str]) -> ExecResult:
        proc = await asyncio.create_subprocess_exec(
            *cmd, stdout=asyncio.subprocess.PIPE, stderr=asyncio.subprocess.PIPE,
        )
        stdout, stderr = await asyncio.wait_for(proc.communicate(), timeout=_TIMEOUT)
        return ExecResult(proc.returncode or 0, stdout.decode(), stderr.decode())
