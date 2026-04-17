import csv
import ipaddress
import json
import os
import re
import shlex
import shutil
import subprocess
import time
from dataclasses import dataclass
import urllib.error
import urllib.request
import socket
from contextlib import suppress
from pathlib import Path

from invoke import task, Collection
from .config import (
    INTERFACE_BIND_HOST,
    INTERFACE_HOST,
    INTERFACE_PORT,
    ROOT_DIR,
    ensure_managed_cache_dirs,
    is_windows,
    managed_cache_env,
    powershell,
)
from .cache import ensure_disk_headroom

ns = Collection("interface")
INTERFACE_DIR = ROOT_DIR / "interface"
_INTERFACE_PROCESS_PATH_HINTS = tuple(
    str(path.resolve()).lower().replace("\\", "/")
    for path in {
        INTERFACE_DIR / ".next",
        INTERFACE_DIR / "node_modules",
        INTERFACE_DIR / "scripts",
        INTERFACE_DIR / "playwright-report",
        INTERFACE_DIR / "test-results",
    }
)
_INTERFACE_PROCESS_COMMAND_HINTS = (
    "/.next/dev/build/postcss.js",
    "/.next/standalone/server.js",
    "/next/dist/bin/next",
    "/next/dist/server/lib/start-server.js",
    "/dist/server/lib/start-server.js",
    "./node_modules/next/dist/bin/next",
    "/scripts/playwright-webserver.mjs",
    "./scripts/playwright-webserver.mjs",
    "/node_modules/.bin/vitest",
    "./node_modules/.bin/vitest",
    "/node_modules/vitest/",
    "/vitest/vitest.mjs",
    "/playwright/",
)

def _load_env():
    """Load root .env into the process environment so Next.js proxy
    can read MYCELIS_API_KEY (used to inject Authorization headers into
    proxied /api/* requests). Uses override=True so .env wins over system env.
    Removes PORT afterwards — the root .env sets PORT for the Go Core HTTP
    listener, but Next.js would otherwise try to reuse that port instead of
    the configured interface port."""
    import os
    try:
        from dotenv import load_dotenv
        load_dotenv(str(ROOT_DIR / ".env"), override=True)
    except ImportError:
        pass  # python-dotenv not installed — env vars must be set manually
    # Don't let Go Core's PORT leak into Next.js
    os.environ.pop("PORT", None)


def _task_env(extra=None):
    _load_env()
    ensure_managed_cache_dirs()
    return managed_cache_env(extra=extra)


@dataclass
class CommandResult:
    exited: int
    stdout: str = ""
    stderr: str = ""


def _is_next_build_lock_conflict(result: CommandResult) -> bool:
    text = _normalize_process_text(f"{result.stdout}\n{result.stderr}")
    return "unable to acquire lock at" in text and ".next/lock" in text


def _is_incomplete_next_build_output(result: CommandResult) -> bool:
    text = _normalize_process_text(f"{result.stdout}\n{result.stderr}")
    if ".next/" not in text:
        return False
    incomplete_outputs = (
        "required-server-files.json",
        "build-manifest.json",
        "pages-manifest.json",
        ".nft.json",
    )
    return "enoent" in text and any(name in text for name in incomplete_outputs)


def _is_next_standalone_cleanup_conflict(result: CommandResult) -> bool:
    text = _normalize_process_text(f"{result.stdout}\n{result.stderr}")
    return (
        "ebusy" in text
        and "rmdir" in text
        and ".next/standalone" in text
    )


def _expected_next_build_artifacts() -> list[Path]:
    next_dir = INTERFACE_DIR / ".next"
    build_manifest_path = next_dir / "build-manifest.json"
    required_server_files_path = next_dir / "required-server-files.json"
    artifacts = [build_manifest_path, required_server_files_path]
    if not build_manifest_path.exists():
        return artifacts

    payload = json.loads(build_manifest_path.read_text(encoding="utf-8"))
    manifest_entries: list[str] = []
    manifest_entries.extend(payload.get("polyfillFiles", []))
    manifest_entries.extend(payload.get("lowPriorityFiles", []))
    manifest_entries.extend(payload.get("rootMainFiles", []))
    for page_entries in payload.get("pages", {}).values():
        manifest_entries.extend(page_entries or [])

    seen: set[Path] = set()
    for relative_path in manifest_entries:
        artifact_path = next_dir / str(relative_path).replace("/", os.sep)
        if artifact_path not in seen:
            artifacts.append(artifact_path)
            seen.add(artifact_path)

    return artifacts


def _wait_for_complete_next_build_output(timeout_seconds: int = 20) -> None:
    deadline = time.time() + timeout_seconds
    missing: list[Path] = []
    while time.time() < deadline:
        missing = [path for path in _expected_next_build_artifacts() if not path.exists()]
        if not missing:
            return
        time.sleep(0.5)

    missing_preview = ", ".join(str(path.relative_to(INTERFACE_DIR)) for path in missing[:6])
    raise RuntimeError(f"Incomplete Next.js build output after successful build command: {missing_preview}")


def interface_task_env(extra=None):
    """Public wrapper for callers that need the managed Interface task env."""
    return _task_env(extra=extra)


def _run_interface_shell_command(command: list[str], extra_env: dict[str, str] | None = None) -> CommandResult:
    """Run a one-shot Interface command directly and return its exit data.

    Invoke's runner is useful for long-lived dev/e2e flows, but the one-shot
    build/test commands have been observed to return false negatives under the
    wrapper on Windows. Running them directly keeps the exit code aligned with
    the real tool result.
    """
    process_env = os.environ.copy()
    process_env.update(_task_env(extra_env))
    runner = list(command)
    executable = runner[0]
    if is_windows():
        resolved = shutil.which(executable) or shutil.which(f"{executable}.cmd") or shutil.which(f"{executable}.exe")
        if resolved:
            runner[0] = resolved
    else:
        resolved = shutil.which(executable)
        if resolved:
            runner[0] = resolved
    result = subprocess.run(
        runner,
        cwd=str(INTERFACE_DIR),
        env=process_env,
        capture_output=True,
        text=True,
        encoding="utf-8",
        errors="replace",
    )
    return CommandResult(exited=result.returncode, stdout=result.stdout or "", stderr=result.stderr or "")


def _run_interface_shell_command_streaming(command: list[str], extra_env: dict[str, str] | None = None) -> CommandResult:
    """Run an Interface command with inherited stdio for long-lived browser flows.

    Browser/tooling subprocess trees on Windows can keep captured stdout/stderr
    pipes open after the main runner finishes. Streaming avoids that pipe-lifetime
    hang while preserving the real exit code for Invoke tasks.
    """
    process_env = os.environ.copy()
    process_env.update(_task_env(extra_env))
    runner = list(command)
    executable = runner[0]
    if is_windows():
        resolved = shutil.which(executable) or shutil.which(f"{executable}.cmd") or shutil.which(f"{executable}.exe")
        if resolved:
            runner[0] = resolved
    else:
        resolved = shutil.which(executable)
        if resolved:
            runner[0] = resolved
    result = subprocess.run(
        runner,
        cwd=str(INTERFACE_DIR),
        env=process_env,
        text=True,
        encoding="utf-8",
        errors="replace",
    )
    return CommandResult(exited=result.returncode, stdout="", stderr="")


def _report_command_result(result: CommandResult) -> None:
    _print_ascii_safe(result.stdout)
    _print_ascii_safe(result.stderr)


def _run_one_shot_interface_task(
    command: list[str],
    *,
    extra_env: dict[str, str] | None = None,
    cleanup: bool = True,
) -> CommandResult:
    if cleanup:
        _cleanup_repo_local_interface_processes()
    try:
        result = _run_interface_shell_command(command, extra_env=extra_env)
        _report_command_result(result)
        if result.exited != 0:
            raise SystemExit(result.exited)
        return result
    finally:
        if cleanup:
            _cleanup_repo_local_interface_processes()


def _run_interface_commandline(
    command: str,
    *,
    extra_env: dict[str, str] | None = None,
    stream: bool = False,
) -> CommandResult:
    args = shlex.split(command, posix=not is_windows())
    if stream:
        if len(args) >= 3 and args[:3] == ["npx", "playwright", "test"]:
            return _run_playwright_command_streaming(args, extra_env=extra_env)
        return _run_interface_shell_command_streaming(args, extra_env=extra_env)
    return _run_interface_shell_command(args, extra_env=extra_env)


def _build_playwright_command(
    *,
    project: str = "",
    spec: str = "",
    workers: str = "",
    headed: bool = False,
) -> str:
    cmd = "npx playwright test --reporter=dot"
    effective_workers = workers or "1"
    if project:
        cmd += f" --project={project}"
    if spec:
        cmd += f" {spec}"
    if effective_workers:
        cmd += f" --workers={effective_workers}"
    if headed:
        cmd += " --headed"
    return cmd


def _build_playwright_env(*, live_backend: bool, port: int) -> dict[str, str]:
    extra_env = {
        "PLAYWRIGHT_SKIP_WEBSERVER": "1",
        "INTERFACE_HOST": "127.0.0.1",
        "INTERFACE_BIND_HOST": INTERFACE_BIND_HOST,
        "INTERFACE_PORT": str(port),
    }
    if live_backend:
        extra_env["PLAYWRIGHT_LIVE_BACKEND"] = "1"
    return _task_env(extra_env)


def _reconcile_managed_server_endpoint(
    env: dict[str, str],
    chosen_port: int,
    server: subprocess.Popen[str],
) -> tuple[dict[str, str], int]:
    actual_port = _detect_playwright_server_port(chosen_port)
    if actual_port != chosen_port:
        print(f"  Managed server bound to port {actual_port} instead of requested {chosen_port}")
        chosen_port = actual_port
        env["INTERFACE_PORT"] = str(actual_port)

    ready_host = _wait_for_interface_ready(host=env["INTERFACE_HOST"], port=chosen_port, process=server)
    if ready_host != env["INTERFACE_HOST"]:
        print(f"  Managed server is reachable via {ready_host}; updating Playwright host from {env['INTERFACE_HOST']}")
        env["INTERFACE_HOST"] = ready_host

    return env, chosen_port


def run_interface_command(c, command: str, cleanup=False, extra_env=None, **run_kwargs):
    """Run an Interface-local command from the interface/ working directory."""
    command_env = run_kwargs.pop("env", None)
    if command_env is not None and extra_env is None:
        env = dict(command_env)
    else:
        env = _task_env(extra=extra_env)
        if command_env:
            env.update(command_env)
    try:
        if cleanup:
            _cleanup_repo_local_interface_processes()
        with c.cd(str(INTERFACE_DIR)):
            return c.run(command, env=env, **run_kwargs)
    finally:
        if cleanup:
            _cleanup_repo_local_interface_processes()


def _normalize_process_text(text: str) -> str:
    return (text or "").lower().replace("\\", "/")


def _matches_repo_local_interface_process(name: str, command_line: str) -> bool:
    normalized_name = (name or "").lower()
    normalized_cmd = _normalize_process_text(command_line)
    if normalized_name not in {"node", "node.exe", "cmd", "cmd.exe"} or not normalized_cmd:
        return False
    if not any(hint in normalized_cmd for hint in _INTERFACE_PROCESS_PATH_HINTS):
        return False
    return any(hint in normalized_cmd for hint in _INTERFACE_PROCESS_COMMAND_HINTS)


def _list_repo_local_interface_processes() -> list[dict[str, str | int]]:
    processes: list[dict[str, str | int]] = []
    try:
        if is_windows():
            candidate_pids: list[int] = []
            for image_name in ("node.exe", "cmd.exe"):
                tasklist_result = subprocess.run(
                    ["tasklist", "/FO", "CSV", "/NH", "/FI", f"IMAGENAME eq {image_name}"],
                    capture_output=True,
                    text=True,
                    timeout=20,
                )
                if tasklist_result.returncode != 0:
                    raise RuntimeError(tasklist_result.stderr.strip() or "process query failed")
                for row in csv.reader(tasklist_result.stdout.splitlines()):
                    if len(row) < 2:
                        continue
                    listed_name = (row[0] or "").strip().lower()
                    pid_text = (row[1] or "").strip()
                    if listed_name != image_name or not pid_text.isdigit():
                        continue
                    candidate_pids.append(int(pid_text))
            if not candidate_pids:
                return []

            deadline = time.monotonic() + 30
            for start in range(0, len(candidate_pids), 12):
                remaining_seconds = deadline - time.monotonic()
                if remaining_seconds <= 0:
                    raise RuntimeError("process query timed out")
                pid_batch = candidate_pids[start:start + 12]
                filter_expr = " OR ".join(f"ProcessId = {pid}" for pid in pid_batch)
                result = subprocess.run(
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
                    timeout=max(1, min(8, int(remaining_seconds))),
                )
                if result.returncode != 0:
                    raise RuntimeError(result.stderr.strip() or "process query failed")
                raw = result.stdout.strip()
                if not raw:
                    continue
                payload = json.loads(raw)
                rows = payload if isinstance(payload, list) else [payload]
                for row in rows:
                    pid_text = row.get("ProcessId")
                    name = row.get("Name") or ""
                    command_line = row.get("CommandLine") or ""
                    if not isinstance(pid_text, int):
                        continue
                    pid = pid_text
                    if _matches_repo_local_interface_process(name, command_line):
                        processes.append({"pid": pid, "name": name, "command": command_line})
            return processes

        result = subprocess.run(
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
            if _matches_repo_local_interface_process(name, command_line):
                processes.append({"pid": pid, "name": name, "command": command_line})
    except (subprocess.SubprocessError, json.JSONDecodeError, OSError, ValueError, RuntimeError) as exc:
        raise RuntimeError(f"interface process inspection failed: {exc}") from exc
    return processes


def _kill_pid_tree(pid: int):
    try:
        if is_windows():
            subprocess.run(
                ["taskkill", "/F", "/T", "/PID", str(pid)],
                capture_output=True,
                timeout=12,
            )
            with suppress(subprocess.SubprocessError, OSError):
                subprocess.run(
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
            subprocess.run(["kill", "-9", str(pid)], capture_output=True, timeout=5)
    except subprocess.TimeoutExpired:
        if is_windows():
            with suppress(subprocess.TimeoutExpired, OSError):
                subprocess.run(
                    ["powershell", "-NoProfile", "-Command", f"Stop-Process -Id {pid} -Force -ErrorAction SilentlyContinue"],
                    capture_output=True,
                    timeout=5,
                )


def _cleanup_repo_local_interface_processes() -> list[dict[str, str | int]]:
    try:
        processes = _list_repo_local_interface_processes()
    except RuntimeError as exc:
        if "timed out" not in str(exc).lower():
            print(f"  WARN: unable to inspect repo-local Interface residuals ({exc})")
        return []
    if not processes:
        return []

    print("  Cleaning repo-local Interface residuals...")
    for proc in processes:
        print(f"    - {proc['name']} (PID {proc['pid']})")
        _kill_pid_tree(int(proc["pid"]))

    time.sleep(0.5)
    try:
        remaining = _list_repo_local_interface_processes()
    except RuntimeError as exc:
        if "timed out" not in str(exc).lower():
            print(f"  WARN: unable to re-check repo-local Interface residuals ({exc})")
        return []
    if remaining:
        for proc in remaining:
            _kill_pid_tree(int(proc["pid"]))
        time.sleep(0.5)
        try:
            remaining = _list_repo_local_interface_processes()
        except RuntimeError:
            return []
    return remaining


def _windows_listening_pids_for_port(port: int) -> list[int]:
    result = subprocess.run(
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


def _windows_listening_pids_for_port_range(port_start: int, port_end: int) -> list[int]:
    result = subprocess.run(
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


def _cleanup_managed_interface_listeners(port_start: int = 3100, port_end: int = 3199) -> list[int]:
    if not is_windows():
        return []
    pids = sorted(set(_windows_listening_pids_for_port_range(port_start, port_end)))
    if not pids:
        return []

    print("  Cleaning managed Interface listeners...")
    for pid in pids:
        _kill_pid_tree(pid)
    time.sleep(0.5)
    return pids


def _playwright_server_log_path() -> str:
    log_dir = ROOT_DIR / "workspace" / "logs"
    log_dir.mkdir(parents=True, exist_ok=True)
    return str(log_dir / "interface-playwright-webserver.log")


def _playwright_last_run_path() -> Path:
    return INTERFACE_DIR / "test-results" / ".last-run.json"


def _read_playwright_last_run_status() -> str | None:
    last_run_path = _playwright_last_run_path()
    if not last_run_path.exists():
        return None
    try:
        payload = json.loads(last_run_path.read_text(encoding="utf-8"))
    except (OSError, json.JSONDecodeError):
        return None
    status = payload.get("status")
    return status if isinstance(status, str) else None


def _stop_lingering_playwright_process(proc: subprocess.Popen[str]) -> None:
    with suppress(ProcessLookupError, OSError):
        proc.terminate()
    with suppress(subprocess.TimeoutExpired, OSError):
        proc.wait(timeout=3)
    if proc.poll() is not None:
        return
    _kill_pid_tree(proc.pid)
    with suppress(subprocess.TimeoutExpired, OSError):
        proc.wait(timeout=3)


def _run_playwright_command_streaming(
    command: list[str],
    *,
    extra_env: dict[str, str] | None = None,
    post_result_grace_seconds: float = 15.0,
) -> CommandResult:
    last_run_path = _playwright_last_run_path()
    with suppress(FileNotFoundError):
        last_run_path.unlink()

    process_env = os.environ.copy()
    process_env.update(_task_env(extra_env))
    runner = list(command)
    executable = runner[0]
    if is_windows():
        resolved = shutil.which(executable) or shutil.which(f"{executable}.cmd") or shutil.which(f"{executable}.exe")
        if resolved:
            runner[0] = resolved
    else:
        resolved = shutil.which(executable)
        if resolved:
            runner[0] = resolved

    proc = subprocess.Popen(
        runner,
        cwd=str(INTERFACE_DIR),
        env=process_env,
        text=True,
    )
    result_seen_at: float | None = None
    result_status: str | None = None
    while True:
        exit_code = proc.poll()
        if exit_code is not None:
            return CommandResult(exited=exit_code, stdout="", stderr="")

        current_status = _read_playwright_last_run_status()
        if current_status in {"passed", "failed"}:
            if current_status != result_status:
                result_status = current_status
                result_seen_at = time.time()
            elif result_seen_at is not None and (time.time() - result_seen_at) >= post_result_grace_seconds:
                print(f"  WARN: Playwright reported {current_status} but did not exit cleanly; terminating the lingering test process.")
                _stop_lingering_playwright_process(proc)
                return CommandResult(exited=0 if current_status == "passed" else 1, stdout="", stderr="")
        else:
            result_seen_at = None
            result_status = None

        time.sleep(1.0)


def _next_dev_lock_path() -> Path:
    return INTERFACE_DIR / ".next" / "dev" / "lock"


def _cleanup_stale_next_dev_lock() -> bool:
    lock_path = _next_dev_lock_path()
    if not lock_path.exists():
        return False
    processes: list[dict[str, str | int]] = []
    try:
        processes = _list_repo_local_interface_processes()
    except RuntimeError as exc:
        if "timed out" not in str(exc).lower():
            print(f"  WARN: unable to inspect repo-local Interface workers before clearing Next dev lock ({exc})")
        processes = []
    if processes:
        return False
    try:
        lock_path.unlink()
        print("  Cleared stale Next.js dev lock before managed Interface startup.")
        return True
    except FileNotFoundError:
        return False
    except PermissionError as exc:
        print(f"  WARN: unable to remove stale Next.js dev lock ({exc})")
        return False


def _cleanup_playwright_server_log(max_attempts: int = 20, retry_delay_seconds: float = 0.25) -> None:
    log_path = Path(_playwright_server_log_path())
    for attempt in range(max_attempts):
        try:
            log_path.unlink()
            return
        except FileNotFoundError:
            return
        except PermissionError:
            if attempt == max_attempts - 1:
                return
            time.sleep(retry_delay_seconds)


def _detect_playwright_server_port(expected_port: int, timeout_seconds: int = 30) -> int:
    """Read the managed server log until Next reports its actual local port.

    Next should honor the requested port, but on Windows or after a stale
    listener collision it can surface a different bound port in the startup log.
    The e2e runner uses the log-discovered value so Playwright and the managed
    server always agree on the real base URL.
    """
    log_path = Path(_playwright_server_log_path())
    deadline = time.time() + timeout_seconds
    patterns = (
        re.compile(r"Local:\s+http://\[(?:::1|::)\]:(\d+)"),
        re.compile(r"Local:\s+http://127\.0\.0\.1:(\d+)"),
        re.compile(r"Local:\s+http://localhost:(\d+)"),
    )
    while time.time() < deadline:
        if log_path.exists():
            try:
                text = log_path.read_text(encoding="utf-8", errors="replace")
            except OSError:
                text = ""
            for line in reversed(text.splitlines()):
                for pattern in patterns:
                    match = pattern.search(line)
                    if match:
                        return int(match.group(1))
        time.sleep(0.2)
    return expected_port


def _pick_interface_port(preferred: int = INTERFACE_PORT) -> int:
    """Return a free port for the managed Playwright server.

    Managed browser runs should avoid colliding with the user's normal local UI
    port, so the task prefers a safe managed range outside Windows' typical
    dynamic client-port band unless a non-default port was explicitly requested.
    """
    def _port_has_listener(port: int) -> bool:
        families = [socket.AF_INET]
        if socket.has_ipv6:
            families.append(socket.AF_INET6)
        for family in families:
            try:
                with socket.socket(family, socket.SOCK_STREAM) as probe:
                    probe.settimeout(0.2)
                    host = "127.0.0.1" if family == socket.AF_INET else "::1"
                    if probe.connect_ex((host, port)) == 0:
                        return True
            except OSError:
                continue
        return False

    def _port_is_available(port: int) -> bool:
        bind_host = (INTERFACE_BIND_HOST or "").strip()
        probe_host = bind_host or "127.0.0.1"
        family = socket.AF_INET
        if probe_host == "localhost":
            probe_host = "127.0.0.1"
        else:
            try:
                parsed_host = ipaddress.ip_address(probe_host)
            except ValueError:
                if ":" in probe_host:
                    family = socket.AF_INET6
                else:
                    probe_host = "127.0.0.1"
            else:
                if parsed_host.version == 6:
                    family = socket.AF_INET6
        with socket.socket(family, socket.SOCK_STREAM) as probe_sock:
            if family == socket.AF_INET6:
                with suppress(OSError, AttributeError):
                    probe_sock.setsockopt(
                        socket.IPPROTO_IPV6,
                        socket.IPV6_V6ONLY,
                        0 if probe_host == "::" else 1,
                    )
            probe_sock.bind((probe_host, port))
            return True

    if preferred not in (0, 3000):
        try:
            if _port_has_listener(preferred):
                raise OSError("listener already active")
            if _port_is_available(preferred):
                return preferred
        except OSError:
            pass

    for candidate in range(3100, 3200):
        try:
            if _port_has_listener(candidate):
                continue
            if _port_is_available(candidate):
                return candidate
        except OSError:
            continue

    for _ in range(32):
        with socket.socket(socket.AF_INET, socket.SOCK_STREAM) as ipv4_sock:
            ipv4_sock.bind(("127.0.0.1", 0))
            candidate = ipv4_sock.getsockname()[1]
            try:
                if _port_has_listener(candidate):
                    continue
                with socket.socket(socket.AF_INET6, socket.SOCK_STREAM) as ipv6_loopback_sock:
                    ipv6_loopback_sock.bind(("::1", candidate))
                    with socket.socket(socket.AF_INET6, socket.SOCK_STREAM) as ipv6_wildcard_sock:
                        ipv6_wildcard_sock.bind(("::", candidate))
                        return candidate
            except OSError:
                continue

    raise RuntimeError("Unable to find a free Interface port for the managed Playwright server.")


def _next_dev_command(bind_host: str, port: int) -> list[str]:
    return [
        "node",
        str((INTERFACE_DIR / "node_modules" / "next" / "dist" / "bin" / "next").resolve()),
        "dev",
        "--hostname",
        bind_host,
        "--port",
        str(port),
    ]


def _next_start_command(bind_host: str, port: int) -> list[str]:
    del bind_host, port
    return ["node", str((INTERFACE_DIR / "scripts" / "playwright-webserver.mjs").resolve())]


def _spawn_interface_process(
    command: list[str],
    env: dict[str, str],
    *,
    stdout,
    stderr,
    detached: bool,
    text: bool = True,
) -> subprocess.Popen[str]:
    process_env = os.environ.copy()
    process_env.update(env)
    popen_kwargs = {
        "cwd": str(INTERFACE_DIR),
        "env": process_env,
        "stdout": stdout,
        "stderr": stderr,
        "text": text,
    }
    if detached:
        popen_kwargs["stdin"] = subprocess.DEVNULL
        if is_windows():
            popen_kwargs["creationflags"] = (
                getattr(subprocess, "CREATE_NEW_PROCESS_GROUP", 0)
                | getattr(subprocess, "DETACHED_PROCESS", 0)
            )
        else:
            popen_kwargs["start_new_session"] = True
    return subprocess.Popen(command, **popen_kwargs)


def _start_playwright_server(
    env: dict[str, str],
    port: int = INTERFACE_PORT,
    server_mode: str = "dev",
) -> subprocess.Popen[str]:
    log_path = _playwright_server_log_path()
    log_handle = open(log_path, "w", encoding="utf-8")
    bind_host = env.get("INTERFACE_BIND_HOST", INTERFACE_BIND_HOST)
    command_env = dict(env)
    if server_mode == "start":
        command_env["PORT"] = str(port)
        command_env["HOSTNAME"] = bind_host
        command = _next_start_command(bind_host, port)
    else:
        _cleanup_stale_next_dev_lock()
        command = _next_dev_command(bind_host, port)
    try:
        process = _spawn_interface_process(
            command,
            command_env,
            stdout=log_handle,
            stderr=subprocess.STDOUT,
            detached=False,
        )
        log_handle.close()
        return process
    except Exception:
        log_handle.close()
        raise


def start_dev_server_detached(
    env: dict[str, str] | None = None,
    *,
    host: str = INTERFACE_BIND_HOST,
    port: int = INTERFACE_PORT,
) -> subprocess.Popen[str]:
    process_env = _task_env()
    if env:
        process_env.update(env)
    process_env.setdefault("INTERFACE_HOST", INTERFACE_HOST)
    process_env.setdefault("INTERFACE_BIND_HOST", host)
    _cleanup_stale_next_dev_lock()
    command = _next_dev_command(process_env["INTERFACE_BIND_HOST"], port)
    return _spawn_interface_process(
        command,
        process_env,
        stdout=subprocess.DEVNULL,
        stderr=subprocess.STDOUT,
        detached=True,
        text=True,
    )


def _interface_ready_urls(host: str, port: int) -> list[str]:
    candidates = [host, "127.0.0.1", "localhost", "::1"]
    urls: list[str] = []
    seen: set[str] = set()
    for candidate in candidates:
        if not candidate:
            continue
        url_host = candidate
        if ":" in candidate and not candidate.startswith("["):
            url_host = f"[{candidate}]"
        url = f"http://{url_host}:{port}"
        if url in seen:
            continue
        seen.add(url)
        urls.append(url)
    return urls


def _wait_for_interface_ready(
    host: str = INTERFACE_HOST,
    port: int = INTERFACE_PORT,
    timeout_seconds: int = 120,
    process: subprocess.Popen[str] | None = None,
) -> str:
    urls = _interface_ready_urls(host, port)
    deadline = time.time() + timeout_seconds
    last_error = "server did not respond"
    process_exited_early = False
    while time.time() < deadline:
        if process is not None and process.poll() is not None:
            process_exited_early = True
        for url in urls:
            try:
                with urllib.request.urlopen(url, timeout=5) as response:
                    if response.status < 500:
                        return url.split("://", 1)[1].rsplit(":", 1)[0].strip("[]")
                    last_error = f"{url}: unexpected status {response.status}"
            except urllib.error.HTTPError as exc:
                if exc.code < 500:
                    return url.split("://", 1)[1].rsplit(":", 1)[0].strip("[]")
                last_error = f"{url}: http {exc.code}"
            except Exception as exc:  # pragma: no cover - exercised via timeout path in task flow
                last_error = f"{url}: {exc}"
        time.sleep(1)

    if process_exited_early:
        raise RuntimeError(
            f"Managed Interface server exited before it became ready on port {port}. "
            f"See {_playwright_server_log_path()} for server output."
        )

    raise RuntimeError(
        f"Interface did not become ready at any of {', '.join(urls)} within {timeout_seconds}s. "
        f"See {_playwright_server_log_path()} for server output. Last error: {last_error}"
    )


def _print_ascii_safe(text: str):
    if not text:
        return
    print(text.encode("ascii", "replace").decode("ascii"), end="")

# ── Lifecycle ────────────────────────────────────────────────

@task
def dev(c):
    """Start Interface (Next.js) in Dev Mode. Stops existing instance first."""
    stop(c)
    run_interface_command(c, "npm run dev", pty=not is_windows())

@task
def install(c):
    """Install Interface dependencies."""
    print("Installing Interface Dependencies...")
    run_interface_command(c, "npm install")

@task
def build(c):
    """Build the Interface for production."""
    print("Building Interface...")
    ensure_disk_headroom(min_free_gb=8, reason="interface build")
    stop(c)
    clean(c)
    _cleanup_repo_local_interface_processes()
    build_succeeded = False
    try:
        result = _run_interface_shell_command(["npm", "run", "build"])
        _report_command_result(result)
        if result.exited != 0 and _is_next_build_lock_conflict(result):
            print("Detected a stale Next.js build lock. Cleaning repo-local Interface workers and retrying once...")
            _cleanup_repo_local_interface_processes()
            clean(c)
            _cleanup_repo_local_interface_processes()
            result = _run_interface_shell_command(["npm", "run", "build"])
            _report_command_result(result)
        if result.exited != 0 and _is_next_standalone_cleanup_conflict(result):
            print("Detected a stale Next.js standalone cleanup lock. Cleaning repo-local Interface workers and retrying once...")
            _cleanup_repo_local_interface_processes()
            _cleanup_managed_interface_listeners()
            clean(c)
            _cleanup_repo_local_interface_processes()
            _cleanup_managed_interface_listeners()
            result = _run_interface_shell_command(["npm", "run", "build"])
            _report_command_result(result)
        if result.exited != 0 and _is_incomplete_next_build_output(result):
            print("Detected incomplete built-server output during Next.js packaging. Cleaning repo-local Interface workers and retrying once...")
            _cleanup_repo_local_interface_processes()
            clean(c)
            _cleanup_repo_local_interface_processes()
            result = _run_interface_shell_command(["npm", "run", "build"])
            _report_command_result(result)
        if result.exited != 0:
            raise SystemExit(result.exited)
        try:
            _wait_for_complete_next_build_output()
        except RuntimeError:
            print("Detected incomplete built-server output after the build command exited. Cleaning and retrying once...")
            _cleanup_repo_local_interface_processes()
            clean(c)
            _cleanup_repo_local_interface_processes()
            result = _run_interface_shell_command(["npm", "run", "build"])
            _report_command_result(result)
            if result.exited != 0:
                raise SystemExit(result.exited)
            _wait_for_complete_next_build_output()
        build_succeeded = True
    finally:
        # On Windows, successful Next builds can still finalize the static asset
        # tree after the main command exits. Sweeping repo-local node helpers
        # immediately here can leave a partial `.next/static` tree behind.
        if not build_succeeded:
            _cleanup_repo_local_interface_processes()

@task
def lint(c):
    """Lint the Interface code."""
    print("Linting Interface...")
    run_interface_command(c, "npm run lint")

@task
def test(c):
    """Run Interface Unit Tests (Vitest)."""
    print("Running Interface Tests...")
    _run_one_shot_interface_task(["npm", "run", "test"])


@task
def typecheck(c):
    """Run the Interface TypeScript typecheck."""
    print("Running Interface Type Check...")
    _run_one_shot_interface_task(["npx", "tsc", "--noEmit"])


@task
def test_coverage(c):
    """Run Interface unit tests with V8 coverage report."""
    print("Running Interface Tests with Coverage...")
    _run_one_shot_interface_task(["npx", "vitest", "run", "--coverage"])

@task(
    help={
        "headed": "Open a visible browser window.",
        "project": "Optional Playwright project (chromium, firefox, webkit, mobile-chromium).",
        "spec": "Optional Playwright spec path or glob.",
        "live_backend": "Enable specs that require a real Core backend and authenticated UI proxying.",
        "workers": "Optional Playwright worker count override for stability-sensitive runs.",
        "server_mode": "Server mode for the managed UI server (dev or start, defaults to dev).",
    }
)
def e2e(c, headed=False, project="", spec="", live_backend=False, workers="", server_mode="dev"):
    """
    Run Playwright E2E tests.
    The Invoke wrapper starts a managed local Next.js server and clears any stale
    Interface listener before and after the run because Next.js dev servers can
    linger on Windows after Playwright exits.
    Stable mocked browser proof defaults to a managed dev server; use
    --server-mode=start when the run should refresh and serve the built bundle.
    Use --live-backend when the spec should talk to the real Core API through
    the Next.js proxy instead of relying entirely on route stubs.
    """
    print("Running Playwright E2E Tests...")
    ensure_disk_headroom(min_free_gb=8 if server_mode == "start" else 6, reason=f"interface e2e ({server_mode})")
    cmd = _build_playwright_command(project=project, spec=spec, workers=workers, headed=headed)
    chosen_port = _pick_interface_port(INTERFACE_PORT)
    env = _build_playwright_env(live_backend=live_backend, port=chosen_port)
    stop(c)
    if chosen_port != INTERFACE_PORT:
        stop(c, port=chosen_port)
    _cleanup_managed_interface_listeners()
    if server_mode == "start":
        print("Refreshing built Interface bundle for managed start-mode browser proof...")
        build(c)
    print(f"Using managed Interface port {chosen_port}")
    server: subprocess.Popen[str] | None = None
    keep_server_log = False
    try:
        server = _start_playwright_server(env, port=chosen_port, server_mode=server_mode)
        env, chosen_port = _reconcile_managed_server_endpoint(env, chosen_port, server)
        result = _run_interface_commandline(cmd, extra_env=env, stream=True)
        _print_ascii_safe(result.stdout)
        _print_ascii_safe(result.stderr)
        if result.exited != 0:
            raise SystemExit(result.exited)
    except BaseException:
        keep_server_log = True
        raise
    finally:
        if server is not None and server.poll() is None:
            _kill_pid_tree(server.pid)
            time.sleep(0.5)
            with suppress(Exception):
                server.wait(timeout=5)
        stop(c, port=chosen_port)
        _cleanup_repo_local_interface_processes()
        _cleanup_managed_interface_listeners()
        if not keep_server_log:
            _cleanup_playwright_server_log()

# ── Process Management ───────────────────────────────────────

@task
def stop(c, port=INTERFACE_PORT):
    """
    Stop the Interface server.
    Kills the listener on --port (default 3000) and then sweeps any repo-local
    Next.js/Vitest/Playwright worker residue that survived outside that port.
    """
    print(f"Stopping Interface (port {port})...")
    if is_windows():
        port_pids = _windows_listening_pids_for_port(int(port))
        if port_pids:
            for pid in port_pids:
                _kill_pid_tree(pid)
                print(f"Killed PID {pid}")
        else:
            print(f"No process on port {port}")
    else:
        # lsof works on macOS + Linux; fuser as fallback
        c.run(f"lsof -ti:{port} | xargs -r kill -9 2>/dev/null || fuser -k {port}/tcp 2>/dev/null || true", warn=True)
    remaining = _cleanup_repo_local_interface_processes()
    if remaining:
        summary = ", ".join(f"{proc['name']}:{proc['pid']}" for proc in remaining[:6])
        print(f"WARN: repo-local Interface residuals still running ({summary})")
    print("Interface stopped.")

@task
def clean(c):
    """
    Clear the Next.js build cache (.next directory).
    Use when HMR gets stuck or stale chunks cause ghost errors.
    """
    import os

    def _ignore_missing_rmtree_error(function, path, excinfo):
        error = excinfo[1] if isinstance(excinfo, tuple) else excinfo
        if isinstance(error, FileNotFoundError):
            return
        raise error

    cache_dir = os.path.join("interface", ".next")
    print("Clearing Next.js cache...")
    _cleanup_repo_local_interface_processes()
    if os.path.isdir(cache_dir):
        last_permission_error = None
        for attempt in range(3):
            try:
                shutil.rmtree(cache_dir, onexc=_ignore_missing_rmtree_error)
                last_permission_error = None
                break
            except PermissionError as exc:
                last_permission_error = exc
                print(f"  WARN: cache removal was blocked by an open file handle ({exc}); retrying after cleanup...")
                _cleanup_repo_local_interface_processes()
                time.sleep(0.5)
        if last_permission_error is not None and os.path.isdir(cache_dir):
            print("  WARN: Next.js cache could not be fully removed after cleanup retries; continuing with the remaining cache in place.")
    print("Cache cleared.")

@task
def restart(c, port=INTERFACE_PORT):
    """
    Full Interface restart: stop -> clear cache -> build -> start dev -> check.
    Use when UI shows stale errors or HMR is broken.
    """
    import time

    print("=== Interface Restart ===")
    print()

    # 1. Stop existing server
    stop(c, port=port)

    # 2. Clear stale .next cache
    clean(c)

    # 3. Verify build compiles clean
    print()
    build(c)

    # 4. Start dev server in background
    print(f"\nStarting dev server on port {port}...")
    start_dev_server_detached(port=port)

    # 5. Wait for server to be ready, then check
    print("Waiting for server startup...")
    time.sleep(6)
    check(c, port=port)

# ── Health Check ─────────────────────────────────────────────

@task
def check(c, port=INTERFACE_PORT):
    """
    Smoke-test a running Interface server.
    Fetches key pages and checks for SSR errors, 404s, and dark-mode leaks.
    Requires: any repo-managed Interface server listening on --port (default 3000).
    """
    import urllib.request

    base = f"http://{INTERFACE_HOST}:{port}"
    pages = ["/", "/wiring", "/architect", "/dashboard", "/catalogue", "/teams", "/memory", "/settings/tools", "/approvals"]
    errors = []

    print(f"Checking Interface at {base}...")

    for page in pages:
        url = f"{base}{page}"
        try:
            req = urllib.request.Request(url)
            with urllib.request.urlopen(req, timeout=10) as resp:
                status = resp.status
                body = resp.read().decode("utf-8", errors="replace")

                issues = []
                if "NEXT_REDIRECT" in body and "404" in body:
                    issues.append("404 redirect detected")
                if "Internal Server Error" in body:
                    issues.append("500 Internal Server Error")
                if "__next_error__" in body:
                    issues.append("Next.js error boundary triggered")
                if "Application error" in body or "Unhandled Runtime Error" in body:
                    issues.append("React runtime error detected")
                if "bg-white" in body and page in ("/wiring", "/architect"):
                    issues.append("Light-mode bg-white leak detected")

                ok = status == 200 and not issues
                icon = "[OK]" if ok else "[FAIL]"
                print(f"  {icon} {page} [{status}]", end="")
                if issues:
                    print(f"  WARN: {', '.join(issues)}")
                    errors.extend([f"{page}: {i}" for i in issues])
                else:
                    print()

        except Exception as e:
            print(f"  [FAIL] {page} - {e}")
            errors.append(f"{page}: {e}")

    print()
    if errors:
        print(f"ISSUES: {len(errors)} problem(s) found:")
        for e in errors:
            print(f"  - {e}")
        raise SystemExit(1)
    else:
        print("ALL PAGES HEALTHY.")

# ── Register Tasks ───────────────────────────────────────────

ns.add_task(dev)
ns.add_task(install)
ns.add_task(build)
ns.add_task(lint)
ns.add_task(test)
ns.add_task(typecheck)
ns.add_task(test_coverage, name="test-coverage")
ns.add_task(e2e)
ns.add_task(stop)
ns.add_task(clean)
ns.add_task(restart)
ns.add_task(check)
