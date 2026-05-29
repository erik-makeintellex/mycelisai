from __future__ import annotations

import json
import os
import subprocess
from pathlib import Path

from .config import API_HOST, API_PORT, CORE_DIR, ROOT_DIR, is_windows


def _normalize_process_text(text: str) -> str:
    return str(text or "").replace("\\", "/").lower()


def _api_host_is_local() -> bool:
    return API_HOST.strip().lower() in {"", "localhost", "127.0.0.1", "::1"}


def repo_local_core_listener_active() -> bool:
    if not is_windows() or not _api_host_is_local():
        return False
    try:
        from . import interface_processes

        pids = interface_processes.windows_listening_pids_for_port(API_PORT)
        if not pids:
            return False
        filter_expr = " OR ".join(f"ProcessId = {pid}" for pid in sorted(set(pids)))
        result = subprocess.run(
            [
                "powershell",
                "-NoProfile",
                "-Command",
                f"Get-CimInstance Win32_Process -Filter \"{filter_expr}\" | "
                "Select-Object ProcessId,Name,CommandLine,ExecutablePath | "
                "ConvertTo-Json -Compress",
            ],
            capture_output=True,
            text=True,
            timeout=8,
        )
        if result.returncode != 0 or not result.stdout.strip():
            return False
        payload = json.loads(result.stdout)
        rows = payload if isinstance(payload, list) else [payload]
    except (OSError, subprocess.SubprocessError, json.JSONDecodeError, ValueError):
        return False

    core_hint = _normalize_process_text(str(CORE_DIR))
    for row in rows:
        command = _normalize_process_text(f"{row.get('CommandLine') or ''} {row.get('ExecutablePath') or ''}")
        if core_hint in command and ("core/bin/server" in command or "cmd/server" in command):
            return True
    return False


def infer_native_backend_workspace_root() -> str | None:
    if os.environ.get("PLAYWRIGHT_BACKEND_WORKSPACE_PROBE", "").strip().lower() == "k8s":
        return None
    if os.environ.get("PLAYWRIGHT_BACKEND_WORKSPACE_ROOT", "").strip():
        return None
    if os.environ.get("MYCELIS_BACKEND_WORKSPACE_ROOT", "").strip():
        return None

    workspace = os.environ.get("MYCELIS_WORKSPACE", "").strip()
    if not workspace or not repo_local_core_listener_active():
        return None

    workspace_path = Path(workspace)
    if workspace_path.is_absolute():
        return str(workspace_path) if workspace_path.exists() else None
    return str((ROOT_DIR / "core" / workspace_path).resolve())
