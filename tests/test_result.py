from cask.result import Result, ExecResult


def test_result_ok():
    r = Result(ok=True, message="done", actions=["installed firefox"])
    assert r.ok
    assert r.message == "done"
    assert r.actions == ["installed firefox"]


def test_result_fail():
    r = Result(ok=False, message="failed", actions=[])
    assert not r.ok


def test_exec_result():
    r = ExecResult(exit_code=0, stdout="output\n", stderr="")
    assert r.exit_code == 0
    assert r.stdout == "output\n"
    assert r.stderr == ""
