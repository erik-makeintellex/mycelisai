"""
Native development infrastructure helpers.

This module manages only repo-required local tools: PostgreSQL reachability,
the Mycelis application database/role, and a local NATS server. It deliberately
does not manage Docker, Rancher Desktop, WSL, or whole host environments.
"""

from __future__ import annotations

import os
import shutil
import socket
import subprocess
import time
import urllib.error
import urllib.request
from pathlib import Path

from invoke import Collection, task

from .config import ROOT_DIR, is_windows, managed_cache_env
from . import native_postgres


DEFAULT_NATS_VERSION = "v2.14.0"
NATS_DATA_DIR = ROOT_DIR / "workspace" / "native-infra" / "nats"
NATS_LOG_PATH = ROOT_DIR / "workspace" / "logs" / "nats-server.log"


def _load_env():
    try:
        from dotenv import load_dotenv
    except ModuleNotFoundError as exc:
        if exc.name != "dotenv":
            raise
        raise SystemExit(
            "Missing python-dotenv in the current invoke environment. "
            "Run tasks with 'uv run inv ...'."
        ) from exc
    load_dotenv(str(ROOT_DIR / ".env"), override=True)


def _env(name: str, default: str) -> str:
    _load_env()
    return os.environ.get(name, default)


def _port_open(port: int, host: str = "127.0.0.1", timeout: float = 1.0) -> bool:
    try:
        with socket.create_connection((host, port), timeout=timeout):
            return True
    except (ConnectionRefusedError, TimeoutError, OSError):
        return False


def _wait_for_port(port: int, label: str, timeout: int = 30) -> bool:
    deadline = time.time() + timeout
    while time.time() < deadline:
        if _port_open(port):
            return True
        time.sleep(0.5)
    print(f"  TIMEOUT waiting for {label} on :{port} ({timeout}s)")
    return False


def _http_get(url: str, timeout: float = 2.0) -> tuple[int, str]:
    try:
        with urllib.request.urlopen(url, timeout=timeout) as resp:
            return resp.status, resp.read().decode("utf-8", errors="replace")
    except urllib.error.HTTPError as exc:
        return exc.code, str(exc)
    except Exception as exc:
        return 0, str(exc)


def _go_bin_dir() -> Path | None:
    try:
        result = subprocess.run(
            ["go", "env", "GOBIN", "GOPATH"],
            capture_output=True,
            text=True,
            timeout=15,
        )
    except (OSError, subprocess.SubprocessError):
        return None
    if result.returncode != 0:
        return None
    lines = [line.strip() for line in result.stdout.splitlines()]
    gobin = Path(lines[0]) if lines and lines[0] else None
    if gobin:
        return gobin
    if len(lines) > 1 and lines[1]:
        return Path(lines[1]) / "bin"
    return None


def _nats_executable() -> str | None:
    found = shutil.which("nats-server")
    if found:
        return found
    go_bin = _go_bin_dir()
    if not go_bin:
        return None
    candidate = go_bin / ("nats-server.exe" if is_windows() else "nats-server")
    if candidate.exists():
        return str(candidate)
    return None


def bootstrap_postgres_database() -> bool:
    return native_postgres.bootstrap_database()


def ensure_nats_running(timeout: int = 30) -> bool:
    """Start a local NATS server if :4222 is not already listening."""
    if _port_open(4222):
        print("  NATS already active on :4222")
        return True

    exe = _nats_executable()
    if not exe:
        print("  NATS server binary not found.")
        print("  Run: uv run inv native-infra.install-nats")
        return False

    NATS_DATA_DIR.mkdir(parents=True, exist_ok=True)
    NATS_LOG_PATH.parent.mkdir(parents=True, exist_ok=True)
    with NATS_LOG_PATH.open("a", encoding="utf-8") as log_file:
        command = [exe, "-js", "-sd", str(NATS_DATA_DIR), "-m", "8222"]
        if is_windows():
            subprocess.Popen(
                command,
                cwd=str(ROOT_DIR),
                stdin=subprocess.DEVNULL,
                stdout=log_file,
                stderr=subprocess.STDOUT,
                creationflags=getattr(subprocess, "CREATE_NEW_PROCESS_GROUP", 0)
                | getattr(subprocess, "DETACHED_PROCESS", 0),
            )
        else:
            subprocess.Popen(
                command,
                cwd=str(ROOT_DIR),
                stdin=subprocess.DEVNULL,
                stdout=log_file,
                stderr=subprocess.STDOUT,
                start_new_session=True,
            )
    return _wait_for_port(4222, "NATS", timeout=timeout)


def ensure_for_lifecycle(timeout: int = 30):
    """Prepare source-mode dependencies for lifecycle.up without owning host services."""
    if _port_open(5432):
        print("  PostgreSQL native dependency active on :5432")
        if bootstrap_postgres_database():
            print("  PostgreSQL app role/database ready")
        else:
            print("  WARN: PostgreSQL is reachable, but app bootstrap failed.")
    else:
        print("  PostgreSQL is not reachable on :5432.")
        print("  Start the local PostgreSQL service, then run: uv run inv native-infra.bootstrap-postgres")
    if not ensure_nats_running(timeout=timeout):
        print("  WARN: NATS native dependency is not ready.")


@task(help={"version": f"NATS server version to install with go install (default: {DEFAULT_NATS_VERSION})"})
def install_nats(c, version=DEFAULT_NATS_VERSION):
    """Install nats-server into the local Go bin directory."""
    print(f"Installing nats-server {version} with Go...")
    env = managed_cache_env()
    c.run(f"go install github.com/nats-io/nats-server/v2@{version}", env=env)
    exe = _nats_executable()
    if not exe:
        go_bin = _go_bin_dir()
        if go_bin:
            print(f"Installed, but {go_bin} is not on PATH for this shell.")
        raise SystemExit("nats-server was installed but could not be resolved from PATH.")
    print(f"nats-server ready: {exe}")


@task
def start_nats(_c):
    """Start local NATS with JetStream and the HTTP monitor."""
    if not ensure_nats_running(timeout=30):
        raise SystemExit(f"NATS did not start. Check {NATS_LOG_PATH}.")
    print("NATS ready on nats://127.0.0.1:4222; monitor http://127.0.0.1:8222")


@task
def bootstrap_postgres(_c):
    """Create/update the native PostgreSQL app role and database."""
    if not bootstrap_postgres_database():
        raise SystemExit("Native PostgreSQL bootstrap failed.")
    print("Native PostgreSQL app role/database ready.")


@task
def up(c, install_nats_if_missing=False):
    """Bring up native development dependencies only: PostgreSQL app DB and NATS."""
    if install_nats_if_missing and not _nats_executable():
        install_nats.body(c)
    print("=== Native Infrastructure Up ===")
    if not bootstrap_postgres_database():
        raise SystemExit("PostgreSQL is not ready for Mycelis.")
    print("  PostgreSQL app database ready")
    if not ensure_nats_running(timeout=30):
        raise SystemExit("NATS is not ready for Mycelis.")
    print("  NATS ready")


@task
def status(_c):
    """Show native PostgreSQL and NATS readiness."""
    _load_env()
    db_host = os.environ.get("DB_HOST", "127.0.0.1")
    db_port = int(os.environ.get("DB_PORT", "5432"))
    db_user = os.environ.get("DB_USER", "mycelis")
    db_password = os.environ.get("DB_PASSWORD", "")
    db_name = os.environ.get("DB_NAME", "cortex")

    print("=== Native Infrastructure Status ===")
    print(f"  PostgreSQL port : {'UP' if _port_open(db_port, db_host) else 'DOWN'}  [{db_host}:{db_port}]")
    if db_password:
        result = native_postgres.psql_command("SELECT 1;", database=db_name, user=db_user, password=db_password)
        print(f"  PostgreSQL query: {'READY' if result.returncode == 0 else 'FAILED'}  [database={db_name}]")
    else:
        print("  PostgreSQL query: SKIPPED  [DB_PASSWORD missing]")

    print(f"  NATS port       : {'UP' if _port_open(4222) else 'DOWN'}  [127.0.0.1:4222]")
    code, _body = _http_get("http://127.0.0.1:8222/healthz")
    print(f"  NATS monitor    : {'READY' if code == 200 else 'DOWN'}  [http://127.0.0.1:8222/healthz]")
    exe = _nats_executable()
    print(f"  nats-server bin : {exe or 'NOT FOUND'}")


ns = Collection("native-infra")
ns.add_task(install_nats, "install-nats")
ns.add_task(start_nats, "start-nats")
ns.add_task(bootstrap_postgres, "bootstrap-postgres")
ns.add_task(up)
ns.add_task(status)
