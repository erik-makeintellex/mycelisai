from __future__ import annotations

from pathlib import Path

from invoke import Context

from ops import cache
from ops import config


def test_managed_cache_env_points_to_workspace_tool_cache(monkeypatch, tmp_path):
    monkeypatch.setattr(config, "PROJECT_CACHE_ROOT", tmp_path / "tool-cache")

    env = config.managed_cache_env()

    assert env["MYCELIS_PROJECT_CACHE_ROOT"] == str(tmp_path / "tool-cache")
    assert env["UV_CACHE_DIR"] == str(tmp_path / "tool-cache" / "uv")
    assert env["PIP_CACHE_DIR"] == str(tmp_path / "tool-cache" / "pip")
    assert env["NPM_CONFIG_CACHE"] == str(tmp_path / "tool-cache" / "npm")
    assert env["GOCACHE"] == str(tmp_path / "tool-cache" / "go-build")
    assert env["GOMODCACHE"] == str(tmp_path / "tool-cache" / "go-mod")
    assert env["PLAYWRIGHT_BROWSERS_PATH"] == str(tmp_path / "tool-cache" / "playwright")
    assert env["NEXT_TELEMETRY_DISABLED"] == "1"
    assert env["PYTHONPYCACHEPREFIX"] == str(tmp_path / "tool-cache" / "pycache")


def test_cache_clean_removes_project_managed_artifacts(monkeypatch, tmp_path):
    project_root = tmp_path / "workspace" / "tool-cache"
    uv_dir = project_root / "uv"
    uv_dir.mkdir(parents=True)
    (uv_dir / "artifact.bin").write_bytes(b"x" * 16)

    interface_next = tmp_path / "interface" / ".next"
    interface_next.mkdir(parents=True)
    (interface_next / "cache.txt").write_text("next", encoding="utf-8")

    monkeypatch.setattr(cache, "PROJECT_CACHE_ROOT", project_root)
    monkeypatch.setattr(
        cache,
        "PROJECT_CACHE_ARTIFACTS",
        (
            interface_next,
        ),
    )

    cache.clean.body(Context(), project=True, user=False)

    assert uv_dir.exists()
    assert list(uv_dir.iterdir()) == []
    assert not interface_next.exists()


def test_apply_user_policy_sets_expected_windows_env(monkeypatch, tmp_path, capsys):
    assigned: dict[str, str] = {}

    monkeypatch.setattr(cache, "is_windows", lambda: True)
    monkeypatch.setattr(cache, "_default_user_cache_root", lambda: tmp_path / "user-cache")
    monkeypatch.setattr(cache, "_set_windows_user_env", lambda name, value: assigned.__setitem__(name, value))
    monkeypatch.setattr(cache, "_broadcast_windows_env_change", lambda: None)

    cache.apply_user_policy.body(Context(), root="")

    output = capsys.readouterr().out
    assert assigned["MYCELIS_USER_CACHE_ROOT"] == str(tmp_path / "user-cache")
    assert assigned["UV_CACHE_DIR"] == str(tmp_path / "user-cache" / "uv")
    assert assigned["PIP_CACHE_DIR"] == str(tmp_path / "user-cache" / "pip")
    assert assigned["NPM_CONFIG_CACHE"] == str(tmp_path / "user-cache" / "npm")
    assert assigned["GOCACHE"] == str(tmp_path / "user-cache" / "go-build")
    assert assigned["GOMODCACHE"] == str(tmp_path / "user-cache" / "go-mod")
    assert assigned["PLAYWRIGHT_BROWSERS_PATH"] == str(tmp_path / "user-cache" / "playwright")
    assert "User cache policy applied:" in output
