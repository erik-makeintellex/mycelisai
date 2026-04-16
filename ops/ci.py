"""
Local CI task entrypoints for operator and workflow use.
GitHub workflows may reuse these task surfaces after workflow-native bootstrap.

Usage:
    uv run inv ci.lint          # Go vet + Next.js lint
    uv run inv ci.test          # Go tests + Interface tests
    uv run inv ci.build         # Go binary + Next.js production build (no Docker)
    uv run inv ci.check         # Full pipeline: lint -> test -> build
    uv run inv ci.deploy        # Build + Docker + K8s deploy (requires cluster)
"""

import os
import ipaddress
import time
from contextlib import suppress
from urllib.parse import urljoin, urlparse
import urllib.error
import urllib.request
from invoke import task, Collection
from .config import (
    CORE_DIR,
    ensure_managed_cache_dirs,
    managed_cache_env,
)
from . import db as db_tasks
from . import cache as cache_tasks
from . import logging as logging_tasks
from . import core as core_tasks
from . import interface as interface_tasks
from . import lifecycle
from . import quality


def _task_env(extra=None):
    ensure_managed_cache_dirs()
    return managed_cache_env(extra=extra)


def _configured_ai_endpoints() -> list[tuple[str, str, str]]:
    endpoints: list[tuple[str, str, str]] = []
    for env_name, label in (
        ("MYCELIS_COMPOSE_OLLAMA_HOST", "compose Ollama host"),
        ("MYCELIS_K8S_TEXT_ENDPOINT", "k8s text endpoint"),
        ("MYCELIS_K8S_MEDIA_ENDPOINT", "k8s media endpoint"),
        ("MYCELIS_PROVIDER_LOCAL_OLLAMA_DEV_ENDPOINT", "local Ollama provider endpoint"),
    ):
        raw = (os.environ.get(env_name, "") or "").strip()
        if raw:
            endpoints.append((env_name, label, raw))
    return endpoints


def _probe_paths_for_endpoint(env_name: str, raw: str) -> tuple[str, str]:
    parsed = urlparse(raw)
    if env_name == "MYCELIS_COMPOSE_OLLAMA_HOST" or parsed.path.rstrip("/").endswith("/api"):
        return ("/api/tags", "/v1/models")
    return ("/models", "/api/tags")


def _is_loopback_or_unspecified_host(host: str) -> bool:
    normalized = (host or "").strip().lower()
    if not normalized:
        return True
    if normalized == "localhost":
        return True
    try:
        address = ipaddress.ip_address(normalized)
    except ValueError:
        return False
    return address.is_loopback or address.is_unspecified


def _probe_http_endpoint(url: str, timeout: float = 3.0) -> tuple[int, str]:
    try:
        request = urllib.request.Request(url, headers={"Accept": "application/json"})
        with urllib.request.urlopen(request, timeout=timeout) as response:
            body = response.read().decode("utf-8", errors="replace")
            return response.status, body
    except urllib.error.HTTPError as exc:
        body = ""
        with suppress(Exception):
            body = exc.read().decode("utf-8", errors="replace")
        return exc.code, body or str(exc)
    except Exception as exc:
        return 0, str(exc)


def _runtime_posture_check(c):
    print("=== RUNTIME POSTURE ===")
    cache_tasks.ensure_disk_headroom(min_free_gb=12, reason="release preflight posture")

    endpoints = _configured_ai_endpoints()
    if not endpoints:
        print("  No explicit AI endpoints configured; skipping endpoint reachability probe.")
        return

    failures: list[str] = []
    for env_name, label, raw in endpoints:
        parsed = urlparse(raw)
        if parsed.scheme not in {"http", "https"} or not parsed.netloc:
            failures.append(f"{env_name}: invalid endpoint URL '{raw}'")
            print(f"  [FAIL] {label}: invalid endpoint URL '{raw}'")
            continue

        host = parsed.hostname or ""
        if _is_loopback_or_unspecified_host(host):
            failures.append(f"{env_name}: loopback or unspecified host '{host}' is not allowed")
            print(f"  [FAIL] {label}: loopback or unspecified host '{host}' is not allowed")
            continue

        probe_paths = _probe_paths_for_endpoint(env_name, raw)
        reachable = False
        for probe_path in probe_paths:
            probe_url = urljoin(raw.rstrip("/") + "/", probe_path.lstrip("/"))
            status, _body = _probe_http_endpoint(probe_url)
            if status in {200, 401, 403}:
                print(f"  [OK]   {label}: {probe_url} [{status}]")
                reachable = True
                break
            print(f"  [WARN] {label}: {probe_url} [{status}]")

        if not reachable:
            failures.append(f"{env_name}: no AI probe path responded successfully")

    if failures:
        raise SystemExit(
            "RUNTIME POSTURE CHECK FAILED: "
            + "; ".join(failures)
        )

    print("RUNTIME POSTURE PASSED")


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
    print("[2/2] interface lint")
    try:
        interface_tasks.lint.body(c)
    except SystemExit:
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
    print("[2/2] interface test")
    try:
        interface_tasks.test.body(c)
    except SystemExit:
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
    cache_tasks.ensure_disk_headroom(min_free_gb=10, reason="ci build")

    # 1. Go binary
    print("[1/2] core compile")
    try:
        core_tasks.compile.body(c)
    except SystemExit:
        errors.append("go build failed")
    else:
        print("  OK")

    # 2. Next.js production build (type-checks + compiles)
    print("[2/2] interface build")
    try:
        interface_tasks.build.body(c)
    except SystemExit:
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
    cache_tasks.ensure_disk_headroom(min_free_gb=10, reason="ci baseline")

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

    print("[5/7] interface build")
    try:
        interface_tasks.build.body(c)
    except SystemExit:
        errors.append("interface build failed")
    else:
        print("  OK")

    print("[6/7] interface typecheck")
    try:
        interface_tasks.typecheck.body(c)
    except SystemExit:
        errors.append("interface typecheck failed")
    else:
        print("  OK")

    print("[7/7] interface test")
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
        if errors:
            print("  SKIP (prerequisites failed)")
        else:
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


@task(help={"live_backend": "Also run the live-backend governed Soma browser contract after health checks."})
def service_check(c, live_backend=False):
    """
    Validate the currently running local stack and optionally prove the live
    backend governed Soma contract through the browser.
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
            print("  SKIP (cortex schema already compatible with the current runtime)")
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

        print("[3/3] interface live-backend governed playwright")
        if errors:
            print("  SKIP (prerequisites failed)")
        else:
            try:
                interface_tasks.build.body(c)
                time.sleep(3)
                interface_tasks.e2e.body(
                    c,
                    project="chromium",
                    spec="e2e/specs/soma-governance-live.spec.ts",
                    live_backend=True,
                    workers="1",
                    server_mode="start",
                )
            except SystemExit:
                errors.append("interface live-backend governed playwright failed")
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
        "runtime_posture": "Also check tighter disk headroom and explicit AI endpoint reachability when configured (default: False).",
    }
)
def release_preflight(
    c,
    e2e=True,
    strict_toolchain=False,
    service_health=False,
    live_backend=False,
    runtime_posture=False,
):
    """
    Enforce release preflight gate:
    - clean working tree
    - toolchain check
    - strict baseline validation
    - optional runtime posture check for storage and explicit AI endpoints
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
    if runtime_posture:
        _runtime_posture_check(c)
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
