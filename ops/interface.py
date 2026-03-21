import json
import os
import subprocess
import time
import urllib.error
import urllib.request

from invoke import task, Collection
from .config import (
    INTERFACE_HOST,
    INTERFACE_PORT,
    ROOT_DIR,
    ensure_managed_cache_dirs,
    is_windows,
    managed_cache_env,
    powershell,
)

ns = Collection("interface")
INTERFACE_DIR = ROOT_DIR / "interface"
_INTERFACE_PROCESS_PATH_HINTS = tuple(
    str(path.resolve()).lower().replace("\\", "/")
    for path in {
        INTERFACE_DIR / ".next",
        INTERFACE_DIR / "node_modules",
        INTERFACE_DIR / "playwright-report",
        INTERFACE_DIR / "test-results",
    }
)
_INTERFACE_PROCESS_COMMAND_HINTS = (
    "/.next/dev/build/postcss.js",
    "/next/dist/bin/next",
    "/node_modules/.bin/vitest",
    "/node_modules/vitest/",
    "/playwright/",
)

def _load_env():
    """Load root .env into the process environment so Next.js proxy
    can read MYCELIS_API_KEY (used to inject Authorization headers into
    proxied /api/* requests). Uses override=True so .env wins over system env.
    Removes PORT afterwards — the root .env sets PORT=8080 for Go Core,
    but Next.js would read it and try to listen on 8080 instead of 3000."""
    import os
    try:
        from dotenv import load_dotenv
        load_dotenv(str(ROOT_DIR / ".env"), override=True)
    except ImportError:
        pass  # python-dotenv not installed — env vars must be set manually
    # Don't let Go Core's PORT leak into Next.js
    os.environ.pop("PORT", None)


def _task_env(extra=None):
    _load_env()
    ensure_managed_cache_dirs()
    return managed_cache_env(extra=extra)


def _normalize_process_text(text: str) -> str:
    return (text or "").lower().replace("\\", "/")


def _matches_repo_local_interface_process(name: str, command_line: str) -> bool:
    normalized_name = (name or "").lower()
    normalized_cmd = _normalize_process_text(command_line)
    if normalized_name not in {"node", "node.exe", "cmd", "cmd.exe"} or not normalized_cmd:
        return False
    if not any(hint in normalized_cmd for hint in _INTERFACE_PROCESS_PATH_HINTS):
        return False
    return any(hint in normalized_cmd for hint in _INTERFACE_PROCESS_COMMAND_HINTS)


def _list_repo_local_interface_processes() -> list[dict[str, str | int]]:
    processes: list[dict[str, str | int]] = []
    try:
        if is_windows():
            result = subprocess.run(
                [
                    "powershell",
                    "-NoProfile",
                    "-Command",
                    "Get-CimInstance Win32_Process -Filter \"Name = 'node.exe' OR Name = 'cmd.exe'\" | "
                    "Select-Object ProcessId,Name,CommandLine | "
                    "ConvertTo-Json -Compress",
                ],
                capture_output=True,
                text=True,
                timeout=20,
            )
            if result.returncode != 0:
                raise RuntimeError(result.stderr.strip() or "process query failed")
            raw = result.stdout.strip()
            if not raw:
                return []
            payload = json.loads(raw)
            rows = payload if isinstance(payload, list) else [payload]
            for row in rows:
                pid = row.get("ProcessId")
                name = row.get("Name") or ""
                command_line = row.get("CommandLine") or ""
                if isinstance(pid, int) and _matches_repo_local_interface_process(name, command_line):
                    processes.append({"pid": pid, "name": name, "command": command_line})
            return processes

        result = subprocess.run(
            ["ps", "-eo", "pid=,comm=,args="],
            capture_output=True,
            text=True,
            timeout=10,
        )
        if result.returncode != 0:
            raise RuntimeError(result.stderr.strip() or "process query failed")
        for line in result.stdout.splitlines():
            parts = line.strip().split(None, 2)
            if len(parts) < 2 or not parts[0].isdigit():
                continue
            pid = int(parts[0])
            name = parts[1]
            command_line = parts[2] if len(parts) > 2 else ""
            if _matches_repo_local_interface_process(name, command_line):
                processes.append({"pid": pid, "name": name, "command": command_line})
    except (subprocess.SubprocessError, json.JSONDecodeError, OSError, ValueError, RuntimeError) as exc:
        raise RuntimeError(f"interface process inspection failed: {exc}") from exc
    return processes


def _kill_pid_tree(pid: int):
    try:
        if is_windows():
            subprocess.run(
                ["taskkill", "/F", "/T", "/PID", str(pid)],
                capture_output=True,
                timeout=8,
            )
        else:
            subprocess.run(["kill", "-9", str(pid)], capture_output=True, timeout=5)
    except subprocess.TimeoutExpired:
        pass


def _cleanup_repo_local_interface_processes() -> list[dict[str, str | int]]:
    try:
        processes = _list_repo_local_interface_processes()
    except RuntimeError as exc:
        print(f"  WARN: unable to inspect repo-local Interface residuals ({exc})")
        return []
    if not processes:
        return []

    print("  Cleaning repo-local Interface residuals...")
    for proc in processes:
        print(f"    - {proc['name']} (PID {proc['pid']})")
        _kill_pid_tree(int(proc["pid"]))

    time.sleep(0.5)
    try:
        remaining = _list_repo_local_interface_processes()
    except RuntimeError as exc:
        print(f"  WARN: unable to re-check repo-local Interface residuals ({exc})")
        return []
    if remaining:
        for proc in remaining:
            _kill_pid_tree(int(proc["pid"]))
        time.sleep(0.5)
        try:
            remaining = _list_repo_local_interface_processes()
        except RuntimeError:
            return []
    return remaining


def _playwright_server_log_path() -> str:
    log_dir = ROOT_DIR / "workspace" / "logs"
    log_dir.mkdir(parents=True, exist_ok=True)
    return str(log_dir / "interface-playwright-webserver.log")


def _start_playwright_server(env: dict[str, str], port: int = INTERFACE_PORT) -> subprocess.Popen[str]:
    log_path = _playwright_server_log_path()
    log_handle = open(log_path, "w", encoding="utf-8")
    process_env = os.environ.copy()
    process_env.update(env)
    host = env.get("INTERFACE_HOST", INTERFACE_HOST)
    try:
        if is_windows():
            command = (
                f'cd /d "{INTERFACE_DIR}" && '
                f"node .\\node_modules\\next\\dist\\bin\\next dev "
                f"--webpack --hostname {host} --port {port}"
            )
            process = subprocess.Popen(
                command,
                shell=True,
                cwd=str(ROOT_DIR),
                env=process_env,
                stdout=log_handle,
                stderr=subprocess.STDOUT,
                text=True,
            )
            log_handle.close()
            return process

        process = subprocess.Popen(
            [
                "node",
                "./node_modules/next/dist/bin/next",
                "dev",
                "--webpack",
                "--hostname",
                host,
                "--port",
                str(port),
            ],
            cwd=str(INTERFACE_DIR),
            env=process_env,
            stdout=log_handle,
            stderr=subprocess.STDOUT,
            text=True,
        )
        log_handle.close()
        return process
    except Exception:
        log_handle.close()
        raise


def _wait_for_interface_ready(host: str = INTERFACE_HOST, port: int = INTERFACE_PORT, timeout_seconds: int = 120):
    url = f"http://{host}:{port}"
    deadline = time.time() + timeout_seconds
    last_error = "server did not respond"
    while time.time() < deadline:
        try:
            with urllib.request.urlopen(url, timeout=5) as response:
                if response.status < 500:
                    return
                last_error = f"unexpected status {response.status}"
        except urllib.error.HTTPError as exc:
            if exc.code < 500:
                return
            last_error = f"http {exc.code}"
        except Exception as exc:  # pragma: no cover - exercised via timeout path in task flow
            last_error = str(exc)
        time.sleep(1)

    raise RuntimeError(
        f"Interface did not become ready at {url} within {timeout_seconds}s. "
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
    c.run("npm run dev --prefix interface", pty=not is_windows(), env=_task_env())

@task
def install(c):
    """Install Interface dependencies."""
    print("Installing Interface Dependencies...")
    c.run("cd interface && npm install", env=_task_env())

@task
def build(c):
    """Build the Interface for production."""
    print("Building Interface...")
    try:
        c.run("cd interface && npm run build", env=_task_env())
    finally:
        _cleanup_repo_local_interface_processes()

@task
def lint(c):
    """Lint the Interface code."""
    print("Linting Interface...")
    c.run("cd interface && npm run lint", env=_task_env())

@task
def test(c):
    """Run Interface Unit Tests (Vitest)."""
    print("Running Interface Tests...")
    try:
        c.run("cd interface && npm run test", env=_task_env())
    finally:
        _cleanup_repo_local_interface_processes()

@task
def test_coverage(c):
    """Run Interface unit tests with V8 coverage report."""
    print("Running Interface Tests with Coverage...")
    try:
        c.run("cd interface && npx vitest run --coverage", pty=not is_windows(), env=_task_env())
    finally:
        _cleanup_repo_local_interface_processes()

@task(
    help={
        "headed": "Open a visible browser window.",
        "project": "Optional Playwright project (chromium, firefox, webkit, mobile-chromium).",
        "spec": "Optional Playwright spec path or glob.",
        "live_backend": "Enable specs that require a real Core backend and authenticated UI proxying.",
    }
)
def e2e(c, headed=False, project="", spec="", live_backend=False):
    """
    Run Playwright E2E tests.
    The Invoke wrapper starts a managed local Next.js server and clears any stale
    Interface listener before and after the run because Next.js dev servers can
    linger on Windows after Playwright exits.
    Use --live-backend when the spec should talk to the real Core API through
    the Next.js proxy instead of relying entirely on route stubs.
    """
    print("Running Playwright E2E Tests...")
    cmd = "cd interface && npx playwright test --reporter=dot"
    if project:
        cmd += f" --project={project}"
    if spec:
        cmd += f" {spec}"
    if headed:
        cmd += " --headed"
    extra_env = {
        "PLAYWRIGHT_SKIP_WEBSERVER": "1",
        "INTERFACE_HOST": "127.0.0.1",
    }
    if live_backend:
        extra_env["PLAYWRIGHT_LIVE_BACKEND"] = "1"
    env = _task_env(extra_env)
    stop(c)
    server: subprocess.Popen[str] | None = None
    try:
        server = _start_playwright_server(env)
        _wait_for_interface_ready(host=env["INTERFACE_HOST"])
        result = c.run(cmd, pty=not is_windows(), env=env, hide=True, warn=True)
        _print_ascii_safe(result.stdout)
        _print_ascii_safe(result.stderr)
        if result.exited != 0:
            raise SystemExit(result.exited)
    finally:
        if server is not None and server.poll() is None:
            _kill_pid_tree(server.pid)
            time.sleep(0.5)
        stop(c)
        _cleanup_repo_local_interface_processes()

# ── Process Management ───────────────────────────────────────

@task
def stop(c, port=INTERFACE_PORT):
    """
    Stop the Interface dev server.
    Kills any node process listening on --port (default 3000).
    """
    print(f"Stopping Interface (port {port})...")
    if is_windows():
        ps_cmd = (
            f"$c = Get-NetTCPConnection -LocalPort {port} -State Listen -ErrorAction SilentlyContinue; "
            f"if ($c) {{ taskkill /F /T /PID $c.OwningProcess | Out-Null; Write-Host Killed PID $c.OwningProcess }} "
            f"else {{ Write-Host No process on port {port} }}"
        )
        c.run(powershell(ps_cmd), warn=True)
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
    import shutil, os
    cache_dir = os.path.join("interface", ".next")
    print("Clearing Next.js cache...")
    if os.path.isdir(cache_dir):
        shutil.rmtree(cache_dir)
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
    if is_windows():
        c.run(
            f'start /B cmd /c "cd interface && npx next dev --port {port} > NUL 2>&1"',
            warn=True, hide=True, env=_task_env(),
        )
    else:
        c.run(f"cd interface && npx next dev --port {port} > /dev/null 2>&1 &", warn=True, env=_task_env())

    # 5. Wait for server to be ready, then check
    print("Waiting for server startup...")
    time.sleep(6)
    check(c, port=port)

# ── Health Check ─────────────────────────────────────────────

@task
def check(c, port=INTERFACE_PORT):
    """
    Smoke-test the running Interface dev server.
    Fetches key pages and checks for SSR errors, 404s, and dark-mode leaks.
    Requires: interface.dev running on --port (default 3000).
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
                if "hydration" in body.lower() and "error" in body.lower():
                    issues.append("Hydration mismatch detected")
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
ns.add_task(test_coverage, name="test-coverage")
ns.add_task(e2e)
ns.add_task(stop)
ns.add_task(clean)
ns.add_task(restart)
ns.add_task(check)
