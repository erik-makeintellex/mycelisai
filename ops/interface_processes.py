from __future__ import annotations

import json
import subprocess
from collections.abc import Callable
from contextlib import suppress
from typing import Any

from .interface_process_support import (
    _INTERFACE_PROCESS_COMMAND_HINTS,
    _INTERFACE_PROCESS_PATH_HINTS,
)


ProcessInfo = dict[str, str | int]
WINDOWS_PROCESS_TREE_KILL_TIMEOUT_SECONDS = 90
WINDOWS_PROCESS_QUERY_TIMEOUT_SECONDS = 180


def matches_repo_local_interface_process(
    name: str,
    command_line: str,
    *,
    normalize_process_text: Callable[[str], str],
) -> bool:
    normalized_name = (name or "").lower()
    normalized_cmd = normalize_process_text(command_line)
    if normalized_name not in {"node", "node.exe", "cmd", "cmd.exe"} or not normalized_cmd:
        return False
    if not any(hint in normalized_cmd for hint in _INTERFACE_PROCESS_PATH_HINTS):
        return False
    return any(hint in normalized_cmd for hint in _INTERFACE_PROCESS_COMMAND_HINTS)


def list_repo_local_interface_processes(
    *,
    is_windows_func: Callable[[], bool],
    normalize_process_text: Callable[[str], str],
    run: Callable[..., Any] = subprocess.run,
) -> list[ProcessInfo]:
    def _matches(name: str, command_line: str) -> bool:
        return matches_repo_local_interface_process(
            name,
            command_line,
            normalize_process_text=normalize_process_text,
        )

    processes: list[ProcessInfo] = []
    try:
        if is_windows_func():
            result = run(
                [
                    "powershell",
                    "-NoProfile",
                    "-Command",
                    "Get-CimInstance Win32_Process "
                    "-Filter \"Name = 'node.exe' OR Name = 'cmd.exe'\" | "
                    "Select-Object ProcessId,Name,CommandLine | "
                    "ConvertTo-Json -Compress",
                ],
                capture_output=True,
                text=True,
                timeout=WINDOWS_PROCESS_QUERY_TIMEOUT_SECONDS,
            )
            if result.returncode != 0:
                raise RuntimeError(result.stderr.strip() or "process query failed")
            raw = result.stdout.strip()
            if not raw:
                return []
            payload = json.loads(raw)
            rows = payload if isinstance(payload, list) else [payload]
            for row in rows:
                pid_text = row.get("ProcessId")
                name = row.get("Name") or ""
                command_line = row.get("CommandLine") or ""
                if not isinstance(pid_text, int):
                    continue
                if _matches(name, command_line):
                    processes.append({"pid": pid_text, "name": name, "command": command_line})
            return processes

        result = run(
            ["ps", "-eo", "pid=,comm=,args="],
            capture_output=True,
            text=True,
            timeout=10,
        )
        if result.returncode != 0:
            raise RuntimeError(result.stderr.strip() or "process query failed")
        for line in result.stdout.splitlines():
            parts = line.strip().split(None, 2)
            if len(parts) < 2 or not parts[0].isdigit():
                continue
            pid = int(parts[0])
            name = parts[1]
            command_line = parts[2] if len(parts) > 2 else ""
            if _matches(name, command_line):
                processes.append({"pid": pid, "name": name, "command": command_line})
    except (subprocess.SubprocessError, json.JSONDecodeError, OSError, ValueError, RuntimeError) as exc:
        raise RuntimeError(f"interface process inspection failed: {exc}") from exc
    return processes


def kill_pid_tree(
    pid: int,
    *,
    is_windows_func: Callable[[], bool],
    run: Callable[..., Any] = subprocess.run,
) -> None:
    try:
        if is_windows_func():
            run(
                ["taskkill", "/F", "/T", "/PID", str(pid)],
                capture_output=True,
                # Busy Windows hosts can spend close to a minute enumerating a
                # Next.js tree. A shorter timeout kills only the parent in the
                # fallback and leaves children holding .next build handles.
                timeout=WINDOWS_PROCESS_TREE_KILL_TIMEOUT_SECONDS,
            )
            with suppress(subprocess.SubprocessError, OSError):
                run(
                    [
                        "powershell",
                        "-NoProfile",
                        "-Command",
                        f"Stop-Process -Id {pid} -Force -ErrorAction SilentlyContinue",
                    ],
                    capture_output=True,
                    timeout=5,
                )
        else:
            run(["kill", "-9", str(pid)], capture_output=True, timeout=5)
    except subprocess.TimeoutExpired:
        if is_windows_func():
            with suppress(subprocess.TimeoutExpired, OSError):
                run(
                    [
                        "powershell",
                        "-NoProfile",
                        "-Command",
                        f"Stop-Process -Id {pid} -Force -ErrorAction SilentlyContinue",
                    ],
                    capture_output=True,
                    timeout=5,
                )


def cleanup_repo_local_interface_processes(
    *,
    list_processes: Callable[[], list[ProcessInfo]],
    kill_pid_tree_func: Callable[[int], None],
    sleep: Callable[[float], None],
) -> list[ProcessInfo]:
    try:
        processes = list_processes()
    except RuntimeError as exc:
        if "timed out" not in str(exc).lower():
            print(f"  WARN: unable to inspect repo-local Interface residuals ({exc})")
        return []
    if not processes:
        return []

    print("  Cleaning repo-local Interface residuals...")
    for proc in processes:
        print(f"    - {proc['name']} (PID {proc['pid']})")
        kill_pid_tree_func(int(proc["pid"]))

    sleep(0.5)
    try:
        remaining = list_processes()
    except RuntimeError as exc:
        if "timed out" not in str(exc).lower():
            print(f"  WARN: unable to re-check repo-local Interface residuals ({exc})")
        return []
    if remaining:
        for proc in remaining:
            kill_pid_tree_func(int(proc["pid"]))
        sleep(0.5)
        try:
            remaining = list_processes()
        except RuntimeError:
            return []
    return remaining


def windows_listening_pids_for_port(
    port: int,
    *,
    run: Callable[..., Any] = subprocess.run,
) -> list[int]:
    result = run(
        ["netstat", "-ano", "-p", "tcp"],
        capture_output=True,
        text=True,
        timeout=15,
    )
    if result.returncode != 0:
        return []

    suffix = f":{port}"
    pids: list[int] = []
    for line in result.stdout.splitlines():
        parts = line.split()
        if len(parts) < 5:
            continue
        local_address = parts[1]
        state = parts[3]
        pid_text = parts[4]
        if not local_address.endswith(suffix):
            continue
        if state.upper() != "LISTENING":
            continue
        if pid_text.isdigit():
            pids.append(int(pid_text))
    return pids


def windows_listening_pids_for_port_range(
    port_start: int,
    port_end: int,
    *,
    run: Callable[..., Any] = subprocess.run,
) -> list[int]:
    result = run(
        ["netstat", "-ano", "-p", "tcp"],
        capture_output=True,
        text=True,
        timeout=15,
    )
    if result.returncode != 0:
        return []

    pids: list[int] = []
    for line in result.stdout.splitlines():
        parts = line.split()
        if len(parts) < 5:
            continue
        local_address = parts[1]
        state = parts[3]
        pid_text = parts[4]
        if state.upper() != "LISTENING" or not pid_text.isdigit():
            continue
        port_text = local_address.rsplit(":", 1)[-1]
        if not port_text.isdigit():
            continue
        port = int(port_text)
        if port_start <= port <= port_end:
            pids.append(int(pid_text))
    return pids


def cleanup_managed_interface_listeners(
    port_start: int,
    port_end: int,
    *,
    is_windows_func: Callable[[], bool],
    windows_listening_pids_for_port_range_func: Callable[[int, int], list[int]],
    kill_pid_tree_func: Callable[[int], None],
    sleep: Callable[[float], None],
) -> list[int]:
    if not is_windows_func():
        return []
    pids = sorted(set(windows_listening_pids_for_port_range_func(port_start, port_end)))
    if not pids:
        return []

    print("  Cleaning managed Interface listeners...")
    for pid in pids:
        kill_pid_tree_func(pid)
    sleep(0.5)
    return pids
