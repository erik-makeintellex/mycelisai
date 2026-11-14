from ops import interface_runtime as interface
from tests.interface_task_support import FakeContext


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
    expected_bind_host = "127.0.0.1" if interface.is_windows() else interface.INTERFACE_BIND_HOST
    assert env_seen["INTERFACE_BIND_HOST"] == expected_bind_host
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
            {"command": command, "env": dict(env), "detached": detached, "text": text}
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
    assert events == ["stop:4313", "start:4313:dev", "ready:127.0.0.1:4313", "kill:6262", "stop:4313", "cleanup", "cleanup-log"]


def test_e2e_live_backend_start_mode_marks_env_and_refreshes_built_bundle(monkeypatch):
    ctx = FakeContext()
    events: list[str] = []
    env_requests: list[tuple[bool, int]] = []
    streamed_env: dict[str, str] = {}

    class FakeServer:
        pid = 7373

        @staticmethod
        def poll():
            return None

    monkeypatch.setattr(interface, "INTERFACE_PORT", 4318)
    monkeypatch.setattr(interface, "_pick_interface_port", lambda preferred=interface.INTERFACE_PORT: 4318)
    monkeypatch.setattr(interface, "stop", lambda _c, port=interface.INTERFACE_PORT: events.append(f"stop:{port}"))
    monkeypatch.setattr(interface, "build", lambda _c: events.append("build"))
    monkeypatch.setattr(
        interface,
        "_build_playwright_env",
        lambda *, live_backend, port: env_requests.append((live_backend, port)) or {
            "PLAYWRIGHT_SKIP_WEBSERVER": "1",
            "PLAYWRIGHT_LIVE_BACKEND": "1" if live_backend else "0",
            "INTERFACE_HOST": "127.0.0.1",
            "INTERFACE_BIND_HOST": interface.INTERFACE_BIND_HOST,
            "INTERFACE_PORT": str(port),
        },
    )
    monkeypatch.setattr(interface, "_cleanup_repo_local_interface_processes", lambda: events.append("cleanup") or [])
    monkeypatch.setattr(interface, "_cleanup_managed_interface_listeners", lambda: events.append("sweep") or [])
    monkeypatch.setattr(interface, "_cleanup_playwright_server_log", lambda: events.append("cleanup-log"))
    monkeypatch.setattr(interface, "_reconcile_managed_server_endpoint", lambda env, chosen_port, server: (env, chosen_port))
    monkeypatch.setattr(interface, "_kill_pid_tree", lambda pid: events.append(f"kill:{pid}"))
    monkeypatch.setattr(interface.time, "sleep", lambda _n: None)

    def fake_start(env, port=interface.INTERFACE_PORT, server_mode="start"):
        events.append(f"start:{port}:{server_mode}")
        return FakeServer()

    monkeypatch.setattr(interface, "_start_playwright_server", fake_start)
    monkeypatch.setattr(
        interface,
        "_run_playwright_command_streaming",
        lambda command, extra_env=None: streamed_env.update(extra_env or {}) or interface.CommandResult(exited=0, stdout="", stderr=""),
    )

    interface.e2e.body(ctx, project="chromium", spec="e2e/specs/soma-governance-live.spec.ts", live_backend=True, server_mode="start")

    assert env_requests == [(True, 4318)]
    assert streamed_env["PLAYWRIGHT_LIVE_BACKEND"] == "1"
    assert streamed_env["INTERFACE_PORT"] == "4318"
    assert events == ["stop:4318", "sweep", "build", "start:4318:start", "kill:7373", "stop:4318", "cleanup", "sweep", "cleanup-log"]
