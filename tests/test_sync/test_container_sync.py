"""Tests for ContainerSync."""
from cask.sync.containers import ContainerSync, ContainerResource
from cask.config.models import PodmanConfig, ContainerConfig, ContainerSecurity


def _make_podman_config(**kwargs) -> PodmanConfig:
    defaults = dict(image="nginx:latest", ports=[], volumes=[])
    defaults.update(kwargs)
    return PodmanConfig(containers={"web": ContainerConfig(**defaults)})


def test_get_config_resources_basic():
    sync = ContainerSync()
    cfg = _make_podman_config()
    resources = sync.get_config_resources(cfg)
    assert len(resources) == 1
    assert resources[0].name == "web"
    assert resources[0].image == "nginx:latest"


def test_get_config_resources_with_ports():
    sync = ContainerSync()
    cfg = _make_podman_config(ports=["8080:80", "443:443"])
    resources = sync.get_config_resources(cfg)
    assert resources[0].ports == ["8080:80", "443:443"]


def test_get_config_resources_with_volumes():
    sync = ContainerSync()
    cfg = _make_podman_config(volumes=["/data:/data:ro"])
    resources = sync.get_config_resources(cfg)
    assert resources[0].volumes == ["/data:/data:ro"]


def test_get_config_resources_read_only_security():
    sync = ContainerSync()
    security = ContainerSecurity(read_only=True)
    cfg = PodmanConfig(containers={"web": ContainerConfig(image="nginx:latest", security=security)})
    resources = sync.get_config_resources(cfg)
    assert resources[0].read_only is True


def test_needs_update_same_image_ports_volumes():
    sync = ContainerSync()
    host = ContainerResource(name="web", image="nginx:latest", ports=["80:80"], volumes=[])
    config = ContainerResource(name="web", image="nginx:latest", ports=["80:80"], volumes=[])
    assert sync.needs_update(host, config) is False


def test_needs_update_image_changed():
    sync = ContainerSync()
    host = ContainerResource(name="web", image="nginx:1.0", ports=[], volumes=[])
    config = ContainerResource(name="web", image="nginx:2.0", ports=[], volumes=[])
    assert sync.needs_update(host, config) is True


def test_needs_update_port_changed():
    sync = ContainerSync()
    host = ContainerResource(name="web", image="nginx:latest", ports=["80:80"], volumes=[])
    config = ContainerResource(name="web", image="nginx:latest", ports=["8080:80"], volumes=[])
    assert sync.needs_update(host, config) is True


def test_needs_update_volume_changed():
    sync = ContainerSync()
    host = ContainerResource(name="web", image="nginx:latest", ports=[], volumes=[])
    config = ContainerResource(name="web", image="nginx:latest", ports=[], volumes=["/data:/data"])
    assert sync.needs_update(host, config) is True


def test_resource_id():
    sync = ContainerSync()
    r = ContainerResource(name="mycontainer", image="alpine:latest")
    assert sync.resource_id(r) == "mycontainer"
