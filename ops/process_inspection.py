from __future__ import annotations

import json
import subprocess
from typing import Any, Callable, TypedDict


WINDOWS_PROCESS_QUERY_TIMEOUT_SECONDS = 180


class ProcessInfo(TypedDict):
    pid: int
    name: str
    command: str


def list_processes_by_pids(
    pids: list[int],
    *,
    is_windows_func: Callable[[], bool],
    run: Callable[..., Any] = subprocess.run,
) -> list[ProcessInfo]:
    """Return process details for supplied PIDs without broad ownership claims."""
    unique_pids = sorted({int(pid) for pid in pids if int(pid) > 0})
    if not unique_pids:
        return []
    try:
        if is_windows_func():
            return _list_windows_processes_by_pids(unique_pids, run=run)
        return _list_posix_processes_by_pids(unique_pids, run=run)
    except (subprocess.SubprocessError, json.JSONDecodeError, OSError, ValueError, RuntimeError) as exc:
        raise RuntimeError(f"process inspection failed: {exc}") from exc


def process_info_for_pid(
    pid: int,
    *,
    is_windows_func: Callable[[], bool],
    run: Callable[..., Any] = subprocess.run,
) -> ProcessInfo | None:
    processes = list_processes_by_pids([pid], is_windows_func=is_windows_func, run=run)
    return processes[0] if processes else None


def _list_windows_processes_by_pids(
    pids: list[int],
    *,
    run: Callable[..., Any],
) -> list[ProcessInfo]:
    filter_expr = " OR ".join(f"ProcessId = {pid}" for pid in pids)
    result = run(
        [
            "powershell",
            "-NoProfile",
            "-Command",
            f"Get-CimInstance Win32_Process -Filter \"{filter_expr}\" | "
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
    return [
        {"pid": row["ProcessId"], "name": row.get("Name") or "", "command": row.get("CommandLine") or ""}
        for row in rows
        if isinstance(row.get("ProcessId"), int)
    ]


def _list_posix_processes_by_pids(
    pids: list[int],
    *,
    run: Callable[..., Any],
) -> list[ProcessInfo]:
    result = run(
        ["ps", "-p", ",".join(str(pid) for pid in pids), "-o", "pid=,comm=,args="],
        capture_output=True,
        text=True,
        timeout=10,
    )
    if result.returncode != 0:
        return []
    processes: list[ProcessInfo] = []
    for line in result.stdout.splitlines():
        parts = line.strip().split(None, 2)
        if len(parts) >= 2 and parts[0].isdigit():
            processes.append({"pid": int(parts[0]), "name": parts[1], "command": parts[2] if len(parts) > 2 else ""})
    return processes
