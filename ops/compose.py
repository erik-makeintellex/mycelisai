import socket
import subprocess
import time
import json
from pathlib import Path
from typing import Callable

from invoke import task, Collection

from . import db as db_tasks
from .config import API_HOST, API_PORT, INTERFACE_HOST, INTERFACE_PORT, ROOT_DIR


COMPOSE_FILE = ROOT_DIR / "docker-compose.yml"
COMPOSE_ENV_FILE = ROOT_DIR / ".env.compose"
COMPOSE_ENV_EXAMPLE = ROOT_DIR / ".env.compose.example"
COMPOSE_PROJECT = "mycelis-home"


def _compose_command(*args: str) -> list[str]:
    return [
        "docker",
        "compose",
        "--project-name",
        COMPOSE_PROJECT,
        "--env-file",
        str(COMPOSE_ENV_FILE),
        "-f",
        str(COMPOSE_FILE),
        *args,
    ]


def _require_compose_env_file():
    if COMPOSE_ENV_FILE.exists():
        return
    raise SystemExit(
        f"Missing {COMPOSE_ENV_FILE.name}. Copy {COMPOSE_ENV_EXAMPLE.name} to "
        f"{COMPOSE_ENV_FILE.name} and set MYCELIS_API_KEY before running compose tasks."
    )


def _load_compose_env() -> dict[str, str]:
    _require_compose_env_file()
    values: dict[str, str] = {}
    for raw_line in COMPOSE_ENV_FILE.read_text(encoding="utf-8").splitlines():
        line = raw_line.strip()
        if not line or line.startswith("#") or "=" not in line:
            continue
        key, value = line.split("=", 1)
        values[key.strip()] = value.strip()
    return values


def _port_open(port: int, host: str = "127.0.0.1", timeout: float = 1.0) -> bool:
    try:
        with socket.create_connection((host, port), timeout=timeout):
            return True
    except (ConnectionRefusedError, TimeoutError, OSError):
        return False


def _normalize_bool(value: str) -> bool:
    return value.strip().lower() in {"1", "true", "yes", "on"}


def _looks_like_container_loopback(url: str) -> bool:
    candidate = url.strip().lower()
    for prefix in ("http://", "https://"):
        if candidate.startswith(prefix):
            candidate = candidate[len(prefix):]
            break
    candidate = candidate.split("/", 1)[0]
    host = candidate.split(":", 1)[0]
    return host in {"127.0.0.1", "localhost", "0.0.0.0"}


def _validate_compose_env(env_values: dict[str, str]):
    ollama_host = env_values.get("MYCELIS_COMPOSE_OLLAMA_HOST", "http://host.docker.internal:11434").strip()
    if ollama_host and _looks_like_container_loopback(ollama_host):
        raise SystemExit(
            "Invalid .env.compose MYCELIS_COMPOSE_OLLAMA_HOST for Docker Compose: "
            f"{ollama_host}. Use a host-reachable address such as "
            "http://host.docker.internal:11434 or another container/service hostname."
        )


def _print_step(step: int, total: int, title: str, expectation: str | None = None):
    print(f"[{step}/{total}] {title}")
    if expectation:
        print(f"  Expect: {expectation}")


def _failure_guidance(message: str, *next_steps: str) -> str:
    lines = [message]
    if next_steps:
        lines.append("Next steps:")
        lines.extend(f"- {step}" for step in next_steps)
    return "\n".join(lines)


def _wait_for_port(port: int, label: str, timeout_seconds: int = 60) -> bool:
    deadline = time.time() + timeout_seconds
    while time.time() < deadline:
        if _port_open(port):
            return True
        time.sleep(1)
    print(f"  TIMEOUT waiting for {label} on localhost:{port} after {timeout_seconds}s")
    return False


def _http_get(url: str, timeout: float = 3.0, headers: dict[str, str] | None = None) -> tuple[int, str]:
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


def _wait_for_http_ok(url: str, label: str, timeout_seconds: int = 60, headers: dict[str, str] | None = None) -> bool:
    deadline = time.time() + timeout_seconds
    while time.time() < deadline:
        status, _body = _http_get(url, timeout=5.0, headers=headers)
        if status == 200:
            return True
        time.sleep(1)
    print(f"  TIMEOUT waiting for {label} at {url} after {timeout_seconds}s")
    return False


def _run_compose(args: list[str], check: bool = True) -> subprocess.CompletedProcess[str]:
    result = subprocess.run(args, text=True, capture_output=True)
    if result.stdout:
        print(result.stdout.rstrip())
    if result.stderr:
        print(result.stderr.rstrip())
    if check and result.returncode != 0:
        raise SystemExit(f"Compose command failed: {' '.join(args)}")
    return result


def _expect_stage(
    check: Callable[..., bool],
    *args,
    failure: str,
    next_steps: list[str],
    **kwargs,
):
    if check(*args, **kwargs):
        return
    raise SystemExit(_failure_guidance(failure, *next_steps))


def _compose_query_succeeds(sql: str) -> bool:
    result = _run_compose(
        _compose_command(
            "exec",
            "-T",
            "postgres",
            "psql",
            "-t",
            "-A",
            "-h",
            "127.0.0.1",
            "-U",
            "mycelis",
            "-d",
            "cortex",
            "-c",
            sql,
        ),
        check=False,
    )
    return result.returncode == 0 and "1" in result.stdout.split()


def _compose_schema_bootstrapped() -> bool:
    for _label, sql in db_tasks.SCHEMA_COMPATIBILITY_CHECKS:
        if not _compose_query_succeeds(sql):
            return False
    return True


def _run_compose_migrations(strict: bool = False):
    if not strict and _compose_schema_bootstrapped():
        print(
            "Compose schema already appears compatible with the current runtime; "
            "skipping forward migration replay."
        )
        print(
            "Use 'uv run inv compose.down --volumes' for a truly fresh compose rebuild "
            "when you need to replay the canonical migration stack end-to-end."
        )
        return

    for migration in db_tasks._migration_files():
        result = _run_compose(
            _compose_command(
                "exec",
                "-T",
                "postgres",
                "psql",
                "-v",
                "ON_ERROR_STOP=1",
                "-h",
                "127.0.0.1",
                "-U",
                "mycelis",
                "-d",
                "cortex",
                "-f",
                f"/migrations/{migration.name}",
            ),
            check=False,
        )
        if result.returncode != 0:
            raise SystemExit(f"Compose migration failed: {migration.name}")


def _wait_for_postgres_ready(timeout_seconds: int = 90) -> bool:
    deadline = time.time() + timeout_seconds
    while time.time() < deadline:
        result = _run_compose(
            _compose_command(
                "exec",
                "-T",
                "postgres",
                "pg_isready",
                "-h",
                "127.0.0.1",
                "-U",
                "mycelis",
                "-d",
                "cortex",
            ),
            check=False,
        )
        if result.returncode == 0:
            return True
        time.sleep(2)
    return False


@task(
    help={
        "build": "Build the core/interface images during bring-up (default: False).",
        "wait_timeout": "Seconds to wait for each bring-up readiness phase (default: 180).",
    }
)
def up(c, build=False, wait_timeout=180):
    """Bring up the home-runtime Docker Compose stack with managed ordering."""
    del c
    _require_compose_env_file()
    env_values = _load_compose_env()
    _validate_compose_env(env_values)
    wait_timeout = int(wait_timeout)
    if wait_timeout < 30:
        raise SystemExit("Compose up wait timeout must be at least 30 seconds.")

    print("=== Mycelis Compose Up ===\n")
    _print_step(
        1,
        4,
        "Starting PostgreSQL and NATS...",
        (
            "Docker should start the infra containers first. With --build, image preparation can take "
            "several minutes before readiness checks begin."
        ),
    )
    infra_cmd = _compose_command("up", "-d")
    if build:
        infra_cmd.append("--build")
    infra_cmd.extend(["postgres", "nats"])
    _run_compose(infra_cmd)

    _expect_stage(
        _wait_for_port,
        5432,
        "PostgreSQL",
        timeout_seconds=wait_timeout,
        failure=f"Compose up failed: PostgreSQL did not become reachable within {wait_timeout}s.",
        next_steps=[
            "Run 'uv run inv compose.status' to inspect container state and host ports.",
            "Run 'uv run inv compose.logs postgres' for PostgreSQL startup logs.",
            "Use 'uv run inv compose.down --volumes' if you need a fully fresh rebuild.",
        ],
    )
    _expect_stage(
        _wait_for_postgres_ready,
        timeout_seconds=max(wait_timeout, 90),
        failure=f"Compose up failed: PostgreSQL did not become query-ready within {max(wait_timeout, 90)}s.",
        next_steps=[
            "Run 'uv run inv compose.logs postgres' to inspect readiness or migration errors.",
            "Verify the local PostgreSQL port is not already occupied by another service.",
        ],
    )
    _expect_stage(
        _wait_for_port,
        4222,
        "NATS",
        timeout_seconds=wait_timeout,
        failure=f"Compose up failed: NATS did not become reachable within {wait_timeout}s.",
        next_steps=[
            "Run 'uv run inv compose.status' to confirm the NATS container is up.",
            "Run 'uv run inv compose.logs nats' for broker startup details.",
        ],
    )

    print()
    _print_step(
        2,
        4,
        "Applying canonical migrations through the PostgreSQL container...",
        "Schema bootstrap may skip replay when the compose database is already compatible with the current runtime.",
    )
    _run_compose_migrations()

    print()
    _print_step(
        3,
        4,
        "Starting Core and Interface...",
        (
            "Core should answer /healthz and the frontend root should load. With --build, app image creation can "
            "take several additional minutes before the HTTP checks pass."
        ),
    )
    app_cmd = _compose_command("up", "-d")
    if build:
        app_cmd.append("--build")
    app_cmd.extend(["core", "interface"])
    _run_compose(app_cmd)

    _expect_stage(
        _wait_for_port,
        API_PORT,
        "Core API",
        timeout_seconds=wait_timeout,
        failure=f"Compose up failed: Core API port did not open within {wait_timeout}s.",
        next_steps=[
            "Run 'uv run inv compose.logs core' to inspect backend startup output.",
            "Run 'uv run inv compose.health' after fixing backend readiness to verify the runtime contract.",
        ],
    )
    _expect_stage(
        _wait_for_http_ok,
        f"http://{API_HOST}:{API_PORT}/healthz",
        "Core health",
        timeout_seconds=wait_timeout,
        failure=f"Compose up failed: Core /healthz did not become ready within {wait_timeout}s.",
        next_steps=[
            "Run 'uv run inv compose.logs core' to inspect healthz startup failures.",
            "Check provider and database configuration in '.env.compose' if Core stays unhealthy.",
        ],
    )
    _expect_stage(
        _wait_for_port,
        INTERFACE_PORT,
        "Frontend",
        timeout_seconds=wait_timeout,
        failure=f"Compose up failed: Frontend port did not open within {wait_timeout}s.",
        next_steps=[
            "Run 'uv run inv compose.logs interface' to inspect frontend startup output.",
            "Verify the host port is not already in use by another local UI server.",
        ],
    )
    _expect_stage(
        _wait_for_http_ok,
        f"http://{INTERFACE_HOST}:{INTERFACE_PORT}/",
        "Frontend",
        timeout_seconds=wait_timeout,
        failure=f"Compose up failed: Frontend root did not become ready within {wait_timeout}s.",
        next_steps=[
            "Run 'uv run inv compose.logs interface' for frontend build or startup failures.",
            "Run 'uv run inv compose.status' to confirm the UI container is still running.",
        ],
    )

    print()
    _print_step(
        4,
        4,
        "Compose stack ready.",
        "Next expected checks: 'uv run inv compose.health' for deep runtime proof, then open http://localhost:3000.",
    )
    status.body(None)


@task(help={"volumes": "Also remove named volumes for a fully fresh rebuild (default: False)."})
def down(c, volumes=False):
    """Stop the Docker Compose home-runtime stack."""
    del c
    _require_compose_env_file()
    print("=== Mycelis Compose Down ===\n")
    cmd = _compose_command("down")
    if volumes:
        cmd.append("--volumes")
    _run_compose(cmd)


@task
def migrate(c):
    """Apply canonical forward migrations through the PostgreSQL compose service."""
    del c
    _require_compose_env_file()
    print("=== Mycelis Compose Migrate ===\n")
    _run_compose_migrations()


@task
def status(c):
    """Show Docker Compose service state plus key host port reachability."""
    del c
    _require_compose_env_file()
    env_values = _load_compose_env()
    _validate_compose_env(env_values)
    print("=== Mycelis Compose Status ===\n")
    _run_compose(_compose_command("ps"), check=False)

    checks = [
        ("PostgreSQL", int(env_values.get("MYCELIS_COMPOSE_POSTGRES_PORT", "5432"))),
        ("NATS", int(env_values.get("MYCELIS_COMPOSE_NATS_PORT", "4222"))),
        ("NATS Monitor", int(env_values.get("MYCELIS_COMPOSE_NATS_MONITOR_PORT", "8222"))),
        ("Core API", int(env_values.get("MYCELIS_COMPOSE_CORE_PORT", str(API_PORT)))),
        ("Frontend", int(env_values.get("MYCELIS_COMPOSE_INTERFACE_PORT", str(INTERFACE_PORT)))),
    ]
    print()
    for label, port in checks:
        state = "UP" if _port_open(port) else "DOWN"
        print(f"  {label:<14}: {state} [:{port}]")


@task
def health(c):
    """Deep health probe for the Docker Compose runtime path."""
    del c
    _require_compose_env_file()
    env_values = _load_compose_env()
    _validate_compose_env(env_values)
    api_key = env_values.get("MYCELIS_API_KEY", "")
    headers = {"Authorization": f"Bearer {api_key}"} if api_key else {}

    checks = [
        ("Core health", f"http://{API_HOST}:{API_PORT}/healthz", None),
        ("Template Engine", f"http://{API_HOST}:{API_PORT}/api/v1/templates", headers),
        ("Brains API", f"http://{API_HOST}:{API_PORT}/api/v1/brains", headers),
        ("Telemetry", f"http://{API_HOST}:{API_PORT}/api/v1/telemetry/compute", headers),
        ("Frontend", f"http://{INTERFACE_HOST}:{INTERFACE_PORT}/", None),
        ("NATS Monitor", f"http://127.0.0.1:{env_values.get('MYCELIS_COMPOSE_NATS_MONITOR_PORT', '8222')}/varz", None),
    ]

    print("=== Mycelis Compose Health ===\n")
    failures: list[str] = []
    for label, url, request_headers in checks:
        status, body = _http_get(url, timeout=5.0, headers=request_headers)
        if status == 200:
            print(f"  [OK] {label:<18} {url}")
        else:
            print(f"  [FAIL] {label:<18} {url} [{status}]")
            failures.append(f"{label}: {body}")

    cognitive_status, cognitive_body = _http_get(
        f"http://{API_HOST}:{API_PORT}/api/v1/cognitive/status",
        timeout=5.0,
        headers=headers,
    )
    if cognitive_status != 200:
        print(f"  [FAIL] {'Cognitive Engine':<18} http://{API_HOST}:{API_PORT}/api/v1/cognitive/status [{cognitive_status}]")
        failures.append(f"Cognitive Engine: {cognitive_body}")
    else:
        try:
            payload = json.loads(cognitive_body)
        except json.JSONDecodeError as exc:
            print(f"  [FAIL] {'Cognitive Engine':<18} http://{API_HOST}:{API_PORT}/api/v1/cognitive/status [invalid-json]")
            failures.append(f"Cognitive Engine: invalid JSON response ({exc})")
        else:
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

    if failures:
        print("\nIssues:")
        for failure in failures:
            print(f"  - {failure}")
        raise SystemExit(1)

    print("\nAll compose endpoints healthy.")


@task(help={"service": "Optional compose service name.", "tail": "Number of log lines to show (default: 200)."})
def logs(c, service="", tail=200):
    """Show compose logs for the full stack or a single service."""
    del c
    _require_compose_env_file()
    cmd = _compose_command("logs", "--tail", str(int(tail)))
    if service:
        cmd.append(service)
    subprocess.run(cmd, check=False)


ns = Collection("compose")
ns.add_task(up)
ns.add_task(down)
ns.add_task(migrate)
ns.add_task(status)
ns.add_task(health)
ns.add_task(logs)
