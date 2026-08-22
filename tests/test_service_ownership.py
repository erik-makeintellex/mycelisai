from __future__ import annotations

from pathlib import Path

import pytest
from invoke import Context

from ops import db as db_tasks
from ops import interface_runtime as interface
from ops import interface_processes, lifecycle
from tests.interface_task_support import FakeContext


def test_cleanup_managed_interface_listeners_skips_foreign_port_range_pids():
    killed: list[int] = []
    sleeps: list[float] = []

    def fake_repo_local_processes(pids: list[int]):
        assert pids == [5101, 5102]
        return [{"pid": 5102, "name": "node.exe", "command": "repo-local interface worker"}]

    cleaned = interface_processes.cleanup_managed_interface_listeners(
        3100,
        3199,
        is_windows_func=lambda: True,
        windows_listening_pids_for_port_range_func=lambda _start, _end: [5101, 5102],
        repo_local_processes_for_pids_func=fake_repo_local_processes,
        kill_pid_tree_func=killed.append,
        sleep=sleeps.append,
    )

    assert cleaned == [5102]
    assert killed == [5102]
    assert sleeps == [0.5]


def test_stop_runs_tree_kill_for_repo_owned_listener_on_windows(monkeypatch):
    cleaned: list[str] = []
    killed: list[int] = []
    port = 4310

    monkeypatch.setattr(interface, "is_windows", lambda: True)
    monkeypatch.setattr(interface, "_cleanup_repo_local_interface_processes", lambda: cleaned.append("repo") or [])
    monkeypatch.setattr(interface, "_cleanup_managed_interface_listeners", lambda: cleaned.append("managed") or [])
    monkeypatch.setattr(interface, "_windows_listening_pids_for_port", lambda _port: [1234])
    monkeypatch.setattr(
        interface,
        "_repo_local_interface_processes_for_pids",
        lambda _pids: [{"pid": 1234, "name": "node.exe", "command": "repo-local interface"}],
    )
    monkeypatch.setattr(interface, "_kill_pid_tree", lambda pid: killed.append(pid))

    interface.stop.body(FakeContext(), port=port)

    assert killed == [1234]
    assert cleaned == ["managed", "repo"]


def test_stop_does_not_kill_foreign_listener_on_requested_port(monkeypatch, capsys):
    cleaned: list[str] = []
    killed: list[int] = []

    monkeypatch.setattr(interface, "is_windows", lambda: True)
    monkeypatch.setattr(interface, "_cleanup_repo_local_interface_processes", lambda: cleaned.append("repo") or [])
    monkeypatch.setattr(interface, "_cleanup_managed_interface_listeners", lambda: cleaned.append("managed") or [])
    monkeypatch.setattr(interface, "_windows_listening_pids_for_port", lambda _port: [7001])
    monkeypatch.setattr(interface, "_repo_local_interface_processes_for_pids", lambda _pids: [])
    monkeypatch.setattr(interface, "_kill_pid_tree", lambda pid: killed.append(pid))

    interface.stop.body(FakeContext(), port=4310)

    assert killed == []
    assert cleaned == ["managed", "repo"]
    assert "non-repo process" in capsys.readouterr().out


def test_lifecycle_up_fails_closed_when_api_port_is_foreign_core(monkeypatch):
    killed_ports: list[tuple[int, str]] = []
    started: list[str] = []

    monkeypatch.setattr(Path, "exists", lambda self: True)
    monkeypatch.setattr(lifecycle.lifecycle_infra, "database_endpoint", lambda _root: ("127.0.0.1", 5432))
    monkeypatch.setattr(lifecycle, "_ensure_bridge", lambda: None)
    monkeypatch.setattr(lifecycle, "_wait_for_port", lambda *args, **kwargs: True)
    monkeypatch.setattr(db_tasks.create, "body", lambda _c: None)
    monkeypatch.setattr(
        lifecycle,
        "_port_open",
        lambda port, host="127.0.0.1", timeout=1.0: port in (5432, 4222, lifecycle.API_PORT),
    )
    monkeypatch.setattr(lifecycle, "_owned_core_pid_on_port", lambda _port=lifecycle.API_PORT: None)
    monkeypatch.setattr(lifecycle, "_kill_port", lambda port, label: killed_ports.append((port, label)) or True)
    monkeypatch.setattr(lifecycle, "_start_core_background", lambda: started.append("core") or True)

    with pytest.raises(SystemExit, match="non-repo process"):
        lifecycle.up.body(Context(), frontend=False, build=False)

    assert killed_ports == []
    assert started == []


def test_lifecycle_down_does_not_kill_foreign_interface_port_listener(monkeypatch, capsys):
    killed: list[int] = []

    from ops import interface as interface_tasks

    monkeypatch.setenv("MYCELIS_DEV_INFRA_MODE", "compose")
    monkeypatch.setattr(lifecycle, "_owned_core_pid_on_port", lambda _port=lifecycle.API_PORT: None)
    monkeypatch.setattr(lifecycle, "_owned_frontend_pid_on_port", lambda _port=lifecycle.INTERFACE_PORT: None)
    monkeypatch.setattr(lifecycle, "_port_open", lambda port, host="127.0.0.1", timeout=1.0: port == lifecycle.INTERFACE_PORT)
    monkeypatch.setattr(lifecycle, "_kill_pid", lambda pid: killed.append(pid))
    monkeypatch.setattr(lifecycle, "_kill_compiled_go_services", lambda: [])
    monkeypatch.setattr(lifecycle, "_remaining_managed_services", lambda: [])
    monkeypatch.setattr(interface_tasks, "_cleanup_repo_local_interface_processes", lambda: [])

    lifecycle.down.body(Context())

    assert killed == []
    assert f"Frontend port {lifecycle.INTERFACE_PORT} is held by a non-repo process" in capsys.readouterr().out
