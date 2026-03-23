"""
Local CI Pipeline — runs on dev machine or configured host.
No GitHub Actions, no auto-triggers. Manual invocation only.

Usage:
    uv run inv ci.lint          # Go vet + Next.js lint
    uv run inv ci.test          # Go tests + Interface tests
    uv run inv ci.build         # Go binary + Next.js production build (no Docker)
    uv run inv ci.check         # Full pipeline: lint -> test -> build
    uv run inv ci.deploy        # Build + Docker + K8s deploy (requires cluster)
"""

import time
from invoke import task, Collection
from .config import (
    API_HOST,
    API_PORT,
    CORE_DIR,
    INTERFACE_PORT,
    ensure_managed_cache_dirs,
    is_windows,
    managed_cache_env,
)
from . import db as db_tasks
from . import logging as logging_tasks
from . import interface as interface_tasks
from . import lifecycle
from . import quality


def _task_env(extra=None):
    ensure_managed_cache_dirs()
    return managed_cache_env(extra=extra)


def _run_interface_command(c, command: str, **run_kwargs):
    return interface_tasks.run_interface_command(c, command, cleanup=True, **run_kwargs)


@task
def lint(c):
    """Lint: Go vet + Next.js lint."""
    errors = []

    print("=== LINT ===")
    print()

    # 1. Go vet
    print("[1/2] go vet ./...")
    with c.cd(str(CORE_DIR)):
        result = c.run("go vet ./...", warn=True, env=_task_env())
        if result.exited != 0:
            errors.append("go vet failed")
        else:
            print("  OK")

    # 2. Next.js lint
    print("[2/2] next lint")
    result = _run_interface_command(c, "npm run lint", warn=True)
    if result.exited != 0:
        errors.append("next lint failed")
    else:
        print("  OK")

    print()
    if errors:
        print(f"LINT FAILED: {', '.join(errors)}")
        raise SystemExit(1)
    print("LINT PASSED")


@task
def test(c):
    """Test: Go unit tests + Interface tests."""
    errors = []

    print("=== TEST ===")
    print()

    # 1. Go tests
    print("[1/2] go test ./...")
    with c.cd(str(CORE_DIR)):
        result = c.run("go test ./...", warn=True, env=_task_env())
        if result.exited != 0:
            errors.append("go tests failed")
        else:
            print("  OK")

    # 2. Interface tests
    print("[2/2] interface vitest run")
    result = _run_interface_command(c, "npx vitest run --reporter=dot", warn=True, hide=True)
    if result.exited != 0:
        errors.append("interface tests failed")
    else:
        print("  OK")

    print()
    if errors:
        print(f"TEST FAILED: {', '.join(errors)}")
        raise SystemExit(1)
    print("TESTS PASSED")


@task
def build(c):
    """Build: Go binary + Next.js production build (no Docker)."""
    errors = []

    print("=== BUILD ===")
    print()

    # 1. Go binary
    print("[1/2] go build")
    with c.cd(str(CORE_DIR)):
        bin_cmd = "go build -v -o bin/server.exe ./cmd/server" if is_windows() else "go build -v -o bin/server ./cmd/server"
        result = c.run(bin_cmd, warn=True, env=_task_env())
        if result.exited != 0:
            errors.append("go build failed")
        else:
            print("  OK")

    # 2. Next.js production build (type-checks + compiles)
    print("[2/2] next build")
    result = _run_interface_command(c, "npx next build", warn=True)
    if result.exited != 0:
        errors.append("next build failed")
    else:
        print("  OK")

    print()
    if errors:
        print(f"BUILD FAILED: {', '.join(errors)}")
        raise SystemExit(1)
    print("BUILD PASSED")


@task
def check(c):
    """
    Full local CI pipeline: lint -> test -> build.
    Run this before pushing code or creating PRs.
    """
    start = time.time()

    print("=" * 60)
    print("  MYCELIS LOCAL CI PIPELINE")
    print("=" * 60)
    print()

    stages = [
        ("LINT", lint),
        ("TEST", test),
        ("BUILD", build),
    ]

    for name, fn in stages:
        stage_start = time.time()
        try:
            fn(c)
        except SystemExit:
            elapsed = time.time() - start
            print()
            print(f"PIPELINE FAILED at stage: {name} ({elapsed:.1f}s)")
            raise SystemExit(1)
        stage_elapsed = time.time() - stage_start
        print(f"  [{name} completed in {stage_elapsed:.1f}s]")
        print()

    elapsed = time.time() - start
    print("=" * 60)
    print(f"  PIPELINE PASSED ({elapsed:.1f}s)")
    print("=" * 60)


@task(help={"e2e": "Include Playwright E2E run (default: True)."})
def baseline(c, e2e=True):
    """
    Strict baseline validation for delivery readiness.
    Runs: core tests, interface build, interface typecheck, vitest, and Playwright by default.
    """
    errors = []

    print("=== BASELINE ===")
    print()

    print("[1/7] logging.check-schema")
    try:
        logging_tasks.check_schema.body(c)
        print("  OK")
    except SystemExit:
        errors.append("logging schema check failed")

    print("[2/7] logging.check-topics")
    try:
        logging_tasks.check_topics.body(c)
        print("  OK")
    except SystemExit:
        errors.append("logging topic check failed")

    print("[3/7] quality.max-lines --limit=350")
    try:
        quality.max_lines.body(c, limit=350, paths=quality.DEFAULT_HOT_PATHS, strict=False)
        print("  OK")
    except SystemExit:
        errors.append("quality max-lines check failed")

    print("[4/7] core go test ./... -count=1")
    with c.cd(str(CORE_DIR)):
        result = c.run("go test ./... -count=1", warn=True, hide=True, env=_task_env())
        if result.exited != 0:
            errors.append("core go tests failed")
        else:
            print("  OK")

    print("[5/7] interface npm run build")
    try:
        interface_tasks.build.body(c)
    except SystemExit:
        errors.append("interface build failed")
    else:
        print("  OK")

    print("[6/7] interface tsc --noEmit")
    result = _run_interface_command(c, "npx tsc --noEmit", warn=True, hide=True)
    if result.exited != 0:
        errors.append("interface typecheck failed")
    else:
        print("  OK")

    print("[7/7] interface vitest run")
    try:
        interface_tasks.stop.body(c)
        interface_tasks.clean.body(c)
        interface_tasks.test.body(c)
    except SystemExit:
        errors.append("interface vitest failed")
    else:
        print("  OK")

    if e2e:
        print("[E2E] interface playwright run")
        try:
            interface_tasks.build.body(c)
            interface_tasks.e2e.body(c, workers="1", server_mode="start")
        except SystemExit:
            errors.append("interface playwright failed")
        else:
            print("  OK")
    else:
        print("[E2E] interface playwright run")
        print("  SKIP (--no-e2e)")

    print()
    if errors:
        print(f"BASELINE FAILED: {', '.join(errors)}")
        raise SystemExit(1)
    print("BASELINE PASSED")


@task(help={"live_backend": "Also run the live-backend workspace Playwright contract after health checks."})
def service_check(c, live_backend=False):
    """
    Validate the currently running local stack and optionally prove the live
    backend workspace contract through the browser.
    """
    errors = []

    print("=== SERVICE CHECK ===")
    print()

    if live_backend:
        print("[1/3] lifecycle.up --frontend=false --build=false")
        try:
            lifecycle.up.body(c, frontend=False, build=False)
            print("  OK")
        except SystemExit:
            errors.append("lifecycle up failed")
        print("[1.5/3] db.migrate")
        if db_tasks.schema_bootstrapped():
            print("  SKIP (cortex schema already initialized)")
        else:
            try:
                db_tasks.migrate.body(c)
                print("  OK")
            except SystemExit:
                errors.append("database migrate failed")

        print("[2/3] lifecycle.health")
        try:
            lifecycle.health.body(c)
            print("  OK")
        except SystemExit:
            errors.append("lifecycle health failed")

        print("[3/3] interface live-backend playwright")
        try:
            interface_tasks.build.body(c)
            time.sleep(3)
            interface_tasks.e2e.body(
                c,
                project="chromium",
                spec="e2e/specs/workspace-live-backend.spec.ts",
                live_backend=True,
                workers="1",
                server_mode="start",
            )
        except SystemExit:
            errors.append("interface live-backend playwright failed")
        else:
            print("  OK")
    else:
        print("[1/2] lifecycle.health")
        try:
            lifecycle.health.body(c)
            print("  OK")
        except SystemExit:
            errors.append("lifecycle health failed")

        print("[2/2] interface live-backend playwright")
        print("  SKIP (--live-backend not set)")

    print()
    if errors:
        print(f"SERVICE CHECK FAILED: {', '.join(errors)}")
        raise SystemExit(1)
    print("SERVICE CHECK PASSED")


@task(help={"strict": "Fail if Go version does not match the locked policy (default: False)."})
def toolchain_check(c, strict=False):
    """
    Report local toolchain versions and optionally enforce Go lock policy.
    """
    print("=== TOOLCHAIN CHECK ===")
    go_result = c.run("go version", hide=True, warn=True)
    node_result = c.run("node -v", hide=True, warn=True)
    npm_result = c.run("npm -v", hide=True, warn=True)

    go_text = (go_result.stdout or "").strip()
    node_text = (node_result.stdout or "").strip()
    npm_text = (npm_result.stdout or "").strip()

    print(f"go:   {go_text or 'unavailable'}")
    print(f"node: {node_text or 'unavailable'}")
    print(f"npm:  {npm_text or 'unavailable'}")

    if go_result.exited != 0:
        raise SystemExit("TOOLCHAIN CHECK FAILED: go is unavailable.")

    locked_go_prefix = "go1.26"
    go_matches = locked_go_prefix in go_text
    if not go_matches:
        message = (
            f"Go version drift: expected {locked_go_prefix} (locked docs), found '{go_text}'."
        )
        if strict:
            raise SystemExit(f"TOOLCHAIN CHECK FAILED: {message}")
        print(f"WARN: {message}")
    else:
        print("Go version matches lock policy.")


@task
def entrypoint_check(c):
    """
    Verify the supported invoke runner matrix.
    - uv run inv ...          => supported primary path
    - uvx inv ...             => unsupported bare alias
    - uvx --from invoke inv   => lightweight compatibility path
    """
    print("=== ENTRYPOINT CHECK ===")

    primary = c.run("uv run inv -l", hide=True, warn=True)
    if primary.exited != 0:
        raise SystemExit("ENTRYPOINT CHECK FAILED: 'uv run inv -l' did not succeed.")
    print("uv run inv -l: OK")

    bare_uvx = c.run("uvx inv -l", hide=True, warn=True)
    bare_uvx_text = f"{bare_uvx.stdout or ''}{bare_uvx.stderr or ''}"
    expected_error = "does not provide any executables"
    if bare_uvx.exited == 0 or expected_error not in bare_uvx_text:
        raise SystemExit(
            "ENTRYPOINT CHECK FAILED: expected bare 'uvx inv -l' to fail with the package-executable message."
        )
    print("uvx inv -l: expected failure confirmed")

    compat = c.run("uvx --from invoke inv -l", hide=True, warn=True)
    if compat.exited != 0:
        raise SystemExit("ENTRYPOINT CHECK FAILED: 'uvx --from invoke inv -l' did not succeed.")
    print("uvx --from invoke inv -l: OK")

    print("ENTRYPOINT CHECK PASSED")


@task(
    help={
        "e2e": "Include Playwright in baseline gate (default: True).",
        "strict_toolchain": "Fail on Go lock mismatch (default: False).",
        "service_health": "Require lifecycle.health against the running local stack (default: False).",
        "live_backend": "Also run the live-backend workspace Playwright contract when service-health is enabled (default: False).",
    }
)
def release_preflight(c, e2e=True, strict_toolchain=False, service_health=False, live_backend=False):
    """
    Enforce release preflight gate:
    - clean working tree
    - toolchain check
    - strict baseline validation
    - optional live service-health / live-backend proof
    """
    print("=== RELEASE PREFLIGHT ===")
    status = c.run("git status --porcelain", hide=True, warn=True)
    dirty_lines = [ln for ln in (status.stdout or "").splitlines() if ln.strip()]
    if dirty_lines:
        print("Working tree is not clean:")
        preview = dirty_lines[:20]
        for ln in preview:
            print(f"  {ln}")
        if len(dirty_lines) > len(preview):
            print(f"  ... and {len(dirty_lines) - len(preview)} more")
        raise SystemExit("RELEASE PREFLIGHT FAILED: clean-tree requirement not met.")

    toolchain_check(c, strict=strict_toolchain)
    baseline(c, e2e=e2e)
    if service_health:
        service_check(c, live_backend=live_backend)
    print("RELEASE PREFLIGHT PASSED")


@task
def deploy(c):
    """
    Build + Docker + K8s deploy.
    Requires: Docker running, Kind cluster active.
    Delegates to k8s.deploy which handles image build + helm upgrade.
    """
    from . import k8s

    print("=== DEPLOY ===")
    print()

    # Run lint + test first
    lint(c)
    test(c)

    # Delegate to k8s.deploy (builds Docker image + helm upgrade)
    k8s.deploy(c)

    print()
    print("DEPLOY COMPLETE")


ns = Collection("ci")
ns.add_task(lint)
ns.add_task(test)
ns.add_task(build)
ns.add_task(check)
ns.add_task(baseline)
ns.add_task(service_check, name="service-check")
ns.add_task(toolchain_check, name="toolchain-check")
ns.add_task(entrypoint_check, name="entrypoint-check")
ns.add_task(release_preflight, name="release-preflight")
ns.add_task(deploy)
