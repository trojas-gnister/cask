from cask.config.models import CaskConfig, PacmanConfig, ContainerConfig, PodmanConfig
from cask.config.validation import validate_config
from cask.config.writer import add_to_config, remove_from_config


def test_validate_empty_config():
    cfg = CaskConfig()
    errors = validate_config(cfg)
    assert errors == []


def test_validate_valid_config():
    cfg = CaskConfig(pacman=PacmanConfig(packages=["firefox"]))
    errors = validate_config(cfg)
    assert errors == []


def test_validate_container_missing_image():
    cfg = CaskConfig(
        podman=PodmanConfig(containers={"bad": ContainerConfig(image="")})
    )
    errors = validate_config(cfg)
    assert any("image" in e.lower() for e in errors)


def test_add_pacman_package(tmp_path):
    cfg_file = tmp_path / "config.toml"
    cfg_file.write_text('[pacman]\npackages = ["git"]\n')
    add_to_config(str(cfg_file), "pacman", "firefox")
    content = cfg_file.read_text()
    assert "firefox" in content
    assert "git" in content


def test_add_to_empty_section(tmp_path):
    cfg_file = tmp_path / "config.toml"
    cfg_file.write_text("# empty\n")
    add_to_config(str(cfg_file), "pacman", "firefox")
    content = cfg_file.read_text()
    assert "firefox" in content


def test_remove_pacman_package(tmp_path):
    cfg_file = tmp_path / "config.toml"
    cfg_file.write_text('[pacman]\npackages = ["firefox", "git"]\n')
    remove_from_config(str(cfg_file), "pacman", "firefox")
    content = cfg_file.read_text()
    assert "firefox" not in content
    assert "git" in content
