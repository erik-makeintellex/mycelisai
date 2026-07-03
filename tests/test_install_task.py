from tests.ci_task_support import FakeContext

import tasks
from ops import cache, config


def test_install_provisions_reticulum_through_uv_and_uvx(monkeypatch):
    monkeypatch.setattr(config, "ensure_managed_cache_dirs", lambda: None)
    monkeypatch.setattr(cache, "ensure_disk_headroom", lambda **_kwargs: None)
    monkeypatch.setattr(config, "managed_cache_env", lambda: {"UV_CACHE_DIR": "workspace/tool-cache/uv"})

    ctx = FakeContext({})

    tasks.install.body(ctx)

    assert "uv sync --all-packages --dev" in ctx.commands
    assert 'uv run python -c "import RNS; print(RNS.__version__)"' in ctx.commands
    assert tasks.RETICULUM_UVX_PROBE in ctx.commands
    assert ctx.commands.index("uv sync --all-packages --dev") < ctx.commands.index(tasks.RETICULUM_UVX_PROBE)
