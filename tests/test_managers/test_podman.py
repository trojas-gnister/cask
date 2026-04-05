import pytest
from cask.managers.podman import PodmanManager
from cask.config.models import ContainerConfig, ContainerSecurity
from cask.executor.mock import MockExecutor


@pytest.mark.asyncio
async def test_create_container():
    mock = MockExecutor()
    cfg = ContainerConfig(image="nginx:latest", ports=["8080:80"])
    mgr = PodmanManager()
    result = await mgr.create("nginx", cfg, mock)
    assert result.ok
    assert mock.was_called("podman run")


@pytest.mark.asyncio
async def test_create_with_security():
    mock = MockExecutor()
    cfg = ContainerConfig(
        image="nginx:latest",
        security=ContainerSecurity(read_only=True, no_new_privileges=True),
    )
    mgr = PodmanManager()
    result = await mgr.create("nginx", cfg, mock)
    assert result.ok
    assert mock.was_called("--read-only")


@pytest.mark.asyncio
async def test_remove_container():
    mock = MockExecutor()
    mgr = PodmanManager()
    result = await mgr.remove("nginx", mock)
    assert result.ok
    assert mock.was_called("podman rm")


@pytest.mark.asyncio
async def test_list_containers():
    mock = MockExecutor()
    mock.add_response("podman", stdout='[{"Names":["nginx"],"Image":"nginx:latest"}]\n')
    mgr = PodmanManager()
    containers = await mgr.list_containers(mock)
    assert "nginx" in containers
