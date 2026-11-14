from __future__ import annotations

from invoke import Context
import pytest
from pathlib import Path

from ops import db as db_tasks
from ops import lifecycle


def test_status_reports_unknown_when_compiled_go_inspection_fails(monkeypatch, capsys):
    class Result:
        ok = True
        exited = 0
        stdout = "mycelis-cluster\n"

    monkeypatch.setattr(lifecycle, "_port_open", lambda *args, **kwargs: False)
    monkeypatch.setattr(lifecycle, "_list_compiled_go_service_processes", lambda: (_ for _ in ()).throw(RuntimeError("process query failed")))

    class DummyContext:
        def run(self, command, hide=True, warn=True):
            return Result()

    lifecycle.status.body(DummyContext())

    output = capsys.readouterr().out
    assert "Compiled Go svc : UNKNOWN" in output


def test_status_reports_docker_down_when_docker_version_fails(monkeypatch, capsys):
    class Result:
        def __init__(self, stdout="", exited=0, ok=True):
            self.stdout = stdout
            self.exited = exited
            self.ok = ok

    monkeypatch.setattr(lifecycle, "_port_open", lambda *args, **kwargs: False)
    monkeypatch.setattr(lifecycle, "_list_compiled_go_service_processes", lambda: [])

    class DummyContext:
        def run(self, command, hide=True, warn=True):
            if command.startswith("docker version"):
                return Result(exited=1, ok=False)
            if command == "kind get clusters":
                return Result(stdout="")
            raise AssertionError(f"unexpected command: {command}")

    lifecycle.status.body(DummyContext())

    output = capsys.readouterr().out
    assert "Docker          : DOWN" in output


def test_down_fails_when_repo_local_interface_residuals_remain(monkeypatch):
    from ops import interface as interface_tasks

    monkeypatch.setattr(lifecycle, "_kill_port", lambda port, label: False)
    monkeypatch.setattr(lifecycle, "_kill_bridges", lambda: None)
    monkeypatch.setattr(lifecycle, "_kill_compiled_go_services", lambda: [])
    monkeypatch.setattr(lifecycle, "_wait_for_port_closed", lambda *args, **kwargs: True)
    monkeypatch.setattr(lifecycle, "_port_open", lambda *args, **kwargs: False)
    monkeypatch.setattr(lifecycle, "is_windows", lambda: True)
    monkeypatch.setattr(lifecycle, "_run_best_effort", lambda cmd, timeout=10: None)
    monkeypatch.setattr(
        interface_tasks,
        "_cleanup_repo_local_interface_processes",
        lambda: [{"pid": 222, "name": "node.exe", "command": "interface\\.next\\dev\\build\\postcss.js"}],
    )
    monkeypatch.setattr(lifecycle.time, "sleep", lambda _n: None)

    with pytest.raises(SystemExit, match="Interface residuals still running"):
        lifecycle.down.body(Context())


def test_wait_for_http_ok_returns_true_on_200(monkeypatch):
    calls: list[str] = []

    def fake_http_get(url: str, timeout: float = 5.0, headers=None):
        calls.append(url)
        return 200, "ok"

    monkeypatch.setattr(lifecycle, "_http_get", fake_http_get)

    assert lifecycle._wait_for_http_ok("http://localhost:8081/healthz", "Core health", timeout=1, interval=0.01)
    assert calls == ["http://localhost:8081/healthz"]


def test_interface_probe_urls_prioritize_probe_host_then_loopback_fallbacks():
    urls = lifecycle._interface_probe_urls("127.0.0.1", 3000)

    assert urls == ["http://127.0.0.1:3000/", "http://localhost:3000/", "http://[::1]:3000/"]


def test_core_startup_log_path_points_to_workspace_logs():
    path = lifecycle._core_startup_log_path()

    assert path.name == "core-startup.log"
    assert "workspace" in str(path)
    assert "logs" in str(path)


def test_up_restarts_running_core_when_dependencies_were_down_before_up(monkeypatch):
    events: list[str] = []
    db_calls: list[str] = []

    monkeypatch.setattr(Path, "exists", lambda self: True)
    monkeypatch.setattr(lifecycle, "_ensure_bridge", lambda: events.append("bridge"))
    monkeypatch.setattr(
        lifecycle,
        "_wait_for_port",
        lambda port, label, timeout=30, interval=1.0: events.append(f"wait:{port}:{label}") or True,
    )
    monkeypatch.setattr(lifecycle, "_wait_for_http_ok", lambda *args, **kwargs: True)
    monkeypatch.setattr(lifecycle, "_core_council_ready", lambda timeout=10, interval=1.0: True)
    monkeypatch.setattr(db_tasks.create, "body", lambda _c: db_calls.append("db.create"))
    monkeypatch.setattr(lifecycle, "_kill_port", lambda port, label: events.append(f"kill:{port}:{label}") or True)
    monkeypatch.setattr(lifecycle, "_start_core_background", lambda: events.append("start_core") or True)

    def fake_port_open(port: int, host: str = "127.0.0.1", timeout: float = 1.0) -> bool:
        return port in (4222, lifecycle.API_PORT)

    monkeypatch.setattr(lifecycle, "_port_open", fake_port_open)

    lifecycle.up.body(Context(), frontend=False, build=False)

    assert db_calls == ["db.create"]
    assert f"kill:{lifecycle.API_PORT}:Core" in events
    assert "start_core" in events
    assert any(item == f"wait:{lifecycle.API_PORT}:Core API" for item in events)


def test_up_restarts_running_core_when_council_routes_are_still_offline(monkeypatch):
    events: list[str] = []
    db_calls: list[str] = []

    monkeypatch.setattr(Path, "exists", lambda self: True)
    monkeypatch.setattr(lifecycle, "_ensure_bridge", lambda: events.append("bridge"))
    monkeypatch.setattr(
        lifecycle,
        "_wait_for_port",
        lambda port, label, timeout=30, interval=1.0: events.append(f"wait:{port}:{label}") or True,
    )
    monkeypatch.setattr(lifecycle, "_wait_for_http_ok", lambda *args, **kwargs: True)
    council_states = iter([False, True])
    monkeypatch.setattr(lifecycle, "_core_council_ready", lambda timeout=10, interval=1.0: next(council_states))
    monkeypatch.setattr(lifecycle, "_kill_port", lambda port, label: events.append(f"kill:{port}:{label}") or True)
    monkeypatch.setattr(lifecycle, "_start_core_background", lambda: events.append("start_core") or True)
    monkeypatch.setattr(db_tasks.create, "body", lambda _c: db_calls.append("db.create"))
    monkeypatch.setattr(
        lifecycle,
        "_port_open",
        lambda port, host="127.0.0.1", timeout=1.0: port in (5432, 4222, lifecycle.API_PORT),
    )

    lifecycle.up.body(Context(), frontend=False, build=False)

    assert db_calls == ["db.create"]
    assert f"kill:{lifecycle.API_PORT}:Core" in events
    assert "start_core" in events


def test_up_keeps_running_core_when_dependencies_were_healthy_before_up(monkeypatch):
    events: list[str] = []
    db_calls: list[str] = []

    monkeypatch.setattr(Path, "exists", lambda self: True)
    monkeypatch.setattr(lifecycle, "_ensure_bridge", lambda: events.append("bridge"))
    monkeypatch.setattr(
        lifecycle,
        "_wait_for_port",
        lambda port, label, timeout=30, interval=1.0: events.append(f"wait:{port}:{label}") or True,
    )
    monkeypatch.setattr(lifecycle, "_wait_for_http_ok", lambda *args, **kwargs: True)
    monkeypatch.setattr(lifecycle, "_core_council_ready", lambda timeout=10, interval=1.0: True)
    monkeypatch.setattr(lifecycle, "_kill_port", lambda port, label: events.append(f"kill:{port}:{label}") or True)
    monkeypatch.setattr(lifecycle, "_start_core_background", lambda: events.append("start_core") or True)
    monkeypatch.setattr(db_tasks.create, "body", lambda _c: db_calls.append("db.create"))
    monkeypatch.setattr(lifecycle, "_port_open", lambda port, host="127.0.0.1", timeout=1.0: port in (5432, 4222, lifecycle.API_PORT))

    lifecycle.up.body(Context(), frontend=False, build=False)

    assert db_calls == ["db.create"]
    assert f"kill:{lifecycle.API_PORT}:Core" not in events
    assert "start_core" not in events


def test_up_fails_when_core_health_recovers_but_council_routes_stay_offline(monkeypatch):
    monkeypatch.setattr(Path, "exists", lambda self: True)
    monkeypatch.setattr(lifecycle, "_ensure_bridge", lambda: None)
    monkeypatch.setattr(lifecycle, "_wait_for_port", lambda *args, **kwargs: True)
    monkeypatch.setattr(db_tasks.create, "body", lambda _c: None)
    monkeypatch.setattr(lifecycle, "_start_core_background", lambda: True)
    monkeypatch.setattr(lifecycle, "_port_open", lambda port, host="127.0.0.1", timeout=1.0: False if port == lifecycle.API_PORT else True)
    monkeypatch.setattr(lifecycle, "_wait_for_http_ok", lambda *args, **kwargs: True)
    monkeypatch.setattr(lifecycle, "_core_council_ready", lambda timeout=10, interval=1.0: False)

    with pytest.raises(SystemExit, match="council routes are still offline"):
        lifecycle.up.body(Context(), frontend=False, build=False)


def test_up_fails_when_core_health_never_becomes_ready(monkeypatch):
    monkeypatch.setattr(Path, "exists", lambda self: True)
    monkeypatch.setattr(lifecycle, "_ensure_bridge", lambda: None)
    monkeypatch.setattr(lifecycle, "_wait_for_port", lambda *args, **kwargs: True)
    monkeypatch.setattr(db_tasks.create, "body", lambda _c: None)
    monkeypatch.setattr(lifecycle, "_start_core_background", lambda: True)
    monkeypatch.setattr(lifecycle, "_port_open", lambda port, host="127.0.0.1", timeout=1.0: False if port == lifecycle.API_PORT else True)
    monkeypatch.setattr(lifecycle, "_wait_for_http_ok", lambda *args, **kwargs: False)

    with pytest.raises(SystemExit, match="core-startup.log"):
        lifecycle.up.body(Context(), frontend=False, build=False)


def test_core_council_ready_requires_success_response(monkeypatch):
    calls: list[tuple[str, dict[str, str] | None]] = []
    monkeypatch.setenv("MYCELIS_API_KEY", "test-key")
    monkeypatch.setattr(
        lifecycle,
        "_http_get",
        lambda url, timeout=5.0, headers=None: calls.append((url, headers)) or (200, '{"ok":true}'),
    )

    assert lifecycle._core_council_ready(timeout=1, interval=0.01)
    assert calls == [(f"http://{lifecycle.API_HOST}:{lifecycle.API_PORT}/api/v1/council/members", {"Authorization": "Bearer test-key"})]


def test_up_frontend_uses_shared_interface_launcher(monkeypatch):
    events: list[str] = []

    monkeypatch.setattr(Path, "exists", lambda self: True)
    monkeypatch.setattr(lifecycle, "_ensure_bridge", lambda: None)
    monkeypatch.setattr(lifecycle, "_wait_for_http_ok", lambda *args, **kwargs: True)
    monkeypatch.setattr(lifecycle, "_core_council_ready", lambda timeout=10, interval=1.0: True)
    monkeypatch.setattr(db_tasks.create, "body", lambda _c: None)
    monkeypatch.setattr(lifecycle, "_start_core_background", lambda: True)
    monkeypatch.setattr(lifecycle, "_wait_for_port", lambda port, label, timeout=30, interval=1.0: events.append(f"wait:{port}:{label}") or True)
    monkeypatch.setattr(lifecycle, "_port_open", lambda port, host="127.0.0.1", timeout=1.0: port in (5432, 4222))

    from ops import interface as interface_tasks

    monkeypatch.setattr(interface_tasks, "interface_task_env", lambda extra=None: {"TEST_ENV": "1"})
    monkeypatch.setattr(
        interface_tasks,
        "start_dev_server_detached",
        lambda env=None, host=lifecycle.INTERFACE_BIND_HOST, port=lifecycle.INTERFACE_PORT: events.append(
            f"frontend:{host}:{port}:{env.get('TEST_ENV') if env else ''}"
        ),
    )

    lifecycle.up.body(Context(), frontend=True, build=False)

    assert f"frontend:{lifecycle.INTERFACE_BIND_HOST}:{lifecycle.INTERFACE_PORT}:1" in events
    assert f"wait:{lifecycle.INTERFACE_PORT}:Frontend" in events


def test_up_with_build_uses_core_compile_task_body(monkeypatch):
    compile_calls: list[str] = []

    monkeypatch.setattr(lifecycle, "_ensure_bridge", lambda: None)
    monkeypatch.setattr(lifecycle, "_wait_for_port", lambda *args, **kwargs: True)
    monkeypatch.setattr(lifecycle, "_wait_for_http_ok", lambda *args, **kwargs: True)
    monkeypatch.setattr(lifecycle, "_core_council_ready", lambda timeout=10, interval=1.0: True)
    monkeypatch.setattr(lifecycle, "_port_open", lambda port, host="127.0.0.1", timeout=1.0: False if port == lifecycle.API_PORT else True)
    monkeypatch.setattr(db_tasks.create, "body", lambda _c: None)
    monkeypatch.setattr(lifecycle, "_start_core_background", lambda: True)

    from ops import core as core_tasks

    monkeypatch.setattr(core_tasks.compile, "body", lambda _c: compile_calls.append("compile"))

    lifecycle.up.body(Context(), frontend=False, build=True)

    assert compile_calls == ["compile"]
