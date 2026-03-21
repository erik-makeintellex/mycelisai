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

    def run(self, command: str, **_kwargs) -> FakeResult:
        self.commands.append(command)
        return self.command_results.get(command, FakeResult())

    @contextmanager
    def cd(self, _path: str):
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
    ctx = FakeContext({"cd interface && npm run build": FakeResult()})

    monkeypatch.setattr(interface, "_cleanup_repo_local_interface_processes", lambda: cleaned.append("cleanup") or [])

    interface.build.body(ctx)

    assert ctx.commands == ["cd interface && npm run build"]
    assert cleaned == ["cleanup"]


def test_stop_runs_tree_kill_and_repo_cleanup_on_windows(monkeypatch):
    cleaned: list[str] = []
    ctx = FakeContext()

    monkeypatch.setattr(interface, "is_windows", lambda: True)
    monkeypatch.setattr(interface, "_cleanup_repo_local_interface_processes", lambda: cleaned.append("cleanup") or [])

    interface.stop.body(ctx, port=3000)

    assert any("taskkill /F /T /PID" in command for command in ctx.commands)
    assert cleaned == ["cleanup"]


def test_e2e_starts_managed_server_and_skips_playwright_webserver(monkeypatch):
    ctx = FakeContext({"cd interface && npx playwright test --reporter=dot --project=chromium e2e/specs/navigation.spec.ts": FakeResult()})
    events: list[str] = []
    env_seen: dict[str, str] = {}

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
        },
    )
    monkeypatch.setattr(interface, "stop", lambda _c, port=3000: events.append(f"stop:{port}"))
    monkeypatch.setattr(interface, "_wait_for_interface_ready", lambda host="127.0.0.1", port=3000, timeout_seconds=120: events.append(f"ready:{host}:{port}"))
    monkeypatch.setattr(interface, "_kill_pid_tree", lambda pid: events.append(f"kill:{pid}"))
    monkeypatch.setattr(interface, "_cleanup_repo_local_interface_processes", lambda: events.append("cleanup") or [])
    monkeypatch.setattr(interface.time, "sleep", lambda _n: None)

    def fake_start(env, port=3000):
        env_seen.update(env)
        events.append(f"start:{port}")
        return FakeServer()

    monkeypatch.setattr(interface, "_start_playwright_server", fake_start)

    interface.e2e.body(ctx, project="chromium", spec="e2e/specs/navigation.spec.ts")

    assert env_seen["PLAYWRIGHT_SKIP_WEBSERVER"] == "1"
    assert ctx.commands == ["cd interface && npx playwright test --reporter=dot --project=chromium e2e/specs/navigation.spec.ts"]
    assert env_seen["INTERFACE_HOST"] == "127.0.0.1"
    assert events == ["stop:3000", "start:3000", "ready:127.0.0.1:3000", "kill:4242", "stop:3000", "cleanup"]
