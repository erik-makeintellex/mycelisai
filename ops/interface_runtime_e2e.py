from __future__ import annotations

from contextlib import suppress
import subprocess

import ops.interface_runtime as _ir


def _playwright_server_log_path() -> str:
    log_dir = _ir.ROOT_DIR / "workspace" / "logs"
    log_dir.mkdir(parents=True, exist_ok=True)
    return str(log_dir / "interface-playwright-webserver.log")


def _stop_lingering_playwright_process(proc: subprocess.Popen[str]) -> None:
    with suppress(ProcessLookupError, OSError):
        proc.terminate()
    with suppress(_ir.subprocess.TimeoutExpired, OSError):
        proc.wait(timeout=3)
    if proc.poll() is not None:
        return
    _ir._kill_pid_tree(proc.pid)
    with suppress(_ir.subprocess.TimeoutExpired, OSError):
        proc.wait(timeout=3)


def _run_playwright_command_streaming(
    command: list[str],
    *,
    extra_env: dict[str, str] | None = None,
    post_result_grace_seconds: float = 15.0,
) -> _ir.CommandResult:
    last_run_path = _ir._playwright_last_run_path()
    with suppress(FileNotFoundError):
        last_run_path.unlink()

    process_env = _ir.os.environ.copy()
    process_env.update(_ir._task_env(extra_env))
    runner = list(command)
    executable = runner[0]
    if _ir.is_windows():
        resolved = (
            _ir.shutil.which(executable)
            or _ir.shutil.which(f"{executable}.cmd")
            or _ir.shutil.which(f"{executable}.exe")
        )
        if resolved:
            runner[0] = resolved
    else:
        resolved = _ir.shutil.which(executable)
        if resolved:
            runner[0] = resolved

    proc = _ir.subprocess.Popen(
        runner,
        cwd=str(_ir.INTERFACE_DIR),
        env=process_env,
        text=True,
    )
    result_seen_at: float | None = None
    result_status: str | None = None
    while True:
        exit_code = proc.poll()
        if exit_code is not None:
            return _ir.CommandResult(exited=exit_code, stdout="", stderr="")

        current_status = _ir._read_playwright_last_run_status()
        if current_status in {"passed", "failed"}:
            if current_status != result_status:
                result_status = current_status
                result_seen_at = _ir.time.time()
            elif result_seen_at is not None and (_ir.time.time() - result_seen_at) >= post_result_grace_seconds:
                print(
                    f"  WARN: Playwright reported {current_status} but did not exit cleanly; "
                    "terminating the lingering test process."
                )
                _ir._stop_lingering_playwright_process(proc)
                return _ir.CommandResult(exited=0 if current_status == "passed" else 1, stdout="", stderr="")
        else:
            result_seen_at = None
            result_status = None

        _ir.time.sleep(1.0)


def _cleanup_playwright_server_log(max_attempts: int = 20, retry_delay_seconds: float = 0.25) -> None:
    log_path = _ir.Path(_ir._playwright_server_log_path())
    for attempt in range(max_attempts):
        try:
            log_path.unlink()
            return
        except FileNotFoundError:
            return
        except PermissionError:
            if attempt == max_attempts - 1:
                return
            _ir.time.sleep(retry_delay_seconds)


def _detect_playwright_server_port(expected_port: int, timeout_seconds: int = 30) -> int:
    """Read the managed server log until Next reports its actual local port."""
    log_path = _ir.Path(_ir._playwright_server_log_path())
    deadline = _ir.time.time() + timeout_seconds
    patterns = (
        _ir.re.compile(r"Local:\s+http://\[(?:::1|::)\]:(\d+)"),
        _ir.re.compile(r"Local:\s+http://127\.0\.0\.1:(\d+)"),
        _ir.re.compile(r"Local:\s+http://localhost:(\d+)"),
    )
    while _ir.time.time() < deadline:
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
        _ir.time.sleep(0.2)
    return expected_port


def _start_playwright_server(
    env: dict[str, str],
    port: int = _ir.INTERFACE_PORT,
    server_mode: str = "dev",
) -> subprocess.Popen[str]:
    log_path = _ir._playwright_server_log_path()
    log_handle = open(log_path, "w", encoding="utf-8")
    bind_host = env.get("INTERFACE_BIND_HOST", _ir.INTERFACE_BIND_HOST)
    command_env = dict(env)
    if server_mode == "start":
        command_env["PORT"] = str(port)
        command_env["HOSTNAME"] = bind_host
        command = _ir._next_start_command(bind_host, port)
    else:
        _ir._cleanup_stale_next_dev_lock()
        command = _ir._next_dev_command(bind_host, port)
    try:
        process = _ir._spawn_interface_process(
            command,
            command_env,
            stdout=log_handle,
            stderr=_ir.subprocess.STDOUT,
            detached=False,
        )
        log_handle.close()
        return process
    except Exception:
        log_handle.close()
        raise


def run_e2e_task(c, headed=False, project="", spec="", live_backend=False, workers="", server_mode="dev"):
    print("Running Playwright E2E Tests...")
    _ir.ensure_disk_headroom(min_free_gb=8 if server_mode == "start" else 6, reason=f"interface e2e ({server_mode})")
    cmd = _ir._build_playwright_command(project=project, spec=spec, workers=workers, headed=headed)
    chosen_port = _ir._pick_interface_port(_ir.INTERFACE_PORT)
    env = _ir._build_playwright_env(live_backend=live_backend, port=chosen_port)
    _ir.stop(c)
    if chosen_port != _ir.INTERFACE_PORT:
        _ir.stop(c, port=chosen_port)
    _ir._cleanup_managed_interface_listeners()
    if server_mode == "start":
        print("Refreshing built Interface bundle for managed start-mode browser proof...")
        _ir.build(c)
    print(f"Using managed Interface port {chosen_port}")
    server: subprocess.Popen[str] | None = None
    keep_server_log = False
    try:
        server = _ir._start_playwright_server(env, port=chosen_port, server_mode=server_mode)
        env, chosen_port = _ir._reconcile_managed_server_endpoint(env, chosen_port, server)
        result = _ir._run_interface_commandline(cmd, extra_env=env, stream=True)
        _ir._print_ascii_safe(result.stdout)
        _ir._print_ascii_safe(result.stderr)
        if result.exited != 0:
            raise SystemExit(result.exited)
    except BaseException:
        keep_server_log = True
        raise
    finally:
        if server is not None and server.poll() is None:
            _ir._kill_pid_tree(server.pid)
            _ir.time.sleep(0.5)
            with suppress(Exception):
                server.wait(timeout=5)
        _ir.stop(c, port=chosen_port)
        _ir._cleanup_repo_local_interface_processes()
        _ir._cleanup_managed_interface_listeners()
        if not keep_server_log:
            _ir._cleanup_playwright_server_log()
