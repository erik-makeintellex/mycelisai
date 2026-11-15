from __future__ import annotations

import ops.interface_runtime as _ir


def _matches_repo_local_interface_process(name: str, command_line: str) -> bool:
    return _ir.interface_processes.matches_repo_local_interface_process(
        name,
        command_line,
        normalize_process_text=_ir._normalize_process_text,
    )


def _list_repo_local_interface_processes() -> list[dict[str, str | int]]:
    return _ir.interface_processes.list_repo_local_interface_processes(
        is_windows_func=_ir.is_windows,
        normalize_process_text=_ir._normalize_process_text,
        run=_ir.subprocess.run,
    )


def _kill_pid_tree(pid: int):
    _ir.interface_processes.kill_pid_tree(pid, is_windows_func=_ir.is_windows, run=_ir.subprocess.run)


def _cleanup_repo_local_interface_processes() -> list[dict[str, str | int]]:
    return _ir.interface_processes.cleanup_repo_local_interface_processes(
        list_processes=_ir._list_repo_local_interface_processes,
        kill_pid_tree_func=_ir._kill_pid_tree,
        sleep=_ir.time.sleep,
    )


def _windows_listening_pids_for_port(port: int) -> list[int]:
    return _ir.interface_processes.windows_listening_pids_for_port(port, run=_ir.subprocess.run)


def _windows_listening_pids_for_port_range(port_start: int, port_end: int) -> list[int]:
    return _ir.interface_processes.windows_listening_pids_for_port_range(port_start, port_end, run=_ir.subprocess.run)


def _cleanup_managed_interface_listeners(port_start: int = 3100, port_end: int = 3199) -> list[int]:
    return _ir.interface_processes.cleanup_managed_interface_listeners(
        port_start,
        port_end,
        is_windows_func=_ir.is_windows,
        windows_listening_pids_for_port_range_func=_ir._windows_listening_pids_for_port_range,
        kill_pid_tree_func=_ir._kill_pid_tree,
        sleep=_ir.time.sleep,
    )


def _print_ascii_safe(text: str):
    if not text:
        return
    print(text.encode("ascii", "replace").decode("ascii"), end="")


def run_build_task(c):
    print("Building Interface...")
    _ir.ensure_disk_headroom(min_free_gb=8, reason="interface build")
    _ir.stop(c)
    _ir.clean(c)
    _ir._cleanup_repo_local_interface_processes()
    build_succeeded = False
    try:
        result = _ir._run_interface_shell_command(["npm", "run", "build"])
        _ir._report_command_result(result)
        if result.exited != 0 and _ir._is_next_build_lock_conflict(result):
            print("Detected a stale Next.js build lock. Cleaning repo-local Interface workers and retrying once...")
            _ir._cleanup_repo_local_interface_processes()
            _ir.clean(c)
            _ir._cleanup_repo_local_interface_processes()
            result = _ir._run_interface_shell_command(["npm", "run", "build"])
            _ir._report_command_result(result)
        if result.exited != 0 and _ir._is_next_standalone_cleanup_conflict(result):
            print("Detected a stale Next.js standalone cleanup lock. Cleaning repo-local Interface workers and retrying once...")
            _ir._cleanup_repo_local_interface_processes()
            _ir._cleanup_managed_interface_listeners()
            _ir.clean(c)
            _ir._cleanup_repo_local_interface_processes()
            _ir._cleanup_managed_interface_listeners()
            result = _ir._run_interface_shell_command(["npm", "run", "build"])
            _ir._report_command_result(result)
        if result.exited != 0 and _ir._is_incomplete_next_build_output(result):
            print("Detected incomplete built-server output during Next.js packaging. Cleaning repo-local Interface workers and retrying once...")
            _ir._cleanup_repo_local_interface_processes()
            _ir.clean(c)
            _ir._cleanup_repo_local_interface_processes()
            result = _ir._run_interface_shell_command(["npm", "run", "build"])
            _ir._report_command_result(result)
        if result.exited != 0:
            raise SystemExit(result.exited)
        try:
            _ir._wait_for_complete_next_build_output()
        except RuntimeError:
            print("Detected incomplete built-server output after the build command exited. Cleaning and retrying once...")
            _ir._cleanup_repo_local_interface_processes()
            _ir.clean(c)
            _ir._cleanup_repo_local_interface_processes()
            result = _ir._run_interface_shell_command(["npm", "run", "build"])
            _ir._report_command_result(result)
            if result.exited != 0:
                raise SystemExit(result.exited)
            _ir._wait_for_complete_next_build_output()
        build_succeeded = True
    finally:
        # On Windows, successful Next builds can still finalize the static asset
        # tree after the main command exits. Sweeping repo-local node helpers
        # immediately here can leave a partial `.next/static` tree behind.
        if not build_succeeded:
            _ir._cleanup_repo_local_interface_processes()


def run_stop_task(c, port):
    print(f"Stopping Interface (port {port})...")
    if _ir.is_windows():
        port_pids = _ir._windows_listening_pids_for_port(int(port))
        if port_pids:
            for pid in port_pids:
                _ir._kill_pid_tree(pid)
                print(f"Killed PID {pid}")
        else:
            print(f"No process on port {port}")
    else:
        c.run(f"lsof -ti:{port} | xargs -r kill -9 2>/dev/null || fuser -k {port}/tcp 2>/dev/null || true", warn=True)
    remaining = _ir._cleanup_repo_local_interface_processes()
    if remaining:
        summary = ", ".join(f"{proc['name']}:{proc['pid']}" for proc in remaining[:6])
        print(f"WARN: repo-local Interface residuals still running ({summary})")
    print("Interface stopped.")


def run_clean_task(c):
    del c

    def _ignore_missing_rmtree_error(function, path, excinfo):
        del function, path
        error = excinfo[1] if isinstance(excinfo, tuple) else excinfo
        if isinstance(error, FileNotFoundError):
            return
        raise error

    cache_dir = _ir.os.path.join("interface", ".next")
    print("Clearing Next.js cache...")
    _ir._cleanup_repo_local_interface_processes()
    if _ir.os.path.isdir(cache_dir):
        last_permission_error = None
        for _attempt in range(3):
            try:
                _ir.shutil.rmtree(cache_dir, onexc=_ignore_missing_rmtree_error)
                last_permission_error = None
                break
            except PermissionError as exc:
                last_permission_error = exc
                print(f"  WARN: cache removal was blocked by an open file handle ({exc}); retrying after cleanup...")
                _ir._cleanup_repo_local_interface_processes()
                _ir.time.sleep(0.5)
        if last_permission_error is not None and _ir.os.path.isdir(cache_dir):
            print("  WARN: Next.js cache could not be fully removed after cleanup retries; continuing with the remaining cache in place.")
    print("Cache cleared.")


def run_restart_task(c, port):
    print("=== Interface Restart ===")
    print()
    _ir.stop(c, port=port)
    _ir.clean(c)
    print()
    _ir.build(c)
    print(f"\nStarting dev server on port {port}...")
    _ir.start_dev_server_detached(port=port)
    print("Waiting for server startup...")
    _ir.time.sleep(6)
    _ir.check(c, port=port)


def run_check_task(c, port):
    del c
    base = f"http://{_ir.INTERFACE_HOST}:{port}"
    pages = ["/", "/wiring", "/architect", "/dashboard", "/catalogue", "/teams", "/memory", "/settings/tools", "/approvals"]
    errors = []

    print(f"Checking Interface at {base}...")

    for page in pages:
        url = f"{base}{page}"
        try:
            req = _ir.urllib.request.Request(url)
            with _ir.urllib.request.urlopen(req, timeout=10) as resp:
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
    print("ALL PAGES HEALTHY.")
