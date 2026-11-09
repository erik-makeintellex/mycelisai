from __future__ import annotations

from ops import interface_runtime as interface
from tests.test_interface_tasks import FakeContext


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
