from __future__ import annotations

import shutil
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
    assert float(env["MYCELIS_CACHE_MIN_FREE_GB"]) > 0
    assert float(env["MYCELIS_CACHE_MAX_GB"]) > 0
    assert float(env["MYCELIS_PLAYWRIGHT_CACHE_MAX_GB"]) > 0
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

    monkeypatch.delenv("MYCELIS_USER_CACHE_ROOT", raising=False)
    monkeypatch.setattr(cache, "is_windows", lambda: True)
    monkeypatch.setattr(cache, "_default_user_cache_root", lambda: tmp_path / "user-cache")
    monkeypatch.setattr(cache, "_user_cache_root", lambda root=None: tmp_path / "user-cache")
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


def test_cache_clean_handles_read_only_files(monkeypatch, tmp_path):
    project_root = tmp_path / "workspace" / "tool-cache"
    go_mod_dir = project_root / "go-mod"
    go_mod_dir.mkdir(parents=True)
    read_only = go_mod_dir / "artifact.txt"
    read_only.write_text("locked", encoding="utf-8")
    read_only.chmod(0o444)

    monkeypatch.setattr(cache, "PROJECT_CACHE_ROOT", project_root)
    monkeypatch.setattr(cache, "PROJECT_CACHE_ARTIFACTS", ())

    cache.clean.body(Context(), project=True, user=False)

    assert go_mod_dir.exists()
    assert list(go_mod_dir.iterdir()) == []


def test_cache_clean_retries_transient_directory_not_empty(monkeypatch, tmp_path):
    project_root = tmp_path / "workspace" / "tool-cache"
    uv_dir = project_root / "uv"
    uv_dir.mkdir(parents=True)
    (uv_dir / "artifact.bin").write_bytes(b"x")
    attempts = {"count": 0}
    real_rmtree = shutil.rmtree

    def fake_rmtree(path, ignore_errors=False, onexc=None):
        attempts["count"] += 1
        if attempts["count"] == 1:
            error = OSError(145, "The directory is not empty")
            error.winerror = 145
            raise error
        Path(path).unlink(missing_ok=True) if Path(path).is_file() else real_rmtree(path, ignore_errors=ignore_errors, onexc=onexc)

    monkeypatch.setattr(cache, "PROJECT_CACHE_ROOT", project_root)
    monkeypatch.setattr(cache, "PROJECT_CACHE_ARTIFACTS", ())
    monkeypatch.setattr(cache, "ensure_managed_cache_dirs", lambda root=None: {"root": project_root, "uv": uv_dir})
    monkeypatch.setattr(cache.time, "sleep", lambda _n: None)
    monkeypatch.setattr(cache.shutil, "rmtree", fake_rmtree)

    cache.clean.body(Context(), project=True, user=False)

    assert attempts["count"] == 2
    assert uv_dir.exists()
    assert list(uv_dir.iterdir()) == []


def test_cache_guard_fails_when_free_space_is_below_threshold(monkeypatch, tmp_path):
    project_root = tmp_path / "workspace" / "tool-cache"
    project_root.mkdir(parents=True)

    monkeypatch.setattr(cache, "PROJECT_CACHE_ROOT", project_root)
    monkeypatch.setattr(cache, "ROOT_DIR", tmp_path)
    monkeypatch.setattr(cache.shutil, "disk_usage", lambda _path: shutil._ntuple_diskusage(100, 95, 5))

    try:
        cache.guard.body(Context(), min_free_gb=8)
    except SystemExit as exc:
        assert "DISK/CACHE POLICY CHECK FAILED" in str(exc)
    else:
        raise AssertionError("cache.guard should fail under low disk headroom")


def test_cache_guard_checks_user_volume_with_default_targets(monkeypatch, tmp_path):
    repo_path = tmp_path / "repo"
    user_path = tmp_path / "user"
    repo_path.mkdir()
    user_path.mkdir()
    gib = 1024 ** 3

    monkeypatch.setattr(
        cache,
        "_disk_targets",
        lambda paths=None: [("repo", repo_path), ("user", user_path)],
    )
    monkeypatch.setattr(
        cache.shutil,
        "disk_usage",
        lambda path: shutil._ntuple_diskusage(100 * gib, 95 * gib, 5 * gib)
        if path == user_path
        else shutil._ntuple_diskusage(100 * gib, 50 * gib, 50 * gib),
    )

    try:
        cache.guard.body(Context(), min_free_gb=8)
    except SystemExit as exc:
        assert "user has only 5.0 GiB free" in str(exc)
    else:
        raise AssertionError("cache.guard should fail when the user volume is low")


def test_cache_status_reports_disk_headroom(monkeypatch, tmp_path, capsys):
    project_root = tmp_path / "workspace" / "tool-cache"
    project_root.mkdir(parents=True)

    monkeypatch.setattr(cache, "PROJECT_CACHE_ROOT", project_root)
    monkeypatch.setattr(cache, "ROOT_DIR", tmp_path)
    monkeypatch.setattr(cache.shutil, "disk_usage", lambda _path: shutil._ntuple_diskusage(200, 100, 100))

    cache.status.body(Context())

    output = capsys.readouterr().out
    assert "Adaptive policy:" in output
    assert "Disk headroom:" in output
    assert "Docker storage shares this volume" in output


def test_managed_cache_policy_scales_and_honors_caps(monkeypatch, tmp_path):
    gib = 1024 ** 3
    monkeypatch.setattr(config.shutil, "disk_usage", lambda _path: shutil._ntuple_diskusage(200 * gib, 100 * gib, 100 * gib))
    monkeypatch.delenv("MYCELIS_CACHE_MIN_FREE_GB", raising=False)
    monkeypatch.delenv("MYCELIS_CACHE_MAX_GB", raising=False)
    monkeypatch.delenv("MYCELIS_PLAYWRIGHT_CACHE_MAX_GB", raising=False)

    policy = config.managed_cache_policy(tmp_path)

    assert policy["reserve_gb"] == 10.0
    assert policy["cache_max_gb"] == 22.5
    assert policy["playwright_max_gb"] == 5.625


def test_cache_guard_fails_when_playwright_exceeds_adaptive_budget(monkeypatch, tmp_path):
    project_root = tmp_path / "workspace" / "tool-cache"
    playwright = project_root / "playwright"
    playwright.mkdir(parents=True)
    monkeypatch.setattr(cache, "PROJECT_CACHE_ROOT", project_root)
    monkeypatch.setattr(cache, "_disk_targets", lambda paths=None: [("repo", tmp_path)])
    monkeypatch.setattr(cache.shutil, "disk_usage", lambda _path: shutil._ntuple_diskusage(100, 20, 80))
    monkeypatch.setattr(cache, "_path_size_bytes", lambda path: 3 * 1024 ** 3 if path == playwright else 4 * 1024 ** 3)
    monkeypatch.setattr(
        cache,
        "managed_cache_policy",
        lambda root=None: {"reserve_gb": 8.0, "cache_max_gb": 10.0, "playwright_max_gb": 2.0},
    )

    try:
        cache.ensure_disk_headroom(min_free_gb=1)
    except SystemExit as exc:
        assert "Playwright cache uses 3.00 GiB; budget is 2.00 GiB" in str(exc)
    else:
        raise AssertionError("cache guard should fail when Playwright exceeds its budget")
