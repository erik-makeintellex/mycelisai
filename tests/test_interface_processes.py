from __future__ import annotations

import json
import subprocess

from ops import interface_processes


def test_windows_process_inventory_is_bounded_and_repo_scoped():
    repo_server = (
        r"node E:\Business\MakeIntellex\Projects\mycelisai\scratch\interface"
        r"\scripts\playwright-webserver.mjs"
    )
    payload = json.dumps(
        [
            {"ProcessId": 4101, "Name": "node.exe", "CommandLine": repo_server},
            {
                "ProcessId": 4102,
                "Name": "node.exe",
                "CommandLine": r"node E:\OtherProject\interface\scripts\playwright-webserver.mjs",
            },
        ]
    )
    calls: list[tuple[list[str], dict[str, object]]] = []

    def fake_run(command, **kwargs):
        calls.append((command, kwargs))
        return subprocess.CompletedProcess(command, 0, stdout=payload, stderr="")

    processes = interface_processes.list_repo_local_interface_processes(
        is_windows_func=lambda: True,
        normalize_process_text=lambda value: value.lower().replace("\\", "/"),
        run=fake_run,
    )

    assert [process["pid"] for process in processes] == [4101]
    assert len(calls) == 1
    assert calls[0][0][0:3] == ["powershell", "-NoProfile", "-Command"]
    assert calls[0][1]["timeout"] == interface_processes.WINDOWS_PROCESS_QUERY_TIMEOUT_SECONDS


def test_windows_tree_kill_allows_slow_descendant_enumeration():
    calls: list[tuple[list[str], dict[str, object]]] = []

    def fake_run(command, **kwargs):
        calls.append((command, kwargs))
        return subprocess.CompletedProcess(command, 0)

    interface_processes.kill_pid_tree(
        4317,
        is_windows_func=lambda: True,
        run=fake_run,
    )

    taskkill, taskkill_kwargs = calls[0]
    assert taskkill == ["taskkill", "/F", "/T", "/PID", "4317"]
    assert taskkill_kwargs["timeout"] == interface_processes.WINDOWS_PROCESS_TREE_KILL_TIMEOUT_SECONDS
    assert calls[1][0][0:3] == ["powershell", "-NoProfile", "-Command"]


def test_windows_tree_kill_falls_back_to_parent_stop_after_timeout():
    calls: list[list[str]] = []

    def fake_run(command, **kwargs):
        calls.append(command)
        if command[0] == "taskkill":
            raise subprocess.TimeoutExpired(command, kwargs["timeout"])
        return subprocess.CompletedProcess(command, 0)

    interface_processes.kill_pid_tree(
        4318,
        is_windows_func=lambda: True,
        run=fake_run,
    )

    assert calls[0] == ["taskkill", "/F", "/T", "/PID", "4318"]
    assert calls[1][0:3] == ["powershell", "-NoProfile", "-Command"]
