"""Bridge port-forwarding, database/memory restart, and stack status helpers for lifecycle tasks."""

from __future__ import annotations

from pathlib import Path

from . import lifecycle as _lc


def _core_council_ready(timeout: int = 10, interval: float = 1.0) -> bool:
    """Return True when council membership is reachable through the Core API."""
    api_key = _lc.os.environ.get("MYCELIS_API_KEY", "")
    if not api_key:
        _lc._load_env()
        api_key = _lc.os.environ.get("MYCELIS_API_KEY", "")
    headers = {"Authorization": f"Bearer {api_key}"} if api_key else {}
    url = f"http://{_lc.API_HOST}:{_lc.API_PORT}/api/v1/council/members"
    deadline = _lc.time.time() + timeout
    while _lc.time.time() < deadline:
        code, _body = _lc._http_get(url, timeout=5.0, headers=headers)
        if code == 200:
            return True
        _lc.time.sleep(interval)
    return False


def _core_startup_log_path() -> Path:
    """Return the log file used for background Core startup diagnostics."""
    return _lc.CORE_STARTUP_LOG


def _start_port_forward(svc: str, forward: str):
    """Start a kubectl port-forward in the background (detached)."""
    if _lc.is_windows():
        _lc.subprocess.Popen(
            ["kubectl", "port-forward", "-n", _lc.NAMESPACE, svc, forward],
            stdout=_lc.subprocess.DEVNULL,
            stderr=_lc.subprocess.DEVNULL,
            creationflags=_lc.subprocess.CREATE_NEW_PROCESS_GROUP | _lc.subprocess.DETACHED_PROCESS,
        )
    else:
        _lc.subprocess.Popen(
            ["kubectl", "port-forward", "-n", _lc.NAMESPACE, svc, forward],
            stdout=_lc.subprocess.DEVNULL,
            stderr=_lc.subprocess.DEVNULL,
            start_new_session=True,
        )


def _ensure_bridge():
    """Ensure port-forwards for PG and NATS are running. Idempotent."""
    for key in ("postgres", "nats"):
        svc = _lc.SERVICES[key]
        port = svc["port"]
        label = svc["label"]

        if _lc._port_open(port):
            print(f"  {label} bridge already active on :{port}")
            continue

        print(f"  Starting {label} port-forward (:{port})...")
        _lc._start_port_forward(svc["kind_svc"], svc["forward"])

        if not _lc._wait_for_port(port, label, timeout=30):
            print(f"  WARN: {label} bridge slow to start — Core will retry for 90s")


def _kill_bridges():
    """Kill kubectl port-forward processes for PG and NATS."""
    for key in ("postgres", "nats"):
        svc = _lc.SERVICES[key]
        _lc._kill_port(svc["port"], f"{svc['label']} bridge")


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
    load_dotenv(str(_lc.ROOT_DIR / ".env"), override=True)


def run_status(c):
    """Show health of every service in the dev stack."""
    print("=== Mycelis Stack Status ===\n")

    try:
        docker_result = c.run("docker version --format {{.Server.Version}}", hide=True, warn=True)
        if getattr(docker_result, "ok", None) is True or getattr(docker_result, "exited", 1) == 0:
            print("  Docker          : UP")
        else:
            print("  Docker          : DOWN")
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

    for _key, svc in _lc.SERVICES.items():
        port = svc["port"]
        label = svc["label"]
        alive = _lc._port_open(port)
        tag = "UP" if alive else "DOWN"
        pid_info = ""
        if alive:
            pid = _lc._find_pid_on_port(port)
            if pid:
                pid_info = f" (PID {pid})"
        print(f"  {label:<16}: {tag}{pid_info}  [:{port}]")

    print()
    if _lc._port_open(_lc.API_PORT):
        _lc._load_env()
        _api_key = _lc.os.environ.get("MYCELIS_API_KEY", "")
        code, _body = _lc._http_get(f"http://{_lc.API_HOST}:{_lc.API_PORT}/api/v1/cognitive/status")
        if code == 200:
            print("  Core API probe  : HEALTHY")
        elif code == 401:
            print("  Core API probe  : UP (auth required - expected)")
        else:
            print(f"  Core API probe  : ERROR ({code})")
    else:
        print("  Core API probe  : OFFLINE")

    try:
        compiled_go = _lc._list_compiled_go_service_processes()
    except RuntimeError as exc:
        print(f"  Compiled Go svc : UNKNOWN ({exc})")
    else:
        if compiled_go:
            summary = ", ".join(f"{proc['name']}:{proc['pid']}" for proc in compiled_go[:4])
            if len(compiled_go) > 4:
                summary += ", ..."
            print(f"  Compiled Go svc : DETECTED ({summary})")
        else:
            print("  Compiled Go svc : CLEAR")

    print()


def run_memory_restart(c, build=False, frontend=False):
    """Fresh memory restart workflow: down -> db.reset -> up -> health -> memory endpoint probes."""
    print("=== Mycelis Memory Fresh Restart ===\n")

    print("[1/6] Stopping stack...")
    _lc.down(c)
    print()
    _lc.time.sleep(2)

    print("[2/6] Restoring database bridge...")
    _lc._ensure_bridge()
    if not _lc._wait_for_port(5432, "PostgreSQL", timeout=30):
        raise SystemExit(
            "MEMORY RESTART FAILED: PostgreSQL bridge not reachable on 127.0.0.1:5432. "
            "Run 'uv run inv k8s.up' or 'uv run inv k8s.bridge' and retry."
        )
    print()

    print("[3/6] Resetting database...")
    _lc.db.reset(c)
    print()

    print("[4/6] Starting stack...")
    _lc.up(c, frontend=frontend, build=build)
    print()

    print("[5/6] Running stack health checks...")
    _lc.health(c)
    print()

    print("[6/6] Verifying memory endpoints...")
    _lc._load_env()
    api_key = _lc.os.environ.get("MYCELIS_API_KEY", "")
    headers = {"Authorization": f"Bearer {api_key}"} if api_key else {}
    base = f"http://{_lc.API_HOST}:{_lc.API_PORT}"

    checks = [
        ("/api/v1/memory/stream", "Memory Stream"),
        ("/api/v1/memory/sitreps?limit=1", "Memory SitReps"),
    ]
    errors: list[str] = []

    for path, label in checks:
        code, body = _lc._http_get(f"{base}{path}", timeout=5.0, headers=headers)
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
