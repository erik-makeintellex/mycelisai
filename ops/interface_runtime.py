import ipaddress
import os
import re
import shutil
import subprocess
import time
import urllib.error
import urllib.request
import socket
from contextlib import suppress
from pathlib import Path

from invoke import Collection, task
from .cache import ensure_disk_headroom
from .config import (
    INTERFACE_BIND_HOST,
    INTERFACE_HOST,
    INTERFACE_PORT,
    ROOT_DIR,
    is_windows,
    powershell,
)
from .interface_env import (
    CommandResult,
    _build_playwright_command,
    _build_playwright_env,
    _expected_next_build_artifacts,
    _is_incomplete_next_build_output,
    _is_next_build_lock_conflict,
    _is_next_standalone_cleanup_conflict,
    _normalize_process_text,
    _playwright_last_run_path,
    _read_playwright_last_run_status,
    _reconcile_managed_server_endpoint,
    _report_command_result,
    _run_interface_commandline,
    _run_interface_shell_command,
    _run_interface_shell_command_streaming,
    _run_one_shot_interface_task,
    _task_env,
    _wait_for_complete_next_build_output,
    interface_task_env,
    run_interface_command,
)
from . import interface_processes
ns = Collection("interface")
INTERFACE_DIR = ROOT_DIR / "interface"

def _matches_repo_local_interface_process(name: str, command_line: str) -> bool:
    return interface_processes.matches_repo_local_interface_process(
        name,
        command_line,
        normalize_process_text=_normalize_process_text,
    )


def _list_repo_local_interface_processes() -> list[dict[str, str | int]]:
    return interface_processes.list_repo_local_interface_processes(
        is_windows_func=is_windows,
        normalize_process_text=_normalize_process_text,
        run=subprocess.run,
    )


def _kill_pid_tree(pid: int):
    interface_processes.kill_pid_tree(pid, is_windows_func=is_windows, run=subprocess.run)


def _cleanup_repo_local_interface_processes() -> list[dict[str, str | int]]:
    return interface_processes.cleanup_repo_local_interface_processes(
        list_processes=_list_repo_local_interface_processes,
        kill_pid_tree_func=_kill_pid_tree,
        sleep=time.sleep,
    )


def _windows_listening_pids_for_port(port: int) -> list[int]:
    return interface_processes.windows_listening_pids_for_port(port, run=subprocess.run)


def _windows_listening_pids_for_port_range(port_start: int, port_end: int) -> list[int]:
    return interface_processes.windows_listening_pids_for_port_range(port_start, port_end, run=subprocess.run)


def _cleanup_managed_interface_listeners(port_start: int = 3100, port_end: int = 3199) -> list[int]:
    return interface_processes.cleanup_managed_interface_listeners(
        port_start,
        port_end,
        is_windows_func=is_windows,
        windows_listening_pids_for_port_range_func=_windows_listening_pids_for_port_range,
        kill_pid_tree_func=_kill_pid_tree,
        sleep=time.sleep,
    )


def _playwright_server_log_path() -> str:
    log_dir = ROOT_DIR / "workspace" / "logs"
    log_dir.mkdir(parents=True, exist_ok=True)
    return str(log_dir / "interface-playwright-webserver.log")


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
        bind_host = ("127.0.0.1" if is_windows() else INTERFACE_BIND_HOST).strip()
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
        "--webpack",
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
    print("Installing Playwright Chromium browser...")
    run_interface_command(c, "npx playwright install chromium")

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
