"""Tests for quadlet unit file generation."""
import pytest
from cask.managers.quadlet import generate_quadlet
from cask.config.models import ContainerConfig, ContainerSecurity


def test_generate_quadlet_basic_structure():
    cfg = ContainerConfig(image="nginx:latest")
    content = generate_quadlet("web", cfg)
    assert "[Unit]" in content
    assert "[Container]" in content
    assert "[Install]" in content
    assert "WantedBy=default.target" in content


def test_generate_quadlet_image():
    cfg = ContainerConfig(image="nginx:latest")
    content = generate_quadlet("web", cfg)
    assert "Image=nginx:latest" in content


def test_generate_quadlet_description():
    cfg = ContainerConfig(image="nginx:latest")
    content = generate_quadlet("myapp", cfg)
    assert "Description=myapp container" in content


def test_generate_quadlet_ports():
    cfg = ContainerConfig(image="nginx:latest", ports=["8080:80", "443:443"])
    content = generate_quadlet("web", cfg)
    assert "PublishPort=8080:80" in content
    assert "PublishPort=443:443" in content


def test_generate_quadlet_volumes():
    cfg = ContainerConfig(image="nginx:latest", volumes=["/data:/data:ro"])
    content = generate_quadlet("web", cfg)
    assert "Volume=/data:/data:ro" in content


def test_generate_quadlet_environment():
    cfg = ContainerConfig(image="nginx:latest", environment={"APP_ENV": "production"})
    content = generate_quadlet("web", cfg)
    assert "Environment=APP_ENV=production" in content


def test_generate_quadlet_no_new_privileges():
    security = ContainerSecurity(no_new_privileges=True)
    cfg = ContainerConfig(image="nginx:latest", security=security)
    content = generate_quadlet("web", cfg)
    assert "NoNewPrivileges=true" in content


def test_generate_quadlet_read_only():
    security = ContainerSecurity(read_only=True)
    cfg = ContainerConfig(image="nginx:latest", security=security)
    content = generate_quadlet("web", cfg)
    assert "ReadOnly=true" in content


def test_generate_quadlet_security_flags_absent_by_default():
    cfg = ContainerConfig(image="nginx:latest")
    content = generate_quadlet("web", cfg)
    assert "NoNewPrivileges" not in content
    assert "ReadOnly" not in content


def test_generate_quadlet_no_ports_when_empty():
    cfg = ContainerConfig(image="nginx:latest", ports=[])
    content = generate_quadlet("web", cfg)
    assert "PublishPort" not in content
