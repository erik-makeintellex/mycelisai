from __future__ import annotations

import json
import os
import shlex
import shutil
import subprocess
import time
from dataclasses import dataclass
from pathlib import Path

from .config import (
    INTERFACE_BIND_HOST,
    ROOT_DIR,
    ensure_managed_cache_dirs,
    is_windows,
    managed_cache_env,
)

INTERFACE_DIR = ROOT_DIR / "interface"


def _load_env():
    """Load Interface env files while keeping root .env as the secret source."""
    try:
        from dotenv import load_dotenv

        load_dotenv(str(ROOT_DIR / ".env.compose"), override=True)
        load_dotenv(str(ROOT_DIR / ".env"), override=True)
    except ImportError:
        pass
    # Root .env owns the Go Core HTTP port; Next.js must use INTERFACE_PORT.
    os.environ.pop("PORT", None)


def _task_env(extra=None):
    _load_env()
    ensure_managed_cache_dirs()
    return managed_cache_env(extra=extra)


@dataclass
class CommandResult:
    exited: int
    stdout: str = ""
    stderr: str = ""


def _is_next_build_lock_conflict(result: CommandResult) -> bool:
    text = _normalize_process_text(f"{result.stdout}\n{result.stderr}")
    return "unable to acquire lock at" in text and ".next/lock" in text


def _is_incomplete_next_build_output(result: CommandResult) -> bool:
    text = _normalize_process_text(f"{result.stdout}\n{result.stderr}")
    if ".next/" not in text:
        return False
    incomplete_outputs = ("required-server-files.json", "build-manifest.json", "pages-manifest.json", ".nft.json")
    if "enoent" in text and any(name in text for name in incomplete_outputs):
        return True
    return ".next/types/" in text and "not found" in text


def _is_next_standalone_cleanup_conflict(result: CommandResult) -> bool:
    text = _normalize_process_text(f"{result.stdout}\n{result.stderr}")
    return "ebusy" in text and "rmdir" in text and ".next/standalone" in text


def _expected_next_build_artifacts() -> list[Path]:
    from . import interface_runtime as runtime

    next_dir = runtime.INTERFACE_DIR / ".next"
    build_manifest_path = next_dir / "build-manifest.json"
    artifacts = [build_manifest_path, next_dir / "required-server-files.json"]
    if not build_manifest_path.exists():
        return artifacts

    payload = json.loads(build_manifest_path.read_text(encoding="utf-8"))
    manifest_entries: list[str] = []
    manifest_entries.extend(payload.get("polyfillFiles", []))
    manifest_entries.extend(payload.get("lowPriorityFiles", []))
    manifest_entries.extend(payload.get("rootMainFiles", []))
    for page_entries in payload.get("pages", {}).values():
        manifest_entries.extend(page_entries or [])

    seen: set[Path] = set()
    for relative_path in manifest_entries:
        artifact_path = next_dir / str(relative_path).replace("/", os.sep)
        if artifact_path not in seen:
            artifacts.append(artifact_path)
            seen.add(artifact_path)

    return artifacts


def _wait_for_complete_next_build_output(timeout_seconds: int = 20) -> None:
    from . import interface_runtime as runtime

    deadline = time.time() + timeout_seconds
    missing: list[Path] = []
    while time.time() < deadline:
        missing = [path for path in _expected_next_build_artifacts() if not path.exists()]
        if not missing:
            return
        time.sleep(0.5)

    missing_preview = ", ".join(str(path.relative_to(runtime.INTERFACE_DIR)) for path in missing[:6])
    raise RuntimeError(f"Incomplete Next.js build output after successful build command: {missing_preview}")


def _playwright_last_run_path() -> Path:
    return INTERFACE_DIR / "test-results" / ".last-run.json"


def _read_playwright_last_run_status() -> str | None:
    last_run_path = _playwright_last_run_path()
    if not last_run_path.exists():
        return None
    try:
        payload = json.loads(last_run_path.read_text(encoding="utf-8"))
    except (OSError, json.JSONDecodeError):
        return None
    status = payload.get("status")
    return status if isinstance(status, str) else None


def interface_task_env(extra=None):
    """Public wrapper for callers that need the managed Interface task env."""
    return _task_env(extra=extra)


def _interface_subprocess_env(extra_env: dict[str, str] | None = None) -> dict[str, str]:
    process_env = os.environ.copy()
    process_env.update(_task_env(extra_env))
    return process_env


def _resolve_interface_runner(command: list[str]) -> list[str]:
    runner = list(command)
    executable = runner[0]
    candidates = [executable, f"{executable}.cmd", f"{executable}.exe"] if is_windows() else [executable]
    for candidate in candidates:
        resolved = shutil.which(candidate)
        if resolved:
            runner[0] = resolved
            break
    return runner


def _run_interface_shell_command(command: list[str], extra_env: dict[str, str] | None = None) -> CommandResult:
    """Run a one-shot Interface command directly and return its exit data."""
    result = subprocess.run(
        _resolve_interface_runner(command),
        cwd=str(INTERFACE_DIR),
        env=_interface_subprocess_env(extra_env),
        capture_output=True,
        text=True,
        encoding="utf-8",
        errors="replace",
    )
    return CommandResult(exited=result.returncode, stdout=result.stdout or "", stderr=result.stderr or "")


def _run_interface_shell_command_streaming(command: list[str], extra_env: dict[str, str] | None = None) -> CommandResult:
    """Run an Interface command with inherited stdio for long-lived browser flows."""
    result = subprocess.run(
        _resolve_interface_runner(command),
        cwd=str(INTERFACE_DIR),
        env=_interface_subprocess_env(extra_env),
        text=True,
        encoding="utf-8",
        errors="replace",
    )
    return CommandResult(exited=result.returncode, stdout="", stderr="")


def _report_command_result(result: CommandResult) -> None:
    _print_ascii_safe(result.stdout)
    _print_ascii_safe(result.stderr)


def _run_one_shot_interface_task(
    command: list[str],
    *,
    extra_env: dict[str, str] | None = None,
    cleanup: bool = True,
) -> CommandResult:
    from . import interface_runtime as runtime

    if cleanup:
        runtime._cleanup_repo_local_interface_processes()
    try:
        result = runtime._run_interface_shell_command(command, extra_env=extra_env)
        runtime._report_command_result(result)
        if result.exited != 0:
            raise SystemExit(result.exited)
        return result
    finally:
        if cleanup:
            runtime._cleanup_repo_local_interface_processes()


def _run_interface_commandline(
    command: str,
    *,
    extra_env: dict[str, str] | None = None,
    stream: bool = False,
) -> CommandResult:
    from . import interface_runtime as runtime

    args = shlex.split(command, posix=not is_windows())
    if stream:
        if len(args) >= 3 and args[:3] == ["npx", "playwright", "test"]:
            return runtime._run_playwright_command_streaming(args, extra_env=extra_env)
        return runtime._run_interface_shell_command_streaming(args, extra_env=extra_env)
    return runtime._run_interface_shell_command(args, extra_env=extra_env)


def _build_playwright_command(*, project: str = "", spec: str = "", workers: str = "", headed: bool = False) -> str:
    cmd = "npx playwright test --reporter=dot"
    effective_workers = workers or "1"
    if project:
        cmd += f" --project={project}"
    if spec:
        cmd += f" {spec}"
    if effective_workers:
        cmd += f" --workers={effective_workers}"
    if headed:
        cmd += " --headed"
    return cmd


def _build_playwright_env(*, live_backend: bool, port: int) -> dict[str, str]:
    bind_host = "127.0.0.1" if is_windows() else INTERFACE_BIND_HOST
    extra_env = {
        "PLAYWRIGHT_SKIP_WEBSERVER": "1",
        "INTERFACE_HOST": "127.0.0.1",
        "INTERFACE_BIND_HOST": bind_host,
        "INTERFACE_PORT": str(port),
    }
    if live_backend:
        extra_env["PLAYWRIGHT_LIVE_BACKEND"] = "1"
    return _task_env(extra_env)


def _reconcile_managed_server_endpoint(
    env: dict[str, str],
    chosen_port: int,
    server: subprocess.Popen[str],
) -> tuple[dict[str, str], int]:
    from . import interface_runtime as runtime

    actual_port = runtime._detect_playwright_server_port(chosen_port)
    if actual_port != chosen_port:
        print(f"  Managed server bound to port {actual_port} instead of requested {chosen_port}")
        chosen_port = actual_port
        env["INTERFACE_PORT"] = str(actual_port)

    ready_host = runtime._wait_for_interface_ready(host=env["INTERFACE_HOST"], port=chosen_port, process=server)
    if ready_host != env["INTERFACE_HOST"]:
        print(f"  Managed server is reachable via {ready_host}; updating Playwright host from {env['INTERFACE_HOST']}")
        env["INTERFACE_HOST"] = ready_host

    return env, chosen_port


def run_interface_command(c, command: str, cleanup=False, extra_env=None, **run_kwargs):
    """Run an Interface-local command from the interface/ working directory."""
    from . import interface_runtime as runtime

    command_env = run_kwargs.pop("env", None)
    if command_env is not None and extra_env is None:
        env = dict(command_env)
    else:
        env = _task_env(extra=extra_env)
        if command_env:
            env.update(command_env)
    try:
        if cleanup:
            runtime._cleanup_repo_local_interface_processes()
        with c.cd(str(INTERFACE_DIR)):
            return c.run(command, env=env, **run_kwargs)
    finally:
        if cleanup:
            runtime._cleanup_repo_local_interface_processes()


def _normalize_process_text(text: str) -> str:
    return (text or "").lower().replace("\\", "/")


def _print_ascii_safe(text: str):
    if not text:
        return
    print(text.encode("ascii", "replace").decode("ascii"), end="")
