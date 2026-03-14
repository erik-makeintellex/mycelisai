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
        raise subprocess.TimeoutExpired(cmd="taskkill", timeout=5)

    monkeypatch.setattr(lifecycle.subprocess, "run", fake_run)

    lifecycle._kill_pid(1234)


def test_run_best_effort_ignores_timeout(monkeypatch):
    def fake_run(*args, **kwargs):
        raise subprocess.TimeoutExpired(cmd="cleanup", timeout=10)

    monkeypatch.setattr(lifecycle.subprocess, "run", fake_run)

    lifecycle._run_best_effort(["dummy"], timeout=10)


def test_wait_for_port_closed_returns_true_when_port_drops(monkeypatch):
    states = iter([True, False])
    monkeypatch.setattr(lifecycle, "_port_open", lambda *args, **kwargs: next(states))
    monkeypatch.setattr(lifecycle.time, "sleep", lambda _n: None)

    assert lifecycle._wait_for_port_closed(8081, "Core", timeout=1, interval=0.01)


def test_down_uses_best_effort_cleanup_without_hanging(monkeypatch):
    commands: list[list[str]] = []

    monkeypatch.setattr(lifecycle, "_kill_port", lambda port, label: False)
    monkeypatch.setattr(lifecycle, "_kill_bridges", lambda: commands.append(["bridges"]))
    monkeypatch.setattr(lifecycle, "_kill_compiled_go_services", lambda: [])
    monkeypatch.setattr(lifecycle, "_wait_for_port_closed", lambda *args, **kwargs: True)
    monkeypatch.setattr(lifecycle, "_port_open", lambda *args, **kwargs: False)
    monkeypatch.setattr(lifecycle, "is_windows", lambda: True)
    monkeypatch.setattr(lifecycle, "_run_best_effort", lambda cmd, timeout=10: commands.append(cmd))
    monkeypatch.setattr(lifecycle.time, "sleep", lambda _n: None)

    lifecycle.down.body(Context())

    assert any(cmd and cmd[:3] == ["powershell", "-NoProfile", "-Command"] for cmd in commands)
    assert any("Get-Process server" in " ".join(cmd) for cmd in commands)
    assert ["bridges"] in commands


def test_down_fails_when_managed_ports_remain(monkeypatch):
    monkeypatch.setattr(lifecycle, "_kill_port", lambda port, label: True)
    monkeypatch.setattr(lifecycle, "_kill_bridges", lambda: None)
    monkeypatch.setattr(lifecycle, "_kill_compiled_go_services", lambda: [])
    monkeypatch.setattr(lifecycle, "_run_best_effort", lambda cmd, timeout=10: None)
    monkeypatch.setattr(lifecycle, "is_windows", lambda: False)
    monkeypatch.setattr(lifecycle.time, "sleep", lambda _n: None)
    monkeypatch.setattr(
        lifecycle,
        "_port_open",
        lambda port, host="127.0.0.1", timeout=1.0: port == lifecycle.INTERFACE_PORT,
    )

    with pytest.raises(SystemExit, match="STACK DOWN INCOMPLETE"):
        lifecycle.down.body(Context())


def test_matches_compiled_go_service_recognizes_known_binaries_and_go_run():
    assert lifecycle._matches_compiled_go_service("server.exe", "")
    assert lifecycle._matches_compiled_go_service("go.exe", "go run ./cmd/server")
    assert lifecycle._matches_compiled_go_service("go", "go run ./cmd/signal_gen")
    assert not lifecycle._matches_compiled_go_service("python.exe", "python -m pytest")


def test_down_kills_detected_compiled_go_services(monkeypatch):
    killed: list[int] = []
    scans = iter(
        [
            [{"pid": 111, "name": "server.exe", "command": "core\\bin\\server.exe"}],
            [],
        ]
    )

    monkeypatch.setattr(lifecycle, "_kill_port", lambda port, label: False)
    monkeypatch.setattr(lifecycle, "_kill_bridges", lambda: None)
    monkeypatch.setattr(lifecycle, "_wait_for_port_closed", lambda *args, **kwargs: True)
    monkeypatch.setattr(lifecycle, "_port_open", lambda *args, **kwargs: False)
    monkeypatch.setattr(lifecycle, "is_windows", lambda: True)
    monkeypatch.setattr(lifecycle, "_run_best_effort", lambda cmd, timeout=10: None)
    monkeypatch.setattr(lifecycle, "_kill_pid", lambda pid: killed.append(pid))
    monkeypatch.setattr(lifecycle, "_list_compiled_go_service_processes", lambda: next(scans))
    monkeypatch.setattr(lifecycle.time, "sleep", lambda _n: None)

    lifecycle.down.body(Context())

    assert killed == [111]


def test_wait_for_http_ok_returns_true_on_200(monkeypatch):
    calls: list[str] = []

    def fake_http_get(url: str, timeout: float = 5.0, headers=None):
        calls.append(url)
        return 200, "ok"

    monkeypatch.setattr(lifecycle, "_http_get", fake_http_get)

    assert lifecycle._wait_for_http_ok("http://localhost:8081/healthz", "Core health", timeout=1, interval=0.01)
    assert calls == ["http://localhost:8081/healthz"]


def test_core_startup_log_path_points_to_workspace_logs():
    path = lifecycle._core_startup_log_path()

    assert path.name == "core-startup.log"
    assert "workspace" in str(path)
    assert "logs" in str(path)


def test_up_restarts_running_core_when_dependencies_were_down_before_up(monkeypatch):
    events: list[str] = []

    monkeypatch.setattr(Path, "exists", lambda self: True)
    monkeypatch.setattr(lifecycle, "_ensure_bridge", lambda: events.append("bridge"))
    monkeypatch.setattr(
        lifecycle,
        "_wait_for_port",
        lambda port, label, timeout=30, interval=1.0: events.append(f"wait:{port}:{label}") or True,
    )
    monkeypatch.setattr(lifecycle, "_wait_for_http_ok", lambda *args, **kwargs: True)
    monkeypatch.setattr(
        lifecycle,
        "_kill_port",
        lambda port, label: events.append(f"kill:{port}:{label}") or True,
    )
    monkeypatch.setattr(
        lifecycle,
        "_start_core_background",
        lambda: events.append("start_core") or True,
    )

    def fake_port_open(port: int, host: str = "127.0.0.1", timeout: float = 1.0) -> bool:
        if port == 5432:
            return False
        if port == 4222:
            return True
        if port == lifecycle.API_PORT:
            return True
        return False

    monkeypatch.setattr(lifecycle, "_port_open", fake_port_open)

    lifecycle.up.body(Context(), frontend=False, build=False)

    assert f"kill:{lifecycle.API_PORT}:Core" in events
    assert "start_core" in events
    assert any(item == f"wait:{lifecycle.API_PORT}:Core API" for item in events)


def test_up_keeps_running_core_when_dependencies_were_healthy_before_up(monkeypatch):
    events: list[str] = []

    monkeypatch.setattr(Path, "exists", lambda self: True)
    monkeypatch.setattr(lifecycle, "_ensure_bridge", lambda: events.append("bridge"))
    monkeypatch.setattr(
        lifecycle,
        "_wait_for_port",
        lambda port, label, timeout=30, interval=1.0: events.append(f"wait:{port}:{label}") or True,
    )
    monkeypatch.setattr(lifecycle, "_wait_for_http_ok", lambda *args, **kwargs: True)
    monkeypatch.setattr(
        lifecycle,
        "_kill_port",
        lambda port, label: events.append(f"kill:{port}:{label}") or True,
    )
    monkeypatch.setattr(
        lifecycle,
        "_start_core_background",
        lambda: events.append("start_core") or True,
    )

    def fake_port_open(port: int, host: str = "127.0.0.1", timeout: float = 1.0) -> bool:
        if port in (5432, 4222, lifecycle.API_PORT):
            return True
        return False

    monkeypatch.setattr(lifecycle, "_port_open", fake_port_open)

    lifecycle.up.body(Context(), frontend=False, build=False)

    assert f"kill:{lifecycle.API_PORT}:Core" not in events
    assert "start_core" not in events


def test_up_fails_when_core_health_never_becomes_ready(monkeypatch):
    monkeypatch.setattr(Path, "exists", lambda self: True)
    monkeypatch.setattr(lifecycle, "_ensure_bridge", lambda: None)
    monkeypatch.setattr(lifecycle, "_wait_for_port", lambda *args, **kwargs: True)
    monkeypatch.setattr(lifecycle, "_start_core_background", lambda: True)
    monkeypatch.setattr(lifecycle, "_port_open", lambda port, host="127.0.0.1", timeout=1.0: False if port == lifecycle.API_PORT else True)
    monkeypatch.setattr(lifecycle, "_wait_for_http_ok", lambda *args, **kwargs: False)

    with pytest.raises(SystemExit, match="core-startup.log"):
        lifecycle.up.body(Context(), frontend=False, build=False)
