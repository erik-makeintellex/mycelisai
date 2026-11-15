import ipaddress
import os
import re
import shutil
import socket
import subprocess
import time
import urllib.error
import urllib.request
from contextlib import suppress
from pathlib import Path

from invoke import Collection, task
from .cache import ensure_disk_headroom
from .config import (
    INTERFACE_BIND_HOST,
    INTERFACE_HOST,
    INTERFACE_PORT,
    ROOT_DIR,
    is_windows,
    powershell,
)
from .interface_env import (
    CommandResult,
    _build_playwright_command,
    _build_playwright_env,
    _expected_next_build_artifacts,
    _is_incomplete_next_build_output,
    _is_next_build_lock_conflict,
    _is_next_standalone_cleanup_conflict,
    _normalize_process_text,
    _playwright_last_run_path,
    _read_playwright_last_run_status,
    _reconcile_managed_server_endpoint,
    _report_command_result,
    _run_interface_commandline,
    _run_interface_shell_command,
    _run_interface_shell_command_streaming,
    _run_one_shot_interface_task,
    _task_env,
    _wait_for_complete_next_build_output,
    interface_task_env,
    run_interface_command,
)
from . import interface_processes

ns = Collection("interface")
INTERFACE_DIR = ROOT_DIR / "interface"

from .interface_runtime_helpers import (  # noqa: E402
    _cleanup_managed_interface_listeners,
    _cleanup_repo_local_interface_processes,
    _kill_pid_tree,
    _list_repo_local_interface_processes,
    _matches_repo_local_interface_process,
    _print_ascii_safe,
    _windows_listening_pids_for_port,
    _windows_listening_pids_for_port_range,
    run_build_task,
    run_check_task,
    run_clean_task,
    run_restart_task,
    run_stop_task,
)
from .interface_runtime_server import (  # noqa: E402
    _cleanup_stale_next_dev_lock,
    _interface_ready_urls,
    _next_dev_command,
    _next_dev_lock_path,
    _next_start_command,
    _pick_interface_port,
    _spawn_interface_process,
    _wait_for_interface_ready,
    start_dev_server_detached,
)
from .interface_runtime_e2e import (  # noqa: E402
    _cleanup_playwright_server_log,
    _detect_playwright_server_port,
    _playwright_server_log_path,
    _run_playwright_command_streaming,
    _start_playwright_server,
    _stop_lingering_playwright_process,
    run_e2e_task,
)


# ── Lifecycle ────────────────────────────────────────────────

@task
def dev(c):
    """Start Interface (Next.js) in Dev Mode. Stops existing instance first."""
    stop(c)
    run_interface_command(c, "npm run dev", pty=not is_windows())


@task
def install(c):
    """Install Interface dependencies."""
    print("Installing Interface Dependencies...")
    run_interface_command(c, "npm install")
    print("Installing Playwright Chromium browser...")
    run_interface_command(c, "npx playwright install chromium")


@task
def build(c):
    """Build the Interface for production."""
    run_build_task(c)


@task
def lint(c):
    """Lint the Interface code."""
    print("Linting Interface...")
    run_interface_command(c, "npm run lint")


@task
def test(c):
    """Run Interface Unit Tests (Vitest)."""
    print("Running Interface Tests...")
    _run_one_shot_interface_task(["npm", "run", "test"])


@task
def typecheck(c):
    """Run the Interface TypeScript typecheck."""
    print("Running Interface Type Check...")
    _run_one_shot_interface_task(["npx", "tsc", "--noEmit"])


@task
def test_coverage(c):
    """Run Interface unit tests with V8 coverage report."""
    print("Running Interface Tests with Coverage...")
    _run_one_shot_interface_task(["npx", "vitest", "run", "--coverage"])


@task(
    help={
        "headed": "Open a visible browser window.",
        "project": "Optional Playwright project (chromium, firefox, webkit, mobile-chromium).",
        "spec": 'Optional Playwright spec path or glob. Pass a single quoted, space-separated string to run several spec files in one invocation, e.g. --spec="e2e/specs/a.spec.ts e2e/specs/b.spec.ts".',
        "live_backend": "Enable specs that require a real Core backend and authenticated UI proxying.",
        "workers": "Optional Playwright worker count override for stability-sensitive runs.",
        "server_mode": "Server mode for the managed UI server (dev or start, defaults to dev).",
    }
)
def e2e(c, headed=False, project="", spec="", live_backend=False, workers="", server_mode="dev"):
    """Run Playwright E2E tests against a managed local Next.js server."""
    run_e2e_task(
        c,
        headed=headed,
        project=project,
        spec=spec,
        live_backend=live_backend,
        workers=workers,
        server_mode=server_mode,
    )


# ── Process Management ───────────────────────────────────────

@task
def stop(c, port=INTERFACE_PORT):
    """Stop the Interface server and sweep repo-local worker residue."""
    run_stop_task(c, port)


@task
def clean(c):
    """Clear the Next.js build cache (.next directory)."""
    run_clean_task(c)


@task
def restart(c, port=INTERFACE_PORT):
    """Full Interface restart: stop -> clear cache -> build -> start dev -> check."""
    run_restart_task(c, port)


# ── Health Check ─────────────────────────────────────────────

@task
def check(c, port=INTERFACE_PORT):
    """Smoke-test a running Interface server."""
    run_check_task(c, port)


# ── Register Tasks ───────────────────────────────────────────

ns.add_task(dev)
ns.add_task(install)
ns.add_task(build)
ns.add_task(lint)
ns.add_task(test)
ns.add_task(typecheck)
ns.add_task(test_coverage, name="test-coverage")
ns.add_task(e2e)
ns.add_task(stop)
ns.add_task(clean)
ns.add_task(restart)
ns.add_task(check)
