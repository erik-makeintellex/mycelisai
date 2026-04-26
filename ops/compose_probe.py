from __future__ import annotations

import json
import subprocess
import time
from typing import Callable


def print_step(step: int, total: int, title: str, expectation: str | None = None):
    print(f"[{step}/{total}] {title}")
    if expectation:
        print(f"  Expect: {expectation}")


def failure_guidance(message: str, *next_steps: str) -> str:
    lines = [message]
    if next_steps:
        lines.append("Next steps:")
        lines.extend(f"- {step}" for step in next_steps)
    return "\n".join(lines)


def wait_for_port(
    port: int,
    label: str,
    *,
    port_open: Callable[[int], bool],
    timeout_seconds: int = 60,
) -> bool:
    deadline = time.time() + timeout_seconds
    while time.time() < deadline:
        if port_open(port):
            return True
        time.sleep(1)
    print(f"  TIMEOUT waiting for {label} on localhost:{port} after {timeout_seconds}s")
    return False


def http_get(url: str, timeout: float = 3.0, headers: dict[str, str] | None = None) -> tuple[int, str]:
    import urllib.error
    import urllib.request

    try:
        req = urllib.request.Request(url, headers=headers or {})
        with urllib.request.urlopen(req, timeout=timeout) as resp:
            return resp.status, resp.read().decode("utf-8", errors="replace")
    except urllib.error.HTTPError as exc:
        return exc.code, str(exc)
    except Exception as exc:
        return 0, str(exc)


def wait_for_http_ok(
    url: str,
    label: str,
    *,
    http_getter: Callable[..., tuple[int, str]],
    timeout_seconds: int = 60,
    headers: dict[str, str] | None = None,
) -> bool:
    deadline = time.time() + timeout_seconds
    while time.time() < deadline:
        status, _body = http_getter(url, timeout=5.0, headers=headers)
        if status == 200:
            return True
        time.sleep(1)
    print(f"  TIMEOUT waiting for {label} at {url} after {timeout_seconds}s")
    return False


def run_compose(
    args: list[str],
    check: bool = True,
    env: dict[str, str] | None = None,
) -> subprocess.CompletedProcess[str]:
    result = subprocess.run(args, text=True, capture_output=True, env=env)
    if result.stdout:
        print(result.stdout.rstrip())
    if result.stderr:
        print(result.stderr.rstrip())
    if check and result.returncode != 0:
        raise SystemExit(f"Compose command failed: {' '.join(args)}")
    return result


def expect_stage(
    check: Callable[..., bool],
    *args,
    failure: str,
    next_steps: list[str],
    failure_guidance: Callable[..., str],
    **kwargs,
):
    if check(*args, **kwargs):
        return
    raise SystemExit(failure_guidance(failure, *next_steps))


def print_status(
    env_values: dict[str, str],
    *,
    run_compose: Callable[..., subprocess.CompletedProcess[str]],
    compose_command: Callable[..., list[str]],
    compose_runtime_env: Callable[[dict[str, str]], dict[str, str] | None],
    port_open: Callable[[int], bool],
    api_port: int,
    interface_port: int,
):
    run_compose(compose_command("ps"), check=False, env=compose_runtime_env(env_values))

    checks = [
        ("PostgreSQL", int(env_values.get("MYCELIS_COMPOSE_POSTGRES_PORT", "5432"))),
        ("NATS", int(env_values.get("MYCELIS_COMPOSE_NATS_PORT", "4222"))),
        ("NATS Monitor", int(env_values.get("MYCELIS_COMPOSE_NATS_MONITOR_PORT", "8222"))),
        ("Core API", int(env_values.get("MYCELIS_COMPOSE_CORE_PORT", str(api_port)))),
        ("Frontend", int(env_values.get("MYCELIS_COMPOSE_INTERFACE_PORT", str(interface_port)))),
    ]
    print()
    for label, port in checks:
        state = "UP" if port_open(port) else "DOWN"
        print(f"  {label:<14}: {state} [:{port}]")


def run_infra_health(
    env_values: dict[str, str],
    *,
    compose_host_port: Callable[[dict[str, str], str, str], int],
    port_open: Callable[[int], bool],
    wait_for_postgres_ready: Callable[..., bool],
    http_getter: Callable[..., tuple[int, str]],
    compose_db_user: Callable[[dict[str, str]], str],
    compose_db_name: Callable[[dict[str, str]], str],
    print_data_plane_connection_guidance: Callable[[dict[str, str]], None],
):
    postgres_port = compose_host_port(env_values, "MYCELIS_COMPOSE_POSTGRES_PORT", "5432")
    nats_port = compose_host_port(env_values, "MYCELIS_COMPOSE_NATS_PORT", "4222")
    nats_monitor_port = compose_host_port(env_values, "MYCELIS_COMPOSE_NATS_MONITOR_PORT", "8222")
    failures: list[str] = []

    if port_open(postgres_port):
        print(f"  [OK] {'PostgreSQL port':<18} 127.0.0.1:{postgres_port}")
    else:
        print(f"  [FAIL] {'PostgreSQL port':<18} 127.0.0.1:{postgres_port}")
        failures.append("PostgreSQL host port is not reachable.")

    db_identity = f"{compose_db_user(env_values)}@postgres/{compose_db_name(env_values)}"
    if wait_for_postgres_ready(timeout_seconds=5, env_values=env_values):
        print(f"  [OK] {'PostgreSQL query':<18} {db_identity}")
    else:
        print(f"  [FAIL] {'PostgreSQL query':<18} {db_identity}")
        failures.append("PostgreSQL container did not accept a query with the configured DB user/name.")

    if port_open(nats_port):
        print(f"  [OK] {'NATS port':<18} nats://127.0.0.1:{nats_port}")
    else:
        print(f"  [FAIL] {'NATS port':<18} nats://127.0.0.1:{nats_port}")
        failures.append("NATS host port is not reachable.")

    nats_monitor_url = f"http://127.0.0.1:{nats_monitor_port}/varz"
    status_code, body = http_getter(nats_monitor_url, timeout=5.0)
    if status_code == 200:
        print(f"  [OK] {'NATS monitor':<18} {nats_monitor_url}")
    else:
        print(f"  [FAIL] {'NATS monitor':<18} {nats_monitor_url} [{status_code}]")
        failures.append(f"NATS monitor did not answer /varz: {body}")

    if failures:
        print("\nIssues:")
        for failure in failures:
            print(f"  - {failure}")
        print("\nNext steps:")
        print("  - Run 'uv run inv compose.status' to inspect service/container state.")
        print("  - Run 'uv run inv compose.logs postgres' or 'uv run inv compose.logs nats' for service logs.")
        raise SystemExit(1)

    print_data_plane_connection_guidance(env_values)
    print("\nData plane healthy. Core/Interface may remain down for infra-only testing.")


def run_storage_health(
    env_values: dict[str, str],
    *,
    compose_storage_check_results: Callable[[dict[str, str]], list[tuple[str, bool]]],
):
    failures: list[str] = []
    for label, ok in compose_storage_check_results(env_values):
        if ok:
            print(f"  [OK] {label}")
        else:
            print(f"  [FAIL] {label}")
            failures.append(label)

    if failures:
        print("\nIssues:")
        for failure in failures:
            print(f"  - Missing or unavailable: {failure}")
        print("\nNext steps:")
        print("  - Run 'uv run inv compose.migrate' to apply canonical schema migrations.")
        print("  - Run 'uv run inv compose.logs postgres' if migrations or storage checks fail.")
        print("  - Use 'uv run inv compose.down --volumes' only when an intentional destructive schema reset is required.")
        raise SystemExit(1)

    print("\nLong-term storage ready for semantic memory, continuity, artifacts, managed exchange, and template recall.")


def run_health(
    env_values: dict[str, str],
    *,
    http_getter: Callable[..., tuple[int, str]],
    api_host: str,
    api_port: int,
    interface_host: str,
    interface_port: int,
):
    api_key = env_values.get("MYCELIS_API_KEY", "")
    headers = {"Authorization": f"Bearer {api_key}"} if api_key else {}
    checks = [
        ("Core health", f"http://{api_host}:{api_port}/healthz", None),
        ("Template Engine", f"http://{api_host}:{api_port}/api/v1/templates", headers),
        ("Brains API", f"http://{api_host}:{api_port}/api/v1/brains", headers),
        ("Telemetry", f"http://{api_host}:{api_port}/api/v1/telemetry/compute", headers),
        ("Frontend", f"http://{interface_host}:{interface_port}/", None),
        ("NATS Monitor", f"http://127.0.0.1:{env_values.get('MYCELIS_COMPOSE_NATS_MONITOR_PORT', '8222')}/varz", None),
    ]

    failures: list[str] = []
    for label, url, request_headers in checks:
        status, body = http_getter(url, timeout=5.0, headers=request_headers)
        if status == 200:
            print(f"  [OK] {label:<18} {url}")
        else:
            print(f"  [FAIL] {label:<18} {url} [{status}]")
            failures.append(f"{label}: {body}")

    cognitive_url = f"http://{api_host}:{api_port}/api/v1/cognitive/status"
    cognitive_status, cognitive_body = http_getter(cognitive_url, timeout=5.0, headers=headers)
    if cognitive_status != 200:
        print(f"  [FAIL] {'Cognitive Engine':<18} {cognitive_url} [{cognitive_status}]")
        failures.append(f"Cognitive Engine: {cognitive_body}")
    else:
        _append_cognitive_health_failures(cognitive_url, cognitive_body, failures)

    if failures:
        print("\nIssues:")
        for failure in failures:
            print(f"  - {failure}")
        raise SystemExit(1)

    print("\nAll compose endpoints healthy.")


def _append_cognitive_health_failures(url: str, body: str, failures: list[str]):
    try:
        payload = json.loads(body)
    except json.JSONDecodeError as exc:
        print(f"  [FAIL] {'Cognitive Engine':<18} {url} [invalid-json]")
        failures.append(f"Cognitive Engine: invalid JSON response ({exc})")
        return

    text_state = str(payload.get("text", {}).get("status", "offline")).lower()
    media_state = str(payload.get("media", {}).get("status", "offline")).lower()
    if text_state == "online":
        print(f"  [OK] {'Cognitive Engine':<18} text={text_state}")
    else:
        print(f"  [FAIL] {'Cognitive Engine':<18} text={text_state}")
        failures.append(
            "Cognitive Engine: text inference is not online. "
            "Check OLLAMA_HOST or your configured provider reachability."
        )
    if media_state != "online":
        print(f"  [NOTE] {'Media Engine':<18} media={media_state}")
