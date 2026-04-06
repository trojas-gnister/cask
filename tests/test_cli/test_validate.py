"""Tests for the validate CLI command."""
import os
import tempfile

import pytest
from typer.testing import CliRunner

from cask.cli.app import app

runner = CliRunner()

FIXTURES_DIR = os.path.join(os.path.dirname(__file__), "..", "fixtures")
VALID_CONFIG = os.path.join(FIXTURES_DIR, "config.toml")


def test_validate_valid_config():
    result = runner.invoke(app, ["-c", VALID_CONFIG, "validate"])
    assert result.exit_code == 0
    assert "Config OK" in result.output


def test_validate_invalid_config(tmp_path):
    bad_config = tmp_path / "bad.toml"
    bad_config.write_text('[pacman]\npackages = "not-a-list"\n')
    result = runner.invoke(app, ["-c", str(bad_config), "validate"])
    # Should either show error or non-zero exit (pydantic validation failure)
    assert result.exit_code != 0 or "error" in result.output.lower()


def test_validate_missing_config(tmp_path):
    missing = str(tmp_path / "nonexistent.toml")
    result = runner.invoke(app, ["-c", missing, "validate"])
    assert "error" in result.output.lower() or "not found" in result.output.lower()


def test_validate_uses_config_flag():
    """Passing --config/-c should override the default path."""
    result = runner.invoke(app, ["validate", "-c", VALID_CONFIG])
    assert result.exit_code == 0
    assert "Config OK" in result.output
