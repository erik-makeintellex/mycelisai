from ops import interface_runtime as interface
from tests.interface_task_support import FakeContext, FakeResult

def test_e2e_starts_managed_server_and_skips_playwright_webserver(monkeypatch):
    ctx = FakeContext()
    events: list[str] = []
    env_seen: dict[str, str] = {}
    shell_calls: list[list[str]] = []
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
    monkeypatch.setattr(interface, "build", lambda _c: events.append("build"))
    monkeypatch.setattr(
        interface,
        "_wait_for_interface_ready",
        lambda host="127.0.0.1", port=interface.INTERFACE_PORT, timeout_seconds=120, process=None: events.append(f"ready:{host}:{port}") or "127.0.0.1",
    )
    monkeypatch.setattr(interface, "_detect_playwright_server_port", lambda expected_port, timeout_seconds=30: expected_port)
    monkeypatch.setattr(interface, "_kill_pid_tree", lambda pid: events.append(f"kill:{pid}"))
    monkeypatch.setattr(interface, "_cleanup_repo_local_interface_processes", lambda: events.append("cleanup") or [])
    monkeypatch.setattr(interface, "_cleanup_managed_interface_listeners", lambda: events.append("sweep") or [])
    monkeypatch.setattr(interface, "_cleanup_playwright_server_log", lambda: events.append("cleanup-log"))
    monkeypatch.setattr(interface.time, "sleep", lambda _n: None)
    monkeypatch.setattr(interface, "_detect_playwright_server_port", lambda expected_port, timeout_seconds=30: expected_port)

    def fake_start(env, port=interface.INTERFACE_PORT, server_mode="start"):
        env_seen.update(env)
        events.append(f"start:{port}:{server_mode}")
        return FakeServer()

    monkeypatch.setattr(interface, "_start_playwright_server", fake_start)
    monkeypatch.setattr(
        interface,
        "_run_playwright_command_streaming",
        lambda command, extra_env=None: shell_calls.append(command) or interface.CommandResult(exited=0, stdout="", stderr=""),
    )

    interface.e2e.body(ctx, project="chromium", spec="e2e/specs/navigation.spec.ts")

    assert env_seen["PLAYWRIGHT_SKIP_WEBSERVER"] == "1"
    assert shell_calls == [["npx", "playwright", "test", "--reporter=dot", "--project=chromium", "e2e/specs/navigation.spec.ts", "--workers=1"]]
    assert env_seen["INTERFACE_HOST"] == "127.0.0.1"
    assert env_seen["INTERFACE_BIND_HOST"] == interface.INTERFACE_BIND_HOST
    assert events == [
        f"stop:{port}",
        "sweep",
        f"start:{port}:dev",
        f"ready:127.0.0.1:{port}",
        "kill:4242",
        f"stop:{port}",
        "cleanup",
        "sweep",
        "cleanup-log",
    ]


def test_e2e_updates_playwright_host_when_managed_server_is_only_ready_on_alt_host(monkeypatch, capsys):
    ctx = FakeContext()
    events: list[str] = []
    env_seen: dict[str, str] = {}
    port = 4312

    class FakeServer:
        pid = 5252

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
    monkeypatch.setattr(interface, "build", lambda _c: events.append("build"))
    monkeypatch.setattr(
        interface,
        "_wait_for_interface_ready",
        lambda host="127.0.0.1", port=interface.INTERFACE_PORT, timeout_seconds=120, process=None: events.append(f"ready:{host}:{port}") or "::1",
    )
    monkeypatch.setattr(interface, "_detect_playwright_server_port", lambda expected_port, timeout_seconds=30: expected_port)
    monkeypatch.setattr(interface, "_kill_pid_tree", lambda pid: events.append(f"kill:{pid}"))
    monkeypatch.setattr(interface, "_cleanup_repo_local_interface_processes", lambda: events.append("cleanup") or [])
    monkeypatch.setattr(interface, "_cleanup_playwright_server_log", lambda: events.append("cleanup-log"))
    monkeypatch.setattr(interface.time, "sleep", lambda _n: None)

    def fake_start(env, port=interface.INTERFACE_PORT, server_mode="start"):
        events.append(f"start:{port}:{server_mode}")
        return FakeServer()

    monkeypatch.setattr(interface, "_start_playwright_server", fake_start)
    monkeypatch.setattr(
        interface,
        "_run_playwright_command_streaming",
        lambda command, extra_env=None: env_seen.update(extra_env or {}) or interface.CommandResult(exited=0, stdout="", stderr=""),
    )

    interface.e2e.body(ctx, project="chromium", spec="e2e/specs/navigation.spec.ts", server_mode="start")

    captured = capsys.readouterr()
    assert "Managed server is reachable via ::1; updating Playwright host from 127.0.0.1" in captured.out
    assert env_seen["INTERFACE_HOST"] == "::1"
    assert events == [
        f"stop:{port}",
        "build",
        f"start:{port}:start",
        f"ready:127.0.0.1:{port}",
        "kill:5252",
        f"stop:{port}",
        "cleanup",
        "cleanup-log",
    ]


def test_start_playwright_server_sets_standalone_host_and_port(monkeypatch, tmp_path):
    captured: dict[str, object] = {}

    class FakeProcess:
        pass

    monkeypatch.setattr(interface, "_playwright_server_log_path", lambda: str(tmp_path / "playwright.log"))
    monkeypatch.setattr(
        interface,
        "_spawn_interface_process",
        lambda command, env, stdout, stderr, detached, text=True: captured.update(
            {
                "command": command,
                "env": dict(env),
                "detached": detached,
                "text": text,
            }
        ) or FakeProcess(),
    )

    interface._start_playwright_server({"INTERFACE_BIND_HOST": "::1"}, port=4314, server_mode="start")

    assert captured["command"] == ["node", str((interface.INTERFACE_DIR / "scripts" / "playwright-webserver.mjs").resolve())]
    assert captured["env"]["HOSTNAME"] == "::1"
    assert captured["env"]["PORT"] == "4314"
    assert captured["detached"] is False


def test_e2e_dev_mode_skips_build(monkeypatch):
    ctx = FakeContext()
    events: list[str] = []

    class FakeServer:
        pid = 6262

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
    monkeypatch.setattr(interface, "INTERFACE_PORT", 4313)
    monkeypatch.setattr(interface, "stop", lambda _c, port=interface.INTERFACE_PORT: events.append(f"stop:{port}"))
    monkeypatch.setattr(interface, "build", lambda _c: events.append("build"))
    monkeypatch.setattr(
        interface,
        "_wait_for_interface_ready",
        lambda host="127.0.0.1", port=interface.INTERFACE_PORT, timeout_seconds=120, process=None: events.append(f"ready:{host}:{port}") or "127.0.0.1",
    )
    monkeypatch.setattr(interface, "_detect_playwright_server_port", lambda expected_port, timeout_seconds=30: expected_port)
    monkeypatch.setattr(interface, "_kill_pid_tree", lambda pid: events.append(f"kill:{pid}"))
    monkeypatch.setattr(interface, "_cleanup_repo_local_interface_processes", lambda: events.append("cleanup") or [])
    monkeypatch.setattr(interface, "_cleanup_playwright_server_log", lambda: events.append("cleanup-log"))
    monkeypatch.setattr(interface.time, "sleep", lambda _n: None)
    monkeypatch.setattr(interface, "_start_playwright_server", lambda env, port=interface.INTERFACE_PORT, server_mode="start": events.append(f"start:{port}:{server_mode}") or FakeServer())
    monkeypatch.setattr(
        interface,
        "_run_playwright_command_streaming",
        lambda command, extra_env=None: interface.CommandResult(exited=0, stdout="", stderr=""),
    )

    interface.e2e.body(ctx, project="chromium", spec="e2e/specs/navigation.spec.ts", server_mode="dev")

    assert "build" not in events
    assert events == [
        "stop:4313",
        "start:4313:dev",
        "ready:127.0.0.1:4313",
        "kill:6262",
        "stop:4313",
        "cleanup",
        "cleanup-log",
    ]


def test_e2e_cleans_dynamic_managed_port_before_and_after_run(monkeypatch):
    ctx = FakeContext()
    events: list[str] = []

    class FakeServer:
        pid = 8181

        @staticmethod
        def poll():
            return None

    monkeypatch.setattr(interface, "INTERFACE_PORT", 3000)
    monkeypatch.setattr(interface, "_pick_interface_port", lambda preferred=interface.INTERFACE_PORT: 4316)
    monkeypatch.setattr(
        interface,
        "_task_env",
        lambda extra=None: {
            "PLAYWRIGHT_SKIP_WEBSERVER": extra["PLAYWRIGHT_SKIP_WEBSERVER"],
            "INTERFACE_HOST": extra["INTERFACE_HOST"],
            "INTERFACE_BIND_HOST": extra["INTERFACE_BIND_HOST"],
        },
    )
    monkeypatch.setattr(interface, "stop", lambda _c, port=interface.INTERFACE_PORT: events.append(f"stop:{port}"))
    monkeypatch.setattr(
        interface,
        "_wait_for_interface_ready",
        lambda host="127.0.0.1", port=interface.INTERFACE_PORT, timeout_seconds=120, process=None: "127.0.0.1",
    )
    monkeypatch.setattr(interface, "_detect_playwright_server_port", lambda expected_port, timeout_seconds=30: expected_port)
    monkeypatch.setattr(interface, "_kill_pid_tree", lambda pid: events.append(f"kill:{pid}"))
    monkeypatch.setattr(interface, "_cleanup_repo_local_interface_processes", lambda: events.append("cleanup") or [])
    monkeypatch.setattr(interface, "_cleanup_playwright_server_log", lambda: events.append("cleanup-log"))
    monkeypatch.setattr(interface.time, "sleep", lambda _n: None)
    monkeypatch.setattr(interface, "_start_playwright_server", lambda env, port=interface.INTERFACE_PORT, server_mode="start": FakeServer())
    monkeypatch.setattr(
        interface,
        "_run_playwright_command_streaming",
        lambda command, extra_env=None: interface.CommandResult(exited=0, stdout="", stderr=""),
    )

    interface.e2e.body(ctx, project="chromium", spec="e2e/specs/navigation.spec.ts", server_mode="dev")

    assert events == [
        "stop:3000",
        "stop:4316",
        "kill:8181",
        "stop:4316",
        "cleanup",
        "cleanup-log",
    ]


def test_e2e_keeps_playwright_server_log_on_failure(monkeypatch):
    ctx = FakeContext()
    events: list[str] = []

    class FakeServer:
        pid = 7171

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
    monkeypatch.setattr(interface, "INTERFACE_PORT", 4314)
    monkeypatch.setattr(interface, "stop", lambda _c, port=interface.INTERFACE_PORT: events.append(f"stop:{port}"))
    monkeypatch.setattr(
        interface,
        "_wait_for_interface_ready",
        lambda host="127.0.0.1", port=interface.INTERFACE_PORT, timeout_seconds=120, process=None: "127.0.0.1",
    )
    monkeypatch.setattr(interface, "_kill_pid_tree", lambda pid: events.append(f"kill:{pid}"))
    monkeypatch.setattr(interface, "_cleanup_repo_local_interface_processes", lambda: events.append("cleanup") or [])
    monkeypatch.setattr(interface, "_cleanup_playwright_server_log", lambda: events.append("cleanup-log"))
    monkeypatch.setattr(interface.time, "sleep", lambda _n: None)
    monkeypatch.setattr(interface, "_start_playwright_server", lambda env, port=interface.INTERFACE_PORT, server_mode="start": FakeServer())
    monkeypatch.setattr(interface, "_detect_playwright_server_port", lambda expected_port, timeout_seconds=30: expected_port)
    monkeypatch.setattr(
        interface,
        "_run_playwright_command_streaming",
        lambda command, extra_env=None: interface.CommandResult(exited=1, stdout="", stderr=""),
    )

    try:
        interface.e2e.body(ctx, project="chromium", spec="e2e/specs/navigation.spec.ts")
    except SystemExit as exc:
        assert exc.code == 1
    else:
        raise AssertionError("expected playwright failure")

    assert "cleanup-log" not in events


def test_stop_lingering_playwright_process_terminates_before_tree_kill(monkeypatch):
    events: list[str] = []

    class FakeProcess:
        pid = 9191
        _exited = False

        def terminate(self):
            events.append("terminate")

        def wait(self, timeout=0):
            events.append(f"wait:{timeout}")
            self._exited = True

        def poll(self):
            return 0 if self._exited else None

    monkeypatch.setattr(interface, "_kill_pid_tree", lambda pid: events.append(f"kill:{pid}"))

    interface._stop_lingering_playwright_process(FakeProcess())

    assert events == ["terminate", "wait:3"]


def test_e2e_keeps_playwright_server_log_on_managed_server_startup_failure(monkeypatch):
    ctx = FakeContext()
    events: list[str] = []

    monkeypatch.setattr(interface, "INTERFACE_PORT", 4317)
    monkeypatch.setattr(interface, "_pick_interface_port", lambda preferred=interface.INTERFACE_PORT: 4317)
    monkeypatch.setattr(interface, "stop", lambda _c, port=interface.INTERFACE_PORT: events.append(f"stop:{port}"))
    monkeypatch.setattr(interface, "_cleanup_repo_local_interface_processes", lambda: events.append("cleanup") or [])
    monkeypatch.setattr(interface, "_cleanup_playwright_server_log", lambda: events.append("cleanup-log"))
    monkeypatch.setattr(interface, "_start_playwright_server", lambda env, port=interface.INTERFACE_PORT, server_mode="start": (_ for _ in ()).throw(RuntimeError("startup failed")))

    try:
        interface.e2e.body(ctx, project="chromium", spec="e2e/specs/navigation.spec.ts")
    except RuntimeError as exc:
        assert str(exc) == "startup failed"
    else:
        raise AssertionError("expected managed server startup failure")

    assert events == [
        "stop:4317",
        "stop:4317",
        "cleanup",
    ]


def test_cleanup_playwright_server_log_retries_permission_error(monkeypatch):
    events: list[str] = []

    class FakePath:
        def __init__(self):
            self.calls = 0

        def unlink(self):
            self.calls += 1
            events.append(f"unlink:{self.calls}")
            if self.calls < 3:
                raise PermissionError("busy")

    fake_path = FakePath()

    monkeypatch.setattr(interface, "_playwright_server_log_path", lambda: "workspace/logs/interface-playwright-webserver.log")
    monkeypatch.setattr(interface, "Path", lambda _value: fake_path)
    monkeypatch.setattr(interface.time, "sleep", lambda _n: events.append("sleep"))

    interface._cleanup_playwright_server_log()

    assert events == ["unlink:1", "sleep", "unlink:2", "sleep", "unlink:3"]


def test_cleanup_stale_next_dev_lock_removes_orphaned_lock(monkeypatch, tmp_path, capsys):
    lock_path = tmp_path / ".next" / "dev" / "lock"
    lock_path.parent.mkdir(parents=True)
    lock_path.write_text("", encoding="utf-8")

    monkeypatch.setattr(interface, "INTERFACE_DIR", tmp_path)
    monkeypatch.setattr(interface, "_list_repo_local_interface_processes", lambda: [])

    removed = interface._cleanup_stale_next_dev_lock()

    assert removed is True
    assert not lock_path.exists()
    assert "Cleared stale Next.js dev lock" in capsys.readouterr().out


def test_cleanup_stale_next_dev_lock_keeps_lock_when_workers_exist(monkeypatch, tmp_path, capsys):
    lock_path = tmp_path / ".next" / "dev" / "lock"
    lock_path.parent.mkdir(parents=True)
    lock_path.write_text("", encoding="utf-8")

    monkeypatch.setattr(interface, "INTERFACE_DIR", tmp_path)
    monkeypatch.setattr(
        interface,
        "_list_repo_local_interface_processes",
        lambda: [{"pid": 404, "name": "node.exe", "command": "interface/.next/dev/build/postcss.js"}],
    )

    removed = interface._cleanup_stale_next_dev_lock()

    assert removed is False
    assert lock_path.exists()
    assert "Cleared stale Next.js dev lock" not in capsys.readouterr().out


def test_start_playwright_server_dev_mode_clears_stale_next_lock(monkeypatch, tmp_path):
    events: list[str] = []

    class FakeProcess:
        pass

    monkeypatch.setattr(interface, "_playwright_server_log_path", lambda: str(tmp_path / "playwright.log"))
    monkeypatch.setattr(interface, "_cleanup_stale_next_dev_lock", lambda: events.append("cleanup-dev-lock") or True)
    monkeypatch.setattr(
        interface,
        "_spawn_interface_process",
        lambda command, env, stdout, stderr, detached, text=True: events.append(f"spawn:{' '.join(command)}") or FakeProcess(),
    )

    interface._start_playwright_server({"INTERFACE_BIND_HOST": "127.0.0.1"}, port=4315, server_mode="dev")

    assert events == [
        "cleanup-dev-lock",
        f"spawn:node {str((interface.INTERFACE_DIR / 'node_modules' / 'next' / 'dist' / 'bin' / 'next').resolve())} dev --webpack --hostname 127.0.0.1 --port 4315",
    ]


def test_start_playwright_server_start_mode_skips_dev_lock_cleanup(monkeypatch, tmp_path):
    events: list[str] = []

    class FakeProcess:
        pass

    monkeypatch.setattr(interface, "_playwright_server_log_path", lambda: str(tmp_path / "playwright.log"))
    monkeypatch.setattr(interface, "_cleanup_stale_next_dev_lock", lambda: events.append("cleanup-dev-lock") or True)
    monkeypatch.setattr(
        interface,
        "_spawn_interface_process",
        lambda command, env, stdout, stderr, detached, text=True: events.append(f"spawn:{command[-1]}") or FakeProcess(),
    )

    interface._start_playwright_server({"INTERFACE_BIND_HOST": "127.0.0.1"}, port=4315, server_mode="start")

    assert events == [f"spawn:{str((interface.INTERFACE_DIR / 'scripts' / 'playwright-webserver.mjs').resolve())}"]


