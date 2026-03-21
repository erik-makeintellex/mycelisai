from invoke import task, Collection
from .config import CORE_DIR, is_windows
from . import core, interface

@task
def all(c):
    """
    Run ALL Unit Tests (Core + Interface).
    """
    print("Executing Full Test Suite...")

    try:
        core.test(c)
        interface.test(c)
        print("All Tests Passed.")
    except Exception as e:
        print(f"Test Failure: {e}")
        exit(1)

@task
def coverage(c):
    """
    Run all tests with coverage reports.
    Core: go test -coverprofile  |  Interface: vitest --coverage
    """
    print("=== Coverage Report ===")
    print()
    print("[Core] Running Go tests with coverage...")
    with c.cd(str(CORE_DIR)):
        c.run("go test -coverprofile=coverage.out ./...", pty=not is_windows())
    print()
    print("[Interface] Running Vitest with V8 coverage...")
    interface.run_interface_command(c, "npx vitest run --coverage", cleanup=True, pty=not is_windows())
    print()
    print("Coverage reports generated.")

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
    Run Playwright E2E tests (alias for interface.e2e).
    Playwright owns the Interface dev server lifecycle. Start Core separately
    only when the spec needs a live backend instead of route stubs. The task
    clears stale Interface listeners before and after the browser run.
    """
    interface.e2e(c, headed=headed, project=project, spec=spec, live_backend=live_backend)

ns = Collection("test")
ns.add_task(all)
ns.add_task(coverage)
ns.add_task(e2e)
