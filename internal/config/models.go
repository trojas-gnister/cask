package config

// CaskConfig is the root configuration structure.
type CaskConfig struct {
	Flatpak         *FlatpakConfig         `toml:"flatpak,omitempty"`
	Podman          *PodmanConfig           `toml:"podman,omitempty"`
	Devbox          *DevboxConfig           `toml:"devbox,omitempty"`
	Tools           *ToolsConfig            `toml:"tools,omitempty"`
	FlatpakHardening *FlatpakHardeningConfig `toml:"flatpak_hardening,omitempty"`
	PodmanRootless  *PodmanRootlessConfig   `toml:"podman_rootless,omitempty"`
}

// ── Flatpak ──────────────────────────────────────────────────────────

// FlatpakRemote defines a Flatpak remote repository.
type FlatpakRemote struct {
	Name string `toml:"name" json:"name"`
	URL  string `toml:"url" json:"url"`
}

// FlatpakConfig defines Flatpak package management settings.
type FlatpakConfig struct {
	Remotes         []FlatpakRemote        `toml:"remotes,omitempty" json:"remotes,omitempty"`
	Packages        []string               `toml:"packages,omitempty" json:"packages,omitempty"`
	ManageOverrides bool                   `toml:"manage_overrides" json:"manage_overrides"`
	Overrides       map[string]map[string]any `toml:"overrides,omitempty" json:"overrides,omitempty"`
}

// ── Containers ───────────────────────────────────────────────────────

// ContainerScope is user or system.
type ContainerScope string

const (
	ScopeUser   ContainerScope = "user"
	ScopeSystem ContainerScope = "system"
)

// TmpfsMount defines a tmpfs mount for a container.
type TmpfsMount struct {
	Path    string `toml:"path" json:"path"`
	Options string `toml:"options,omitempty" json:"options,omitempty"`
}

// ContainerSecurityOptions defines security settings for a container.
type ContainerSecurityOptions struct {
	ReadOnlyRootfs  bool       `toml:"read_only_rootfs" json:"read_only_rootfs"`
	DropAllCaps     bool       `toml:"drop_all_caps" json:"drop_all_caps"`
	AddCaps         []string   `toml:"add_caps,omitempty" json:"add_caps,omitempty"`
	NoNewPrivileges bool       `toml:"no_new_privileges" json:"no_new_privileges"`
	SeccompProfile  string     `toml:"seccomp_profile,omitempty" json:"seccomp_profile,omitempty"`
	User            string     `toml:"user,omitempty" json:"user,omitempty"`
	AppArmorProfile string     `toml:"apparmor_profile,omitempty" json:"apparmor_profile,omitempty"`
	DNS             []string   `toml:"dns,omitempty" json:"dns,omitempty"`
	DNSSearch       []string   `toml:"dns_search,omitempty" json:"dns_search,omitempty"`
	DNSOptions      []string   `toml:"dns_options,omitempty" json:"dns_options,omitempty"`
	Tmpfs           []TmpfsMount `toml:"tmpfs,omitempty" json:"tmpfs,omitempty"`
}

// ContainerBuildConfig defines how to build a container image.
type ContainerBuildConfig struct {
	Context    string            `toml:"context" json:"context"`
	Dockerfile string            `toml:"dockerfile,omitempty" json:"dockerfile,omitempty"`
	BuildArgs  map[string]string `toml:"build_args,omitempty" json:"build_args,omitempty"`
	ExtraFlags []string          `toml:"extra_flags,omitempty" json:"extra_flags,omitempty"`
}

// SetupCommand is a pre-container-setup command with a description.
type SetupCommand struct {
	Description string `toml:"description" json:"description"`
	Command     string `toml:"command" json:"command"`
}

// Container defines a single Podman container.
type Container struct {
	Name       string                    `toml:"name" json:"name"`
	Image      string                    `toml:"image" json:"image"`
	RawFlags   string                    `toml:"raw_flags,omitempty" json:"raw_flags,omitempty"`
	Autostart  *bool                     `toml:"autostart,omitempty" json:"autostart,omitempty"`
	Build      *ContainerBuildConfig     `toml:"build,omitempty" json:"build,omitempty"`
	Scope      ContainerScope            `toml:"scope" json:"scope"`
	RawQuadlet string                    `toml:"raw_quadlet,omitempty" json:"raw_quadlet,omitempty"`
	Security   *ContainerSecurityOptions `toml:"security,omitempty" json:"security,omitempty"`
}

// PodmanConfig defines Podman container management settings.
type PodmanConfig struct {
	PreContainerSetup []SetupCommand `toml:"pre_container_setup,omitempty" json:"pre_container_setup,omitempty"`
	Containers        []Container    `toml:"containers,omitempty" json:"containers,omitempty"`
}

// ── Devbox ───────────────────────────────────────────────────────────

// DevboxInstance defines a Distrobox instance.
type DevboxInstance struct {
	Name               string            `toml:"name" json:"name"`
	Image              string            `toml:"image" json:"image"`
	Home               string            `toml:"home,omitempty" json:"home,omitempty"`
	AdditionalPackages []string          `toml:"additional_packages,omitempty" json:"additional_packages,omitempty"`
	InitHooks          []string          `toml:"init_hooks,omitempty" json:"init_hooks,omitempty"`
	Packages           []string          `toml:"packages,omitempty" json:"packages,omitempty"`
	PostCreate         []string          `toml:"post_create,omitempty" json:"post_create,omitempty"`
	Environment        map[string]string `toml:"environment,omitempty" json:"environment,omitempty"`
	ExportApps         []string          `toml:"export_apps,omitempty" json:"export_apps,omitempty"`
	Flags              []string          `toml:"flags,omitempty" json:"flags,omitempty"`
}

// DevboxProject maps a filesystem path to a Distrobox instance for auto-enter.
type DevboxProject struct {
	Path      string `toml:"path" json:"path"`
	BoxName   string `toml:"box_name" json:"box_name"`
	AutoEnter bool   `toml:"auto_enter" json:"auto_enter"`
	Hook      string `toml:"hook,omitempty" json:"hook,omitempty"`
}

// DevboxConfig defines Distrobox management settings.
type DevboxConfig struct {
	Instances []DevboxInstance `toml:"instances,omitempty" json:"instances,omitempty"`
	Projects  []DevboxProject `toml:"projects,omitempty" json:"projects,omitempty"`
}

// ── Tools ────────────────────────────────────────────────────────────

// ToolVersion defines a tool managed by mise.
type ToolVersion struct {
	Name          string `toml:"name" json:"name"`
	Version       string `toml:"version" json:"version"`
	GlobalInstall bool   `toml:"global_install" json:"global_install"`
}

// ToolsConfig defines mise tool version management settings.
type ToolsConfig struct {
	Tools            []ToolVersion `toml:"tools,omitempty" json:"tools,omitempty"`
	ShellIntegration bool          `toml:"shell_integration" json:"shell_integration"`
}

// ── Flatpak Hardening ────────────────────────────────────────────────

// FlatpakNetworkPolicy controls Flatpak network access.
type FlatpakNetworkPolicy string

const (
	NetworkAllow FlatpakNetworkPolicy = "allow"
	NetworkDeny  FlatpakNetworkPolicy = "deny"
)

// FlatpakHardeningConfig defines global Flatpak permission restrictions.
type FlatpakHardeningConfig struct {
	Enabled            bool                 `toml:"enabled" json:"enabled"`
	RestrictFilesystem bool                 `toml:"restrict_filesystem" json:"restrict_filesystem"`
	NetworkPolicy      FlatpakNetworkPolicy `toml:"network_policy" json:"network_policy"`
	DefaultDenials     []string             `toml:"default_denials,omitempty" json:"default_denials,omitempty"`
}

// ── Podman Rootless ──────────────────────────────────────────────────

// PodmanRootlessConfig configures rootless Podman sysctl settings.
type PodmanRootlessConfig struct {
	Enabled           bool `toml:"enabled" json:"enabled"`
	MaxUserNamespaces int  `toml:"max_user_namespaces" json:"max_user_namespaces"`
}
