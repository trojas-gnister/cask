import pytest
from cask.executor.mock import MockExecutor
from cask.executor.protocol import Executor


@pytest.mark.asyncio
async def test_mock_records_execute():
    mock = MockExecutor()
    r = await mock.execute(["pacman", "-S", "firefox"])
    assert r.exit_code == 0
    assert mock.call_count == 1
    assert mock.was_called("pacman")


@pytest.mark.asyncio
async def test_mock_records_sudo():
    mock = MockExecutor()
    r = await mock.execute_sudo(["pacman", "-S", "firefox"])
    assert r.exit_code == 0
    assert mock.was_called("[sudo] pacman")


@pytest.mark.asyncio
async def test_mock_write_and_read_file():
    mock = MockExecutor()
    mock.write_file("/etc/test.conf", "key=value\n")
    assert mock.file_exists("/etc/test.conf")
    assert mock.read_file("/etc/test.conf") == "key=value\n"
    assert not mock.file_exists("/nonexistent")


@pytest.mark.asyncio
async def test_mock_preconfigured_response():
    mock = MockExecutor()
    mock.add_response("pacman", exit_code=0, stdout="firefox 138.0-1\n")
    r = await mock.execute(["pacman", "-Q", "firefox"])
    assert r.stdout == "firefox 138.0-1\n"


def test_mock_is_executor():
    mock = MockExecutor()
    assert isinstance(mock, Executor)


@pytest.mark.asyncio
async def test_mock_dry_run():
    mock = MockExecutor(dry_run=True)
    assert mock.dry_run
    r = await mock.execute(["echo", "hello"])
    assert r.exit_code == 0
    assert mock.call_count == 1
