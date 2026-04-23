from __future__ import annotations

from contextlib import contextmanager
from dataclasses import dataclass
from invoke import Context

from ops import interface_runtime as interface


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


def test_install_provisions_npm_and_playwright(monkeypatch):
    commands: list[str] = []

    monkeypatch.setattr(interface, "run_interface_command", lambda _ctx, command, **_kwargs: commands.append(command))

    interface.install.body(FakeContext())

    assert commands == [
        "npm install",
        "npx playwright install chromium",
    ]


