from cask.state.lockfile import Lockfile


def test_lockfile_save_load(tmp_path):
    lf = Lockfile(str(tmp_path / "lock.json"))
    lf.pin("firefox", "138.0-1")
    lf.pin("git", "2.47.0-1")
    lf.save()

    lf2 = Lockfile(str(tmp_path / "lock.json"))
    lf2.load()
    assert lf2.get("firefox") == "138.0-1"
    assert lf2.get("git") == "2.47.0-1"
    assert lf2.get("missing") is None


def test_lockfile_verify():
    lf = Lockfile("/tmp/test_lock.json")
    lf.pin("firefox", "138.0-1")
    assert lf.verify("firefox", "138.0-1") is True
    assert lf.verify("firefox", "139.0-1") is False
    assert lf.verify("missing", "1.0") is True  # Not pinned = OK
