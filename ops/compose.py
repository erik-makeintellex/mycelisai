import subprocess
from pathlib import Path

from invoke import Collection, task

from . import compose_env, compose_helpers, compose_probe, compose_storage, compose_wsl_relay
from . import db as db_tasks
from .compose_helpers import (
    _clean_env_value,
    _compose_command,
    _compose_effective_env,
    _compose_runtime_env,
    _docker_run,
    _ensure_wsl_ollama_relay,
    _expect_stage,
    _failure_guidance,
    _http_get,
    _inspect_wsl_ollama_relay_labels,
    _load_compose_env,
    _looks_like_container_loopback,
    _normalize_bool,
    _parse_network_endpoint,
    _port_open,
    _prepare_wsl_ollama_host,
    _print_step,
    _require_compose_env_file,
    _resolve_host_path,
    _run_compose,
    _stop_wsl_ollama_relay,
    _validate_compose_env,
    _validate_output_block_config,
    _wait_for_http_ok,
    _wait_for_port,
    _wsl_exec_command,
    _wsl_http_available,
    _wsl_ollama_relay_port,
)
from .compose_storage import (
    COMPOSE_LONG_TERM_STORAGE_CHECKS,
    COMPOSE_STORAGE_MIGRATIONS_BY_CHECK,
    _compose_check_results,
    _compose_db_name,
    _compose_db_user,
    _compose_host_port,
    _compose_query_succeeds,
    _compose_schema_bootstrapped,
    _compose_storage_check_results,
    _print_data_plane_connection_guidance,
    _run_compose_migration_file,
    _run_compose_migrations,
    _run_compose_psql,
    _run_missing_compose_storage_migrations,
    _wait_for_postgres_ready,
)
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
        1, 3, "Starting PostgreSQL and NATS only...",
        "Core and Interface stay down; this exposes the shared data services for a separate app bring-up or connectivity test.",
    )
    _run_compose(
        _compose_command("up", "-d", "postgres", "nats"),
        env=_compose_runtime_env(env_values),
    )

    _expect_stage(
        _wait_for_port, postgres_port, "PostgreSQL", timeout_seconds=wait_timeout,
        failure=f"Compose infra-up failed: PostgreSQL did not become reachable within {wait_timeout}s.",
        next_steps=[
            "Run 'uv run inv compose.status' to inspect container state and host ports.",
            "Run 'uv run inv compose.logs postgres' for PostgreSQL startup logs.",
            "Verify MYCELIS_COMPOSE_POSTGRES_PORT is not already occupied.",
        ],
    )
    _expect_stage(
        _wait_for_postgres_ready, timeout_seconds=max(wait_timeout, 90), env_values=env_values,
        failure=f"Compose infra-up failed: PostgreSQL did not become query-ready within {max(wait_timeout, 90)}s.",
        next_steps=[
            "Run 'uv run inv compose.logs postgres' to inspect readiness errors.",
            "Use 'uv run inv compose.down --volumes' only when you intentionally want to reset persisted database state.",
        ],
    )
    _expect_stage(
        _wait_for_port, nats_port, "NATS", timeout_seconds=wait_timeout,
        failure=f"Compose infra-up failed: NATS did not become reachable within {wait_timeout}s.",
        next_steps=[
            "Run 'uv run inv compose.status' to confirm the NATS container is up.",
            "Run 'uv run inv compose.logs nats' for broker startup details.",
            "Verify MYCELIS_COMPOSE_NATS_PORT is not already occupied.",
        ],
    )

    print()
    _print_step(
        2, 3, "Data plane reachable.",
        "Use the printed endpoints to configure Core/UI app services that connect to these shared data services.",
    )
    if migrate:
        print("  Applying canonical migrations because --migrate was requested.")
        _run_compose_migrations()
    else:
        print("  Migrations skipped. Run 'uv run inv compose.migrate' when the app schema needs bootstrap before Core starts.")

    print()
    _print_step(
        3, 3, "Connection handoff.",
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
    compose_helpers.run_up(build=build, wait_timeout=wait_timeout)


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
        env_values, run_compose=_run_compose, compose_command=_compose_command,
        compose_runtime_env=_compose_runtime_env, port_open=_port_open,
        api_port=API_PORT, interface_port=INTERFACE_PORT,
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
        env_values, compose_host_port=_compose_host_port, port_open=_port_open,
        wait_for_postgres_ready=_wait_for_postgres_ready, http_getter=_http_get,
        compose_db_user=_compose_db_user, compose_db_name=_compose_db_name,
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
        env_values, compose_storage_check_results=_compose_storage_check_results,
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
        env_values, http_getter=_http_get, api_host=API_HOST,
        api_port=API_PORT, interface_host=INTERFACE_HOST, interface_port=INTERFACE_PORT,
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
