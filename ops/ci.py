"""
Local CI task entrypoints for operator and workflow use.

Usage:
    uv run inv ci.lint
    uv run inv ci.test
    uv run inv ci.build
    uv run inv ci.check
    uv run inv ci.deploy
"""

import ipaddress
import os
import re
import urllib.error
import urllib.request
from contextlib import suppress
from pathlib import Path
from urllib.parse import urljoin, urlparse

from invoke import Collection, task

from . import cache as cache_tasks
from . import ci_pipeline, ci_release
from . import core as core_tasks
from . import db as db_tasks
from . import interface as interface_tasks
from . import lifecycle, logging as logging_tasks, quality
from .config import CORE_DIR, ROOT_DIR, ensure_managed_cache_dirs, managed_cache_env, running_in_wsl


def _task_env(extra=None):
    ensure_managed_cache_dirs()
    return managed_cache_env(extra=extra)


def _read_env_file(path: Path) -> dict[str, str]:
    if not path.exists():
        return {}
    values: dict[str, str] = {}
    for raw_line in path.read_text(encoding="utf-8").splitlines():
        line = raw_line.strip()
        if not line or line.startswith("#") or "=" not in raw_line:
            continue
        key, value = raw_line.split("=", 1)
        cleaned = value.strip().strip('"').strip("'")
        if cleaned:
            values[key.strip()] = cleaned
    return values


def _runtime_posture_env_values() -> dict[str, str]:
    values: dict[str, str] = {}
    for path_name in (".env", ".env.compose"):
        values.update(_read_env_file(ROOT_DIR / path_name))
    values.update({key: value for key, value in os.environ.items() if value})
    return values


def _configured_ai_endpoints(env_values: dict[str, str] | None = None) -> list[tuple[str, str, str]]:
    values = env_values or _runtime_posture_env_values()
    endpoints: list[tuple[str, str, str]] = []
    seen_env_names: set[str] = set()
    for env_name, label in (
        ("MYCELIS_COMPOSE_OLLAMA_HOST", "compose Ollama host"),
        ("MYCELIS_K8S_TEXT_ENDPOINT", "k8s text endpoint"),
        ("MYCELIS_K8S_MEDIA_ENDPOINT", "k8s media endpoint"),
        ("MYCELIS_PROVIDER_LOCAL_OLLAMA_DEV_ENDPOINT", "local Ollama provider endpoint"),
    ):
        raw = (values.get(env_name, "") or "").strip()
        if raw:
            endpoints.append((env_name, label, raw))
            seen_env_names.add(env_name)

    for env_name in sorted(values):
        if env_name in seen_env_names:
            continue
        if not re.fullmatch(r"MYCELIS_PROVIDER_[A-Z0-9_]+_ENDPOINT", env_name):
            continue
        raw = (values.get(env_name, "") or "").strip()
        if not raw:
            continue
        provider_name = env_name.removeprefix("MYCELIS_PROVIDER_").removesuffix("_ENDPOINT").replace("_", " ").lower()
        endpoints.append((env_name, f"{provider_name} provider endpoint", raw))
    return endpoints


def _probe_paths_for_endpoint(env_name: str, raw: str) -> tuple[str, str]:
    parsed = urlparse(raw)
    if env_name == "MYCELIS_COMPOSE_OLLAMA_HOST" or parsed.path.rstrip("/").endswith("/api"):
        return ("/api/tags", "/v1/models")
    return ("/models", "/api/tags")


def _probe_urls_for_endpoint(env_name: str, raw: str) -> list[str]:
    urls = [urljoin(raw.rstrip("/") + "/", path.lstrip("/")) for path in _probe_paths_for_endpoint(env_name, raw)]
    parsed = urlparse(raw)
    if (
        env_name == "MYCELIS_COMPOSE_OLLAMA_HOST"
        and running_in_wsl()
        and (parsed.hostname or "").strip().lower() == "host.docker.internal"
    ):
        local_base = f"{parsed.scheme}://127.0.0.1:{parsed.port or (443 if parsed.scheme == 'https' else 80)}"
        urls.extend(urljoin(local_base.rstrip("/") + "/", path.lstrip("/")) for path in _probe_paths_for_endpoint(env_name, raw))

    deduped: list[str] = []
    seen: set[str] = set()
    for url in urls:
        if url not in seen:
            seen.add(url)
            deduped.append(url)
    return deduped


def _is_loopback_or_unspecified_host(host: str) -> bool:
    normalized = (host or "").strip().lower()
    if not normalized or normalized == "localhost":
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
            return response.status, response.read().decode("utf-8", errors="replace")
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
        raise SystemExit(
            "RUNTIME POSTURE CHECK FAILED: no explicit AI endpoint configured in process env, .env.compose, or .env. "
            "Set MYCELIS_COMPOSE_OLLAMA_HOST, a Kubernetes endpoint, or a provider-specific endpoint override."
        )

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

        reachable = False
        for probe_url in _probe_urls_for_endpoint(env_name, raw):
            status, _body = _probe_http_endpoint(probe_url)
            if status in {200, 401, 403}:
                print(f"  [OK]   {label}: {probe_url} [{status}]")
                reachable = True
                break
            print(f"  [WARN] {label}: {probe_url} [{status}]")
        if not reachable:
            failures.append(f"{env_name}: no AI probe path responded successfully")

    if failures:
        raise SystemExit("RUNTIME POSTURE CHECK FAILED: " + "; ".join(failures))
    print("RUNTIME POSTURE PASSED")


@task
def lint(c):
    """Lint: Go vet + Next.js lint."""
    ci_pipeline.run_lint(c, core_dir=CORE_DIR, task_env=_task_env, interface_tasks=interface_tasks)


@task
def test(c):
    """Test: Go unit tests + Interface tests."""
    ci_pipeline.run_test(c, core_dir=CORE_DIR, task_env=_task_env, interface_tasks=interface_tasks)


@task
def build(c):
    """Build: Go binary + Next.js production build (no Docker)."""
    ci_pipeline.run_build(c, cache_tasks=cache_tasks, core_tasks=core_tasks, interface_tasks=interface_tasks)


@task
def check(c):
    """Full local CI pipeline: lint -> test -> build."""
    ci_pipeline.run_check(c, lint_task=lint, test_task=test, build_task=build)


@task(help={"e2e": "Include Playwright E2E run (default: True)."})
def baseline(c, e2e=True):
    """Strict baseline validation for delivery readiness."""
    ci_pipeline.run_baseline(
        c,
        e2e=e2e,
        cache_tasks=cache_tasks,
        logging_tasks=logging_tasks,
        quality=quality,
        core_dir=CORE_DIR,
        task_env=_task_env,
        interface_tasks=interface_tasks,
    )


@task(help={"live_backend": "Also run the live-backend governed Soma browser contract after health checks."})
def service_check(c, live_backend=False):
    """Validate the currently running local stack."""
    ci_release.run_service_check(c, live_backend=live_backend, lifecycle=lifecycle, db_tasks=db_tasks, interface_tasks=interface_tasks)


@task(help={"strict": "Fail if Go version does not match the locked policy (default: False)."})
def toolchain_check(c, strict=False):
    """Report local toolchain versions and optionally enforce Go lock policy."""
    ci_release.run_toolchain_check(c, strict=strict)


@task
def entrypoint_check(c):
    """Verify the supported invoke runner matrix."""
    ci_release.run_entrypoint_check(c)


@task(
    help={
        "lane": "Preset gate lane: baseline, runtime, service, or release (default: baseline).",
        "e2e": "Include Playwright in baseline gate (default: True).",
        "strict_toolchain": "Fail on Go lock mismatch (default: False).",
        "service_health": "Require lifecycle.health against the running local stack (default: False).",
        "live_backend": "Also run the live-backend workspace Playwright contract when service-health is enabled (default: False).",
        "runtime_posture": "Also check tighter disk headroom and explicit AI endpoint reachability when configured (default: False).",
    }
)
def release_preflight(c, lane="baseline", e2e=True, strict_toolchain=False, service_health=False, live_backend=False, runtime_posture=False):
    """Enforce release preflight gate."""
    ci_release.run_release_preflight(
        c,
        lane=lane,
        e2e=e2e,
        strict_toolchain=strict_toolchain,
        service_health=service_health,
        live_backend=live_backend,
        runtime_posture=runtime_posture,
        runtime_posture_check=_runtime_posture_check,
        toolchain_check=toolchain_check,
        baseline=baseline,
        service_check=service_check,
    )


@task
def deploy(c):
    """Build + Docker + K8s deploy."""
    ci_release.run_deploy(c, lint_task=lint, test_task=test)


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
