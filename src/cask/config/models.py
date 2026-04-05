"""Pydantic config models mapping 1:1 to TOML structure."""
from __future__ import annotations

from pydantic import BaseModel


class RepoConfig(BaseModel):
    """Custom pacman repository."""
    name: str
    url: str
    key_id: str | None = None
    keyserver: str | None = None


class AppOverride(BaseModel):
    """Per-app flatpak permission overrides."""
    filesystems: list[str] = []
    sockets: list[str] = []


class ContainerSecurity(BaseModel):
    """Per-container security options."""
    read_only: bool = False
    no_new_privileges: bool = False
    drop_capabilities: list[str] = []
    add_capabilities: list[str] = []


class BuildConfig(BaseModel):
    """Container image build config."""
    dockerfile: str
    context: str
    args: dict[str, str] = {}


class ContainerConfig(BaseModel):
    """Podman container definition."""
    image: str
    ports: list[str] = []
    volumes: list[str] = []
    environment: dict[str, str] = {}
    security: ContainerSecurity = ContainerSecurity()
    quadlet: bool = False
    build: BuildConfig | None = None


class DevboxInstanceConfig(BaseModel):
    """Distrobox instance definition."""
    image: str
    packages: list[str] = []
    home: str = "host"
    export_apps: list[str] = []


class PacmanConfig(BaseModel):
    """Pacman package configuration."""
    packages: list[str] = []
    aur_packages: list[str] = []
    custom_repos: list[RepoConfig] = []


class FlatpakConfig(BaseModel):
    """Flatpak configuration."""
    remotes: list[str] = ["flathub"]
    packages: list[str] = []
    overrides: dict[str, AppOverride] = {}


class PodmanConfig(BaseModel):
    """Podman container configuration."""
    containers: dict[str, ContainerConfig] = {}


class DevboxConfig(BaseModel):
    """Distrobox configuration."""
    instances: dict[str, DevboxInstanceConfig] = {}
    hooks: dict[str, str] = {}


class CaskConfig(BaseModel):
    """Root configuration model."""
    pacman: PacmanConfig | None = None
    flatpak: FlatpakConfig | None = None
    podman: PodmanConfig | None = None
    devbox: DevboxConfig | None = None
    tools: dict[str, str] | None = None
