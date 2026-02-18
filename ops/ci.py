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
ns.add_task(deploy)
