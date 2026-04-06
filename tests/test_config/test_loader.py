import os
import pytest
from pathlib import Path
from cask.config.loader import load_config
from cask.config.paths import config_dir, state_dir, default_config_path


def test_load_simple_config(tmp_path):
    cfg_file = tmp_path / "config.toml"
    cfg_file.write_text('[pacman]\npackages = ["firefox"]\n')
    cfg = load_config(str(cfg_file))
    assert cfg.pacman is not None
    assert cfg.pacman.packages == ["firefox"]


def test_load_empty_config(tmp_path):
    cfg_file = tmp_path / "config.toml"
    cfg_file.write_text("# empty\n")
    cfg = load_config(str(cfg_file))
    assert cfg.pacman is None
    assert cfg.tools is None


def test_load_with_includes():
    fixture = str(Path(__file__).parent.parent / "fixtures" / "config.toml")
    cfg = load_config(fixture)
    assert cfg.pacman is not None
    assert "firefox" in cfg.pacman.packages
    assert cfg.flatpak is not None
    assert "org.signal.Signal" in cfg.flatpak.packages
    assert cfg.podman is not None
    assert "nginx" in cfg.podman.containers
    assert cfg.tools == {"node": "22.0.0"}


def test_load_with_env_expansion(tmp_path, monkeypatch):
    monkeypatch.setenv("MY_PORT", "9090")
    cfg_file = tmp_path / "config.toml"
    cfg_file.write_text(
        '[podman.containers.app]\nimage = "myapp"\nports = ["${MY_PORT}:80"]\n'
    )
    cfg = load_config(str(cfg_file))
    assert cfg.podman.containers["app"].ports == ["9090:80"]


def test_load_missing_file():
    with pytest.raises(FileNotFoundError):
        load_config("/nonexistent/config.toml")


def test_config_dir():
    d = config_dir()
    assert d.endswith("cask")


def test_default_config_path():
    p = default_config_path()
    assert p.endswith("config.toml")
