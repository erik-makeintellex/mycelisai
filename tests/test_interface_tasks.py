from __future__ import annotations

from contextlib import contextmanager
from dataclasses import dataclass
from invoke import Context

from ops import interface_env
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


def test_interface_env_keeps_root_env_as_secret_store(monkeypatch, tmp_path):
    (tmp_path / ".env").write_text(
        "MYCELIS_API_KEY=root-key\nMYCELIS_API_HOST=root-host\nPORT=8081\n",
        encoding="utf-8",
    )
    (tmp_path / ".env.compose").write_text(
        "MYCELIS_API_KEY=compose-key\nMYCELIS_API_HOST=compose-host\n",
        encoding="utf-8",
    )

    monkeypatch.setattr(interface_env, "ROOT_DIR", tmp_path)
    monkeypatch.delenv("MYCELIS_API_KEY", raising=False)
    monkeypatch.delenv("MYCELIS_API_HOST", raising=False)
    monkeypatch.delenv("PORT", raising=False)

    interface_env._load_env()

    assert interface_env.os.environ["MYCELIS_API_KEY"] == "root-key"
    assert interface_env.os.environ["MYCELIS_API_HOST"] == "root-host"
    assert "PORT" not in interface_env.os.environ
