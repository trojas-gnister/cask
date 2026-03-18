package containers

import (
	"context"
	"strings"
	"testing"

	"github.com/iskry/cask/internal/config"
	"github.com/iskry/cask/internal/executor"
)

func TestBuildSecurityFlagsNil(t *testing.T) {
	flags := BuildSecurityFlags(nil)
	if len(flags) != 0 {
		t.Error("nil security should produce no flags")
	}
}

func TestBuildSecurityFlagsFull(t *testing.T) {
	sec := &config.ContainerSecurityOptions{
		ReadOnlyRootfs:  true,
		DropAllCaps:     true,
		AddCaps:         []string{"NET_BIND_SERVICE", "SYS_TIME"},
		NoNewPrivileges: true,
		SeccompProfile:  "/path/to/profile.json",
		User:            "1000:1000",
		AppArmorProfile: "custom-profile",
		DNS:             []string{"1.1.1.1"},
		DNSSearch:       []string{"example.com"},
		DNSOptions:      []string{"ndots:1"},
		Tmpfs: []config.TmpfsMount{
			{Path: "/tmp", Options: "size=100m"},
		},
	}

	flags := BuildSecurityFlags(sec)
	joined := strings.Join(flags, " ")

	expected := []string{
		"--read-only",
		"--cap-drop=ALL",
		"--cap-add=NET_BIND_SERVICE",
		"--cap-add=SYS_TIME",
		"--security-opt=no-new-privileges",
		"--security-opt=seccomp=/path/to/profile.json",
		"--user", "1000:1000",
		"--security-opt=apparmor=custom-profile",
		"--dns", "1.1.1.1",
		"--dns-search", "example.com",
		"--dns-option", "ndots:1",
		"--tmpfs", "/tmp:size=100m",
	}
	for _, e := range expected {
		if !strings.Contains(joined, e) {
			t.Errorf("missing flag: %s in %s", e, joined)
		}
	}
}

func TestBuildSecurityFlagsDefaultSeccomp(t *testing.T) {
	sec := &config.ContainerSecurityOptions{SeccompProfile: "default"}
	flags := BuildSecurityFlags(sec)
	for _, f := range flags {
		if strings.Contains(f, "seccomp") {
			t.Error("default seccomp should not produce a flag")
		}
	}
}

func TestDetectPackageManager(t *testing.T) {
	tests := []struct {
		image    string
		expected string
	}{
		{"docker.io/library/ubuntu:22.04", "apt"},
		{"fedora:39", "dnf"},
		{"archlinux:latest", "pacman"},
		{"alpine:3.18", "apk"},
		{"opensuse/tumbleweed", "zypper"},
		{"centos:8", "dnf"},
		{"unknown-distro:1", "dnf"},
	}

	for _, tt := range tests {
		name, _ := DetectPackageManager(tt.image)
		if name != tt.expected {
			t.Errorf("DetectPackageManager(%s): expected %s, got %s", tt.image, tt.expected, name)
		}
	}
}

func TestBuildHomeFlagsHost(t *testing.T) {
	inst := &config.DevboxInstance{Name: "test", Home: "host"}
	flags := BuildHomeFlags(inst)
	if len(flags) != 0 {
		t.Error("host home should produce no flags")
	}
}

func TestBuildHomeFlagsEmpty(t *testing.T) {
	inst := &config.DevboxInstance{Name: "test"}
	flags := BuildHomeFlags(inst)
	if len(flags) != 0 {
		t.Error("empty home should produce no flags")
	}
}

func TestBuildHomeFlagsIsolated(t *testing.T) {
	inst := &config.DevboxInstance{Name: "dev", Home: "isolated"}
	flags := BuildHomeFlags(inst)
	if len(flags) != 3 {
		t.Fatalf("expected 3 flags, got %d", len(flags))
	}
	if flags[0] != "--no-entry" {
		t.Error("first flag should be --no-entry")
	}
	if !strings.Contains(flags[2], "dev") {
		t.Error("home path should contain instance name")
	}
}

func TestBuildHomeFlagsCustom(t *testing.T) {
	inst := &config.DevboxInstance{Name: "dev", Home: "/custom/home"}
	flags := BuildHomeFlags(inst)
	if len(flags) != 2 || flags[1] != "/custom/home" {
		t.Errorf("expected [--home /custom/home], got %v", flags)
	}
}

func TestGenerateQuadletBasic(t *testing.T) {
	autostart := true
	c := &config.Container{
		Name:      "postgres",
		Image:     "docker.io/library/postgres:16",
		Autostart: &autostart,
	}
	content := GenerateQuadlet(c)
	if !strings.Contains(content, "ContainerName=postgres") {
		t.Error("should contain container name")
	}
	if !strings.Contains(content, "Image=docker.io/library/postgres:16") {
		t.Error("should contain image")
	}
	if !strings.Contains(content, "WantedBy=default.target") {
		t.Error("should contain WantedBy for autostart")
	}
}

func TestGenerateQuadletWithRawFlags(t *testing.T) {
	c := &config.Container{
		Name:     "web",
		Image:    "nginx",
		RawFlags: `-p 8080:80 -v /data:/usr/share/nginx/html -e NGINX_HOST=localhost --network host`,
	}
	content := GenerateQuadlet(c)
	if !strings.Contains(content, "PublishPort=8080:80") {
		t.Error("should parse -p into PublishPort")
	}
	if !strings.Contains(content, "Volume=/data:/usr/share/nginx/html") {
		t.Error("should parse -v into Volume")
	}
	if !strings.Contains(content, "Environment=NGINX_HOST=localhost") {
		t.Error("should parse -e into Environment")
	}
	if !strings.Contains(content, "Network=host") {
		t.Error("should parse --network into Network")
	}
}

func TestGenerateQuadletWithSecurity(t *testing.T) {
	c := &config.Container{
		Name:  "secure",
		Image: "alpine",
		Security: &config.ContainerSecurityOptions{
			ReadOnlyRootfs: true,
			DropAllCaps:    true,
			AddCaps:        []string{"NET_BIND_SERVICE"},
			User:           "1000",
		},
	}
	content := GenerateQuadlet(c)
	if !strings.Contains(content, "ReadOnly=true") {
		t.Error("should contain ReadOnly")
	}
	if !strings.Contains(content, "DropCapability=ALL") {
		t.Error("should contain DropCapability")
	}
	if !strings.Contains(content, "AddCapability=NET_BIND_SERVICE") {
		t.Error("should contain AddCapability")
	}
	if !strings.Contains(content, "User=1000") {
		t.Error("should contain User")
	}
}

func TestParseRawFlagsEmpty(t *testing.T) {
	d := ParseRawFlags("")
	if d != nil {
		t.Error("empty flags should return nil")
	}
}

func TestParseRawFlagsNameSkipped(t *testing.T) {
	d := ParseRawFlags("--name mycontainer -p 8080:80")
	if _, ok := d["ContainerName"]; ok {
		t.Error("--name should be skipped")
	}
	if len(d["PublishPort"]) != 1 {
		t.Error("should still parse other flags")
	}
}

func TestListContainers(t *testing.T) {
	mock := executor.NewMockExecutor()
	mock.SetResponse("podman ps -a --format {{.Names}}\t{{.Image}}\t{{.Status}}", &executor.CommandResult{
		Success: true,
		Stdout:  "postgres\tdocker.io/library/postgres:16\tUp 2 hours\nredis\tdocker.io/library/redis:7\tExited (0) 1 hour ago\n",
	})

	containers, err := ListContainers(context.Background(), mock)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(containers) != 2 {
		t.Errorf("expected 2 containers, got %d", len(containers))
	}
}

func TestListDevboxes(t *testing.T) {
	mock := executor.NewMockExecutor()
	mock.SetResponse("distrobox list --no-color", &executor.CommandResult{
		Success: true,
		Stdout:  "ID | NAME | STATUS | IMAGE\nabc123 | dev | Up 1 hour | ubuntu:22.04\n",
	})

	instances, err := ListDevboxes(context.Background(), mock)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(instances) != 1 {
		t.Fatalf("expected 1 instance, got %d", len(instances))
	}
	if instances[0]["name"] != "dev" {
		t.Errorf("expected name 'dev', got '%s'", instances[0]["name"])
	}
}
