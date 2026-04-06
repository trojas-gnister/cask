"""Distrobox instance management."""
from __future__ import annotations

from cask.config.models import DevboxInstanceConfig
from cask.executor.protocol import Executor
from cask.result import Result


def _detect_pkg_manager(image: str) -> tuple[str, list[str]]:
    """Return (pkg_manager_name, install_cmd_prefix) based on image."""
    image_lower = image.lower()
    if any(d in image_lower for d in ("fedora", "centos", "rhel", "rocky", "alma")):
        return "dnf", ["sudo", "dnf", "install", "-y"]
    elif any(d in image_lower for d in ("ubuntu", "debian", "mint")):
        return "apt", ["sudo", "apt-get", "install", "-y"]
    elif any(d in image_lower for d in ("arch", "manjaro", "endeavour")):
        return "pacman", ["sudo", "pacman", "-S", "--noconfirm"]
    elif any(d in image_lower for d in ("opensuse", "suse", "tumbleweed")):
        return "zypper", ["sudo", "zypper", "install", "-y"]
    elif any(d in image_lower for d in ("alpine",)):
        return "apk", ["sudo", "apk", "add"]
    else:
        return "dnf", ["sudo", "dnf", "install", "-y"]  # fallback


class DistroboxManager:
    """Manages Distrobox instances."""

    async def create(self, name: str, cfg: DevboxInstanceConfig, exec: Executor) -> Result:
        args = ["distrobox", "create", "--name", name, "--image", cfg.image, "--yes"]
        if cfg.home == "isolated":
            args.extend(["--home", f"~/.local/share/distrobox/{name}"])
        r = await exec.execute(args)
        if r.exit_code != 0:
            return Result(ok=False, message=f"Failed to create {name}: {r.stderr}", actions=[])

        actions = [f"Created distrobox {name}"]

        # Install packages if specified
        if cfg.packages:
            _, install_prefix = _detect_pkg_manager(cfg.image)
            pkg_cmd = f"distrobox enter {name} -- {' '.join(install_prefix)} {' '.join(cfg.packages)}"
            await exec.execute_shell(pkg_cmd)
            actions.append(f"Installed {len(cfg.packages)} packages in {name}")

        # Export apps
        for app in cfg.export_apps:
            await exec.execute(["distrobox", "enter", name, "--", "distrobox-export", "--app", app])
            actions.append(f"Exported {app}")

        return Result(ok=True, message=f"Created distrobox {name}", actions=actions)

    async def remove(self, name: str, exec: Executor) -> Result:
        r = await exec.execute(["distrobox", "rm", "--force", name])
        if r.exit_code == 0:
            return Result(ok=True, message=f"Removed distrobox {name}", actions=[])
        return Result(ok=False, message=f"Failed to remove {name}: {r.stderr}", actions=[])

    async def list_instances(self, exec: Executor) -> dict[str, str]:
        """Return {name: image} of distrobox instances."""
        r = await exec.execute(["distrobox", "list", "--no-color"])
        instances = {}
        if r.exit_code == 0:
            for line in r.stdout.strip().splitlines()[1:]:  # Skip header
                parts = [p.strip() for p in line.split("|")]
                if len(parts) >= 4:
                    instances[parts[1]] = parts[3]
        return instances
