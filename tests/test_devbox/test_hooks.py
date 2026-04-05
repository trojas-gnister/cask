from cask.devbox.hooks import generate_hook


def test_generate_zsh_hook():
    hooks = {"~/projects/myapp": "fedora-dev"}
    content = generate_hook("zsh", hooks)
    assert "chpwd" in content
    assert "fedora-dev" in content
    assert "projects/myapp" in content


def test_generate_bash_hook():
    hooks = {"~/projects": "dev"}
    content = generate_hook("bash", hooks)
    assert "dev" in content
    assert "projects" in content


def test_generate_fish_hook():
    hooks = {"~/work": "work-env"}
    content = generate_hook("fish", hooks)
    assert "work-env" in content
