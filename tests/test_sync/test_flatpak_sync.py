"""Tests for FlatpakSync."""
from cask.sync.flatpak import FlatpakSync, FlatpakResource
from cask.config.models import FlatpakConfig, AppOverride


def test_get_config_resources_returns_resources():
    sync = FlatpakSync()
    config = FlatpakConfig(packages=["org.signal.Signal", "com.spotify.Client"])
    resources = sync.get_config_resources(config)
    assert len(resources) == 2
    ids = [r.app_id for r in resources]
    assert "org.signal.Signal" in ids
    assert "com.spotify.Client" in ids


def test_get_config_resources_uses_first_remote():
    sync = FlatpakSync()
    config = FlatpakConfig(remotes=["my-remote"], packages=["org.signal.Signal"])
    resources = sync.get_config_resources(config)
    assert resources[0].remote == "my-remote"


def test_get_config_resources_defaults_to_flathub():
    sync = FlatpakSync()
    config = FlatpakConfig(remotes=[], packages=["org.signal.Signal"])
    resources = sync.get_config_resources(config)
    assert resources[0].remote == "flathub"


def test_get_config_resources_empty_packages():
    sync = FlatpakSync()
    config = FlatpakConfig(packages=[])
    resources = sync.get_config_resources(config)
    assert resources == []


def test_resource_id():
    sync = FlatpakSync()
    r = FlatpakResource(app_id="org.signal.Signal")
    assert sync.resource_id(r) == "org.signal.Signal"


def test_needs_update_with_override():
    overrides = {"org.signal.Signal": AppOverride(filesystems=["home"])}
    sync = FlatpakSync(overrides=overrides)
    host = FlatpakResource(app_id="org.signal.Signal")
    config = FlatpakResource(app_id="org.signal.Signal")
    assert sync.needs_update(host, config) is True


def test_needs_update_without_override():
    sync = FlatpakSync()
    host = FlatpakResource(app_id="org.signal.Signal")
    config = FlatpakResource(app_id="org.signal.Signal")
    assert sync.needs_update(host, config) is False
