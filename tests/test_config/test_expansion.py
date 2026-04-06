import os
from cask.config.expansion import expand_value, expand_config


def test_expand_tilde():
    result = expand_value("~/projects")
    assert result == os.path.expanduser("~/projects")


def test_expand_env_var(monkeypatch):
    monkeypatch.setenv("MY_VAR", "hello")
    assert expand_value("${MY_VAR}") == "hello"


def test_expand_env_var_with_default(monkeypatch):
    monkeypatch.delenv("MISSING_VAR", raising=False)
    assert expand_value("${MISSING_VAR:-fallback}") == "fallback"


def test_expand_env_var_missing_no_default(monkeypatch):
    monkeypatch.delenv("MISSING_VAR", raising=False)
    assert expand_value("${MISSING_VAR}") == ""


def test_expand_mixed(monkeypatch):
    monkeypatch.setenv("DB_PASS", "secret")
    result = expand_value("~/data/${DB_PASS}/db")
    assert result == os.path.expanduser("~/data/secret/db")


def test_expand_no_expansion():
    assert expand_value("plain string") == "plain string"


def test_expand_config_recursive(monkeypatch):
    monkeypatch.setenv("PORT", "8080")
    data = {
        "name": "test",
        "path": "~/config",
        "port": "${PORT}",
        "nested": {"inner": "${PORT:-3000}"},
        "items": ["~/a", "${PORT}"],
    }
    result = expand_config(data)
    assert result["path"] == os.path.expanduser("~/config")
    assert result["port"] == "8080"
    assert result["nested"]["inner"] == "8080"
    assert result["items"][0] == os.path.expanduser("~/a")
    assert result["items"][1] == "8080"
    assert result["name"] == "test"
