"""Semantic validation for config models."""
from __future__ import annotations

from cask.config.models import CaskConfig


def validate_config(cfg: CaskConfig) -> list[str]:
    """Return list of validation error messages. Empty list = valid."""
    errors: list[str] = []

    if cfg.podman:
        for name, container in cfg.podman.containers.items():
            if not container.image:
                errors.append(f"Container '{name}': image is required")

    if cfg.devbox:
        for name, instance in cfg.devbox.instances.items():
            if not instance.image:
                errors.append(f"Devbox '{name}': image is required")

    if cfg.flatpak:
        for remote in cfg.flatpak.remotes:
            if not remote:
                errors.append("Flatpak: empty remote name")

    return errors
