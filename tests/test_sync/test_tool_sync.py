"""Tests for ToolSync."""
from cask.sync.tools import ToolSync, ToolResource


def test_get_config_resources_basic():
    sync = ToolSync()
    config = {"node": "22.0.0", "python": "3.12.0"}
    resources = sync.get_config_resources(config)
    assert len(resources) == 2
    names = {r.name for r in resources}
    assert names == {"node", "python"}


def test_get_config_resources_versions():
    sync = ToolSync()
    config = {"node": "22.0.0"}
    resources = sync.get_config_resources(config)
    assert resources[0].version == "22.0.0"


def test_get_config_resources_empty():
    sync = ToolSync()
    resources = sync.get_config_resources({})
    assert resources == []


def test_needs_update_same_version():
    sync = ToolSync()
    host = ToolResource(name="node", version="22.0.0")
    config = ToolResource(name="node", version="22.0.0")
    assert sync.needs_update(host, config) is False


def test_needs_update_version_mismatch():
    sync = ToolSync()
    host = ToolResource(name="node", version="20.0.0")
    config = ToolResource(name="node", version="22.0.0")
    assert sync.needs_update(host, config) is True


def test_resource_id():
    sync = ToolSync()
    r = ToolResource(name="node", version="22.0.0")
    assert sync.resource_id(r) == "node"
