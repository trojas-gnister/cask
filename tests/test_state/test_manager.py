import json
from cask.state.manager import StateManager


def test_state_manager_save_load(tmp_path):
    mgr = StateManager(str(tmp_path / "state"))
    mgr.mark_applied("pacman", "abc123")
    mgr.save()

    mgr2 = StateManager(str(tmp_path / "state"))
    mgr2.load()
    assert mgr2.has_changed("pacman", "abc123") is False
    assert mgr2.has_changed("pacman", "different") is True


def test_state_manager_empty(tmp_path):
    mgr = StateManager(str(tmp_path / "state"))
    mgr.load()
    assert mgr.has_changed("pacman", "anything") is True
