"""
Lifecycle management for the Mycelis development stack.

Provides unified start/stop/status/health commands that handle the full
dependency graph: port-forwards -> core server -> frontend.

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

CORE_STARTUP_LOG = ROOT_DIR / "workspace" / "logs" / "core-startup.log"


# ── Low-Level Probes ─────────────────────────────────────────────────

def _port_open(port: int, host: str = "127.0.0.1", timeout: float = 1.0) -> bool:
    """Check if a TCP port is accepting connections."""
    try:
        with socket.create_connection((host, port), timeout=timeout):
            return True
    except (ConnectionRefusedError, TimeoutError, OSError):
        return False


def _http_get(url: str, timeout: float = 3.0, headers: dict[str, str] | None = None) -> tuple[int, str]:
    """HTTP GET returning (status_code, body). Returns (0, error) on failure."""
    import urllib.request
    import urllib.error
    try:
        req = urllib.request.Request(url, headers=headers or {})
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


def _core_startup_log_path() -> Path:
    """Return the log file used for background Core startup diagnostics."""
    return CORE_STARTUP_LOG


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
    """Kill a process by PID without failing the workflow on slow OS cleanup."""
    try:
        if is_windows():
            subprocess.run(
                ["powershell", "-NoProfile", "-Command", f"Stop-Process -Id {pid} -Force"],
                capture_output=True,
                timeout=5,
            )
        else:
            subprocess.run(["kill", "-9", str(pid)],
                           capture_output=True, timeout=5)
    except subprocess.TimeoutExpired:
        # Windows taskkill can hang after the target is already terminating.
        # Treat this as best-effort and let the caller re-check the port.
        pass


def _run_best_effort(cmd: list[str], timeout: int = 5):
    """Run a cleanup command without letting a hung subprocess block the workflow."""
    try:
        subprocess.run(cmd, capture_output=True, timeout=timeout)
    except subprocess.TimeoutExpired:
        pass


def _remaining_managed_services() -> list[str]:
    """Return managed service labels that still have listening ports."""
    return [
        svc["label"]
        for key, svc in SERVICES.items()
        if key in ("postgres", "nats", "core", "frontend") and _port_open(svc["port"])
    ]


def _service_keys_by_label(labels: list[str]) -> list[str]:
    keys: list[str] = []
    for key, svc in SERVICES.items():
        if key in ("postgres", "nats", "core", "frontend") and svc["label"] in labels:
            keys.append(key)
    return keys


def _kill_port(port: int, label: str) -> bool:
    """Kill whatever is listening on the given port. Returns True if killed."""
    pid = _find_pid_on_port(port)
    if pid:
        _kill_pid(pid)
        if not _wait_for_port_closed(port, label):
            return False
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
    try:
        from dotenv import load_dotenv
    except ModuleNotFoundError as exc:
        if exc.name != "dotenv":
            raise
        raise SystemExit(
            "Missing python-dotenv in the current invoke environment. "
            "Run tasks with 'uv run inv ...' or '.\\.venv\\Scripts\\inv.exe ...'; "
            "do not use 'uvx --from invoke inv ...'."
        ) from exc
    load_dotenv(str(ROOT_DIR / ".env"), override=True)


def _start_core_background():
    """Start the core server binary in the background."""
    _load_env()
    bin_path = CORE_DIR / ("bin/server.exe" if is_windows() else "bin/server")
    if not bin_path.exists():
        print(f"  ERROR: Binary not found at {bin_path}. Run 'uv run inv core.build' first.")
        return False

    env = {**os.environ, "PYTHONIOENCODING": "utf-8"}
    log_path = _core_startup_log_path()
    log_path.parent.mkdir(parents=True, exist_ok=True)

    with log_path.open("w", encoding="utf-8") as log_file:
        if is_windows():
            subprocess.Popen(
                [str(bin_path)],
                cwd=str(CORE_DIR),
                env=env,
                stdin=subprocess.DEVNULL,
                stdout=log_file,
                stderr=subprocess.STDOUT,
                creationflags=subprocess.CREATE_NEW_PROCESS_GROUP | subprocess.DETACHED_PROCESS,
            )
        else:
            subprocess.Popen(
                [str(bin_path)],
                cwd=str(CORE_DIR),
                env=env,
                stdin=subprocess.DEVNULL,
                stdout=log_file,
                stderr=subprocess.STDOUT,
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
    Order: port-forwards -> core server -> (optional) frontend.
    """
    print("=== Mycelis Stack Up ===\n")

    # Capture dependency state before we touch bridges. If Core is already up while one
    # of these is down, Core is likely running in degraded mode and should be restarted
    # after dependencies are healthy.
    deps_were_down_before_up = (not _port_open(5432)) or (not _port_open(4222))

    # 0. Optionally build
    if build:
        print("[1/4] Building core binary...")
        from .core import build as core_build
        core_build(c)
    else:
        bin_path = CORE_DIR / ("bin/server.exe" if is_windows() else "bin/server")
        if not bin_path.exists():
            print("ERROR: No binary found. Run with --build or 'uv run inv core.build' first.")
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
        print("        If this persists, check: uv run inv k8s.status")
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
        if deps_were_down_before_up:
            print("  Restarting Core to exit degraded startup mode after dependency recovery...")
            _kill_port(API_PORT, "Core")
            if _start_core_background():
                if _wait_for_port(API_PORT, "Core API", timeout=120):
                    print(f"  Core restarted on :{API_PORT}")
                else:
                    raise SystemExit(
                        "STACK UP FAILED: Core restart did not open its port in time. "
                        f"Check {_core_startup_log_path()} or run 'uv run inv core.run'."
                    )
    else:
        if _start_core_background():
            if _wait_for_port(API_PORT, "Core API", timeout=120):
                print(f"  Core port open on :{API_PORT}")
            else:
                raise SystemExit(
                    "STACK UP FAILED: Core did not open its port in time. "
                    f"Check {_core_startup_log_path()} or run 'uv run inv core.run'."
                )

    if _wait_for_http_ok(f"http://{API_HOST}:{API_PORT}/healthz", "Core health", timeout=45):
        print(f"  Core healthy on :{API_PORT}")
    else:
        raise SystemExit(
            "STACK UP FAILED: Core port opened but /healthz never became ready. "
            f"Check {_core_startup_log_path()} or run 'uv run inv core.run'."
        )
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

    print("\nStack ready. Run 'uv run inv lifecycle.status' to verify.")


@task
def down(c):
    """
    Stop all dev stack services cleanly.
    Order: core -> frontend -> port-forwards.
    """
    print("=== Mycelis Stack Down ===\n")

    # 1. Core
    print("[1/3] Stopping Core...")
    if _kill_port(API_PORT, "Core"):
        pass
    else:
        # Fallback: kill by process name
        if is_windows():
            _run_best_effort(
                [
                    "powershell",
                    "-NoProfile",
                    "-Command",
                    "Get-Process server -ErrorAction SilentlyContinue | Stop-Process -Force",
                ]
            )
        else:
            _run_best_effort(["pkill", "-f", "bin/server"])
        _wait_for_port_closed(API_PORT, "Core")
        print("  Core stopped (by name)")

    # 2. Frontend
    print("[2/3] Stopping Frontend...")
    if not _kill_port(INTERFACE_PORT, "Frontend"):
        if is_windows():
            _run_best_effort(
                [
                    "powershell",
                    "-NoProfile",
                    "-Command",
                    "Get-CimInstance Win32_Process -Filter \"Name = 'node.exe'\" | "
                    "Where-Object { $_.CommandLine -match 'next dev --port 3000' } | "
                    "ForEach-Object { Stop-Process -Id $_.ProcessId -Force }",
                ],
            )
        else:
            _run_best_effort(["pkill", "-f", "next dev --port 3000"])

        if _wait_for_port_closed(INTERFACE_PORT, "Frontend"):
            print("  Frontend stopped (by command line)")
        else:
            print("  Frontend not running")

    # 3. Port-forwards
    print("[3/3] Stopping port-forwards...")
    _kill_bridges()

    # Also kill any remaining kubectl port-forward processes
    if is_windows():
        _run_best_effort(
            ["powershell", "-NoProfile", "-Command",
             "Get-Process kubectl -ErrorAction SilentlyContinue | "
             "Where-Object { $_.CommandLine -match 'port-forward' } | "
             "Stop-Process -Force"],
        )
    else:
        _run_best_effort(["pkill", "-f", "kubectl port-forward"])

    remaining = []
    deadline = time.time() + 8
    while time.time() < deadline:
        remaining = _remaining_managed_services()
        if not remaining:
            break
        time.sleep(0.5)
    if remaining:
        # Windows can report the listener a moment after the first kill wave; retry by port.
        for key in _service_keys_by_label(remaining):
            svc = SERVICES[key]
            _kill_port(svc["port"], svc["label"])

        deadline = time.time() + 5
        while time.time() < deadline:
            remaining = _remaining_managed_services()
            if not remaining:
                break
            time.sleep(0.5)
    if remaining:
        raise SystemExit(
            "STACK DOWN INCOMPLETE: remaining listeners on "
            + ", ".join(remaining)
            + ". Re-run lifecycle.down or inspect the reported ports/PIDs with lifecycle.status."
        )

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
    Full stack restart: down -> (optional build) -> up.
    """
    down(c)
    print()
    time.sleep(2)  # brief settle
    up(c, frontend=frontend, build=build)


@task(
    help={
        "build": "Build the Go binary before restart (default: False).",
        "frontend": "Also start frontend after restart (default: False).",
    }
)
def memory_restart(c, build=False, frontend=False):
    """
    Fresh memory restart workflow:
    down -> db.reset -> up -> health -> memory endpoint probes.
    """
    print("=== Mycelis Memory Fresh Restart ===\n")

    print("[1/6] Stopping stack...")
    down(c)
    print()
    time.sleep(2)

    print("[2/6] Restoring database bridge...")
    _ensure_bridge()
    if not _wait_for_port(5432, "PostgreSQL", timeout=30):
        raise SystemExit(
            "MEMORY RESTART FAILED: PostgreSQL bridge not reachable on 127.0.0.1:5432. "
            "Run 'uv run inv k8s.up' or 'uv run inv k8s.bridge' and retry."
        )
    print()

    print("[3/6] Resetting database...")
    from . import db as db_tasks
    db_tasks.reset(c)
    print()

    print("[4/6] Starting stack...")
    up(c, frontend=frontend, build=build)
    print()

    print("[5/6] Running stack health checks...")
    health(c)
    print()

    print("[6/6] Verifying memory endpoints...")
    _load_env()
    api_key = os.environ.get("MYCELIS_API_KEY", "")
    headers = {"Authorization": f"Bearer {api_key}"} if api_key else {}
    base = f"http://{API_HOST}:{API_PORT}"

    checks = [
        ("/api/v1/memory/stream", "Memory Stream"),
        ("/api/v1/memory/sitreps?limit=1", "Memory SitReps"),
    ]
    errors: list[str] = []

    for path, label in checks:
        code, body = _http_get(f"{base}{path}", timeout=5.0, headers=headers)
        if code == 200:
            print(f"  [OK] {label:<20} {path} [200]")
        else:
            print(f"  [FAIL] {label:<20} {path} [{code}]")
            errors.append(f"{label}: HTTP {code} ({body[:120]})")

    if errors:
        print()
        for item in errors:
            print(f"  - {item}")
        raise SystemExit("MEMORY RESTART FAILED: one or more memory probes failed.")

    print()
    print("MEMORY RESTART READY.")


# ── Collection ───────────────────────────────────────────────────────

ns = Collection("lifecycle")
ns.add_task(status)
ns.add_task(up)
ns.add_task(down)
ns.add_task(health)
ns.add_task(restart)
ns.add_task(memory_restart, name="memory-restart")
