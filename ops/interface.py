from invoke import task, Collection
from .config import is_windows, powershell, INTERFACE_HOST, INTERFACE_PORT

ns = Collection("interface")

# ── Lifecycle ────────────────────────────────────────────────

@task
def dev(c):
    """Start Interface (Next.js) in Dev Mode. Stops existing instance first."""
    stop(c)
    c.run("npm run dev --prefix interface", pty=not is_windows())

@task
def install(c):
    """Install Interface dependencies."""
    print("Installing Interface Dependencies...")
    c.run("cd interface && npm install")

@task
def build(c):
    """Build the Interface for production."""
    print("Building Interface...")
    c.run("cd interface && npm run build")

@task
def lint(c):
    """Lint the Interface code."""
    print("Linting Interface...")
    c.run("cd interface && npm run lint")

@task
def test(c):
    """Run Interface Unit Tests (Vitest)."""
    print("Running Interface Tests...")
    c.run("cd interface && npm run test")

@task
def test_coverage(c):
    """Run Interface unit tests with V8 coverage report."""
    print("Running Interface Tests with Coverage...")
    c.run("cd interface && npx vitest run --coverage", pty=not is_windows())

@task
def e2e(c, headed=False):
    """
    Run Playwright E2E tests against a running dev server.
    Requires: interface.dev running on port 3000.
    Use --headed to see the browser.
    """
    print("Running Playwright E2E Tests...")
    cmd = "cd interface && npx playwright test"
    if headed:
        cmd += " --headed"
    c.run(cmd, pty=not is_windows())

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
            f"if ($c) {{ Stop-Process -Id $c.OwningProcess -Force; Write-Host Killed PID $c.OwningProcess }} "
            f"else {{ Write-Host No process on port {port} }}"
        )
        c.run(powershell(ps_cmd), warn=True)
    else:
        # lsof works on macOS + Linux; fuser as fallback
        c.run(f"lsof -ti:{port} | xargs -r kill -9 2>/dev/null || fuser -k {port}/tcp 2>/dev/null || true", warn=True)
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
            warn=True, hide=True,
        )
    else:
        c.run(f"cd interface && npx next dev --port {port} > /dev/null 2>&1 &", warn=True)

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
