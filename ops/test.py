from invoke import task, Collection
from .config import is_windows
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
    c.run("cd core && go test -coverprofile=coverage.out ./...", pty=not is_windows())
    print()
    print("[Interface] Running Vitest with V8 coverage...")
    c.run("cd interface && npx vitest run --coverage", pty=not is_windows())
    print()
    print("Coverage reports generated.")

@task
def e2e(c, headed=False):
    """
    Run Playwright E2E tests (alias for interface.e2e).
    Requires: core.run + interface.dev running.
    """
    interface.e2e(c, headed=headed)

ns = Collection("test")
ns.add_task(all)
ns.add_task(coverage)
ns.add_task(e2e)
