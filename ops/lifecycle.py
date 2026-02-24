"""
Lifecycle management for the Mycelis development stack.

Provides unified start/stop/status/health commands that handle the full
dependency graph: port-forwards → core server → frontend.

All probes are Python-native (socket/urllib) — no shell wrappers.
"""

import os
import socket
import subprocess
import time
from pathlib import Path

from invoke import task, Collection
from .config import (
    ROOT_DIR,
    CORE_DIR,
    NAMESPACE,
    API_HOST,
    API_PORT,
    INTERFACE_HOST,
    INTERFACE_PORT,
    is_windows,
    powershell,
)


# ── Port / Service Definitions ───────────────────────────────────────

SERVICES = {
    "postgres": {"port": 5432, "label": "PostgreSQL", "kind_svc": f"svc/mycelis-core-postgresql", "forward": "5432:5432"},
    "nats":     {"port": 4222, "label": "NATS",       "kind_svc": f"svc/mycelis-core-nats",       "forward": "4222:4222"},
    "core":     {"port": API_PORT,       "label": "Core API"},
    "frontend": {"port": INTERFACE_PORT, "label": "Frontend"},
    "ollama":   {"port": 11434,          "label": "Ollama"},
}


# ── Low-Level Probes ─────────────────────────────────────────────────

def _port_open(port: int, host: str = "127.0.0.1", timeout: float = 1.0) -> bool:
    """Check if a TCP port is accepting connections."""
    try:
        with socket.create_connection((host, port), timeout=timeout):
            return True
    except (ConnectionRefusedError, TimeoutError, OSError):
        return False


def _http_get(url: str, timeout: float = 3.0) -> tuple[int, str]:
    """HTTP GET returning (status_code, body). Returns (0, error) on failure."""
    import urllib.request
    import urllib.error
    try:
        req = urllib.request.Request(url)
        with urllib.request.urlopen(req, timeout=timeout) as resp:
            return resp.status, resp.read().decode("utf-8", errors="replace")
    except urllib.error.HTTPError as e:
        return e.code, str(e)
    except Exception as e:
        return 0, str(e)


def _wait_for_port(port: int, label: str, timeout: int = 30, interval: float = 1.0) -> bool:
    """Block until a port is open or timeout expires. Returns True on success."""
    deadline = time.time() + timeout
    while time.time() < deadline:
        if _port_open(port):
            return True
        time.sleep(interval)
    print(f"  TIMEOUT waiting for {label} on port {port} ({timeout}s)")
    return False


def _find_pid_on_port(port: int) -> int | None:
    """Find the PID listening on a given port. Returns None if not found."""
    if is_windows():
        try:
            # Use netstat which works reliably in all Windows shells
            result = subprocess.run(
                ["netstat", "-ano"],
                capture_output=True, text=True, timeout=5,
            )
            for line in result.stdout.splitlines():
                # Match LISTENING state on our port
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
            result = subprocess.run(
                ["lsof", "-ti", f":{port}"],
                capture_output=True, text=True, timeout=5,
            )
            pid_str = result.stdout.strip().split("\n")[0]
            if pid_str.isdigit():
                return int(pid_str)
        except Exception:
            pass
    return None


def _kill_pid(pid: int):
    """Kill a process by PID."""
    if is_windows():
        subprocess.run(["taskkill", "/F", "/PID", str(pid)],
                       capture_output=True, timeout=5)
    else:
        subprocess.run(["kill", "-9", str(pid)],
                       capture_output=True, timeout=5)


def _kill_port(port: int, label: str) -> bool:
    """Kill whatever is listening on the given port. Returns True if killed."""
    pid = _find_pid_on_port(port)
    if pid:
        _kill_pid(pid)
        print(f"  Killed {label} (PID {pid}) on port {port}")
        return True
    return False


# ── Port-Forward Management ──────────────────────────────────────────

def _start_port_forward(svc: str, forward: str):
    """Start a kubectl port-forward in the background (detached)."""
    if is_windows():
        # Use cmd /c start to launch detached on Windows (kubectl on Windows PATH)
        subprocess.Popen(
            ["cmd", "/c", "start", "/B", "kubectl", "port-forward", "-n", NAMESPACE, svc, forward],
            stdout=subprocess.DEVNULL,
            stderr=subprocess.DEVNULL,
            creationflags=subprocess.CREATE_NEW_PROCESS_GROUP | subprocess.DETACHED_PROCESS,
        )
    else:
        subprocess.Popen(
            ["kubectl", "port-forward", "-n", NAMESPACE, svc, forward],
            stdout=subprocess.DEVNULL,
            stderr=subprocess.DEVNULL,
            start_new_session=True,
        )


def _ensure_bridge():
    """Ensure port-forwards for PG and NATS are running. Idempotent."""
    for key in ("postgres", "nats"):
        svc = SERVICES[key]
        port = svc["port"]
        label = svc["label"]

        if _port_open(port):
            print(f"  {label} bridge already active on :{port}")
            continue

        print(f"  Starting {label} port-forward (:{port})...")
        _start_port_forward(svc["kind_svc"], svc["forward"])

        if not _wait_for_port(port, label, timeout=30):
            print(f"  WARN: {label} bridge slow to start — Core will retry for 90s")


def _kill_bridges():
    """Kill kubectl port-forward processes for PG and NATS."""
    for key in ("postgres", "nats"):
        svc = SERVICES[key]
        _kill_port(svc["port"], f"{svc['label']} bridge")


# ── Core Server Management ───────────────────────────────────────────

def _load_env():
    """Load .env into the process environment."""
    from dotenv import load_dotenv
    load_dotenv(str(ROOT_DIR / ".env"), override=True)


def _start_core_background():
    """Start the core server binary in the background."""
    _load_env()
    bin_path = CORE_DIR / ("bin/server.exe" if is_windows() else "bin/server")
    if not bin_path.exists():
        print(f"  ERROR: Binary not found at {bin_path}. Run 'uvx inv core.build' first.")
        return False

    env = {**os.environ, "PYTHONIOENCODING": "utf-8"}

    if is_windows():
        subprocess.Popen(
            [str(bin_path)],
            cwd=str(CORE_DIR),
            env=env,
            stdout=subprocess.DEVNULL,
            stderr=subprocess.DEVNULL,
            creationflags=subprocess.CREATE_NEW_PROCESS_GROUP | subprocess.DETACHED_PROCESS,
        )
    else:
        subprocess.Popen(
            [str(bin_path)],
            cwd=str(CORE_DIR),
            env=env,
            stdout=subprocess.DEVNULL,
            stderr=subprocess.DEVNULL,
            start_new_session=True,
        )
    return True


# ── Tasks ────────────────────────────────────────────────────────────

@task
def status(c):
    """
    Show health of every service in the dev stack.
    Checks: Docker, Kind cluster, port-forwards, core, frontend, Ollama.
    """
    print("=== Mycelis Stack Status ===\n")

    # Docker / Kind — use invoke's c.run() which resolves PATH correctly
    try:
        c.run("docker version --format {{.Server.Version}}", hide=True, warn=True)
        print("  Docker          : UP")
    except Exception:
        print("  Docker          : DOWN")

    try:
        result = c.run("kind get clusters", hide=True, warn=True)
        if result and "mycelis-cluster" in result.stdout:
            print("  Kind Cluster    : UP")
        else:
            print("  Kind Cluster    : NOT FOUND")
    except Exception:
        print("  Kind Cluster    : ERROR")

    print()

    # Service ports
    for key, svc in SERVICES.items():
        port = svc["port"]
        label = svc["label"]
        alive = _port_open(port)
        tag = "UP" if alive else "DOWN"
        pid_info = ""
        if alive:
            pid = _find_pid_on_port(port)
            if pid:
                pid_info = f" (PID {pid})"
        print(f"  {label:<16}: {tag}{pid_info}  [:{port}]")

    # Deep probe: Core API health
    print()
    if _port_open(API_PORT):
        _load_env()
        api_key = os.environ.get("MYCELIS_API_KEY", "")
        code, body = _http_get(f"http://{API_HOST}:{API_PORT}/api/v1/cognitive/status")
        if code == 200:
            print("  Core API probe  : HEALTHY")
        elif code == 401:
            print("  Core API probe  : UP (auth required - expected)")
        else:
            print(f"  Core API probe  : ERROR ({code})")
    else:
        print("  Core API probe  : OFFLINE")

    print()


@task(
    help={
        "frontend": "Also start the frontend dev server (default: False)",
        "build": "Build the Go binary before starting (default: False)",
    }
)
def up(c, frontend=False, build=False):
    """
    Bring up the full dev stack (idempotent).
    Order: port-forwards → core server → (optional) frontend.
    """
    print("=== Mycelis Stack Up ===\n")

    # 0. Optionally build
    if build:
        print("[1/4] Building core binary...")
        from .core import build as core_build
        core_build(c)
    else:
        bin_path = CORE_DIR / ("bin/server.exe" if is_windows() else "bin/server")
        if not bin_path.exists():
            print("ERROR: No binary found. Run with --build or 'uvx inv core.build' first.")
            return

    # 1. Port-forwards
    print("[1/4] Ensuring port-forwards...")
    _ensure_bridge()
    print()

    # 2. Wait for dependencies — Core will also retry internally (up to 90s each),
    #    so these timeouts are just for the console status message, not hard gates.
    print("[2/4] Waiting for dependencies...")
    pg_ok = _wait_for_port(5432, "PostgreSQL", timeout=30)
    nats_ok = _wait_for_port(4222, "NATS", timeout=30)

    if not pg_ok:
        print("  WARN: PostgreSQL not yet reachable — Core will retry for 90s after start.")
        print("        If this persists, check: uvx inv k8s.status")
    else:
        print("  PostgreSQL ready")

    if not nats_ok:
        print("  WARN: NATS not yet reachable — Core will retry for 90s after start.")
        print("        Real-time features will activate automatically when NATS connects.")
    else:
        print("  NATS ready")
    print()

    # 3. Core server — allow up to 120s since Core waits up to 90s for its own deps
    print("[3/4] Starting Core server...")
    if _port_open(API_PORT):
        print(f"  Core already running on :{API_PORT}")
    else:
        if _start_core_background():
            if _wait_for_port(API_PORT, "Core API", timeout=120):
                print(f"  Core online on :{API_PORT}")
            else:
                print("  WARN: Core did not come up in time. Check logs with: uvx inv core.run")
    print()

    # 4. Frontend (optional)
    if frontend:
        print("[4/4] Starting Frontend...")
        if _port_open(INTERFACE_PORT):
            print(f"  Frontend already running on :{INTERFACE_PORT}")
        else:
            from .interface import dev as interface_dev
            # Run in background via subprocess
            if is_windows():
                subprocess.Popen(
                    ["cmd", "/c", f"cd interface && npx next dev --port {INTERFACE_PORT}"],
                    cwd=str(ROOT_DIR),
                    stdout=subprocess.DEVNULL,
                    stderr=subprocess.DEVNULL,
                    creationflags=subprocess.CREATE_NEW_PROCESS_GROUP | subprocess.DETACHED_PROCESS,
                )
            else:
                subprocess.Popen(
                    ["npx", "next", "dev", "--port", str(INTERFACE_PORT)],
                    cwd=str(ROOT_DIR / "interface"),
                    stdout=subprocess.DEVNULL,
                    stderr=subprocess.DEVNULL,
                    start_new_session=True,
                )
            if _wait_for_port(INTERFACE_PORT, "Frontend", timeout=30):
                print(f"  Frontend online on :{INTERFACE_PORT}")
            else:
                print("  WARN: Frontend did not come up in time.")
    else:
        print("[4/4] Frontend: skipped (use --frontend to include)")

    print("\nStack ready. Run 'uvx inv lifecycle.status' to verify.")


@task
def down(c):
    """
    Stop all dev stack services cleanly.
    Order: core → frontend → port-forwards.
    """
    print("=== Mycelis Stack Down ===\n")

    # 1. Core
    print("[1/3] Stopping Core...")
    if _kill_port(API_PORT, "Core"):
        pass
    else:
        # Fallback: kill by process name
        if is_windows():
            subprocess.run(["taskkill", "/F", "/IM", "server.exe"],
                           capture_output=True)
        else:
            subprocess.run(["pkill", "-f", "bin/server"], capture_output=True)
        print("  Core stopped (by name)")

    # 2. Frontend
    print("[2/3] Stopping Frontend...")
    if not _kill_port(INTERFACE_PORT, "Frontend"):
        print("  Frontend not running")

    # 3. Port-forwards
    print("[3/3] Stopping port-forwards...")
    _kill_bridges()

    # Also kill any remaining kubectl port-forward processes
    if is_windows():
        subprocess.run(
            ["powershell", "-NoProfile", "-Command",
             "Get-Process kubectl -ErrorAction SilentlyContinue | "
             "Where-Object { $_.CommandLine -match 'port-forward' } | "
             "Stop-Process -Force"],
            capture_output=True,
        )
    else:
        subprocess.run(["pkill", "-f", "kubectl port-forward"],
                       capture_output=True)

    print("\nAll services stopped.")


@task
def health(c):
    """
    Deep health check — probes actual API endpoints.
    Requires core server to be running.
    """
    print("=== Mycelis Health Check ===\n")

    _load_env()
    api_key = os.environ.get("MYCELIS_API_KEY", "")
    base = f"http://{API_HOST}:{API_PORT}"
    errors = []

    endpoints = [
        ("/api/v1/cognitive/status", "Cognitive Engine"),
        ("/api/v1/templates", "Template Engine"),
        ("/api/v1/brains", "Brains API"),
        ("/api/v1/telemetry/compute", "Telemetry"),
    ]

    for path, label in endpoints:
        import urllib.request
        try:
            req = urllib.request.Request(
                f"{base}{path}",
                headers={"Authorization": f"Bearer {api_key}"},
            )
            with urllib.request.urlopen(req, timeout=5) as resp:
                status = resp.status
                icon = "OK" if status == 200 else "WARN"
                print(f"  [{icon}] {label:<20} {path} [{status}]")
                if status != 200:
                    errors.append(f"{label}: HTTP {status}")
        except Exception as e:
            print(f"  [FAIL] {label:<20} {path} — {e}")
            errors.append(f"{label}: {e}")

    # Frontend check
    print()
    if _port_open(INTERFACE_PORT):
        code, body = _http_get(f"http://{INTERFACE_HOST}:{INTERFACE_PORT}/")
        if code == 200:
            print(f"  [OK] Frontend              / [200]")
        else:
            print(f"  [WARN] Frontend            / [{code}]")
    else:
        print(f"  [SKIP] Frontend            (not running)")

    # Ollama check
    if _port_open(11434):
        code, body = _http_get("http://127.0.0.1:11434/api/tags")
        if code == 200:
            print(f"  [OK] Ollama                /api/tags [200]")
        else:
            print(f"  [WARN] Ollama              /api/tags [{code}]")
    else:
        print(f"  [SKIP] Ollama              (not running)")

    print()
    if errors:
        print(f"ISSUES: {len(errors)} problem(s)")
        for e in errors:
            print(f"  - {e}")
    else:
        print("ALL ENDPOINTS HEALTHY.")


@task
def restart(c, build=False, frontend=False):
    """
    Full stack restart: down → (optional build) → up.
    """
    down(c)
    print()
    time.sleep(2)  # brief settle
    up(c, frontend=frontend, build=build)


# ── Collection ───────────────────────────────────────────────────────

ns = Collection("lifecycle")
ns.add_task(status)
ns.add_task(up)
ns.add_task(down)
ns.add_task(health)
ns.add_task(restart)
