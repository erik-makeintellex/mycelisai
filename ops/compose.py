import subprocess
from pathlib import Path

from invoke import task, Collection

from . import compose_env
from . import compose_probe
from . import compose_storage
from . import compose_wsl_relay
from . import db as db_tasks
from .config import (
    API_HOST,
    API_PORT,
    INTERFACE_HOST,
    INTERFACE_PORT,
    ROOT_DIR,
    docker_command,
    docker_host_mode,
    docker_host_path,
    running_in_wsl,
)


COMPOSE_FILE = ROOT_DIR / "docker-compose.yml"
COMPOSE_ENV_FILE = ROOT_DIR / ".env.compose"
COMPOSE_ENV_EXAMPLE = ROOT_DIR / ".env.compose.example"
COMPOSE_PROJECT = "mycelis-home"
DEFAULT_OUTPUT_HOST_PATH = ROOT_DIR / "workspace" / "docker-compose" / "data"
OUTPUT_BLOCK_MODES = compose_env.OUTPUT_BLOCK_MODES
WSL_OLLAMA_RELAY_NAME = f"{COMPOSE_PROJECT}-ollama-relay"
WSL_OLLAMA_RELAY_IMAGE = "alpine:3.21"
DEFAULT_WSL_OLLAMA_RELAY_PORT = 11435
COMPOSE_RUNTIME_OVERRIDE_KEYS = compose_env.COMPOSE_RUNTIME_OVERRIDE_KEYS


def _compose_command(*args: str) -> list[str]:
    return compose_env.compose_command(
        ROOT_DIR,
        COMPOSE_PROJECT,
        COMPOSE_ENV_FILE,
        COMPOSE_FILE,
        docker_command,
        docker_host_path,
        *args,
    )


def _require_compose_env_file():
    return compose_env.require_compose_env_file(COMPOSE_ENV_FILE, COMPOSE_ENV_EXAMPLE)


def _load_compose_env() -> dict[str, str]:
    return compose_env.load_compose_env(COMPOSE_ENV_FILE, _require_compose_env_file)


def _compose_effective_env(env_values: dict[str, str] | None = None) -> dict[str, str]:
    return compose_env.compose_effective_env(env_values, _load_compose_env)


def _wsl_exec_command(*args: str) -> list[str]:
    return compose_env.wsl_exec_command(*args)


def _port_open(port: int, host: str = "127.0.0.1", timeout: float = 1.0) -> bool:
    return compose_env.port_open(port, host=host, timeout=timeout)


def _normalize_bool(value: str) -> bool:
    return compose_env.normalize_bool(value)


def _looks_like_container_loopback(url: str) -> bool:
    return compose_env.looks_like_container_loopback(url)


def _clean_env_value(value: str) -> str:
    return compose_env.clean_env_value(value)


def _resolve_host_path(path_value: str) -> Path:
    return compose_env.resolve_host_path(path_value)


def _parse_network_endpoint(url: str) -> tuple[str, int]:
    return compose_env.parse_network_endpoint(url)


def _wsl_http_available(url: str) -> bool:
    return compose_env.wsl_http_available(url, _wsl_exec_command)


def _wsl_ollama_relay_port(env_values: dict[str, str]) -> int:
    return compose_env.wsl_ollama_relay_port(env_values, DEFAULT_WSL_OLLAMA_RELAY_PORT)


def _docker_run(args: list[str], check: bool = True) -> subprocess.CompletedProcess[str]:
    return compose_wsl_relay.docker_run(
        args,
        docker_command=docker_command,
        root_dir=ROOT_DIR,
        check=check,
    )


def _inspect_wsl_ollama_relay_labels() -> dict[str, str] | None:
    return compose_wsl_relay.inspect_labels(
        relay_name=WSL_OLLAMA_RELAY_NAME,
        docker_runner=_docker_run,
    )


def _stop_wsl_ollama_relay():
    compose_wsl_relay.stop(
        relay_name=WSL_OLLAMA_RELAY_NAME,
        docker_runner=_docker_run,
        docker_host_mode=docker_host_mode,
        running_in_wsl=running_in_wsl,
    )


def _ensure_wsl_ollama_relay(target_host: str, target_port: int, relay_port: int):
    compose_wsl_relay.ensure(
        target_host,
        target_port,
        relay_port,
        relay_name=WSL_OLLAMA_RELAY_NAME,
        relay_image=WSL_OLLAMA_RELAY_IMAGE,
        inspect_relay_labels=_inspect_wsl_ollama_relay_labels,
        stop_relay=_stop_wsl_ollama_relay,
        docker_runner=_docker_run,
    )


def _prepare_wsl_ollama_host(env_values: dict[str, str]) -> dict[str, str]:
    return compose_wsl_relay.prepare_host(
        env_values,
        docker_host_mode=docker_host_mode,
        running_in_wsl=running_in_wsl,
        clean_env_value=_clean_env_value,
        parse_network_endpoint=_parse_network_endpoint,
        relay_port=_wsl_ollama_relay_port,
        inspect_relay_labels=_inspect_wsl_ollama_relay_labels,
        wsl_http_available=_wsl_http_available,
        ensure_relay=_ensure_wsl_ollama_relay,
    )


def _validate_output_block_config(env_values: dict[str, str]):
    return compose_env.validate_output_block_config(
        env_values,
        DEFAULT_OUTPUT_HOST_PATH,
        output_block_modes=OUTPUT_BLOCK_MODES,
    )


def _validate_compose_env(env_values: dict[str, str]):
    return compose_env.validate_compose_env(env_values, _validate_output_block_config)


def _compose_runtime_env(env_values: dict[str, str] | None = None) -> dict[str, str] | None:
    return compose_env.compose_runtime_env(
        env_values,
        docker_host_mode,
        running_in_wsl,
        _compose_effective_env,
        DEFAULT_OUTPUT_HOST_PATH,
        docker_host_path,
    )


def _print_step(step: int, total: int, title: str, expectation: str | None = None):
    return compose_probe.print_step(step, total, title, expectation)


def _failure_guidance(message: str, *next_steps: str) -> str:
    return compose_probe.failure_guidance(message, *next_steps)


def _wait_for_port(port: int, label: str, timeout_seconds: int = 60) -> bool:
    return compose_probe.wait_for_port(
        port,
        label,
        port_open=_port_open,
        timeout_seconds=timeout_seconds,
    )


def _http_get(url: str, timeout: float = 3.0, headers: dict[str, str] | None = None) -> tuple[int, str]:
    return compose_probe.http_get(url, timeout=timeout, headers=headers)


def _wait_for_http_ok(url: str, label: str, timeout_seconds: int = 60, headers: dict[str, str] | None = None) -> bool:
    return compose_probe.wait_for_http_ok(
        url,
        label,
        http_getter=_http_get,
        timeout_seconds=timeout_seconds,
        headers=headers,
    )


def _run_compose(
    args: list[str],
    check: bool = True,
    env: dict[str, str] | None = None,
) -> subprocess.CompletedProcess[str]:
    return compose_probe.run_compose(args, check=check, env=env)


def _expect_stage(
    check,
    *args,
    failure: str,
    next_steps: list[str],
    **kwargs,
):
    return compose_probe.expect_stage(
        check,
        *args,
        failure=failure,
        next_steps=next_steps,
        failure_guidance=_failure_guidance,
        **kwargs,
    )


def _compose_db_user(env_values: dict[str, str]) -> str:
    return compose_storage.compose_db_user(env_values, _clean_env_value)


def _compose_db_name(env_values: dict[str, str]) -> str:
    return compose_storage.compose_db_name(env_values, _clean_env_value)


def _run_compose_psql(sql: str, env_values: dict[str, str]) -> subprocess.CompletedProcess[str]:
    return compose_storage.run_compose_psql(
        sql,
        env_values,
        compose_command=_compose_command,
        compose_runtime_env=_compose_runtime_env,
        db_user=_compose_db_user,
        db_name=_compose_db_name,
    )


def _compose_query_succeeds(sql: str, env_values: dict[str, str] | None = None) -> bool:
    env_values = env_values or _compose_effective_env()
    return compose_storage.compose_query_succeeds(sql, env_values, _run_compose_psql)


def _compose_check_results(checks: tuple[tuple[str, str], ...], env_values: dict[str, str]) -> list[tuple[str, bool]]:
    return compose_storage.compose_check_results(
        checks,
        env_values,
        run_psql=_run_compose_psql,
        failure_guidance=_failure_guidance,
    )


def _compose_schema_bootstrapped(env_values: dict[str, str] | None = None) -> bool:
    env_values = env_values or _compose_effective_env()
    return all(ok for _label, ok in _compose_check_results(db_tasks.SCHEMA_COMPATIBILITY_CHECKS, env_values))


COMPOSE_LONG_TERM_STORAGE_CHECKS = compose_storage.COMPOSE_LONG_TERM_STORAGE_CHECKS
COMPOSE_STORAGE_MIGRATIONS_BY_CHECK = compose_storage.COMPOSE_STORAGE_MIGRATIONS_BY_CHECK


def _compose_storage_check_results(env_values: dict[str, str]) -> list[tuple[str, bool]]:
    return _compose_check_results(COMPOSE_LONG_TERM_STORAGE_CHECKS, env_values)


def _compose_host_port(env_values: dict[str, str], key: str, default: str) -> int:
    return compose_storage.compose_host_port(env_values, key, default)


def _print_data_plane_connection_guidance(env_values: dict[str, str]):
    compose_storage.print_data_plane_connection_guidance(
        env_values,
        host_port=_compose_host_port,
        db_user=_compose_db_user,
        db_name=_compose_db_name,
    )


def _run_compose_migration_file(migration: Path, env_values: dict[str, str]):
    compose_storage.run_compose_migration_file(
        migration,
        env_values,
        run_compose=_run_compose,
        compose_command=_compose_command,
        compose_runtime_env=_compose_runtime_env,
        db_user=_compose_db_user,
        db_name=_compose_db_name,
    )


def _run_missing_compose_storage_migrations(env_values: dict[str, str]) -> bool:
    return compose_storage.run_missing_compose_storage_migrations(
        env_values,
        storage_check_results=_compose_storage_check_results,
        migration_files=db_tasks._migration_files,
        run_migration_file=_run_compose_migration_file,
    )


def _run_compose_migrations(strict: bool = False):
    compose_storage.run_compose_migrations(
        strict,
        effective_env=_compose_effective_env,
        schema_bootstrapped=_compose_schema_bootstrapped,
        run_missing_storage_migrations=_run_missing_compose_storage_migrations,
        migration_files=db_tasks._migration_files,
        run_migration_file=_run_compose_migration_file,
    )


def _wait_for_postgres_ready(timeout_seconds: int = 90, env_values: dict[str, str] | None = None) -> bool:
    env_values = env_values or _compose_effective_env()
    return compose_storage.wait_for_postgres_ready(
        timeout_seconds,
        env_values,
        run_compose=_run_compose,
        compose_command=_compose_command,
        compose_runtime_env=_compose_runtime_env,
        db_user=_compose_db_user,
        db_name=_compose_db_name,
    )


@task(
    help={
        "wait_timeout": "Seconds to wait for PostgreSQL and NATS readiness (default: 180).",
        "migrate": "Apply canonical migrations after PostgreSQL is ready (default: False).",
    }
)
def infra_up(c, wait_timeout=180, migrate=False):
    """Start only PostgreSQL and NATS for a Compose data-plane test."""
    del c
    _require_compose_env_file()
    env_values = _compose_effective_env()
    _validate_compose_env(env_values)
    wait_timeout = int(wait_timeout)
    if wait_timeout < 30:
        raise SystemExit("Compose infra-up wait timeout must be at least 30 seconds.")

    postgres_port = _compose_host_port(env_values, "MYCELIS_COMPOSE_POSTGRES_PORT", "5432")
    nats_port = _compose_host_port(env_values, "MYCELIS_COMPOSE_NATS_PORT", "4222")

    print("=== Mycelis Compose Data Plane Up ===\n")
    _print_step(
        1,
        3,
        "Starting PostgreSQL and NATS only...",
        "Core and Interface stay down; this exposes the shared data services for a separate app bring-up or connectivity test.",
    )
    _run_compose(
        _compose_command("up", "-d", "postgres", "nats"),
        env=_compose_runtime_env(env_values),
    )

    _expect_stage(
        _wait_for_port,
        postgres_port,
        "PostgreSQL",
        timeout_seconds=wait_timeout,
        failure=f"Compose infra-up failed: PostgreSQL did not become reachable within {wait_timeout}s.",
        next_steps=[
            "Run 'uv run inv compose.status' to inspect container state and host ports.",
            "Run 'uv run inv compose.logs postgres' for PostgreSQL startup logs.",
            "Verify MYCELIS_COMPOSE_POSTGRES_PORT is not already occupied.",
        ],
    )
    _expect_stage(
        _wait_for_postgres_ready,
        timeout_seconds=max(wait_timeout, 90),
        env_values=env_values,
        failure=f"Compose infra-up failed: PostgreSQL did not become query-ready within {max(wait_timeout, 90)}s.",
        next_steps=[
            "Run 'uv run inv compose.logs postgres' to inspect readiness errors.",
            "Use 'uv run inv compose.down --volumes' only when you intentionally want to reset persisted database state.",
        ],
    )
    _expect_stage(
        _wait_for_port,
        nats_port,
        "NATS",
        timeout_seconds=wait_timeout,
        failure=f"Compose infra-up failed: NATS did not become reachable within {wait_timeout}s.",
        next_steps=[
            "Run 'uv run inv compose.status' to confirm the NATS container is up.",
            "Run 'uv run inv compose.logs nats' for broker startup details.",
            "Verify MYCELIS_COMPOSE_NATS_PORT is not already occupied.",
        ],
    )

    print()
    _print_step(
        2,
        3,
        "Data plane reachable.",
        "Use the printed endpoints to configure Core/UI app services that connect to these shared data services.",
    )
    if migrate:
        print("  Applying canonical migrations because --migrate was requested.")
        _run_compose_migrations()
    else:
        print("  Migrations skipped. Run 'uv run inv compose.migrate' when the app schema needs bootstrap before Core starts.")

    print()
    _print_step(
        3,
        3,
        "Connection handoff.",
        "Keep credentials in .env.compose or the consuming deployment's configuration surface; do not bake them into images.",
    )
    _print_data_plane_connection_guidance(env_values)


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
    env_values = _compose_effective_env()
    _validate_compose_env(env_values)
    env_values = _prepare_wsl_ollama_host(env_values)
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
    _run_compose(infra_cmd, env=_compose_runtime_env(env_values))

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
        env_values=env_values,
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
    _run_compose(app_cmd, env=_compose_runtime_env(env_values))

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
    _run_compose(cmd, env=_compose_runtime_env())
    _stop_wsl_ollama_relay()


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
    env_values = _compose_effective_env()
    _validate_compose_env(env_values)
    print("=== Mycelis Compose Status ===\n")
    compose_probe.print_status(
        env_values,
        run_compose=_run_compose,
        compose_command=_compose_command,
        compose_runtime_env=_compose_runtime_env,
        port_open=_port_open,
        api_port=API_PORT,
        interface_port=INTERFACE_PORT,
    )


@task
def infra_health(c):
    """Health probe for the Compose PostgreSQL + NATS data plane only."""
    del c
    _require_compose_env_file()
    env_values = _compose_effective_env()
    _validate_compose_env(env_values)

    print("=== Mycelis Compose Data Plane Health ===\n")
    compose_probe.run_infra_health(
        env_values,
        compose_host_port=_compose_host_port,
        port_open=_port_open,
        wait_for_postgres_ready=_wait_for_postgres_ready,
        http_getter=_http_get,
        compose_db_user=_compose_db_user,
        compose_db_name=_compose_db_name,
        print_data_plane_connection_guidance=_print_data_plane_connection_guidance,
    )


@task
def storage_health(c):
    """Probe Compose PostgreSQL long-term Mycelis storage after migrations."""
    del c
    _require_compose_env_file()
    env_values = _compose_effective_env()
    _validate_compose_env(env_values)

    print("=== Mycelis Compose Long-Term Storage Health ===\n")
    compose_probe.run_storage_health(
        env_values,
        compose_storage_check_results=_compose_storage_check_results,
    )


@task
def health(c):
    """Deep health probe for the Docker Compose runtime path."""
    del c
    _require_compose_env_file()
    env_values = _compose_effective_env()
    _validate_compose_env(env_values)
    env_values = _prepare_wsl_ollama_host(env_values)

    print("=== Mycelis Compose Health ===\n")
    compose_probe.run_health(
        env_values,
        http_getter=_http_get,
        api_host=API_HOST,
        api_port=API_PORT,
        interface_host=INTERFACE_HOST,
        interface_port=INTERFACE_PORT,
    )


@task(help={"service": "Optional compose service name.", "tail": "Number of log lines to show (default: 200)."})
def logs(c, service="", tail=200):
    """Show compose logs for the full stack or a single service."""
    del c
    _require_compose_env_file()
    cmd = _compose_command("logs", "--tail", str(int(tail)))
    if service:
        cmd.append(service)
    subprocess.run(cmd, check=False, env=_compose_runtime_env())


ns = Collection("compose")
ns.add_task(infra_up)
ns.add_task(infra_health)
ns.add_task(storage_health)
ns.add_task(up)
ns.add_task(down)
ns.add_task(migrate)
ns.add_task(status)
ns.add_task(health)
ns.add_task(logs)
