import pytest
from cask.managers.flatpak import FlatpakManager
from cask.executor.mock import MockExecutor


@pytest.mark.asyncio
async def test_install_package():
    mock = MockExecutor()
    mgr = FlatpakManager()
    result = await mgr.install("org.signal.Signal", "flathub", mock)
    assert result.ok
    assert mock.was_called("flatpak install")


@pytest.mark.asyncio
async def test_remove_package():
    mock = MockExecutor()
    mgr = FlatpakManager()
    result = await mgr.remove("org.signal.Signal", mock)
    assert result.ok
    assert mock.was_called("flatpak uninstall")


@pytest.mark.asyncio
async def test_list_installed():
    mock = MockExecutor()
    mock.add_response("flatpak", stdout="org.signal.Signal\tSignal\t7.32\tflathub\n")
    mgr = FlatpakManager()
    apps = await mgr.list_installed(mock)
    assert "org.signal.Signal" in apps
