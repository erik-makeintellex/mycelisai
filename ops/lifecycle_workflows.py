"""Stack up, down, and deep health workflow implementations for lifecycle tasks."""

from __future__ import annotations

from . import lifecycle as _lc


def run_up(c, frontend=False, build=False):
    """Bring up the full dev stack (idempotent)."""
    print("=== Mycelis Stack Up ===\n")

    deps_were_down_before_up = (not _lc._port_open(5432)) or (not _lc._port_open(4222))

    if build:
        print("[1/4] Building core binary...")
        _lc.core.compile.body(c)
    else:
        bin_path = _lc.CORE_DIR / ("bin/server.exe" if _lc.is_windows() else "bin/server")
        if not bin_path.exists():
            print("ERROR: No binary found. Run with --build or 'uv run inv core.compile' first.")
            return

    print("[1/4] Ensuring port-forwards...")
    _lc._ensure_bridge()
    print()

    print("[2/4] Waiting for dependencies...")
    pg_ok = _lc._wait_for_port(5432, "PostgreSQL", timeout=30)
    nats_ok = _lc._wait_for_port(4222, "NATS", timeout=30)

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

    print("[2.5/4] Ensuring cortex database exists...")
    _lc.db.create.body(c)
    print("  cortex database ready")
    print()

    print("[3/4] Starting Core server...")
    if _lc._port_open(_lc.API_PORT):
        print(f"  Core already running on :{_lc.API_PORT}")
        council_ready = _lc._core_council_ready(timeout=3, interval=0.5)
        if deps_were_down_before_up or not council_ready:
            print("  Restarting Core to exit degraded startup mode after dependency recovery...")
            _lc._kill_port(_lc.API_PORT, "Core")
            if _lc._start_core_background():
                if _lc._wait_for_port(_lc.API_PORT, "Core API", timeout=120):
                    print(f"  Core restarted on :{_lc.API_PORT}")
                else:
                    raise SystemExit(
                        "STACK UP FAILED: Core restart did not open its port in time. "
                        f"Check {_lc._core_startup_log_path()} or run 'uv run inv core.run'."
                    )
    else:
        if _lc._start_core_background():
            if _lc._wait_for_port(_lc.API_PORT, "Core API", timeout=120):
                print(f"  Core port open on :{_lc.API_PORT}")
            else:
                raise SystemExit(
                    "STACK UP FAILED: Core did not open its port in time. "
                    f"Check {_lc._core_startup_log_path()} or run 'uv run inv core.run'."
                )

    if _lc._wait_for_http_ok(f"http://{_lc.API_HOST}:{_lc.API_PORT}/healthz", "Core health", timeout=45):
        print(f"  Core healthy on :{_lc.API_PORT}")
        if not _lc._core_council_ready(timeout=20, interval=1.0):
            raise SystemExit(
                "STACK UP FAILED: Core health recovered but council routes are still offline. "
                f"Check {_lc._core_startup_log_path()} or rerun 'uv run inv lifecycle.up'."
            )
    else:
        raise SystemExit(
            "STACK UP FAILED: Core port opened but /healthz never became ready. "
            f"Check {_lc._core_startup_log_path()} or run 'uv run inv core.run'."
        )
    print()

    if frontend:
        print("[4/4] Starting Frontend...")
        if _lc._port_open(_lc.INTERFACE_PORT):
            print(f"  Frontend already running on :{_lc.INTERFACE_PORT}")
        else:
            _lc.interface.start_dev_server_detached(
                _lc.interface.interface_task_env(),
                host=_lc.INTERFACE_BIND_HOST,
                port=_lc.INTERFACE_PORT,
            )
            if _lc._wait_for_port(_lc.INTERFACE_PORT, "Frontend", timeout=30):
                print(f"  Frontend online on :{_lc.INTERFACE_PORT}")
            else:
                print("  WARN: Frontend did not come up in time.")
    else:
        print("[4/4] Frontend: skipped (use --frontend to include)")

    print("\nStack ready. Run 'uv run inv lifecycle.status' to verify.")


def run_down(_c):
    """Stop all dev stack services cleanly."""
    print("=== Mycelis Stack Down ===\n")

    print("[1/4] Stopping Core...")
    if not _lc._kill_port(_lc.API_PORT, "Core"):
        if _lc.is_windows():
            _lc._run_best_effort(
                [
                    "powershell",
                    "-NoProfile",
                    "-Command",
                    "Get-Process server -ErrorAction SilentlyContinue | Stop-Process -Force",
                ]
            )
        else:
            _lc._run_best_effort(["pkill", "-f", "bin/server"])
        _lc._wait_for_port_closed(_lc.API_PORT, "Core")
        print("  Core stopped (by name)")

    print("[2/4] Stopping Frontend...")
    if not _lc._kill_port(_lc.INTERFACE_PORT, "Frontend"):
        if _lc.is_windows():
            _lc._run_best_effort(
                [
                    "powershell",
                    "-NoProfile",
                    "-Command",
                    "Get-CimInstance Win32_Process -Filter \"Name = 'node.exe'\" | "
                    f"Where-Object {{ $_.CommandLine -match 'next (dev|start).*--port {_lc.INTERFACE_PORT}' }} | "
                    "ForEach-Object { Stop-Process -Id $_.ProcessId -Force }",
                ],
            )
        else:
            _lc._run_best_effort(["pkill", "-f", f"next (dev|start).*--port {_lc.INTERFACE_PORT}"])

        if _lc._wait_for_port_closed(_lc.INTERFACE_PORT, "Frontend"):
            print("  Frontend stopped (by command line)")
        else:
            print("  Frontend not running")

    remaining_interface = _lc.interface._cleanup_repo_local_interface_processes()
    if remaining_interface:
        summary = ", ".join(f"{proc['name']}:{proc['pid']}" for proc in remaining_interface[:6])
        raise SystemExit(
            "STACK DOWN INCOMPLETE: repo-local Interface residuals still running ("
            + summary
            + "). Re-run lifecycle.down or inspect Interface node workers."
        )

    print("[3/4] Cleaning compiled Go services...")
    remaining_compiled = _lc._kill_compiled_go_services()

    print("[4/4] Stopping port-forwards...")
    _lc._kill_bridges()

    if _lc.is_windows():
        _lc._run_best_effort(
            [
                "powershell",
                "-NoProfile",
                "-Command",
                "Get-Process kubectl -ErrorAction SilentlyContinue | Stop-Process -Force",
            ],
        )
    else:
        _lc._run_best_effort(["pkill", "-f", "kubectl port-forward"])

    remaining = []
    deadline = _lc.time.time() + 8
    while _lc.time.time() < deadline:
        remaining = _lc._remaining_managed_services()
        if not remaining:
            break
        _lc.time.sleep(0.5)
    if remaining:
        for key in _lc._service_keys_by_label(remaining):
            svc = _lc.SERVICES[key]
            _lc._kill_port(svc["port"], svc["label"])

        deadline = _lc.time.time() + 5
        while _lc.time.time() < deadline:
            remaining = _lc._remaining_managed_services()
            if not remaining:
                break
            _lc.time.sleep(0.5)
    if remaining:
        raise SystemExit(
            "STACK DOWN INCOMPLETE: remaining listeners on "
            + ", ".join(remaining)
            + ". Re-run lifecycle.down or inspect the reported ports/PIDs with lifecycle.status."
        )
    if remaining_compiled:
        compiled_summary = ", ".join(f"{proc['name']}:{proc['pid']}" for proc in remaining_compiled)
        raise SystemExit(
            "STACK DOWN INCOMPLETE: stray compiled Go services still running ("
            + compiled_summary
            + "). Re-run lifecycle.down or inspect the reported ports/PIDs with lifecycle.status."
        )

    print("\nAll services stopped.")


def run_health(_c):
    """Deep health check — probes actual API endpoints."""
    print("=== Mycelis Health Check ===\n")

    _lc._load_env()
    api_key = _lc.os.environ.get("MYCELIS_API_KEY", "")
    base = f"http://{_lc.API_HOST}:{_lc.API_PORT}"
    errors = []

    endpoints = [
        ("/api/v1/cognitive/status", "Cognitive Engine"),
        ("/api/v1/templates", "Template Engine"),
        ("/api/v1/brains", "Brains API"),
        ("/api/v1/telemetry/compute", "Telemetry"),
    ]

    for path, label in endpoints:
        headers = {"Authorization": f"Bearer {api_key}"} if api_key else {}
        status, body = _lc._http_get(f"{base}{path}", timeout=5.0, headers=headers)
        if status == 200:
            print(f"  [OK] {label:<20} {path} [200]")
        else:
            detail = body or f"HTTP {status}"
            print(f"  [FAIL] {label:<20} {path} — {detail}")
            errors.append(f"{label}: {detail}")

    print()
    if _lc._port_open(_lc.INTERFACE_PORT):
        code = 0
        for url in _lc._interface_probe_urls():
            code, body = _lc._http_get(url)
            if code == 200:
                print(f"  [OK] Frontend              {url} [200]")
                break
        else:
            print(f"  [WARN] Frontend            / [{code}]")
    else:
        print("  [SKIP] Frontend            (not running)")

    if _lc._port_open(11434):
        code, body = _lc._http_get("http://127.0.0.1:11434/api/tags")
        if code == 200:
            print("  [OK] Ollama                /api/tags [200]")
        else:
            print(f"  [WARN] Ollama              /api/tags [{code}]")
    else:
        print("  [SKIP] Ollama              (not running)")

    print()
    if errors:
        print(f"ISSUES: {len(errors)} problem(s)")
        for err in errors:
            print(f"  - {err}")
        raise SystemExit(1)
    print("ALL ENDPOINTS HEALTHY.")
