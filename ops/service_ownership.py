from __future__ import annotations

import subprocess
from typing import Callable

from . import interface_processes, process_inspection


def pid_matches_compiled_go_service(
    pid: int,
    *,
    is_windows_func: Callable[[], bool],
    matches_compiled_go_service_func: Callable[[str, str], bool],
    run=subprocess.run,
) -> bool:
    try:
        process = process_inspection.process_info_for_pid(pid, is_windows_func=is_windows_func, run=run)
    except RuntimeError:
        return False
    if not process:
        return False
    return matches_compiled_go_service_func(str(process["name"]), str(process["command"]))


def owned_core_pid_on_port(
    port: int,
    *,
    find_pid_on_port_func: Callable[[int], int | None],
    is_windows_func: Callable[[], bool],
    matches_compiled_go_service_func: Callable[[str, str], bool],
    run=subprocess.run,
) -> int | None:
    pid = find_pid_on_port_func(port)
    if not pid:
        return None
    if pid_matches_compiled_go_service(
        pid,
        is_windows_func=is_windows_func,
        matches_compiled_go_service_func=matches_compiled_go_service_func,
        run=run,
    ):
        return pid
    return None


def repo_local_interface_processes_for_pids(
    pids: list[int],
    *,
    is_windows_func: Callable[[], bool],
    normalize_process_text: Callable[[str], str],
    run=subprocess.run,
) -> list[dict[str, str | int]]:
    return interface_processes.repo_local_interface_processes_for_pids(
        pids,
        is_windows_func=is_windows_func,
        normalize_process_text=normalize_process_text,
        run=run,
    )


def owned_frontend_pid_on_port(
    port: int,
    *,
    is_windows_func: Callable[[], bool],
    normalize_process_text: Callable[[str], str],
    run=subprocess.run,
) -> int | None:
    pids: list[int] = []
    if is_windows_func():
        pids = interface_processes.windows_listening_pids_for_port(port, run=run)
    else:
        try:
            result = run(["lsof", "-ti", f":{port}"], capture_output=True, text=True, timeout=5)
            pids = [int(line) for line in result.stdout.splitlines() if line.strip().isdigit()]
        except (subprocess.SubprocessError, OSError, ValueError):
            pids = []
    processes = repo_local_interface_processes_for_pids(
        pids,
        is_windows_func=is_windows_func,
        normalize_process_text=normalize_process_text,
        run=run,
    )
    return int(processes[0]["pid"]) if processes else None
