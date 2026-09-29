from invoke import task, Collection
from .config import CORE_DIR, is_windows
from . import compose_images, interface


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


@task(help={"allow_stale": "Run even if core/interface containers run a stale image (proof then tests old code)."})
def probe(c, allow_stale=False):
    """
    Run the live user-journey delivery probe (ops/live_journey_probe.py) against a running stack.
    Use after `lifecycle.up`/`compose.up` to check real outcomes (ask, deliverable, memory
    recall, sensors, search), not just health status; it is the delivery metric in
    .state/V8_DEV_STATE.md. Requires MYCELIS_API_KEY in .env and a reachable Core on :8081.
    Refuses (non-zero) when the Compose core/interface containers run a stale image; pass
    --allow-stale to override.
    """
    compose_images.require_current("test.probe", allow_stale=allow_stale)
    c.run("uv run python ops/live_journey_probe.py", pty=not is_windows())


ns = Collection("test")
ns.add_task(coverage)
ns.add_task(probe)
