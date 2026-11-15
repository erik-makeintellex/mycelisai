"""Process discovery, port cleanup, and background Core launch helpers for lifecycle tasks."""

from __future__ import annotations

import csv
import json

from . import lifecycle as _lc


def _find_pid_on_port(port: int) -> int | None:
    """Find the PID listening on a given port. Returns None if not found."""
    if _lc.is_windows():
        try:
            result = _lc.subprocess.run(
                ["netstat", "-ano"],
                capture_output=True,
                text=True,
                timeout=5,
            )
            for line in result.stdout.splitlines():
                parts = line.split()
                if len(parts) >= 5 and "LISTENING" in line:
                    local_addr = parts[1]
                    if local_addr.endswith(f":{port}"):
                        pid_str = parts[-1]
                        if pid_str.isdigit():
                            return int(pid_str)
        except Exception:
            pass
    else:
        try:
            result = _lc.subprocess.run(
                ["lsof", "-ti", f":{port}"],
                capture_output=True,
                text=True,
                timeout=5,
            )
            pid_str = result.stdout.strip().split("\n")[0]
            if pid_str.isdigit():
                return int(pid_str)
        except Exception:
            pass
    return None


def _kill_pid(pid: int):
    """Kill a process by PID without failing the workflow on slow OS cleanup."""
    try:
        if _lc.is_windows():
            _lc.subprocess.run(
                ["taskkill", "/F", "/T", "/PID", str(pid)],
                capture_output=True,
                timeout=10,
            )
        else:
            _lc.subprocess.run(
                ["kill", "-9", str(pid)],
                capture_output=True,
                timeout=5,
            )
    except _lc.subprocess.TimeoutExpired:
        if _lc.is_windows():
            _lc._run_best_effort(
                [
                    "powershell",
                    "-NoProfile",
                    "-Command",
                    f"Stop-Process -Id {pid} -Force -ErrorAction SilentlyContinue",
                ],
                timeout=5,
            )


def _run_best_effort(cmd: list[str], timeout: int = 5):
    """Run a cleanup command without letting a hung subprocess block the workflow."""
    try:
        _lc.subprocess.run(cmd, capture_output=True, timeout=timeout)
    except _lc.subprocess.TimeoutExpired:
        pass


def _remaining_managed_services() -> list[str]:
    """Return managed service labels that still have listening ports."""
    return [
        svc["label"]
        for key, svc in _lc.SERVICES.items()
        if key in ("postgres", "nats", "core", "frontend") and _lc._port_open(svc["port"])
    ]


def _matches_compiled_go_service(name: str, command_line: str) -> bool:
    """Return True when a process looks like a repo-local Go service run."""
    _normalized_name = (name or "").lower()
    normalized_cmd = (command_line or "").lower().replace("\\", "/")
    if not normalized_cmd:
        return False
    return any(hint.replace("\\", "/") in normalized_cmd for hint in _lc.COMPILED_GO_PROCESS_HINTS)


def _list_compiled_go_service_processes() -> list[dict[str, str | int]]:
    """Find repo-local Go service processes that may outlive prior test runs."""
    processes: list[dict[str, str | int]] = []
    try:
        if _lc.is_windows():
            go_candidate_pids: list[int] = []
            for image_name in _lc.WINDOWS_COMPILED_GO_PROCESS_NAMES:
                try:
                    result = _lc.subprocess.run(
                        ["tasklist", "/FO", "CSV", "/NH", "/FI", f"IMAGENAME eq {image_name}"],
                        capture_output=True,
                        text=True,
                        timeout=5,
                    )
                except _lc.subprocess.TimeoutExpired:
                    continue
                if result.returncode != 0:
                    raise RuntimeError(result.stderr.strip() or "process query failed")
                rows = [row for row in csv.reader(result.stdout.splitlines()) if len(row) >= 2]
                for row in rows:
                    listed_name = (row[0] or "").strip().lower()
                    pid_str = (row[1] or "").strip()
                    if listed_name != image_name or not pid_str.isdigit():
                        continue
                    pid = int(pid_str)
                    process_name = listed_name.removesuffix(".exe")
                    if listed_name == "go.exe":
                        go_candidate_pids.append(pid)
                        continue
                    processes.append({"pid": pid, "name": process_name, "command": listed_name})

            if not go_candidate_pids:
                return processes

            filter_expr = " OR ".join(f"ProcessId = {pid}" for pid in go_candidate_pids)
            try:
                result = _lc.subprocess.run(
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
                    timeout=8,
                )
                if result.returncode != 0:
                    raise RuntimeError(result.stderr.strip() or "process query failed")
                raw = result.stdout.strip()
                if not raw:
                    return []
                payload = json.loads(raw)
                rows = payload if isinstance(payload, list) else [payload]
                for row in rows:
                    pid = row.get("ProcessId")
                    name = row.get("Name") or ""
                    command_line = row.get("CommandLine") or ""
                    if isinstance(pid, int) and _lc._matches_compiled_go_service(name, command_line):
                        processes.append({"pid": pid, "name": name, "command": command_line})
            except (_lc.subprocess.SubprocessError, json.JSONDecodeError, OSError, ValueError, RuntimeError):
                pass
            return processes

        result = _lc.subprocess.run(
            ["ps", "-eo", "pid=,comm=,args="],
            capture_output=True,
            text=True,
            timeout=8,
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
            if _lc._matches_compiled_go_service(name, command_line):
                processes.append({"pid": pid, "name": name, "command": command_line})
    except (_lc.subprocess.SubprocessError, json.JSONDecodeError, OSError, ValueError, RuntimeError) as exc:
        raise RuntimeError(f"compiled Go service inspection failed: {exc}") from exc
    return processes


def _kill_compiled_go_services() -> list[dict[str, str | int]]:
    """Terminate stray compiled or go-run Mycelis services from prior runs."""
    try:
        processes = _lc._list_compiled_go_service_processes()
    except RuntimeError as exc:
        raise SystemExit(
            "STACK DOWN INCOMPLETE: unable to inspect compiled Go services. "
            + str(exc)
            + ". Resolve the local process-inspection failure before running tests."
        ) from exc
    if not processes:
        print("  No stray compiled Go services detected")
        return []

    print("  Cleaning stray compiled Go services from prior runs...")
    for proc in processes:
        print(f"    - {proc['name']} (PID {proc['pid']})")
        _lc._kill_pid(int(proc["pid"]))

    _lc.time.sleep(0.5)
    remaining = _lc._list_compiled_go_service_processes()
    if remaining:
        for proc in remaining:
            _lc._kill_pid(int(proc["pid"]))
        _lc.time.sleep(0.5)
        remaining = _lc._list_compiled_go_service_processes()
    return remaining


def _service_keys_by_label(labels: list[str]) -> list[str]:
    keys: list[str] = []
    for key, svc in _lc.SERVICES.items():
        if key in ("postgres", "nats", "core", "frontend") and svc["label"] in labels:
            keys.append(key)
    return keys


def _kill_port(port: int, label: str) -> bool:
    """Kill whatever is listening on the given port. Returns True if killed."""
    pid = _lc._find_pid_on_port(port)
    if pid:
        _lc._kill_pid(pid)
        if not _lc._wait_for_port_closed(port, label):
            return False
        print(f"  Killed {label} (PID {pid}) on port {port}")
        return True
    return False


def _start_core_background():
    """Start the core server binary in the background."""
    _lc._load_env()
    bin_path = _lc.CORE_DIR / ("bin/server.exe" if _lc.is_windows() else "bin/server")
    if not bin_path.exists():
        print(f"  ERROR: Binary not found at {bin_path}. Run 'uv run inv core.compile' first.")
        return False

    env = {**_lc.os.environ, "PYTHONIOENCODING": "utf-8"}
    log_path = _lc._core_startup_log_path()
    log_path.parent.mkdir(parents=True, exist_ok=True)

    with log_path.open("w", encoding="utf-8") as log_file:
        if _lc.is_windows():
            _lc.subprocess.Popen(
                [str(bin_path)],
                cwd=str(_lc.CORE_DIR),
                env=env,
                stdin=_lc.subprocess.DEVNULL,
                stdout=log_file,
                stderr=_lc.subprocess.STDOUT,
                creationflags=_lc.subprocess.CREATE_NEW_PROCESS_GROUP | _lc.subprocess.DETACHED_PROCESS,
            )
        else:
            _lc.subprocess.Popen(
                [str(bin_path)],
                cwd=str(_lc.CORE_DIR),
                env=env,
                stdin=_lc.subprocess.DEVNULL,
                stdout=log_file,
                stderr=_lc.subprocess.STDOUT,
                start_new_session=True,
            )
    return True
