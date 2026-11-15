"""
Lifecycle management for the Mycelis development stack.

Provides unified start/stop/status/health commands that handle the full
dependency graph: port-forwards -> core server -> frontend.

All probes are Python-native (socket/urllib) — no shell wrappers.
"""

import csv  # noqa: F401
import json  # noqa: F401
import os
import socket
import subprocess  # noqa: F401
import time
from pathlib import Path  # noqa: F401

from invoke import Collection, task

from . import core as core_tasks, db as db_tasks, interface as interface_tasks
from .config import (
    API_HOST, API_PORT, CORE_DIR, INTERFACE_BIND_HOST, INTERFACE_HOST,
    INTERFACE_PORT, NAMESPACE, ROOT_DIR, is_windows, powershell,  # noqa: F401
)

core = core_tasks
db = db_tasks
interface = interface_tasks

SERVICES = {
    "postgres": {"port": 5432, "label": "PostgreSQL", "kind_svc": "svc/mycelis-core-postgresql", "forward": "5432:5432"},
    "nats":     {"port": 4222, "label": "NATS",       "kind_svc": "svc/mycelis-core-nats",       "forward": "4222:4222"},
    "core":     {"port": API_PORT,       "label": "Core API"},
    "frontend": {"port": INTERFACE_PORT, "label": "Frontend"},
    "ollama":   {"port": 11434,          "label": "Ollama"},
}

CORE_STARTUP_LOG = ROOT_DIR / "workspace" / "logs" / "core-startup.log"
WINDOWS_COMPILED_GO_PROCESS_NAMES = ("go.exe", "server.exe", "probe.exe", "signal_gen.exe", "smoke.exe")
COMPILED_GO_PROCESS_HINTS = tuple(
    hint.lower()
    for hint in {
        "go run ./cmd/server", "go run .\\cmd\\server",
        "go run ./cmd/probe", "go run .\\cmd\\probe",
        "go run ./cmd/signal_gen", "go run .\\cmd\\signal_gen",
        "go run ./cmd/smoke/main.go", "go run .\\cmd\\smoke\\main.go",
        "cmd/server", "cmd\\server", "cmd/probe", "cmd\\probe",
        "cmd/signal_gen", "cmd\\signal_gen", "cmd/smoke", "cmd\\smoke",
        "bin/server", "bin\\server", "bin/probe", "bin\\probe",
        "bin/signal_gen", "bin\\signal_gen", "bin/smoke", "bin\\smoke",
        str((CORE_DIR / "bin" / "server").resolve()),
        str((CORE_DIR / "bin" / "server.exe").resolve()),
        str((CORE_DIR / "bin" / "probe").resolve()),
        str((CORE_DIR / "bin" / "probe.exe").resolve()),
        str((CORE_DIR / "bin" / "signal_gen").resolve()),
        str((CORE_DIR / "bin" / "signal_gen.exe").resolve()),
        str((CORE_DIR / "cmd" / "server").resolve()),
        str((CORE_DIR / "cmd" / "probe").resolve()),
        str((CORE_DIR / "cmd" / "signal_gen").resolve()),
        str((CORE_DIR / "cmd" / "smoke").resolve()),
    }
)


def _port_open(port: int, host: str = "127.0.0.1", timeout: float = 1.0) -> bool:
    """Check if a TCP port is accepting connections."""
    try:
        with socket.create_connection((host, port), timeout=timeout):
            return True
    except (ConnectionRefusedError, TimeoutError, OSError):
        return False


def _http_get(url: str, timeout: float = 3.0, headers: dict[str, str] | None = None) -> tuple[int, str]:
    """HTTP GET returning (status_code, body). Returns (0, error) on failure."""
    import urllib.error
    import urllib.request

    try:
        req = urllib.request.Request(url, headers=headers or {})
        with urllib.request.urlopen(req, timeout=timeout) as resp:
            return resp.status, resp.read().decode("utf-8", errors="replace")
    except urllib.error.HTTPError as e:
        return e.code, str(e)
    except Exception as e:
        return 0, str(e)


def _interface_probe_urls(host: str = INTERFACE_HOST, port: int = INTERFACE_PORT) -> list[str]:
    candidates = [host, "127.0.0.1", "localhost", "::1"]
    urls: list[str] = []
    seen: set[str] = set()
    for candidate in candidates:
        if not candidate:
            continue
        url_host = f"[{candidate}]" if (":" in candidate and not candidate.startswith("[")) else candidate
        url = f"http://{url_host}:{port}/"
        if url not in seen:
            seen.add(url)
            urls.append(url)
    return urls


def _wait_for_port(port: int, label: str, timeout: int = 30, interval: float = 1.0) -> bool:
    """Block until a port is open or timeout expires. Returns True on success."""
    deadline = time.time() + timeout
    while time.time() < deadline:
        if _port_open(port):
            return True
        time.sleep(interval)
    print(f"  TIMEOUT waiting for {label} on port {port} ({timeout}s)")
    return False


def _wait_for_port_closed(port: int, label: str, timeout: int = 3, interval: float = 0.25) -> bool:
    """Block until a port is no longer open or timeout expires."""
    deadline = time.time() + timeout
    while time.time() < deadline:
        if not _port_open(port):
            return True
        time.sleep(interval)
    print(f"  WARN: {label} still holds port {port} after {timeout}s")
    return False


def _wait_for_http_ok(
    url: str,
    label: str,
    timeout: int = 30,
    interval: float = 1.0,
    headers: dict[str, str] | None = None,
) -> bool:
    """Block until an HTTP endpoint returns 200 or timeout expires."""
    deadline = time.time() + timeout
    while time.time() < deadline:
        code, _body = _http_get(url, timeout=5.0, headers=headers)
        if code == 200:
            return True
        time.sleep(interval)
    print(f"  TIMEOUT waiting for {label} at {url} ({timeout}s)")
    return False


from .lifecycle_postgres import (  # noqa: E402,F401
    _core_council_ready, _core_startup_log_path, _ensure_bridge, _kill_bridges,
    _load_env, _start_port_forward, run_memory_restart, run_status,
)
from .lifecycle_services import (  # noqa: E402,F401
    _find_pid_on_port, _kill_compiled_go_services, _kill_pid, _kill_port,
    _list_compiled_go_service_processes, _matches_compiled_go_service,
    _remaining_managed_services, _run_best_effort, _service_keys_by_label,
    _start_core_background,
)
from .lifecycle_workflows import run_down, run_health, run_up  # noqa: E402,F401


@task
def status(c):
    """Show health of every service in the dev stack."""
    run_status(c)


@task(
    help={
        "frontend": "Also start the frontend dev server (default: False)",
        "build": "Build the Go binary before starting (default: False)",
    }
)
def up(c, frontend=False, build=False):
    """Bring up the full dev stack (idempotent)."""
    run_up(c, frontend=frontend, build=build)


@task
def down(c):
    """Stop all dev stack services cleanly."""
    run_down(c)


@task
def health(c):
    """Deep health check — probes actual API endpoints."""
    run_health(c)


@task
def restart(c, build=False, frontend=False):
    """Full stack restart: down -> (optional build) -> up."""
    down(c)
    print()
    time.sleep(2)
    up(c, frontend=frontend, build=build)


@task(
    help={
        "build": "Build the Go binary before restart (default: False).",
        "frontend": "Also start frontend after restart (default: False).",
    }
)
def memory_restart(c, build=False, frontend=False):
    """Fresh memory restart workflow: down -> db.reset -> up -> health -> memory endpoint probes."""
    run_memory_restart(c, build=build, frontend=frontend)


ns = Collection("lifecycle")
ns.add_task(status)
ns.add_task(up)
ns.add_task(down)
ns.add_task(health)
ns.add_task(restart)
ns.add_task(memory_restart, name="memory-restart")
