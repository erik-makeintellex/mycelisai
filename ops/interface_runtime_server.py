from __future__ import annotations

from contextlib import suppress
from pathlib import Path
import subprocess

import ops.interface_runtime as _ir


def _next_dev_lock_path() -> Path:
    return _ir.INTERFACE_DIR / ".next" / "dev" / "lock"


def _cleanup_stale_next_dev_lock() -> bool:
    lock_path = _ir._next_dev_lock_path()
    if not lock_path.exists():
        return False
    processes: list[dict[str, str | int]] = []
    try:
        processes = _ir._list_repo_local_interface_processes()
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


def _pick_interface_port(preferred: int = _ir.INTERFACE_PORT) -> int:
    """Return a free port for the managed Playwright server."""
    def _port_has_listener(port: int) -> bool:
        families = [_ir.socket.AF_INET]
        if _ir.socket.has_ipv6:
            families.append(_ir.socket.AF_INET6)
        for family in families:
            try:
                with _ir.socket.socket(family, _ir.socket.SOCK_STREAM) as probe:
                    probe.settimeout(0.2)
                    host = "127.0.0.1" if family == _ir.socket.AF_INET else "::1"
                    if probe.connect_ex((host, port)) == 0:
                        return True
            except OSError:
                continue
        return False

    def _port_is_available(port: int) -> bool:
        bind_host = ("127.0.0.1" if _ir.is_windows() else _ir.INTERFACE_BIND_HOST).strip()
        probe_host = bind_host or "127.0.0.1"
        family = _ir.socket.AF_INET
        if probe_host == "localhost":
            probe_host = "127.0.0.1"
        else:
            try:
                parsed_host = _ir.ipaddress.ip_address(probe_host)
            except ValueError:
                if ":" in probe_host:
                    family = _ir.socket.AF_INET6
                else:
                    probe_host = "127.0.0.1"
            else:
                if parsed_host.version == 6:
                    family = _ir.socket.AF_INET6
        with _ir.socket.socket(family, _ir.socket.SOCK_STREAM) as probe_sock:
            if family == _ir.socket.AF_INET6:
                with suppress(OSError, AttributeError):
                    probe_sock.setsockopt(
                        _ir.socket.IPPROTO_IPV6,
                        _ir.socket.IPV6_V6ONLY,
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
        with _ir.socket.socket(_ir.socket.AF_INET, _ir.socket.SOCK_STREAM) as ipv4_sock:
            ipv4_sock.bind(("127.0.0.1", 0))
            candidate = ipv4_sock.getsockname()[1]
            try:
                if _port_has_listener(candidate):
                    continue
                with _ir.socket.socket(_ir.socket.AF_INET6, _ir.socket.SOCK_STREAM) as ipv6_loopback_sock:
                    ipv6_loopback_sock.bind(("::1", candidate))
                    with _ir.socket.socket(_ir.socket.AF_INET6, _ir.socket.SOCK_STREAM) as ipv6_wildcard_sock:
                        ipv6_wildcard_sock.bind(("::", candidate))
                        return candidate
            except OSError:
                continue

    raise RuntimeError("Unable to find a free Interface port for the managed Playwright server.")


def _next_dev_command(bind_host: str, port: int) -> list[str]:
    return [
        "node",
        str((_ir.INTERFACE_DIR / "node_modules" / "next" / "dist" / "bin" / "next").resolve()),
        "dev",
        "--webpack",
        "--hostname",
        bind_host,
        "--port",
        str(port),
    ]


def _next_start_command(bind_host: str, port: int) -> list[str]:
    del bind_host, port
    return ["node", str((_ir.INTERFACE_DIR / "scripts" / "playwright-webserver.mjs").resolve())]


def _spawn_interface_process(
    command: list[str],
    env: dict[str, str],
    *,
    stdout,
    stderr,
    detached: bool,
    text: bool = True,
) -> subprocess.Popen[str]:
    process_env = _ir.os.environ.copy()
    process_env.update(env)
    popen_kwargs = {
        "cwd": str(_ir.INTERFACE_DIR),
        "env": process_env,
        "stdout": stdout,
        "stderr": stderr,
        "text": text,
    }
    if detached:
        popen_kwargs["stdin"] = _ir.subprocess.DEVNULL
        if _ir.is_windows():
            popen_kwargs["creationflags"] = (
                getattr(_ir.subprocess, "CREATE_NEW_PROCESS_GROUP", 0)
                | getattr(_ir.subprocess, "DETACHED_PROCESS", 0)
            )
        else:
            popen_kwargs["start_new_session"] = True
    return _ir.subprocess.Popen(command, **popen_kwargs)


def start_dev_server_detached(
    env: dict[str, str] | None = None,
    *,
    host: str = _ir.INTERFACE_BIND_HOST,
    port: int = _ir.INTERFACE_PORT,
) -> subprocess.Popen[str]:
    process_env = _ir._task_env()
    if env:
        process_env.update(env)
    process_env.setdefault("INTERFACE_HOST", _ir.INTERFACE_HOST)
    process_env.setdefault("INTERFACE_BIND_HOST", host)
    _ir._cleanup_stale_next_dev_lock()
    command = _ir._next_dev_command(process_env["INTERFACE_BIND_HOST"], port)
    return _ir._spawn_interface_process(
        command,
        process_env,
        stdout=_ir.subprocess.DEVNULL,
        stderr=_ir.subprocess.STDOUT,
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
    host: str = _ir.INTERFACE_HOST,
    port: int = _ir.INTERFACE_PORT,
    timeout_seconds: int = 120,
    process: subprocess.Popen[str] | None = None,
) -> str:
    urls = _ir._interface_ready_urls(host, port)
    deadline = _ir.time.time() + timeout_seconds
    last_error = "server did not respond"
    process_exited_early = False
    while _ir.time.time() < deadline:
        if process is not None and process.poll() is not None:
            process_exited_early = True
        for url in urls:
            try:
                with _ir.urllib.request.urlopen(url, timeout=5) as response:
                    if response.status < 500:
                        return url.split("://", 1)[1].rsplit(":", 1)[0].strip("[]")
                    last_error = f"{url}: unexpected status {response.status}"
            except _ir.urllib.error.HTTPError as exc:
                if exc.code < 500:
                    return url.split("://", 1)[1].rsplit(":", 1)[0].strip("[]")
                last_error = f"{url}: http {exc.code}"
            except Exception as exc:  # pragma: no cover - exercised via timeout path in task flow
                last_error = f"{url}: {exc}"
        _ir.time.sleep(1)

    if process_exited_early:
        raise RuntimeError(
            f"Managed Interface server exited before it became ready on port {port}. "
            f"See {_ir._playwright_server_log_path()} for server output."
        )

    raise RuntimeError(
        f"Interface did not become ready at any of {', '.join(urls)} within {timeout_seconds}s. "
        f"See {_ir._playwright_server_log_path()} for server output. Last error: {last_error}"
    )
