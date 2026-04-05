import pytest
from cask.managers.distrobox import DistroboxManager
from cask.config.models import DevboxInstanceConfig
from cask.executor.mock import MockExecutor


@pytest.mark.asyncio
async def test_create_instance():
    mock = MockExecutor()
    cfg = DevboxInstanceConfig(image="fedora:41", packages=["gcc"])
    mgr = DistroboxManager()
    result = await mgr.create("dev", cfg, mock)
    assert result.ok
    assert mock.was_called("distrobox create")


@pytest.mark.asyncio
async def test_remove_instance():
    mock = MockExecutor()
    mgr = DistroboxManager()
    result = await mgr.remove("dev", mock)
    assert result.ok
    assert mock.was_called("distrobox rm")


@pytest.mark.asyncio
async def test_list_instances():
    mock = MockExecutor()
    mock.add_response("distrobox", stdout="ID | NAME | STATUS | IMAGE\n1 | dev | running | fedora:41\n")
    mgr = DistroboxManager()
    instances = await mgr.list_instances(mock)
    assert "dev" in instances
