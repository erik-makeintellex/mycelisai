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


def test_start_port_forward_uses_direct_detached_kubectl_on_windows(monkeypatch):
    captured: dict[str, object] = {}

    def fake_popen(command, **kwargs):
        captured["command"] = command
        captured["kwargs"] = kwargs

        class DummyProcess:
            pass

        return DummyProcess()

    monkeypatch.setattr(lifecycle, "is_windows", lambda: True)
    monkeypatch.setattr(lifecycle.subprocess, "Popen", fake_popen)

    lifecycle._start_port_forward("svc/mycelis-core-nats", "4222:4222")

    assert captured["command"] == ["kubectl", "port-forward", "-n", lifecycle.NAMESPACE, "svc/mycelis-core-nats", "4222:4222"]
    assert captured["kwargs"]["stdout"] is lifecycle.subprocess.DEVNULL
    assert captured["kwargs"]["stderr"] is lifecycle.subprocess.DEVNULL
    assert captured["kwargs"]["creationflags"] == lifecycle.subprocess.CREATE_NEW_PROCESS_GROUP | lifecycle.subprocess.DETACHED_PROCESS


def test_health_raises_when_any_probe_fails(monkeypatch):
    monkeypatch.setattr(lifecycle, "_load_env", lambda: None)
    monkeypatch.setattr(lifecycle, "_port_open", lambda port, host="127.0.0.1", timeout=1.0: port == 11434)

    def fake_http_get(url: str, timeout: float = 3.0, headers: dict[str, str] | None = None):
        if "11434" in url:
            return 200, "ok"
        return 503, "down"

    monkeypatch.setattr(lifecycle, "_http_get", fake_http_get)

    with pytest.raises(SystemExit):
        lifecycle.health.body(Context())


def test_health_passes_when_all_probes_are_healthy(monkeypatch):
    monkeypatch.setattr(lifecycle, "_load_env", lambda: None)
    monkeypatch.setattr(lifecycle, "_port_open", lambda *args, **kwargs: True)
    monkeypatch.setattr(lifecycle, "_http_get", lambda url, timeout=3.0, headers=None: (200, "ok"))

    lifecycle.health.body(Context())


def test_down_uses_best_effort_cleanup_without_hanging(monkeypatch):
    commands: list[list[str]] = []
    port = 4312

    from ops import interface as interface_tasks

    monkeypatch.setattr(lifecycle, "_kill_port", lambda port, label: False)
    monkeypatch.setattr(lifecycle, "_kill_bridges", lambda: commands.append(["bridges"]))
    monkeypatch.setattr(lifecycle, "_kill_compiled_go_services", lambda: [])
    monkeypatch.setattr(lifecycle, "_wait_for_port_closed", lambda *args, **kwargs: True)
    monkeypatch.setattr(lifecycle, "_port_open", lambda *args, **kwargs: False)
    monkeypatch.setattr(lifecycle, "is_windows", lambda: True)
    monkeypatch.setattr(lifecycle, "INTERFACE_PORT", port)
    monkeypatch.setattr(lifecycle, "_run_best_effort", lambda cmd, timeout=10: commands.append(cmd))
    monkeypatch.setattr(interface_tasks, "_cleanup_repo_local_interface_processes", lambda: [])
    monkeypatch.setattr(lifecycle.time, "sleep", lambda _n: None)

    lifecycle.down.body(Context())

    assert any(cmd and cmd[:3] == ["powershell", "-NoProfile", "-Command"] for cmd in commands)
    assert any("Get-Process server" in " ".join(cmd) for cmd in commands)
    assert any(f"next (dev|start).*--port {port}" in " ".join(cmd) for cmd in commands if isinstance(cmd, list))
    assert ["bridges"] in commands


def test_down_fails_when_managed_ports_remain(monkeypatch):
    from ops import interface as interface_tasks

    monkeypatch.setattr(lifecycle, "_kill_port", lambda port, label: True)
    monkeypatch.setattr(lifecycle, "_kill_bridges", lambda: None)
    monkeypatch.setattr(lifecycle, "_kill_compiled_go_services", lambda: [])
    monkeypatch.setattr(lifecycle, "_run_best_effort", lambda cmd, timeout=10: None)
    monkeypatch.setattr(lifecycle, "is_windows", lambda: False)
    monkeypatch.setattr(interface_tasks, "_cleanup_repo_local_interface_processes", lambda: [])
    monkeypatch.setattr(lifecycle.time, "sleep", lambda _n: None)
    monkeypatch.setattr(
        lifecycle,
        "_port_open",
        lambda port, host="127.0.0.1", timeout=1.0: port == lifecycle.INTERFACE_PORT,
    )

    with pytest.raises(SystemExit, match="STACK DOWN INCOMPLETE"):
        lifecycle.down.body(Context())


def test_matches_compiled_go_service_recognizes_known_binaries_and_go_run():
    assert lifecycle._matches_compiled_go_service("server.exe", str(Path("core/bin/server.exe")))
    assert lifecycle._matches_compiled_go_service("go.exe", "go run ./cmd/server")
    assert lifecycle._matches_compiled_go_service("go", "go run ./cmd/signal_gen")
    assert not lifecycle._matches_compiled_go_service("server.exe", "")
    assert not lifecycle._matches_compiled_go_service("python.exe", "python -m pytest")


def test_matches_compiled_go_binary_path_recognizes_repo_local_binaries():
    assert lifecycle._matches_compiled_go_binary_path("server.exe", str(Path("core/bin/server.exe").resolve()))
    assert lifecycle._matches_compiled_go_binary_path("probe.exe", "")
    assert not lifecycle._matches_compiled_go_binary_path("python.exe", str(Path("core/bin/server.exe").resolve()))


def test_down_kills_detected_compiled_go_services(monkeypatch):
    killed: list[int] = []
    scans = iter(
        [
            [{"pid": 111, "name": "server.exe", "command": "core\\bin\\server.exe"}],
            [],
        ]
    )

    from ops import interface as interface_tasks

    monkeypatch.setattr(lifecycle, "_kill_port", lambda port, label: False)
    monkeypatch.setattr(lifecycle, "_kill_bridges", lambda: None)
    monkeypatch.setattr(lifecycle, "_wait_for_port_closed", lambda *args, **kwargs: True)
    monkeypatch.setattr(lifecycle, "_port_open", lambda *args, **kwargs: False)
    monkeypatch.setattr(lifecycle, "is_windows", lambda: True)
    monkeypatch.setattr(lifecycle, "_run_best_effort", lambda cmd, timeout=10: None)
    monkeypatch.setattr(lifecycle, "_kill_pid", lambda pid: killed.append(pid))
    monkeypatch.setattr(lifecycle, "_list_compiled_go_service_processes", lambda: next(scans))
    monkeypatch.setattr(interface_tasks, "_cleanup_repo_local_interface_processes", lambda: [])
    monkeypatch.setattr(lifecycle.time, "sleep", lambda _n: None)

    lifecycle.down.body(Context())

    assert killed == [111]


def test_down_fails_when_compiled_go_inspection_fails(monkeypatch):
    from ops import interface as interface_tasks

    monkeypatch.setattr(lifecycle, "_kill_port", lambda port, label: False)
    monkeypatch.setattr(lifecycle, "_kill_bridges", lambda: None)
    monkeypatch.setattr(lifecycle, "_wait_for_port_closed", lambda *args, **kwargs: True)
    monkeypatch.setattr(lifecycle, "_port_open", lambda *args, **kwargs: False)
    monkeypatch.setattr(lifecycle, "is_windows", lambda: True)
    monkeypatch.setattr(lifecycle, "_run_best_effort", lambda cmd, timeout=10: None)
    monkeypatch.setattr(
        lifecycle,
        "_list_compiled_go_service_processes",
        lambda: (_ for _ in ()).throw(RuntimeError("process query failed")),
    )
    monkeypatch.setattr(interface_tasks, "_cleanup_repo_local_interface_processes", lambda: [])
    monkeypatch.setattr(lifecycle.time, "sleep", lambda _n: None)

    with pytest.raises(SystemExit, match="unable to inspect compiled Go services"):
        lifecycle.down.body(Context())


def test_list_compiled_go_service_processes_falls_back_when_cim_times_out(monkeypatch):
    calls: list[list[str]] = []

    def fake_run(cmd, capture_output=True, text=True, timeout=0):
        calls.append(cmd)

        class Result:
            def __init__(self, stdout="", returncode=0, stderr=""):
                self.stdout = stdout
                self.returncode = returncode
                self.stderr = stderr

        if cmd[:4] == ["tasklist", "/FO", "CSV", "/NH"]:
            return Result(stdout='"server.exe","111","Console","1","12,000 K"')
        raise subprocess.TimeoutExpired(cmd="Get-CimInstance", timeout=timeout)

    monkeypatch.setattr(lifecycle, "is_windows", lambda: True)
    monkeypatch.setattr(lifecycle.subprocess, "run", fake_run)

    processes = lifecycle._list_compiled_go_service_processes()

    assert processes == [
        {
            "pid": 111,
            "name": "server",
            "command": "server.exe",
        }
    ]
    assert calls[0][:4] == ["tasklist", "/FO", "CSV", "/NH"]
    assert "/FI" in calls[0]


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

    assert urls == [
        "http://127.0.0.1:3000/",
        "http://localhost:3000/",
        "http://[::1]:3000/",
    ]


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
    monkeypatch.setattr(db_tasks.create, "body", lambda _c: db_calls.append("db.create"))

    def fake_port_open(port: int, host: str = "127.0.0.1", timeout: float = 1.0) -> bool:
        if port in (5432, 4222, lifecycle.API_PORT):
            return True
        return False

    monkeypatch.setattr(lifecycle, "_port_open", fake_port_open)

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
    assert calls == [
        (
            f"http://{lifecycle.API_HOST}:{lifecycle.API_PORT}/api/v1/council/members",
            {"Authorization": "Bearer test-key"},
        )
    ]


def test_up_frontend_uses_shared_interface_launcher(monkeypatch):
    events: list[str] = []

    monkeypatch.setattr(Path, "exists", lambda self: True)
    monkeypatch.setattr(lifecycle, "_ensure_bridge", lambda: None)
    monkeypatch.setattr(lifecycle, "_wait_for_http_ok", lambda *args, **kwargs: True)
    monkeypatch.setattr(lifecycle, "_core_council_ready", lambda timeout=10, interval=1.0: True)
    monkeypatch.setattr(db_tasks.create, "body", lambda _c: None)
    monkeypatch.setattr(lifecycle, "_start_core_background", lambda: True)

    def fake_wait_for_port(port, label, timeout=30, interval=1.0):
        events.append(f"wait:{port}:{label}")
        return True

    def fake_port_open(port: int, host: str = "127.0.0.1", timeout: float = 1.0) -> bool:
        if port in (5432, 4222):
            return True
        if port in (lifecycle.API_PORT, lifecycle.INTERFACE_PORT):
            return False
        return False

    monkeypatch.setattr(lifecycle, "_wait_for_port", fake_wait_for_port)
    monkeypatch.setattr(lifecycle, "_port_open", fake_port_open)

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
