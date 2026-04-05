import pytest
from dataclasses import dataclass
from cask.sync.protocol import ResourceSync, SyncOptions, SyncStats
from cask.sync.algorithm import run_sync
from cask.executor.mock import MockExecutor
from cask.result import Result


@dataclass
class FakeResource:
    name: str
    version: str = "1.0"


class FakeSync:
    """Minimal ResourceSync implementation for testing."""

    def __init__(self, host: list[FakeResource], config: list[FakeResource]):
        self._host = host
        self._config = config
        self.applied: list[str] = []
        self.removed: list[str] = []

    async def get_host_resources(self, exec) -> list[FakeResource]:
        return self._host

    def get_config_resources(self, config) -> list[FakeResource]:
        return self._config

    async def apply(self, resource: FakeResource, exec) -> Result:
        self.applied.append(resource.name)
        return Result(ok=True, message=f"applied {resource.name}", actions=[])

    async def remove(self, resource_id: str, exec) -> Result:
        self.removed.append(resource_id)
        return Result(ok=True, message=f"removed {resource_id}", actions=[])

    def needs_update(self, host: FakeResource, config: FakeResource) -> bool:
        return host.version != config.version

    def resource_id(self, resource: FakeResource) -> str:
        return resource.name


@pytest.mark.asyncio
async def test_sync_apply_new():
    sync = FakeSync(host=[], config=[FakeResource("firefox")])
    mock = MockExecutor()
    stats = await run_sync(sync, None, mock, SyncOptions(yes=True))
    assert stats.applied == 1
    assert "firefox" in sync.applied


@pytest.mark.asyncio
async def test_sync_remove_undeclared_with_no():
    sync = FakeSync(host=[FakeResource("old")], config=[])
    mock = MockExecutor()
    stats = await run_sync(sync, None, mock, SyncOptions(no=True))
    assert stats.removed == 1
    assert "old" in sync.removed


@pytest.mark.asyncio
async def test_sync_keep_undeclared_with_yes():
    sync = FakeSync(host=[FakeResource("old")], config=[])
    mock = MockExecutor()
    stats = await run_sync(sync, None, mock, SyncOptions(yes=True))
    assert stats.removed == 0
    assert stats.kept == 1


@pytest.mark.asyncio
async def test_sync_update_changed():
    sync = FakeSync(
        host=[FakeResource("firefox", "1.0")],
        config=[FakeResource("firefox", "2.0")],
    )
    mock = MockExecutor()
    stats = await run_sync(sync, None, mock, SyncOptions(yes=True))
    assert stats.updated == 1
    assert "firefox" in sync.applied


@pytest.mark.asyncio
async def test_sync_no_changes():
    sync = FakeSync(
        host=[FakeResource("firefox", "1.0")],
        config=[FakeResource("firefox", "1.0")],
    )
    mock = MockExecutor()
    stats = await run_sync(sync, None, mock, SyncOptions(yes=True))
    assert stats.applied == 0
    assert stats.updated == 0
    assert stats.removed == 0
