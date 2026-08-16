from invoke import task, Collection
from .config import CORE_DIR, is_windows
from . import interface


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


ns = Collection("test")
ns.add_task(coverage)
