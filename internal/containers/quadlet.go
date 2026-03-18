package containers

import (
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/iskry/cask/internal/config"
)

// QuadletDir returns the appropriate Quadlet directory based on container scope.
func QuadletDir(scope config.ContainerScope) string {
	if scope == config.ScopeSystem {
		return "/etc/containers/systemd"
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "containers", "systemd")
}

// ParseRawFlags parses podman raw_flags into Quadlet directives.
func ParseRawFlags(rawFlags string) map[string][]string {
	if rawFlags == "" {
		return nil
	}

	directives := make(map[string][]string)
	tokens := splitFlags(rawFlags)
	i := 0
	for i < len(tokens) {
		flag := tokens[i]
		switch {
		case (flag == "-p" || flag == "--publish") && i+1 < len(tokens):
			directives["PublishPort"] = append(directives["PublishPort"], tokens[i+1])
			i += 2
		case (flag == "-v" || flag == "--volume") && i+1 < len(tokens):
			directives["Volume"] = append(directives["Volume"], tokens[i+1])
			i += 2
		case (flag == "-e" || flag == "--env") && i+1 < len(tokens):
			directives["Environment"] = append(directives["Environment"], tokens[i+1])
			i += 2
		case flag == "--network" && i+1 < len(tokens):
			directives["Network"] = append(directives["Network"], tokens[i+1])
			i += 2
		case flag == "--label" && i+1 < len(tokens):
			directives["Label"] = append(directives["Label"], tokens[i+1])
			i += 2
		case flag == "--name" && i+1 < len(tokens):
			i += 2 // Skip — name comes from container config
		default:
			directives["PodmanArgs"] = append(directives["PodmanArgs"], flag)
			i++
		}
	}
	return directives
}

// GenerateQuadlet generates the content of a .container Quadlet file.
func GenerateQuadlet(c *config.Container) string {
	var lines []string
	lines = append(lines, "[Container]")
	lines = append(lines, fmt.Sprintf("ContainerName=%s", c.Name))
	lines = append(lines, fmt.Sprintf("Image=%s", c.Image))

	// Parse raw_flags into Quadlet directives
	directives := ParseRawFlags(c.RawFlags)
	for key, values := range directives {
		if key == "PodmanArgs" {
			lines = append(lines, fmt.Sprintf("PodmanArgs=%s", strings.Join(values, " ")))
		} else {
			for _, val := range values {
				lines = append(lines, fmt.Sprintf("%s=%s", key, val))
			}
		}
	}

	// Security options
	if c.Security != nil {
		sec := c.Security
		if sec.ReadOnlyRootfs {
			lines = append(lines, "ReadOnly=true")
		}
		if sec.NoNewPrivileges {
			lines = append(lines, "SecurityLabelNested=true")
		}
		if sec.User != "" {
			lines = append(lines, fmt.Sprintf("User=%s", sec.User))
		}
		for _, cap := range sec.AddCaps {
			lines = append(lines, fmt.Sprintf("AddCapability=%s", cap))
		}
		if sec.DropAllCaps {
			lines = append(lines, "DropCapability=ALL")
		}
		if sec.AppArmorProfile != "" {
			lines = append(lines, fmt.Sprintf("SecurityLabelType=%s", sec.AppArmorProfile))
		}
		if sec.SeccompProfile != "" && sec.SeccompProfile != "default" {
			lines = append(lines, fmt.Sprintf("PodmanArgs=--security-opt=seccomp=%s", sec.SeccompProfile))
		}
		for _, ip := range sec.DNS {
			lines = append(lines, fmt.Sprintf("PodmanArgs=--dns %s", ip))
		}
		for _, domain := range sec.DNSSearch {
			lines = append(lines, fmt.Sprintf("PodmanArgs=--dns-search %s", domain))
		}
		for _, opt := range sec.DNSOptions {
			lines = append(lines, fmt.Sprintf("PodmanArgs=--dns-option %s", opt))
		}
		for _, mount := range sec.Tmpfs {
			val := mount.Path
			if mount.Options != "" {
				val = fmt.Sprintf("%s:%s", mount.Path, mount.Options)
			}
			lines = append(lines, fmt.Sprintf("Tmpfs=%s", val))
		}
	}

	// Raw quadlet passthrough
	if c.RawQuadlet != "" {
		lines = append(lines, "")
		lines = append(lines, c.RawQuadlet)
	}

	// Install section
	lines = append(lines, "")
	lines = append(lines, "[Install]")
	if c.Autostart != nil && *c.Autostart {
		lines = append(lines, "WantedBy=default.target")
	}

	return strings.Join(lines, "\n") + "\n"
}

// WriteQuadlet writes a .container Quadlet file if content has changed.
// Returns (path, changed, error).
func WriteQuadlet(c *config.Container) (string, bool, error) {
	dir := QuadletDir(c.Scope)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", false, err
	}

	path := filepath.Join(dir, c.Name+".container")
	content := GenerateQuadlet(c)

	// Idempotency: skip if file exists with identical content
	if existing, err := os.ReadFile(path); err == nil {
		if contentHash(string(existing)) == contentHash(content) {
			return path, false, nil
		}
	}

	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		return "", false, err
	}
	return path, true, nil
}

func contentHash(s string) string {
	h := sha256.Sum256([]byte(s))
	return fmt.Sprintf("%x", h[:8])
}
