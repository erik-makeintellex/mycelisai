from __future__ import annotations

from invoke import Context
import pytest
import subprocess
from pathlib import Path

from ops import db as db_tasks
from ops import lifecycle


def test_memory_restart_runs_order_and_probes(monkeypatch):
    order: list[str] = []
    probe_calls: list[tuple[str, float, dict[str, str] | None]] = []

    monkeypatch.setenv("MYCELIS_API_KEY", "test-key")
    monkeypatch.setattr(lifecycle, "down", lambda _c: order.append("down"))
    monkeypatch.setattr(lifecycle, "_ensure_bridge", lambda: order.append("bridge"))
    monkeypatch.setattr(
        lifecycle,
        "_wait_for_port",
        lambda port, label, timeout=30, interval=1.0: order.append(f"wait:{port}") or True,
    )
    monkeypatch.setattr(db_tasks, "reset", lambda _c: order.append("db.reset"))
    monkeypatch.setattr(lifecycle, "up", lambda _c, frontend=False, build=False: order.append(f"up:{build}:{frontend}"))
    monkeypatch.setattr(lifecycle, "health", lambda _c: order.append("health"))
    monkeypatch.setattr(lifecycle.time, "sleep", lambda _n: None)
    monkeypatch.setattr(lifecycle, "_load_env", lambda: None)

    def fake_http_get(url: str, timeout: float = 3.0, headers: dict[str, str] | None = None):
        probe_calls.append((url, timeout, headers))
        return 200, "ok"

    monkeypatch.setattr(lifecycle, "_http_get", fake_http_get)

    lifecycle.memory_restart.body(Context(), build=True, frontend=True)

    assert order == ["down", "bridge", "wait:5432", "db.reset", "up:True:True", "health"]
    assert len(probe_calls) == 2
    assert probe_calls[0][0].endswith("/api/v1/memory/stream")
    assert probe_calls[1][0].endswith("/api/v1/memory/sitreps?limit=1")
    assert probe_calls[0][2] == {"Authorization": "Bearer test-key"}


def test_memory_restart_fails_when_memory_probe_fails(monkeypatch):
    monkeypatch.setenv("MYCELIS_API_KEY", "test-key")
    monkeypatch.setattr(lifecycle, "down", lambda _c: None)
    monkeypatch.setattr(lifecycle, "_ensure_bridge", lambda: None)
    monkeypatch.setattr(lifecycle, "_wait_for_port", lambda *args, **kwargs: True)
    monkeypatch.setattr(db_tasks, "reset", lambda _c: None)
    monkeypatch.setattr(lifecycle, "up", lambda _c, frontend=False, build=False: None)
    monkeypatch.setattr(lifecycle, "health", lambda _c: None)
    monkeypatch.setattr(lifecycle.time, "sleep", lambda _n: None)
    monkeypatch.setattr(lifecycle, "_load_env", lambda: None)
    monkeypatch.setattr(
        lifecycle,
        "_http_get",
        lambda _url, timeout=3.0, headers=None: (503, "memory unavailable"),
    )

    with pytest.raises(SystemExit):
        lifecycle.memory_restart.body(Context(), build=False, frontend=False)


def test_memory_restart_fails_when_postgres_bridge_does_not_return(monkeypatch):
    monkeypatch.setattr(lifecycle, "down", lambda _c: None)
    monkeypatch.setattr(lifecycle, "_ensure_bridge", lambda: None)
    monkeypatch.setattr(lifecycle, "_wait_for_port", lambda *args, **kwargs: False)
    monkeypatch.setattr(lifecycle.time, "sleep", lambda _n: None)

    with pytest.raises(SystemExit, match="PostgreSQL bridge not reachable"):
        lifecycle.memory_restart.body(Context(), build=False, frontend=False)


def test_kill_pid_ignores_windows_taskkill_timeout(monkeypatch):
    monkeypatch.setattr(lifecycle, "is_windows", lambda: True)

    def fake_run(*args, **kwargs):
        raise subprocess.TimeoutExpired(cmd="taskkill", timeout=15)

    monkeypatch.setattr(lifecycle.subprocess, "run", fake_run)

    lifecycle._kill_pid(1234)


def test_wait_for_http_ok_returns_true_on_200(monkeypatch):
    calls: list[str] = []

    def fake_http_get(url: str, timeout: float = 5.0, headers=None):
        calls.append(url)
        return 200, "ok"

    monkeypatch.setattr(lifecycle, "_http_get", fake_http_get)

    assert lifecycle._wait_for_http_ok("http://localhost:8081/healthz", "Core health", timeout=1, interval=0.01)
    assert calls == ["http://localhost:8081/healthz"]


def test_up_fails_when_core_health_never_becomes_ready(monkeypatch):
    monkeypatch.setattr(Path, "exists", lambda self: True)
    monkeypatch.setattr(lifecycle, "_ensure_bridge", lambda: None)
    monkeypatch.setattr(lifecycle, "_wait_for_port", lambda *args, **kwargs: True)
    monkeypatch.setattr(lifecycle, "_start_core_background", lambda: True)
    monkeypatch.setattr(lifecycle, "_port_open", lambda port, host="127.0.0.1", timeout=1.0: False if port == lifecycle.API_PORT else True)
    monkeypatch.setattr(lifecycle, "_wait_for_http_ok", lambda *args, **kwargs: False)

    with pytest.raises(SystemExit, match="Core port opened but /healthz never became ready"):
        lifecycle.up.body(Context(), frontend=False, build=False)
