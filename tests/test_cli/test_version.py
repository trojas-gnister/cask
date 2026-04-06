"""Tests for the version CLI command."""
from typer.testing import CliRunner
from cask.cli.app import app

runner = CliRunner()


def test_version_exit_code():
    result = runner.invoke(app, ["version"])
    assert result.exit_code == 0


def test_version_output_contains_version_string():
    result = runner.invoke(app, ["version"])
    assert "0.1.0" in result.output


def test_version_output_contains_cask():
    result = runner.invoke(app, ["version"])
    assert "cask" in result.output
