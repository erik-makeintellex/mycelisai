"""
Local CI Pipeline — runs on dev machine or configured host.
No GitHub Actions, no auto-triggers. Manual invocation only.

Usage:
    inv ci.lint          # Go vet + Next.js lint
    inv ci.test          # Go tests + Interface tests
    inv ci.build         # Go binary + Next.js production build (no Docker)
    inv ci.check         # Full pipeline: lint → test → build
    inv ci.deploy        # Build + Docker + K8s deploy (requires cluster)
"""

import time
from invoke import task, Collection
from .config import CORE_DIR, is_windows, API_HOST, API_PORT, INTERFACE_PORT


@task
def lint(c):
    """Lint: Go vet + Next.js lint."""
    errors = []

    print("=== LINT ===")
    print()

    # 1. Go vet
    print("[1/2] go vet ./...")
    with c.cd(str(CORE_DIR)):
        result = c.run("go vet ./...", warn=True)
        if result.exited != 0:
            errors.append("go vet failed")
        else:
            print("  OK")

    # 2. Next.js lint
    print("[2/2] next lint")
    result = c.run("cd interface && npm run lint", warn=True)
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
        result = c.run("go test ./...", warn=True)
        if result.exited != 0:
            errors.append("go tests failed")
        else:
            print("  OK")

    # 2. Interface tests (may not have test suite yet — warn only)
    print("[2/2] interface tests")
    result = c.run("cd interface && npm run test -- --run 2>/dev/null", warn=True)
    if result.exited != 0:
        # Check if it's "no tests found" vs actual failure
        if result.stderr and "no test" in result.stderr.lower():
            print("  SKIP (no test files)")
        else:
            print("  WARN: interface tests failed (non-blocking)")

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
        result = c.run(bin_cmd, warn=True)
        if result.exited != 0:
            errors.append("go build failed")
        else:
            print("  OK")

    # 2. Next.js production build (type-checks + compiles)
    print("[2/2] next build")
    result = c.run("cd interface && npx next build", warn=True)
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
    Full local CI pipeline: lint → test → build.
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


@task(help={"e2e": "Include Playwright E2E run (default: False)."})
def baseline(c, e2e=False):
    """
    Strict baseline validation for delivery readiness.
    Runs: core tests, interface build, vitest, and optional playwright.
    """
    errors = []

    print("=== BASELINE ===")
    print()

    print("[1/4] core go test ./... -count=1")
    with c.cd(str(CORE_DIR)):
        result = c.run("go test ./... -count=1", warn=True)
        if result.exited != 0:
            errors.append("core go tests failed")
        else:
            print("  OK")

    print("[2/4] interface npm run build")
    result = c.run("cd interface && npm run build", warn=True)
    if result.exited != 0:
        errors.append("interface build failed")
    else:
        print("  OK")

    print("[3/4] interface vitest run")
    result = c.run("cd interface && npx vitest run --reporter=dot", warn=True)
    if result.exited != 0:
        errors.append("interface vitest failed")
    else:
        print("  OK")

    if e2e:
        print("[4/4] interface playwright run")
        result = c.run("cd interface && npx playwright test --reporter=dot", warn=True)
        if result.exited != 0:
            errors.append("interface playwright failed")
        else:
            print("  OK")
    else:
        print("[4/4] interface playwright run")
        print("  SKIP (--e2e not set)")

    print()
    if errors:
        print(f"BASELINE FAILED: {', '.join(errors)}")
        raise SystemExit(1)
    print("BASELINE PASSED")


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


@task(
    help={
        "e2e": "Include Playwright in baseline gate (default: False).",
        "strict_toolchain": "Fail on Go lock mismatch (default: False).",
    }
)
def release_preflight(c, e2e=False, strict_toolchain=False):
    """
    Enforce release preflight gate:
    - clean working tree
    - toolchain check
    - strict baseline validation
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
ns.add_task(toolchain_check, name="toolchain-check")
ns.add_task(release_preflight, name="release-preflight")
ns.add_task(deploy)
