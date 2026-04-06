from cask.state.hashing import hash_section


def test_hash_section_deterministic():
    data = {"packages": ["firefox", "git"]}
    h1 = hash_section(data)
    h2 = hash_section(data)
    assert h1 == h2
    assert len(h1) == 64  # SHA256 hex


def test_hash_section_changes():
    h1 = hash_section({"packages": ["firefox"]})
    h2 = hash_section({"packages": ["firefox", "git"]})
    assert h1 != h2


def test_hash_section_order_independent():
    h1 = hash_section({"b": 2, "a": 1})
    h2 = hash_section({"a": 1, "b": 2})
    assert h1 == h2
