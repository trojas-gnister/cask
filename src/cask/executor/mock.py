"""Mock executor for testing."""
from __future__ import annotations

from dataclasses import dataclass, field

from cask.result import ExecResult


@dataclass
class _MockResponse:
    cmd_prefix: str
    exit_code: int
    stdout: str


class MockExecutor:
    """Records calls and returns preconfigured responses."""

    def __init__(self, dry_run: bool = False) -> None:
        self.dry_run = dry_run
        self._calls: list[str] = []
        self._files: dict[str, str] = {}
        self._responses: list[_MockResponse] = []

    @property
    def call_count(self) -> int:
        return len(self._calls)

    @property
    def calls(self) -> list[str]:
        return list(self._calls)

    def was_called(self, substring: str) -> bool:
        return any(substring in c for c in self._calls)

    def call_at(self, index: int) -> str:
        return self._calls[index]

    def add_response(self, cmd_prefix: str, exit_code: int = 0, stdout: str = "") -> None:
        self._responses.append(_MockResponse(cmd_prefix, exit_code, stdout))

    def _find_response(self, cmd: list[str]) -> ExecResult:
        cmd_str = cmd[0] if cmd else ""
        for resp in self._responses:
            if cmd_str.startswith(resp.cmd_prefix):
                return ExecResult(resp.exit_code, resp.stdout, "")
        return ExecResult(0, "", "")

    def _record(self, prefix: str | None, cmd: list[str]) -> None:
        parts = []
        if prefix:
            parts.append(prefix)
        parts.extend(cmd)
        self._calls.append(" ".join(parts))

    async def execute(self, cmd: list[str]) -> ExecResult:
        self._record(None, cmd)
        return self._find_response(cmd)

    async def execute_sudo(self, cmd: list[str]) -> ExecResult:
        self._record("[sudo]", cmd)
        return self._find_response(cmd)

    async def execute_shell(self, cmd: str) -> ExecResult:
        self._record("[shell]", [cmd])
        return self._find_response([cmd])

    def file_exists(self, path: str) -> bool:
        return path in self._files

    def read_file(self, path: str) -> str:
        return self._files.get(path, "")

    def write_file(self, path: str, content: str, sudo: bool = False) -> None:
        self._files[path] = content
