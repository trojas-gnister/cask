import pytest
from cask.managers.distrobox import DistroboxManager, _detect_pkg_manager
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


# _detect_pkg_manager tests

def test_detect_pkg_manager_fedora():
    name, cmd = _detect_pkg_manager("fedora:41")
    assert name == "dnf"
    assert "dnf" in cmd

def test_detect_pkg_manager_ubuntu():
    name, cmd = _detect_pkg_manager("ubuntu:22.04")
    assert name == "apt"
    assert "apt-get" in cmd

def test_detect_pkg_manager_debian():
    name, cmd = _detect_pkg_manager("debian:bookworm")
    assert name == "apt"
    assert "apt-get" in cmd

def test_detect_pkg_manager_arch():
    name, cmd = _detect_pkg_manager("archlinux:latest")
    assert name == "pacman"
    assert "pacman" in cmd

def test_detect_pkg_manager_opensuse():
    name, cmd = _detect_pkg_manager("opensuse/tumbleweed:latest")
    assert name == "zypper"
    assert "zypper" in cmd

def test_detect_pkg_manager_alpine():
    name, cmd = _detect_pkg_manager("alpine:3.19")
    assert name == "apk"
    assert "apk" in cmd

def test_detect_pkg_manager_rocky():
    name, cmd = _detect_pkg_manager("rockylinux:9")
    assert name == "dnf"
    assert "dnf" in cmd

def test_detect_pkg_manager_fallback():
    name, cmd = _detect_pkg_manager("unknown:image")
    assert name == "dnf"
    assert "dnf" in cmd

def test_detect_pkg_manager_used_in_create():
    """Verify create uses detected pkg manager, not hardcoded dnf."""
    import asyncio
    mock = MockExecutor()
    cfg = DevboxInstanceConfig(image="ubuntu:22.04", packages=["vim"])
    mgr = DistroboxManager()
    asyncio.run(mgr.create("mybox", cfg, mock))
    shell_calls = [c for c in mock.calls if "apt-get" in c]
    assert shell_calls, f"Expected apt-get call, got: {mock.calls}"
