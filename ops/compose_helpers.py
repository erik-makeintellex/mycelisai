from __future__ import annotations

import subprocess
from pathlib import Path

from . import compose_env, compose_probe, compose_wsl_relay


def _compose_command(*args: str) -> list[str]:
    from . import compose as _c
    return compose_env.compose_command(
        _c.ROOT_DIR, _c.COMPOSE_PROJECT, _c.COMPOSE_ENV_FILE, _c.COMPOSE_FILE,
        _c.docker_command, _c.docker_host_path, *args,
    )


def _require_compose_env_file():
    from . import compose as _c
    return compose_env.require_compose_env_file(_c.COMPOSE_ENV_FILE, _c.COMPOSE_ENV_EXAMPLE)


def _load_compose_env() -> dict[str, str]:
    from . import compose as _c
    return compose_env.load_compose_env(_c.COMPOSE_ENV_FILE, _c._require_compose_env_file)


def _compose_effective_env(env_values: dict[str, str] | None = None) -> dict[str, str]:
    from . import compose as _c
    return compose_env.compose_effective_env(env_values, _c._load_compose_env)


def _wsl_exec_command(*args: str) -> list[str]:
    from . import compose as _c
    return list(args) if _c.running_in_wsl() else compose_env.wsl_exec_command(*args)


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
    from . import compose as _c
    return compose_env.wsl_http_available(url, _c._wsl_exec_command)


def _wsl_ollama_relay_port(env_values: dict[str, str]) -> int:
    from . import compose as _c
    return compose_env.wsl_ollama_relay_port(env_values, _c.DEFAULT_WSL_OLLAMA_RELAY_PORT)


def _docker_run(args: list[str], check: bool = True) -> subprocess.CompletedProcess[str]:
    from . import compose as _c
    return compose_wsl_relay.docker_run(
        args, docker_command=_c.docker_command, root_dir=_c.ROOT_DIR, check=check,
    )


def _inspect_wsl_ollama_relay_labels() -> dict[str, str] | None:
    from . import compose as _c
    return compose_wsl_relay.inspect_labels(
        relay_name=_c.WSL_OLLAMA_RELAY_NAME, docker_runner=_c._docker_run,
    )


def _stop_wsl_ollama_relay():
    from . import compose as _c
    compose_wsl_relay.stop(
        relay_name=_c.WSL_OLLAMA_RELAY_NAME, docker_runner=_c._docker_run,
        docker_host_mode=_c.docker_host_mode, running_in_wsl=_c.running_in_wsl,
    )


def _ensure_wsl_ollama_relay(target_host: str, target_port: int, relay_port: int):
    from . import compose as _c
    compose_wsl_relay.ensure(
        target_host, target_port, relay_port,
        relay_name=_c.WSL_OLLAMA_RELAY_NAME, relay_image=_c.WSL_OLLAMA_RELAY_IMAGE,
        inspect_relay_labels=_c._inspect_wsl_ollama_relay_labels,
        stop_relay=_c._stop_wsl_ollama_relay, docker_runner=_c._docker_run,
    )


def _prepare_wsl_ollama_host(env_values: dict[str, str]) -> dict[str, str]:
    from . import compose as _c
    return compose_wsl_relay.prepare_host(
        env_values, docker_host_mode=_c.docker_host_mode, running_in_wsl=_c.running_in_wsl,
        clean_env_value=_c._clean_env_value, parse_network_endpoint=_c._parse_network_endpoint,
        relay_port=_c._wsl_ollama_relay_port, inspect_relay_labels=_c._inspect_wsl_ollama_relay_labels,
        wsl_http_available=_c._wsl_http_available, ensure_relay=_c._ensure_wsl_ollama_relay,
    )


def _validate_output_block_config(env_values: dict[str, str]):
    from . import compose as _c
    return compose_env.validate_output_block_config(
        env_values, _c.DEFAULT_OUTPUT_HOST_PATH, output_block_modes=_c.OUTPUT_BLOCK_MODES,
    )


def _validate_compose_env(env_values: dict[str, str]):
    from . import compose as _c
    return compose_env.validate_compose_env(env_values, _c._validate_output_block_config)


def _compose_runtime_env(env_values: dict[str, str] | None = None) -> dict[str, str] | None:
    from . import compose as _c
    return compose_env.compose_runtime_env(
        env_values, _c.docker_host_mode, _c.running_in_wsl,
        _c._compose_effective_env, _c.DEFAULT_OUTPUT_HOST_PATH, _c.docker_host_path,
    )


def _print_step(step: int, total: int, title: str, expectation: str | None = None):
    return compose_probe.print_step(step, total, title, expectation)


def _failure_guidance(message: str, *next_steps: str) -> str:
    return compose_probe.failure_guidance(message, *next_steps)


def _wait_for_port(port: int, label: str, timeout_seconds: int = 60) -> bool:
    from . import compose as _c
    return compose_probe.wait_for_port(
        port, label, port_open=_c._port_open, timeout_seconds=timeout_seconds,
    )


def _http_get(url: str, timeout: float = 3.0, headers: dict[str, str] | None = None) -> tuple[int, str]:
    return compose_probe.http_get(url, timeout=timeout, headers=headers)


def _wait_for_http_ok(
    url: str, label: str, timeout_seconds: int = 60, headers: dict[str, str] | None = None
) -> bool:
    from . import compose as _c
    return compose_probe.wait_for_http_ok(
        url, label, http_getter=_c._http_get, timeout_seconds=timeout_seconds, headers=headers,
    )


def _run_compose(
    args: list[str], check: bool = True, env: dict[str, str] | None = None
) -> subprocess.CompletedProcess[str]:
    return compose_probe.run_compose(args, check=check, env=env)


def _expect_stage(check, *args, failure: str, next_steps: list[str], **kwargs):
    from . import compose as _c
    return compose_probe.expect_stage(
        check, *args, failure=failure, next_steps=next_steps,
        failure_guidance=_c._failure_guidance, **kwargs,
    )


def run_up(build: bool = False, wait_timeout: int = 180):
    from . import compose as _c

    _c._require_compose_env_file()
    env_values = _c._compose_effective_env()
    _c._validate_compose_env(env_values)
    env_values = _c._prepare_wsl_ollama_host(env_values)
    wait_timeout = int(wait_timeout)
    if wait_timeout < 30:
        raise SystemExit("Compose up wait timeout must be at least 30 seconds.")

    print("=== Mycelis Compose Up ===\n")
    _c._print_step(
        1, 4, "Starting PostgreSQL and NATS...",
        "Docker should start the infra containers first. With --build, image preparation can take several minutes before readiness checks begin.",
    )
    infra_cmd = _c._compose_command("up", "-d")
    if build:
        infra_cmd.append("--build")
    infra_cmd.extend(["postgres", "nats"])
    _c._run_compose(infra_cmd, env=_c._compose_runtime_env(env_values))

    _c._expect_stage(
        _c._wait_for_port, 5432, "PostgreSQL", timeout_seconds=wait_timeout,
        failure=f"Compose up failed: PostgreSQL did not become reachable within {wait_timeout}s.",
        next_steps=[
            "Run 'uv run inv compose.status' to inspect container state and host ports.",
            "Run 'uv run inv compose.logs postgres' for PostgreSQL startup logs.",
            "Use 'uv run inv compose.down --volumes' if you need a fully fresh rebuild.",
        ],
    )
    _c._expect_stage(
        _c._wait_for_postgres_ready, timeout_seconds=max(wait_timeout, 90), env_values=env_values,
        failure=f"Compose up failed: PostgreSQL did not become query-ready within {max(wait_timeout, 90)}s.",
        next_steps=[
            "Run 'uv run inv compose.logs postgres' to inspect readiness or migration errors.",
            "Verify the local PostgreSQL port is not already occupied by another service.",
        ],
    )
    _c._expect_stage(
        _c._wait_for_port, 4222, "NATS", timeout_seconds=wait_timeout,
        failure=f"Compose up failed: NATS did not become reachable within {wait_timeout}s.",
        next_steps=[
            "Run 'uv run inv compose.status' to confirm the NATS container is up.",
            "Run 'uv run inv compose.logs nats' for broker startup details.",
        ],
    )

    print()
    _c._print_step(
        2, 4, "Applying canonical migrations through the PostgreSQL container...",
        "Schema bootstrap may skip replay when the compose database is already compatible with the current runtime.",
    )
    _c._run_compose_migrations()

    print()
    _c._print_step(
        3, 4, "Starting Core and Interface...",
        "Core should answer /healthz and the frontend root should load. With --build, app image creation can take several additional minutes before the HTTP checks pass.",
    )
    app_cmd = _c._compose_command("up", "-d")
    if build:
        app_cmd.append("--build")
    app_cmd.extend(["core", "interface"])
    _c._run_compose(app_cmd, env=_c._compose_runtime_env(env_values))

    _c._expect_stage(
        _c._wait_for_port, _c.API_PORT, "Core API", timeout_seconds=wait_timeout,
        failure=f"Compose up failed: Core API port did not open within {wait_timeout}s.",
        next_steps=[
            "Run 'uv run inv compose.logs core' to inspect backend startup output.",
            "Run 'uv run inv compose.health' after fixing backend readiness to verify the runtime contract.",
        ],
    )
    _c._expect_stage(
        _c._wait_for_http_ok, f"http://{_c.API_HOST}:{_c.API_PORT}/healthz", "Core health", timeout_seconds=wait_timeout,
        failure=f"Compose up failed: Core /healthz did not become ready within {wait_timeout}s.",
        next_steps=[
            "Run 'uv run inv compose.logs core' to inspect healthz startup failures.",
            "Check provider and database configuration in '.env.compose' if Core stays unhealthy.",
        ],
    )
    _c._expect_stage(
        _c._wait_for_port, _c.INTERFACE_PORT, "Frontend", timeout_seconds=wait_timeout,
        failure=f"Compose up failed: Frontend port did not open within {wait_timeout}s.",
        next_steps=[
            "Run 'uv run inv compose.logs interface' to inspect frontend startup output.",
            "Verify the host port is not already in use by another local UI server.",
        ],
    )
    _c._expect_stage(
        _c._wait_for_http_ok, f"http://{_c.INTERFACE_HOST}:{_c.INTERFACE_PORT}/", "Frontend", timeout_seconds=wait_timeout,
        failure=f"Compose up failed: Frontend root did not become ready within {wait_timeout}s.",
        next_steps=[
            "Run 'uv run inv compose.logs interface' for frontend build or startup failures.",
            "Run 'uv run inv compose.status' to confirm the UI container is still running.",
        ],
    )

    print()
    _c._print_step(
        4, 4, "Compose stack ready.",
        "Next expected checks: 'uv run inv compose.health' for deep runtime proof, then open http://localhost:3000.",
    )
    _c.status.body(None)
