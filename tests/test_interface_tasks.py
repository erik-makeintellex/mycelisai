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
    command = f'"node" {interface.INTERFACE_DIR / ".next" / "dev" / "build" / "postcss.js"} 52847'

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


def test_cleanup_repo_local_interface_processes_suppresses_timeout_warnings(monkeypatch, capsys):
    monkeypatch.setattr(
        interface,
        "_list_repo_local_interface_processes",
        lambda: (_ for _ in ()).throw(RuntimeError("process query failed: timed out after 5 seconds")),
    )

    remaining = interface._cleanup_repo_local_interface_processes()

    assert remaining == []
    assert "unable to inspect repo-local Interface residuals" not in capsys.readouterr().out


def test_list_repo_local_interface_processes_windows_queries_tasklist_then_cim(monkeypatch):
    commands: list[list[str]] = []
    timeouts: list[int] = []

    class Result:
        def __init__(self, stdout="", returncode=0, stderr=""):
            self.stdout = stdout
            self.returncode = returncode
            self.stderr = stderr

    def fake_run(command, capture_output=True, text=True, timeout=0):
        commands.append(command)
        timeouts.append(timeout)
        if command[:4] == ["tasklist", "/FO", "CSV", "/NH"]:
            image_name = command[-1].removeprefix("IMAGENAME eq ")
            if image_name == "node.exe":
                return Result(
                    stdout='"node.exe","101","Console","1","12,000 K"\n'
                    '"node.exe","103","Console","1","12,000 K"\n'
                    '"node.exe","104","Console","1","12,000 K"\n'
                    '"node.exe","102","Console","1","12,000 K"',
                )
            return Result(
                stdout='"cmd.exe","201","Console","1","12,000 K"',
            )
        return Result(
            stdout='[{"ProcessId":101,"Name":"node.exe","CommandLine":"D:/MakeIntellex/Projects/mycelisai/scratch/interface/.next/dev/build/postcss.js"},'
            '{"ProcessId":103,"Name":"node.exe","CommandLine":"D:/MakeIntellex/Projects/mycelisai/scratch/interface/scripts/playwright-webserver.mjs"},'
            '{"ProcessId":104,"Name":"node.exe","CommandLine":"D:/MakeIntellex/Projects/mycelisai/scratch/interface/.next/standalone/server.js"},'
            '{"ProcessId":102,"Name":"node.exe","CommandLine":"C:/other-app/node_modules/vite/bin/vite.js"},'
            '{"ProcessId":201,"Name":"cmd.exe","CommandLine":"C:/Windows/System32/cmd.exe /d /c echo unrelated"}]',
        )

    monkeypatch.setattr(interface, "is_windows", lambda: True)
    monkeypatch.setattr(interface.subprocess, "run", fake_run)

    processes = interface._list_repo_local_interface_processes()

    assert processes == [
        {
            "pid": 101,
            "name": "node.exe",
            "command": "D:/MakeIntellex/Projects/mycelisai/scratch/interface/.next/dev/build/postcss.js",
        },
        {
            "pid": 103,
            "name": "node.exe",
            "command": "D:/MakeIntellex/Projects/mycelisai/scratch/interface/scripts/playwright-webserver.mjs",
        },
        {
            "pid": 104,
            "name": "node.exe",
            "command": "D:/MakeIntellex/Projects/mycelisai/scratch/interface/.next/standalone/server.js",
        },
    ]
    assert commands == [
        ["tasklist", "/FO", "CSV", "/NH", "/FI", "IMAGENAME eq node.exe"],
        ["tasklist", "/FO", "CSV", "/NH", "/FI", "IMAGENAME eq cmd.exe"],
        [
            "powershell",
            "-NoProfile",
            "-Command",
            "Get-CimInstance Win32_Process -Filter \"ProcessId = 101 OR ProcessId = 103 OR ProcessId = 104 OR ProcessId = 102 OR ProcessId = 201\" | "
            "Select-Object ProcessId,Name,CommandLine | "
            "ConvertTo-Json -Compress",
        ]
    ]
    assert timeouts == [20, 20, 8]


def test_windows_listening_pids_for_port_range_filters_managed_ports(monkeypatch):
    class Result:
        def __init__(self, stdout="", returncode=0, stderr=""):
            self.stdout = stdout
            self.returncode = returncode
            self.stderr = stderr

    def fake_run(command, capture_output=True, text=True, timeout=0):
        return Result(
            stdout="\n".join(
                [
                    "  TCP    0.0.0.0:3099           0.0.0.0:0              LISTENING       100",
                    "  TCP    0.0.0.0:3100           0.0.0.0:0              LISTENING       101",
                    "  TCP    127.0.0.1:3105         0.0.0.0:0              LISTENING       105",
                    "  TCP    [::]:3110              [::]:0                 LISTENING       110",
                    "  TCP    0.0.0.0:3199           0.0.0.0:0              LISTENING       199",
                    "  TCP    0.0.0.0:3200           0.0.0.0:0              LISTENING       200",
                ]
            ),
        )

    monkeypatch.setattr(interface.subprocess, "run", fake_run)

    assert interface._windows_listening_pids_for_port_range(3100, 3199) == [101, 105, 110, 199]


def test_build_cleans_residual_interface_workers(monkeypatch):
    cleaned: list[str] = []
    stopped: list[str] = []
    shell_calls: list[list[str]] = []
    ctx = FakeContext()

    monkeypatch.setattr(interface, "stop", lambda _c, port=interface.INTERFACE_PORT: stopped.append(f"stop:{port}"))
    monkeypatch.setattr(interface, "clean", lambda _c: cleaned.append("clean") or None)
    monkeypatch.setattr(interface, "_cleanup_repo_local_interface_processes", lambda: cleaned.append("cleanup") or [])
    monkeypatch.setattr(interface, "_wait_for_complete_next_build_output", lambda timeout_seconds=20: None)
    monkeypatch.setattr(
        interface,
        "_run_interface_shell_command",
        lambda command, extra_env=None: shell_calls.append(command) or interface.CommandResult(exited=0, stdout="", stderr=""),
    )

    interface.build.body(ctx)

    assert stopped == [f"stop:{interface.INTERFACE_PORT}"]
    assert cleaned == ["clean", "cleanup"]
    assert shell_calls == [["npm", "run", "build"]]


def test_build_retries_once_after_next_lock_conflict(monkeypatch):
    cleaned: list[str] = []
    stopped: list[str] = []
    shell_calls: list[list[str]] = []
    shell_results = iter(
        [
            interface.CommandResult(
                exited=1,
                stdout="",
                stderr="Unable to acquire lock at D:\\repo\\interface\\.next\\lock, is another instance of next build running?",
            ),
            interface.CommandResult(exited=0, stdout="", stderr=""),
        ]
    )
    ctx = FakeContext()

    monkeypatch.setattr(interface, "stop", lambda _c, port=interface.INTERFACE_PORT: stopped.append(f"stop:{port}"))
    monkeypatch.setattr(interface, "clean", lambda _c: cleaned.append("clean") or None)
    monkeypatch.setattr(interface, "_cleanup_repo_local_interface_processes", lambda: cleaned.append("cleanup") or [])
    monkeypatch.setattr(interface, "_wait_for_complete_next_build_output", lambda timeout_seconds=20: None)
    monkeypatch.setattr(
        interface,
        "_run_interface_shell_command",
        lambda command, extra_env=None: shell_calls.append(command) or next(shell_results),
    )

    interface.build.body(ctx)

    assert stopped == [f"stop:{interface.INTERFACE_PORT}"]
    assert cleaned == ["clean", "cleanup", "cleanup", "clean", "cleanup"]
    assert shell_calls == [["npm", "run", "build"], ["npm", "run", "build"]]


def test_build_retries_once_after_incomplete_next_build_output(monkeypatch):
    cleaned: list[str] = []
    stopped: list[str] = []
    shell_calls: list[list[str]] = []
    shell_results = iter(
        [
            interface.CommandResult(
                exited=1,
                stdout="",
                stderr="Error: ENOENT: no such file or directory, open 'D:\\repo\\interface\\.next\\build-manifest.json'",
            ),
            interface.CommandResult(exited=0, stdout="", stderr=""),
        ]
    )
    ctx = FakeContext()

    monkeypatch.setattr(interface, "stop", lambda _c, port=interface.INTERFACE_PORT: stopped.append(f"stop:{port}"))
    monkeypatch.setattr(interface, "clean", lambda _c: cleaned.append("clean") or None)
    monkeypatch.setattr(interface, "_cleanup_repo_local_interface_processes", lambda: cleaned.append("cleanup") or [])
    monkeypatch.setattr(interface, "_wait_for_complete_next_build_output", lambda timeout_seconds=20: None)
    monkeypatch.setattr(
        interface,
        "_run_interface_shell_command",
        lambda command, extra_env=None: shell_calls.append(command) or next(shell_results),
    )

    interface.build.body(ctx)

    assert stopped == [f"stop:{interface.INTERFACE_PORT}"]
    assert cleaned == ["clean", "cleanup", "cleanup", "clean", "cleanup"]
    assert shell_calls == [["npm", "run", "build"], ["npm", "run", "build"]]


def test_build_retries_once_after_next_standalone_cleanup_conflict(monkeypatch):
    cleaned: list[str] = []
    stopped: list[str] = []
    shell_calls: list[list[str]] = []
    shell_results = iter(
        [
            interface.CommandResult(
                exited=1,
                stdout="",
                stderr="Error: EBUSY: resource busy or locked, rmdir 'D:\\repo\\interface\\.next\\standalone'",
            ),
            interface.CommandResult(exited=0, stdout="", stderr=""),
        ]
    )
    ctx = FakeContext()

    monkeypatch.setattr(interface, "stop", lambda _c, port=interface.INTERFACE_PORT: stopped.append(f"stop:{port}"))
    monkeypatch.setattr(interface, "clean", lambda _c: cleaned.append("clean") or None)
    monkeypatch.setattr(interface, "_cleanup_repo_local_interface_processes", lambda: cleaned.append("cleanup") or [])
    monkeypatch.setattr(interface, "_wait_for_complete_next_build_output", lambda timeout_seconds=20: None)
    monkeypatch.setattr(
        interface,
        "_run_interface_shell_command",
        lambda command, extra_env=None: shell_calls.append(command) or next(shell_results),
    )

    interface.build.body(ctx)

    assert stopped == [f"stop:{interface.INTERFACE_PORT}"]
    assert cleaned == ["clean", "cleanup", "cleanup", "clean", "cleanup"]
    assert shell_calls == [["npm", "run", "build"], ["npm", "run", "build"]]


def test_build_retries_once_after_next_standalone_cleanup_conflict_sweeps_managed_listeners(monkeypatch):
    cleaned: list[str] = []
    swept: list[str] = []
    stopped: list[str] = []
    shell_calls: list[list[str]] = []
    shell_results = iter(
        [
            interface.CommandResult(
                exited=1,
                stdout="",
                stderr="Error: EBUSY: resource busy or locked, rmdir 'D:\\repo\\interface\\.next\\standalone'",
            ),
            interface.CommandResult(exited=0, stdout="", stderr=""),
        ]
    )
    ctx = FakeContext()

    monkeypatch.setattr(interface, "stop", lambda _c, port=interface.INTERFACE_PORT: stopped.append(f"stop:{port}"))
    monkeypatch.setattr(interface, "clean", lambda _c: cleaned.append("clean") or None)
    monkeypatch.setattr(interface, "_cleanup_repo_local_interface_processes", lambda: cleaned.append("cleanup") or [])
    monkeypatch.setattr(interface, "_cleanup_managed_interface_listeners", lambda: swept.append("sweep") or [3100, 3101])
    monkeypatch.setattr(interface, "_wait_for_complete_next_build_output", lambda timeout_seconds=20: None)
    monkeypatch.setattr(
        interface,
        "_run_interface_shell_command",
        lambda command, extra_env=None: shell_calls.append(command) or next(shell_results),
    )

    interface.build.body(ctx)

    assert swept == ["sweep", "sweep"]
    assert stopped == [f"stop:{interface.INTERFACE_PORT}"]
    assert cleaned == ["clean", "cleanup", "cleanup", "clean", "cleanup"]
    assert shell_calls == [["npm", "run", "build"], ["npm", "run", "build"]]


def test_wait_for_complete_next_build_output_accepts_present_manifest_artifacts(monkeypatch, tmp_path):
    next_dir = tmp_path / ".next"
    static_dir = next_dir / "static" / "chunks"
    static_dir.mkdir(parents=True)
    (next_dir / "required-server-files.json").write_text("{}", encoding="utf-8")
    (next_dir / "build-manifest.json").write_text(
        '{"polyfillFiles":["static/chunks/polyfills.js"],"lowPriorityFiles":["static/build/_buildManifest.js"],"rootMainFiles":["static/chunks/main.js"],"pages":{"/_app":["static/chunks/app.js"]}}',
        encoding="utf-8",
    )
    (next_dir / "static" / "chunks" / "polyfills.js").write_text("", encoding="utf-8")
    (next_dir / "static" / "chunks" / "main.js").write_text("", encoding="utf-8")
    (next_dir / "static" / "chunks" / "app.js").write_text("", encoding="utf-8")
    (next_dir / "static" / "build").mkdir(parents=True)
    (next_dir / "static" / "build" / "_buildManifest.js").write_text("", encoding="utf-8")

    monkeypatch.setattr(interface, "INTERFACE_DIR", tmp_path)

    interface._wait_for_complete_next_build_output(timeout_seconds=1)


def test_wait_for_complete_next_build_output_reports_missing_manifest_artifacts(monkeypatch, tmp_path):
    next_dir = tmp_path / ".next"
    next_dir.mkdir(parents=True)
    (next_dir / "required-server-files.json").write_text("{}", encoding="utf-8")
    (next_dir / "build-manifest.json").write_text(
        '{"rootMainFiles":["static/chunks/main.js"]}',
        encoding="utf-8",
    )

    monkeypatch.setattr(interface, "INTERFACE_DIR", tmp_path)
    monkeypatch.setattr(interface.time, "sleep", lambda _n: None)

    try:
        interface._wait_for_complete_next_build_output(timeout_seconds=1)
    except RuntimeError as exc:
        assert "main.js" in str(exc)
    else:
        raise AssertionError("expected incomplete build output failure")


def test_clean_ignores_missing_files_during_rmtree(monkeypatch):
    removed: list[str] = []

    monkeypatch.setattr(interface.os.path, "isdir", lambda path: True)
    monkeypatch.setattr(interface, "_cleanup_repo_local_interface_processes", lambda: [])

    def fake_rmtree(path, onexc=None):
        removed.append(path)
        assert onexc is not None
        try:
            raise FileNotFoundError("gone")
        except FileNotFoundError:
            onexc(None, path, __import__("sys").exc_info())

    monkeypatch.setattr(interface.shutil, "rmtree", fake_rmtree)

    interface.clean.body(FakeContext())

    assert removed == [interface.os.path.join("interface", ".next")]


def test_clean_ignores_missing_files_when_rmtree_passes_exception_object(monkeypatch):
    removed: list[str] = []

    monkeypatch.setattr(interface.os.path, "isdir", lambda path: True)
    monkeypatch.setattr(interface, "_cleanup_repo_local_interface_processes", lambda: [])

    def fake_rmtree(path, onexc=None):
        removed.append(path)
        assert onexc is not None
        onexc(None, path, FileNotFoundError("gone"))

    monkeypatch.setattr(interface.shutil, "rmtree", fake_rmtree)

    interface.clean.body(FakeContext())

    assert removed == [interface.os.path.join("interface", ".next")]


def test_clean_warns_and_continues_when_cache_directory_stays_locked(monkeypatch, capsys):
    attempts: list[str] = []
    cache_dir = interface.os.path.join("interface", ".next")

    monkeypatch.setattr(interface.os.path, "isdir", lambda path: path == cache_dir)
    monkeypatch.setattr(interface, "_cleanup_repo_local_interface_processes", lambda: [])
    monkeypatch.setattr(interface.time, "sleep", lambda _n: None)

    def fake_rmtree(path, onexc=None):
        attempts.append(path)
        raise PermissionError("busy")

    monkeypatch.setattr(interface.shutil, "rmtree", fake_rmtree)

    interface.clean.body(FakeContext())

    assert attempts == [cache_dir, cache_dir, cache_dir]
    output = capsys.readouterr().out
    assert "could not be fully removed" in output
    assert "Cache cleared." in output


def test_stop_runs_tree_kill_and_repo_cleanup_on_windows(monkeypatch):
    cleaned: list[str] = []
    killed: list[int] = []
    ctx = FakeContext()
    port = 4310

    monkeypatch.setattr(interface, "is_windows", lambda: True)
    monkeypatch.setattr(interface, "_cleanup_repo_local_interface_processes", lambda: cleaned.append("cleanup") or [])
    monkeypatch.setattr(interface, "_windows_listening_pids_for_port", lambda _port: [1234])
    monkeypatch.setattr(interface, "_kill_pid_tree", lambda pid: killed.append(pid))

    interface.stop.body(ctx, port=port)

    assert killed == [1234]
    assert cleaned == ["cleanup"]


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


def test_typecheck_uses_direct_shell_command(monkeypatch):
    shell_calls: list[list[str]] = []
    ctx = FakeContext()

    monkeypatch.setattr(
        interface,
        "_run_interface_shell_command",
        lambda command, extra_env=None: shell_calls.append(command) or interface.CommandResult(exited=0, stdout="", stderr=""),
    )

    interface.typecheck.body(ctx)

    assert shell_calls == [["npx", "tsc", "--noEmit"]]


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
    occupied_ports = {3000}

    class FakeSocket:
        def __init__(self, family, *args, **kwargs):
            self.family = family
            self.bind_calls: list[tuple[str, int]] = []
            self.closed = False

        def setsockopt(self, *args, **kwargs):
            return None

        def settimeout(self, *args, **kwargs):
            return None

        def connect_ex(self, address):
            return 0 if address[1] in occupied_ports else 111

        def bind(self, address):
            self.bind_calls.append(address)
            if address[1] in occupied_ports:
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


def test_pick_interface_port_uses_ipv6_bind_host_without_dual_binding(monkeypatch):
    occupied_ports: set[int] = set()
    bind_calls: list[tuple[int, tuple[str, int]]] = []
    sockopts: list[tuple[int, int, int]] = []

    class FakeSocket:
        def __init__(self, family, *args, **kwargs):
            self.family = family
            self.port = 43210

        def setsockopt(self, level, option, value):
            sockopts.append((level, option, value))

        def settimeout(self, *args, **kwargs):
            return None

        def connect_ex(self, address):
            return 0 if address[1] in occupied_ports else 111

        def bind(self, address):
            bind_calls.append((self.family, address))
            if address[1] in occupied_ports:
                raise OSError("address in use")
            self.port = address[1] or 43210

        def getsockname(self):
            if self.family == interface.socket.AF_INET6:
                return ("::", self.port)
            return ("127.0.0.1", self.port)

        def close(self):
            return None

        def __enter__(self):
            return self

        def __exit__(self, exc_type, exc, tb):
            self.close()

    monkeypatch.setattr(interface, "INTERFACE_BIND_HOST", "::")
    monkeypatch.setattr(interface.socket, "socket", lambda family, *args, **kwargs: FakeSocket(family))

    assert interface._pick_interface_port(3000) == 3100
    assert bind_calls == [(interface.socket.AF_INET6, ("::", 3100))]
    assert sockopts == [(interface.socket.IPPROTO_IPV6, interface.socket.IPV6_V6ONLY, 0)]


def test_wait_for_interface_ready_fails_when_managed_server_exits(monkeypatch):
    class FakeServer:
        @staticmethod
        def poll():
            return 1

    monkeypatch.setattr(interface.time, "sleep", lambda _n: None)

    try:
        interface._wait_for_interface_ready("127.0.0.1", 4310, timeout_seconds=1, process=FakeServer())
    except RuntimeError as exc:
        assert "Managed Interface server exited before it became ready" in str(exc)
        assert "4310" in str(exc)
    else:
        raise AssertionError("expected managed server startup failure")


def test_wait_for_interface_ready_prefers_reachable_port_over_exited_parent(monkeypatch):
    class FakeServer:
        @staticmethod
        def poll():
            return 1

    class FakeHTTPResponse:
        status = 200

        def __enter__(self):
            return self

        def __exit__(self, exc_type, exc, tb):
            return False

    monkeypatch.setattr(interface.time, "sleep", lambda _n: None)
    monkeypatch.setattr(interface.urllib.request, "urlopen", lambda url, timeout=5: FakeHTTPResponse())

    assert interface._wait_for_interface_ready("127.0.0.1", 4310, timeout_seconds=1, process=FakeServer()) == "127.0.0.1"


def test_check_does_not_treat_plain_html_words_as_hydration_failure(monkeypatch, capsys):
    class FakeHTTPResponse:
        def __init__(self, body: str):
            self.status = 200
            self._body = body.encode("utf-8")

        def read(self):
            return self._body

        def __enter__(self):
            return self

        def __exit__(self, exc_type, exc, tb):
            return False

    monkeypatch.setattr(
        interface.urllib.request,
        "urlopen",
        lambda req, timeout=10: FakeHTTPResponse("<html><body>hydration and error words in static docs text</body></html>"),
    )

    interface.check.body(FakeContext(), port=3000)

    out = capsys.readouterr().out
    assert "ALL PAGES HEALTHY." in out
