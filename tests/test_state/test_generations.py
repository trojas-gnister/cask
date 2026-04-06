import json
from cask.state.generations import GenerationManager


def test_generations_create_and_list(tmp_path):
    gm = GenerationManager(str(tmp_path / "generations"))
    gm.create({"pacman": {"hash": "abc123"}})
    gm.create({"pacman": {"hash": "def456"}})

    gens = gm.list_generations()
    assert len(gens) == 2


def test_generations_empty(tmp_path):
    gm = GenerationManager(str(tmp_path / "generations"))
    assert gm.list_generations() == []
