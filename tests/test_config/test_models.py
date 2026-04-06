from cask.config.models import (
    CaskConfig, PacmanConfig, FlatpakConfig, PodmanConfig,
    ContainerConfig, ContainerSecurity, DevboxConfig,
    DevboxInstanceConfig, RepoConfig, AppOverride, BuildConfig,
)


def test_empty_config():
    cfg = CaskConfig()
    assert cfg.pacman is None
    assert cfg.flatpak is None
    assert cfg.podman is None
    assert cfg.devbox is None
    assert cfg.tools is None


def test_pacman_config():
    cfg = PacmanConfig(
        packages=["firefox", "git"],
        aur_packages=["yay"],
        custom_repos=[RepoConfig(name="chaotic", url="https://example.com")],
    )
    assert cfg.packages == ["firefox", "git"]
    assert cfg.aur_packages == ["yay"]
    assert cfg.custom_repos[0].name == "chaotic"


def test_flatpak_config():
    cfg = FlatpakConfig(
        packages=["org.signal.Signal"],
        overrides={"org.signal.Signal": AppOverride(filesystems=["!home"])},
    )
    assert cfg.remotes == ["flathub"]
    assert cfg.packages == ["org.signal.Signal"]
    assert cfg.overrides["org.signal.Signal"].filesystems == ["!home"]


def test_container_config():
    cfg = ContainerConfig(
        image="nginx:latest",
        ports=["8080:80"],
        security=ContainerSecurity(read_only=True, no_new_privileges=True),
        quadlet=True,
    )
    assert cfg.image == "nginx:latest"
    assert cfg.security.read_only
    assert cfg.quadlet


def test_devbox_config():
    cfg = DevboxConfig(
        instances={"dev": DevboxInstanceConfig(image="fedora:41", packages=["gcc"])},
        hooks={"~/projects": "dev"},
    )
    assert cfg.instances["dev"].image == "fedora:41"
    assert cfg.hooks["~/projects"] == "dev"


def test_full_config_from_dict():
    data = {
        "pacman": {"packages": ["firefox"]},
        "flatpak": {"packages": ["org.signal.Signal"]},
        "tools": {"node": "22.0.0"},
    }
    cfg = CaskConfig.model_validate(data)
    assert cfg.pacman.packages == ["firefox"]
    assert cfg.flatpak.packages == ["org.signal.Signal"]
    assert cfg.tools == {"node": "22.0.0"}
