# Cask

Declarative package and container management CLI for Linux. Define your system state in TOML — Cask reconciles what's installed against what's declared, installs missing resources, and reports undeclared ones.

## Features

- Declarative TOML configuration with include support and env var expansion
- Bidirectional sync: installs missing packages, reports undeclared ones
- Supports: pacman, AUR (yay/paru), Flatpak, Podman containers, Quadlet systemd units, Distrobox instances, mise tools
- Dry-run mode (`--dry-run`) previews all changes without applying them
- State tracking via SHA256 hashing for idempotent operations
- Lockfile for version pinning
- Generation snapshots after each successful sync
- Shell auto-enter hooks for Distrobox instances (bash, zsh, fish)

## Installation

```bash
pip install cask
# or in dev mode:
git clone https://github.com/trojas-gnister/cask
cd cask
python -m venv .venv && source .venv/bin/activate
pip install -e ".[dev]"
```

## Configuration

Default config: `~/.config/cask/config.toml`

```toml
# ~/.config/cask/config.toml
include = ["packages.toml"]  # Optional: split into multiple files

[pacman]
packages = ["firefox", "git", "neovim", "ripgrep"]
aur_packages = ["yay", "1password"]

[flatpak]
remotes = ["flathub"]
packages = [
    "org.signal.Signal",
    "com.spotify.Client",
]

[podman.containers.nginx]
image = "docker.io/library/nginx:latest"
ports = ["8080:80"]
volumes = ["/srv/www:/usr/share/nginx/html:ro"]
quadlet = true  # Generate systemd unit via Quadlet

[devbox.instances.dev]
image = "ghcr.io/ublue-os/fedora-toolbox:41"
packages = ["gcc", "clang", "python3"]
export_apps = ["code"]

[devbox.hooks]
"~/projects" = "dev"  # Auto-enter 'dev' distrobox when cd-ing into ~/projects

[tools]
node = "22.0.0"
python = "3.12.0"
```

## Usage

```bash
# Initialize config
cask config init

# Preview changes (no system modification)
cask diff

# Apply all declared resources
cask sync

# Apply only specific sections
cask sync pacman flatpak

# Add and immediately install
cask add pacman firefox git
cask add flatpak org.signal.Signal
cask add tool node 22.0.0

# Remove from config and uninstall
cask remove pacman firefox
cask remove flatpak org.signal.Signal

# List managed resources
cask list
cask list pacman flatpak

# Update all resources to latest
cask update

# Validate config file
cask validate
cask validate --config /path/to/config.toml

# Version pinning
cask lock create   # Pin current versions
cask lock verify   # Check versions match lockfile

# State management
cask state show
cask state reset
cask state generations

# Devbox shell hooks
cask devbox install   # Append hooks to shell rc file
cask devbox remove    # Instructions for manual removal

# Global options (place before subcommand)
cask --dry-run sync
cask --config /path/to/config.toml sync
cask --yes sync    # Auto-keep all undeclared resources
cask --no sync     # Auto-remove all undeclared resources
cask --verbose sync
```

## Architecture

```
cask/
  src/cask/
    executor/          # Executor protocol: abstracts all system calls
      protocol.py      # Executor Protocol (structural subtyping)
      system.py        # Real system executor (asyncio subprocess)
      mock.py          # Mock executor for testing
    config/            # TOML config loading and models
      models.py        # Pydantic models (CaskConfig, PacmanConfig, ...)
      loader.py        # Load + merge TOML with include support
      expansion.py     # Env var and tilde expansion
      validation.py    # Semantic validation
      writer.py        # Add/remove items in config files
    managers/          # Low-level package manager wrappers
      pacman.py        # pacman -S / -Rs / -Qe
      aur.py           # yay/paru AUR helper
      flatpak.py       # flatpak install/uninstall/list
      podman.py        # podman run/rm/ps
      quadlet.py       # Quadlet .container unit files
      distrobox.py     # distrobox create/rm/list
      mise.py          # mise install/list
    sync/              # Bidirectional sync engine
      protocol.py      # ResourceSync Protocol + SyncOptions + SyncStats
      algorithm.py     # Generic apply/remove/keep reconciliation loop
      flatpak.py       # FlatpakSync implementation
      containers.py    # ContainerSync implementation
      devbox.py        # DevboxSync implementation
      tools.py         # ToolSync implementation
    state/             # State persistence
      manager.py       # StateManager (global.json)
      hashing.py       # SHA256 config hashing
      lockfile.py      # Version pin lockfile (lock.json)
      generations.py   # Snapshot per successful sync
    devbox/
      hooks.py         # Shell auto-enter hook generation
    cli/               # Typer CLI (thin layer over library)
      app.py           # App factory, global options, command registration
      commands/        # One file per command group
```

**Key design principle:** All system calls go through the `Executor` protocol. Tests use `MockExecutor` — no real system calls in tests. The `SystemExecutor` uses `asyncio.create_subprocess_exec` with a 120-second timeout.

## Known Limitations

- **Pacman sync** does not currently run `cask sync pacman` — pacman sync is left to the manager's `list_installed` vs config diff. A future task would wire PacmanSync into the sync engine.
- **AUR helper selection** defaults to `yay`; `paru` support requires configuration.
- **Quadlet units** are generated but systemd reload (`systemctl --user daemon-reload`) must be run manually after first sync.
- **lock apply** (install exact locked versions) is not yet implemented.
- **Interactive mode** (prompting on undeclared resources) requires a TTY; use `--yes` or `--no` in scripts.
- **Global `--config` flag** must precede the subcommand: `cask --config /path sync` not `cask sync --config /path`. The `validate` and `config init` subcommands accept their own `--config` flag as a convenience.
- Tested on Arch Linux with Python 3.11+. Other distros may require adjusting package manager paths.

## Development

```bash
source .venv/bin/activate
pytest -v              # Run all 67 tests
cask version           # cask 0.1.0
cask --help
```

## License

MIT
