from __future__ import annotations

from contextlib import contextmanager
from dataclasses import dataclass

from invoke import Context

from ops import interface


@dataclass
class FakeResult:
    exited: int = 0
    stdout: str = ""
    stderr: str = ""


class FakeContext(Context):
    def __init__(self, command_results: dict[str, FakeResult] | None = None):
        super().__init__()
        self.command_results = command_results or {}
        self.commands: list[str] = []
        self.cd_paths: list[str] = []

    def run(self, command: str, **_kwargs) -> FakeResult:
        self.commands.append(command)
        return self.command_results.get(command, FakeResult())

    @contextmanager
    def cd(self, path: str):
        self.cd_paths.append(path)
        yield


def test_matches_repo_local_interface_process_accepts_repo_postcss_worker():
    command = r'"node" D:\MakeIntellex\Projects\mycelisai\scratch\interface\.next\dev\build\postcss.js 52847'

    assert interface._matches_repo_local_interface_process("node.exe", command)


def test_matches_repo_local_interface_process_rejects_unrelated_node_work():
    command = r'"C:\Program Files\nodejs\node.exe" C:\other-app\node_modules\vite\bin\vite.js'

    assert not interface._matches_repo_local_interface_process("node.exe", command)


def test_cleanup_repo_local_interface_processes_kills_detected_workers(monkeypatch):
    killed: list[int] = []
    scans = iter(
        [
            [{"pid": 101, "name": "node.exe", "command": "interface\\.next\\dev\\build\\postcss.js"}],
            [],
        ]
    )

    monkeypatch.setattr(interface, "_list_repo_local_interface_processes", lambda: next(scans))
    monkeypatch.setattr(interface, "_kill_pid_tree", lambda pid: killed.append(pid))
    monkeypatch.setattr(interface.time, "sleep", lambda _n: None)

    remaining = interface._cleanup_repo_local_interface_processes()

    assert killed == [101]
    assert remaining == []


def test_build_cleans_residual_interface_workers(monkeypatch):
    cleaned: list[str] = []
    stopped: list[str] = []
    shell_calls: list[list[str]] = []
    ctx = FakeContext()

    monkeypatch.setattr(interface, "stop", lambda _c, port=interface.INTERFACE_PORT: stopped.append(f"stop:{port}"))
    monkeypatch.setattr(interface, "clean", lambda _c: cleaned.append("clean") or None)
    monkeypatch.setattr(interface, "_cleanup_repo_local_interface_processes", lambda: cleaned.append("cleanup") or [])
    monkeypatch.setattr(
        interface,
        "_run_interface_shell_command",
        lambda command, extra_env=None: shell_calls.append(command) or interface.CommandResult(exited=0, stdout="", stderr=""),
    )

    interface.build.body(ctx)

    assert stopped == [f"stop:{interface.INTERFACE_PORT}"]
    assert cleaned == ["clean", "cleanup", "cleanup"]
    assert shell_calls == [["npm", "run", "build"]]


def test_stop_runs_tree_kill_and_repo_cleanup_on_windows(monkeypatch):
    cleaned: list[str] = []
    ctx = FakeContext()
    port = 4310

    monkeypatch.setattr(interface, "is_windows", lambda: True)
    monkeypatch.setattr(interface, "_cleanup_repo_local_interface_processes", lambda: cleaned.append("cleanup") or [])

    interface.stop.body(ctx, port=port)

    assert any("taskkill /F /T /PID" in command for command in ctx.commands)
    assert cleaned == ["cleanup"]


def test_e2e_starts_managed_server_and_skips_playwright_webserver(monkeypatch):
    ctx = FakeContext({"npx playwright test --reporter=dot --project=chromium e2e/specs/navigation.spec.ts --workers=1": FakeResult()})
    events: list[str] = []
    env_seen: dict[str, str] = {}
    port = 4311

    class FakeServer:
        pid = 4242

        @staticmethod
        def poll():
            return None

    monkeypatch.setattr(
        interface,
        "_task_env",
        lambda extra=None: {
            "PLAYWRIGHT_SKIP_WEBSERVER": extra["PLAYWRIGHT_SKIP_WEBSERVER"],
            "INTERFACE_HOST": extra["INTERFACE_HOST"],
            "INTERFACE_BIND_HOST": extra["INTERFACE_BIND_HOST"],
        },
    )
    monkeypatch.setattr(interface, "INTERFACE_PORT", port)
    monkeypatch.setattr(interface, "stop", lambda _c, port=interface.INTERFACE_PORT: events.append(f"stop:{port}"))
    monkeypatch.setattr(
        interface,
        "_wait_for_interface_ready",
        lambda host="127.0.0.1", port=interface.INTERFACE_PORT, timeout_seconds=120: events.append(f"ready:{host}:{port}") or "127.0.0.1",
    )
    monkeypatch.setattr(interface, "_detect_playwright_server_port", lambda expected_port, timeout_seconds=30: expected_port)
    monkeypatch.setattr(interface, "_kill_pid_tree", lambda pid: events.append(f"kill:{pid}"))
    monkeypatch.setattr(interface, "_cleanup_repo_local_interface_processes", lambda: events.append("cleanup") or [])
    monkeypatch.setattr(interface.time, "sleep", lambda _n: None)

    def fake_start(env, port=interface.INTERFACE_PORT, server_mode="start"):
        env_seen.update(env)
        events.append(f"start:{port}:{server_mode}")
        return FakeServer()

    monkeypatch.setattr(interface, "_start_playwright_server", fake_start)

    interface.e2e.body(ctx, project="chromium", spec="e2e/specs/navigation.spec.ts")

    assert env_seen["PLAYWRIGHT_SKIP_WEBSERVER"] == "1"
    assert ctx.commands == ["npx playwright test --reporter=dot --project=chromium e2e/specs/navigation.spec.ts --workers=1"]
    assert ctx.cd_paths == [str(interface.INTERFACE_DIR)]
    assert env_seen["INTERFACE_HOST"] == "127.0.0.1"
    assert env_seen["INTERFACE_BIND_HOST"] == interface.INTERFACE_BIND_HOST
    assert events == [
        f"stop:{port}",
        f"start:{port}:start",
        f"ready:127.0.0.1:{port}",
        "kill:4242",
        f"stop:{port}",
        "cleanup",
    ]


def test_test_uses_direct_shell_command(monkeypatch):
    shell_calls: list[list[str]] = []
    ctx = FakeContext()

    monkeypatch.setattr(
        interface,
        "_run_interface_shell_command",
        lambda command, extra_env=None: shell_calls.append(command) or interface.CommandResult(exited=0, stdout="", stderr=""),
    )

    interface.test.body(ctx)

    assert shell_calls == [["npm", "run", "test"]]


def test_test_coverage_uses_direct_shell_command(monkeypatch):
    shell_calls: list[list[str]] = []
    ctx = FakeContext()

    monkeypatch.setattr(
        interface,
        "_run_interface_shell_command",
        lambda command, extra_env=None: shell_calls.append(command) or interface.CommandResult(exited=0, stdout="", stderr=""),
    )

    interface.test_coverage.body(ctx)

    assert shell_calls == [["npx", "vitest", "run", "--coverage"]]


def test_interface_ready_urls_prioritize_requested_host_then_loopback_fallbacks():
    urls = interface._interface_ready_urls("127.0.0.1", 3000)

    assert urls == [
        "http://127.0.0.1:3000",
        "http://localhost:3000",
        "http://[::1]:3000",
    ]


def test_pick_interface_port_falls_back_when_preferred_port_is_busy(monkeypatch):
    class FakeSocket:
        def __init__(self, family, *args, **kwargs):
            self.family = family
            self.bind_calls: list[tuple[str, int]] = []
            self.closed = False

        def setsockopt(self, *args, **kwargs):
            return None

        def bind(self, address):
            self.bind_calls.append(address)
            if address[1] == 3000 and self.family == interface.socket.AF_INET6:
                raise OSError("address in use")
            self.port = 43210 if address[1] == 3000 else 43211

        def getsockname(self):
            return ("127.0.0.1", getattr(self, "port", 43210))

        def close(self):
            self.closed = True

        def __enter__(self):
            return self

        def __exit__(self, exc_type, exc, tb):
            self.close()

    monkeypatch.setattr(interface.socket, "socket", lambda family, *args, **kwargs: FakeSocket(family))

    assert interface._pick_interface_port(3000) == 3100
