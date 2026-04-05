import pytest
from cask.managers.pacman import PacmanManager
from cask.executor.mock import MockExecutor


@pytest.mark.asyncio
async def test_install_packages():
    mock = MockExecutor()
    mgr = PacmanManager()
    result = await mgr.install(["firefox", "git"], mock)
    assert result.ok
    assert mock.was_called("[sudo] pacman")
    assert mock.was_called("firefox")


@pytest.mark.asyncio
async def test_remove_package():
    mock = MockExecutor()
    mgr = PacmanManager()
    result = await mgr.remove("firefox", mock)
    assert result.ok
    assert mock.was_called("[sudo] pacman")


@pytest.mark.asyncio
async def test_list_installed():
    mock = MockExecutor()
    mock.add_response("pacman", stdout="firefox 138.0-1\ngit 2.47.0-1\n")
    mgr = PacmanManager()
    packages = await mgr.list_installed(mock)
    assert len(packages) == 2
    assert packages["firefox"] == "138.0-1"
