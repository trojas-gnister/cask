import pytest
from cask.managers.mise import MiseManager
from cask.executor.mock import MockExecutor


@pytest.mark.asyncio
async def test_install_tool():
    mock = MockExecutor()
    mgr = MiseManager()
    result = await mgr.install("node", "22.0.0", mock)
    assert result.ok
    assert mock.was_called("mise install")


@pytest.mark.asyncio
async def test_list_tools():
    mock = MockExecutor()
    mock.add_response("mise", stdout='{"node": {"version": "22.0.0"}}\n')
    mgr = MiseManager()
    tools = await mgr.list_tools(mock)
    assert "node" in tools
