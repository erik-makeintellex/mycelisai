import socket
import subprocess
import time
import json
import os
import shlex
from pathlib import Path
from typing import Callable
from urllib.parse import urlparse

from invoke import task, Collection

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
)


COMPOSE_FILE = ROOT_DIR / "docker-compose.yml"
COMPOSE_ENV_FILE = ROOT_DIR / ".env.compose"
COMPOSE_ENV_EXAMPLE = ROOT_DIR / ".env.compose.example"
COMPOSE_PROJECT = "mycelis-home"
DEFAULT_OUTPUT_HOST_PATH = ROOT_DIR / "workspace" / "docker-compose" / "data"
OUTPUT_BLOCK_MODES = {"local_hosted", "cluster_generated"}
WSL_OLLAMA_RELAY_NAME = f"{COMPOSE_PROJECT}-ollama-relay"
WSL_OLLAMA_RELAY_IMAGE = "alpine:3.21"
DEFAULT_WSL_OLLAMA_RELAY_PORT = 11435
COMPOSE_RUNTIME_OVERRIDE_KEYS = {
    "CORS_ORIGIN",
    "DATA_DIR",
    "DB_HOST",
    "DB_NAME",
    "DB_PASSWORD",
    "DB_PORT",
    "DB_USER",
    "MYCELIS_API_KEY",
    "MYCELIS_BOOTSTRAP_TEMPLATE_ID",
    "MYCELIS_COMPOSE_CORE_PORT",
    "MYCELIS_COMPOSE_INTERFACE_PORT",
    "MYCELIS_COMPOSE_NATS_MONITOR_PORT",
    "MYCELIS_COMPOSE_NATS_PORT",
    "MYCELIS_COMPOSE_OLLAMA_HOST",
    "MYCELIS_COMPOSE_POSTGRES_PORT",
    "MYCELIS_COMPOSE_WSL_OLLAMA_RELAY_PORT",
    "MYCELIS_DISABLE_DEFAULT_MCP_BOOTSTRAP",
    "MYCELIS_OUTPUT_BLOCK_MODE",
    "MYCELIS_OUTPUT_HOST_PATH",
    "MYCELIS_WORKSPACE",
    "NATS_URL",
    "POSTGRES_DB",
    "POSTGRES_PASSWORD",
    "POSTGRES_USER",
}


def _compose_command(*args: str) -> list[str]:
    return docker_command(
        "compose",
        "--project-name",
        COMPOSE_PROJECT,
        "--env-file",
        docker_host_path(COMPOSE_ENV_FILE),
        "-f",
        docker_host_path(COMPOSE_FILE),
        *args,
        cwd=ROOT_DIR,
    )


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


def _compose_effective_env(env_values: dict[str, str] | None = None) -> dict[str, str]:
    values = dict(env_values or _load_compose_env())
    for key in COMPOSE_RUNTIME_OVERRIDE_KEYS:
        override = os.environ.get(key)
        if override is None:
            continue
        cleaned = _clean_env_value(override)
        if cleaned:
            values[key] = cleaned
    return values


def _wsl_exec_command(*args: str) -> list[str]:
    command = ["wsl.exe"]
    distro = os.environ.get("MYCELIS_WSL_DISTRO", "").strip()
    if distro:
        command.extend(["-d", distro])
    command.extend(["--exec", *args])
    return command


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


def _clean_env_value(value: str) -> str:
    return value.strip().strip('"').strip("'")


def _resolve_host_path(path_value: str) -> Path:
    raw_path = _clean_env_value(path_value)
    expanded = os.path.expandvars(os.path.expanduser(raw_path))
    return Path(expanded).resolve(strict=False)


def _parse_network_endpoint(url: str) -> tuple[str, int]:
    parsed = urlparse(url.strip())
    if parsed.scheme not in {"http", "https"} or not parsed.hostname:
        raise SystemExit(
            "Invalid .env.compose MYCELIS_COMPOSE_OLLAMA_HOST: "
            f"{url}. Use an http(s) endpoint such as http://host.docker.internal:11434."
        )
    port = parsed.port or (443 if parsed.scheme == "https" else 80)
    return parsed.hostname, port


def _wsl_http_available(url: str) -> bool:
    probe = url.rstrip("/") + "/api/tags"
    result = subprocess.run(
        _wsl_exec_command(
            "sh",
            "-lc",
            f"curl -fsS --max-time 5 -o /dev/null {shlex.quote(probe)}",
        ),
        capture_output=True,
        text=True,
        timeout=15,
    )
    return result.returncode == 0


def _wsl_ollama_relay_port(env_values: dict[str, str]) -> int:
    raw = _clean_env_value(
        env_values.get("MYCELIS_COMPOSE_WSL_OLLAMA_RELAY_PORT", str(DEFAULT_WSL_OLLAMA_RELAY_PORT))
    )
    try:
        return int(raw)
    except ValueError as exc:
        raise SystemExit(
            "Invalid .env.compose MYCELIS_COMPOSE_WSL_OLLAMA_RELAY_PORT: "
            f"{raw!r} must be an integer port."
        ) from exc


def _docker_run(args: list[str], check: bool = True) -> subprocess.CompletedProcess[str]:
    result = subprocess.run(
        docker_command(*args, cwd=ROOT_DIR),
        capture_output=True,
        text=True,
    )
    if check and result.returncode != 0:
        detail = result.stderr.strip() or result.stdout.strip() or "unknown docker error"
        raise SystemExit(f"Docker command failed: {' '.join(args)} ({detail})")
    return result


def _inspect_wsl_ollama_relay_labels() -> dict[str, str] | None:
    result = _docker_run(
        ["inspect", WSL_OLLAMA_RELAY_NAME, "--format", "{{json .Config.Labels}}"],
        check=False,
    )
    if result.returncode != 0:
        return None
    raw = result.stdout.strip()
    if not raw:
        return {}
    try:
        return json.loads(raw)
    except json.JSONDecodeError:
        return None


def _stop_wsl_ollama_relay():
    if docker_host_mode() != "wsl":
        return
    _docker_run(["rm", "-f", WSL_OLLAMA_RELAY_NAME], check=False)


def _ensure_wsl_ollama_relay(target_host: str, target_port: int, relay_port: int):
    labels = _inspect_wsl_ollama_relay_labels()
    expected = {
        "mycelis.relay.target_host": target_host,
        "mycelis.relay.target_port": str(target_port),
        "mycelis.relay.listen_port": str(relay_port),
    }
    if labels and all(labels.get(key) == value for key, value in expected.items()):
        return

    _stop_wsl_ollama_relay()
    relay_cmd = (
        "apk add --no-cache socat >/dev/null && "
        f"exec socat TCP-LISTEN:{relay_port},fork,reuseaddr,bind=0.0.0.0 "
        f"TCP:{target_host}:{target_port}"
    )
    result = _docker_run(
        [
            "run",
            "-d",
            "--rm",
            "--name",
            WSL_OLLAMA_RELAY_NAME,
            "--network",
            "host",
            "--label",
            f"mycelis.relay.target_host={target_host}",
            "--label",
            f"mycelis.relay.target_port={target_port}",
            "--label",
            f"mycelis.relay.listen_port={relay_port}",
            WSL_OLLAMA_RELAY_IMAGE,
            "sh",
            "-lc",
            relay_cmd,
        ]
    )
    if not result.stdout.strip():
        raise SystemExit("Failed to start the WSL Ollama relay container.")
    time.sleep(2)


def _prepare_wsl_ollama_host(env_values: dict[str, str]) -> dict[str, str]:
    values = dict(env_values)
    if docker_host_mode() != "wsl":
        return values

    configured = _clean_env_value(values.get("MYCELIS_COMPOSE_OLLAMA_HOST", "http://host.docker.internal:11434"))
    configured_host, configured_port = _parse_network_endpoint(configured)
    relay_target_host = configured_host
    relay_target_port = configured_port

    if not _wsl_http_available(configured):
        localhost_candidate = f"http://127.0.0.1:{configured_port}"
        if _wsl_http_available(localhost_candidate):
            relay_target_host = "127.0.0.1"
        else:
            raise SystemExit(
                "WSL Docker could not reach the configured MYCELIS_COMPOSE_OLLAMA_HOST and "
                "no mirrored localhost Ollama fallback was reachable from WSL. "
                "Verify the Windows Ollama service is running and reachable from the Docker-owning WSL distro."
            )

    relay_port = _wsl_ollama_relay_port(values)
    _ensure_wsl_ollama_relay(relay_target_host, relay_target_port, relay_port)
    print(
        "  WSL Ollama relay: "
        f"http://host.docker.internal:{relay_port} -> {relay_target_host}:{relay_target_port}"
    )
    values["MYCELIS_COMPOSE_OLLAMA_HOST"] = f"http://host.docker.internal:{relay_port}"
    return values


def _validate_output_block_config(env_values: dict[str, str]):
    mode_explicit = "MYCELIS_OUTPUT_BLOCK_MODE" in env_values
    mode = _clean_env_value(env_values.get("MYCELIS_OUTPUT_BLOCK_MODE", "local_hosted")).lower()
    if mode not in OUTPUT_BLOCK_MODES:
        raise SystemExit(
            "Invalid .env.compose MYCELIS_OUTPUT_BLOCK_MODE: "
            f"{mode}. Use one of: {', '.join(sorted(OUTPUT_BLOCK_MODES))}."
        )

    path_explicit = "MYCELIS_OUTPUT_HOST_PATH" in env_values
    raw_path = _clean_env_value(env_values.get("MYCELIS_OUTPUT_HOST_PATH", ""))
    if not raw_path:
        if mode == "local_hosted" and mode_explicit:
            raise SystemExit(
                "MYCELIS_OUTPUT_HOST_PATH is required when MYCELIS_OUTPUT_BLOCK_MODE=local_hosted. "
                "Set it to the host directory Docker should mount as Core /data."
            )
        raw_path = str(DEFAULT_OUTPUT_HOST_PATH)

    host_path = _resolve_host_path(raw_path)
    if host_path.exists() and not host_path.is_dir():
        raise SystemExit(
            "Invalid .env.compose MYCELIS_OUTPUT_HOST_PATH: "
            f"{host_path} exists but is not a directory."
        )
    if not host_path.exists():
        if mode == "local_hosted" and (mode_explicit or path_explicit):
            raise SystemExit(
                "Invalid .env.compose MYCELIS_OUTPUT_HOST_PATH: "
                f"{host_path} does not exist. Create the directory first so Docker mounts the intended output block."
            )
        host_path.mkdir(parents=True, exist_ok=True)


def _validate_compose_env(env_values: dict[str, str]):
    ollama_host = env_values.get("MYCELIS_COMPOSE_OLLAMA_HOST", "http://host.docker.internal:11434").strip()
    if ollama_host and _looks_like_container_loopback(ollama_host):
        raise SystemExit(
            "Invalid .env.compose MYCELIS_COMPOSE_OLLAMA_HOST for Docker Compose: "
            f"{ollama_host}. Use a host-reachable address such as "
            "http://host.docker.internal:11434 or another container/service hostname."
        )
    _validate_output_block_config(env_values)


def _compose_runtime_env(env_values: dict[str, str] | None = None) -> dict[str, str] | None:
    if docker_host_mode() != "wsl":
        return None

    values = _compose_effective_env(env_values)
    raw_host_path = _clean_env_value(values.get("MYCELIS_OUTPUT_HOST_PATH", ""))
    resolved_host_path = _resolve_host_path(raw_host_path) if raw_host_path else DEFAULT_OUTPUT_HOST_PATH

    env = os.environ.copy()
    passthrough_keys = sorted(values.keys())
    for key in passthrough_keys:
        env[key] = values[key]
    env["MYCELIS_OUTPUT_HOST_PATH"] = docker_host_path(resolved_host_path)
    passthrough_entries = [entry for entry in env.get("WSLENV", "").split(":") if entry]
    if "MYCELIS_OUTPUT_HOST_PATH" not in passthrough_keys:
        passthrough_keys.append("MYCELIS_OUTPUT_HOST_PATH")
    for key in passthrough_keys:
        if key not in passthrough_entries:
            passthrough_entries.append(key)
    env["WSLENV"] = ":".join(passthrough_entries)
    return env


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


def _run_compose(
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


def _compose_db_user(env_values: dict[str, str]) -> str:
    return _clean_env_value(env_values.get("DB_USER") or env_values.get("POSTGRES_USER") or "mycelis")


def _compose_db_name(env_values: dict[str, str]) -> str:
    return _clean_env_value(env_values.get("DB_NAME") or env_values.get("POSTGRES_DB") or "cortex")


def _run_compose_psql(sql: str, env_values: dict[str, str]) -> subprocess.CompletedProcess[str]:
    return subprocess.run(
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
            _compose_db_user(env_values),
            "-d",
            _compose_db_name(env_values),
            "-c",
            sql,
        ),
        text=True,
        capture_output=True,
        env=_compose_runtime_env(env_values),
    )


def _compose_query_succeeds(sql: str, env_values: dict[str, str] | None = None) -> bool:
    env_values = env_values or _compose_effective_env()
    result = _run_compose_psql(sql, env_values)
    return result.returncode == 0 and "1" in result.stdout.split()


def _compose_check_results(checks: tuple[tuple[str, str], ...], env_values: dict[str, str]) -> list[tuple[str, bool]]:
    values = []
    for index, (label, sql) in enumerate(checks, start=1):
        exists_sql = sql.strip().rstrip(";")
        escaped_label = label.replace("'", "''")
        values.append(f"({index}, '{escaped_label}', EXISTS({exists_sql}))")
    query = (
        "WITH checks(ord, label, ok) AS (VALUES "
        + ", ".join(values)
        + ") SELECT label || E'\\t' || CASE WHEN ok THEN 'ok' ELSE 'missing' END FROM checks ORDER BY ord;"
    )
    result = _run_compose_psql(query, env_values)
    if result.returncode != 0:
        detail = result.stderr.strip() or result.stdout.strip() or "unknown psql error"
        raise SystemExit(_failure_guidance(
            f"Compose PostgreSQL check failed: {detail}",
            "Run 'uv run inv compose.infra-health' to confirm the data plane is reachable.",
            "Run 'uv run inv compose.logs postgres' to inspect database service logs.",
        ))

    parsed: list[tuple[str, bool]] = []
    for raw_line in result.stdout.splitlines():
        line = raw_line.strip()
        if not line or "\t" not in line:
            continue
        label, state = line.split("\t", 1)
        parsed.append((label, state == "ok"))
    return parsed


def _compose_schema_bootstrapped(env_values: dict[str, str] | None = None) -> bool:
    env_values = env_values or _compose_effective_env()
    return all(ok for _label, ok in _compose_check_results(db_tasks.SCHEMA_COMPATIBILITY_CHECKS, env_values))


COMPOSE_LONG_TERM_STORAGE_CHECKS = (
    (
        "pgvector extension",
        "SELECT 1 FROM pg_extension WHERE extname = 'vector';",
    ),
    (
        "semantic context vectors",
        "SELECT 1 FROM information_schema.tables WHERE table_schema = 'public' AND table_name = 'context_vectors';",
    ),
    (
        "durable agent memory",
        "SELECT 1 FROM information_schema.tables WHERE table_schema = 'public' AND table_name = 'agent_memories';",
    ),
    (
        "conversation continuity",
        "SELECT 1 FROM information_schema.tables WHERE table_schema = 'public' AND table_name = 'conversation_turns';",
    ),
    (
        "retained artifacts",
        "SELECT 1 FROM information_schema.tables WHERE table_schema = 'public' AND table_name = 'artifacts';",
    ),
    (
        "temporary continuity",
        "SELECT 1 FROM information_schema.tables WHERE table_schema = 'public' AND table_name = 'temp_memory_channels';",
    ),
    (
        "collaboration groups",
        "SELECT 1 FROM information_schema.tables WHERE table_schema = 'public' AND table_name = 'collaboration_groups';",
    ),
    (
        "managed exchange channels",
        "SELECT 1 FROM information_schema.tables WHERE table_schema = 'public' AND table_name = 'exchange_channels';",
    ),
    (
        "managed exchange items",
        "SELECT 1 FROM information_schema.tables WHERE table_schema = 'public' AND table_name = 'exchange_items';",
    ),
    (
        "conversation templates",
        "SELECT 1 FROM information_schema.tables WHERE table_schema = 'public' AND table_name = 'conversation_templates';",
    ),
)


COMPOSE_STORAGE_MIGRATIONS_BY_CHECK = {
    "semantic context vectors": ("008_context_engine.up.sql",),
    "durable agent memory": ("019_agent_memories.up.sql", "037_scoped_memory_visibility.up.sql"),
    "conversation continuity": ("030_conversation_turns.up.sql",),
    "retained artifacts": ("018_artifacts.up.sql",),
    "temporary continuity": ("033_temp_memory_channels.up.sql",),
    "collaboration groups": ("034_collaboration_groups.up.sql",),
    "managed exchange channels": ("035_managed_exchange.up.sql", "036_managed_exchange_security.up.sql"),
    "managed exchange items": ("035_managed_exchange.up.sql", "036_managed_exchange_security.up.sql"),
    "conversation templates": ("038_conversation_templates.up.sql",),
}


def _compose_storage_check_results(env_values: dict[str, str]) -> list[tuple[str, bool]]:
    return _compose_check_results(COMPOSE_LONG_TERM_STORAGE_CHECKS, env_values)


def _compose_host_port(env_values: dict[str, str], key: str, default: str) -> int:
    try:
        return int(env_values.get(key, default))
    except ValueError as exc:
        raise SystemExit(f"Invalid .env.compose {key}: {env_values.get(key)!r} must be an integer port.") from exc


def _print_data_plane_connection_guidance(env_values: dict[str, str]):
    postgres_port = _compose_host_port(env_values, "MYCELIS_COMPOSE_POSTGRES_PORT", "5432")
    nats_port = _compose_host_port(env_values, "MYCELIS_COMPOSE_NATS_PORT", "4222")
    nats_monitor_port = _compose_host_port(env_values, "MYCELIS_COMPOSE_NATS_MONITOR_PORT", "8222")
    db_user = _compose_db_user(env_values)
    db_name = _compose_db_name(env_values)

    print("\nData service connection settings:")
    print("  Same compose project app containers:")
    print("    DB_HOST=postgres")
    print("    DB_PORT=5432")
    print("    NATS_URL=nats://nats:4222")
    print("  Host-native clients:")
    print("    DB_HOST=127.0.0.1")
    print(f"    DB_PORT={postgres_port}")
    print(f"    NATS_URL=nats://127.0.0.1:{nats_port}")
    print("  Separate Docker Compose app project:")
    print("    DB_HOST=host.docker.internal")
    print(f"    DB_PORT={postgres_port}")
    print(f"    NATS_URL=nats://host.docker.internal:{nats_port}")
    print("  Credentials:")
    print(f"    DB_USER={db_user}")
    print("    DB_PASSWORD=<from .env.compose; not printed>")
    print(f"    DB_NAME={db_name}")
    print(f"    NATS monitor=http://127.0.0.1:{nats_monitor_port}/varz")


def _run_compose_migration_file(migration: Path, env_values: dict[str, str]):
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
            _compose_db_user(env_values),
            "-d",
            _compose_db_name(env_values),
            "-f",
            f"/migrations/{migration.name}",
        ),
        check=False,
        env=_compose_runtime_env(env_values),
    )
    if result.returncode != 0:
        raise SystemExit(f"Compose migration failed: {migration.name}")


def _run_missing_compose_storage_migrations(env_values: dict[str, str]) -> bool:
    missing = [label for label, ok in _compose_storage_check_results(env_values) if not ok]
    migration_names: list[str] = []
    for label in missing:
        for migration_name in COMPOSE_STORAGE_MIGRATIONS_BY_CHECK.get(label, ()):
            if migration_name not in migration_names:
                migration_names.append(migration_name)

    if not migration_names:
        return False

    migrations_by_name = {migration.name: migration for migration in db_tasks._migration_files()}
    print("Applying missing long-term storage migrations:")
    for migration_name in migration_names:
        migration = migrations_by_name.get(migration_name)
        if migration is None:
            raise SystemExit(f"Missing migration file required for Compose storage bootstrap: {migration_name}")
        print(f"  - {migration_name}")
        _run_compose_migration_file(migration, env_values)
    return True


def _run_compose_migrations(strict: bool = False):
    env_values = _compose_effective_env()
    if not strict and _compose_schema_bootstrapped(env_values):
        print(
            "Compose schema already appears compatible with the current runtime; "
            "skipping forward migration replay."
        )
        print(
            "Use 'uv run inv compose.down --volumes' for a truly fresh compose rebuild "
            "when you need to replay the canonical migration stack end-to-end."
        )
        _run_missing_compose_storage_migrations(env_values)
        return

    for migration in db_tasks._migration_files():
        _run_compose_migration_file(migration, env_values)


def _wait_for_postgres_ready(timeout_seconds: int = 90, env_values: dict[str, str] | None = None) -> bool:
    env_values = env_values or _compose_effective_env()
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
                _compose_db_user(env_values),
                "-d",
                _compose_db_name(env_values),
            ),
            check=False,
            env=_compose_runtime_env(env_values),
        )
        if result.returncode == 0:
            return True
        time.sleep(2)
    return False


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
    _run_compose(_compose_command("ps"), check=False, env=_compose_runtime_env(env_values))

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
def infra_health(c):
    """Health probe for the Compose PostgreSQL + NATS data plane only."""
    del c
    _require_compose_env_file()
    env_values = _compose_effective_env()
    _validate_compose_env(env_values)
    postgres_port = _compose_host_port(env_values, "MYCELIS_COMPOSE_POSTGRES_PORT", "5432")
    nats_port = _compose_host_port(env_values, "MYCELIS_COMPOSE_NATS_PORT", "4222")
    nats_monitor_port = _compose_host_port(env_values, "MYCELIS_COMPOSE_NATS_MONITOR_PORT", "8222")

    print("=== Mycelis Compose Data Plane Health ===\n")
    failures: list[str] = []

    if _port_open(postgres_port):
        print(f"  [OK] {'PostgreSQL port':<18} 127.0.0.1:{postgres_port}")
    else:
        print(f"  [FAIL] {'PostgreSQL port':<18} 127.0.0.1:{postgres_port}")
        failures.append("PostgreSQL host port is not reachable.")

    if _wait_for_postgres_ready(timeout_seconds=5, env_values=env_values):
        print(f"  [OK] {'PostgreSQL query':<18} {_compose_db_user(env_values)}@postgres/{_compose_db_name(env_values)}")
    else:
        print(f"  [FAIL] {'PostgreSQL query':<18} {_compose_db_user(env_values)}@postgres/{_compose_db_name(env_values)}")
        failures.append("PostgreSQL container did not accept a query with the configured DB user/name.")

    if _port_open(nats_port):
        print(f"  [OK] {'NATS port':<18} nats://127.0.0.1:{nats_port}")
    else:
        print(f"  [FAIL] {'NATS port':<18} nats://127.0.0.1:{nats_port}")
        failures.append("NATS host port is not reachable.")

    nats_monitor_url = f"http://127.0.0.1:{nats_monitor_port}/varz"
    status_code, body = _http_get(nats_monitor_url, timeout=5.0)
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

    _print_data_plane_connection_guidance(env_values)
    print("\nData plane healthy. Core/Interface may remain down for infra-only testing.")


@task
def storage_health(c):
    """Probe Compose PostgreSQL long-term Mycelis storage after migrations."""
    del c
    _require_compose_env_file()
    env_values = _compose_effective_env()
    _validate_compose_env(env_values)

    print("=== Mycelis Compose Long-Term Storage Health ===\n")
    failures: list[str] = []
    for label, ok in _compose_storage_check_results(env_values):
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


@task
def health(c):
    """Deep health probe for the Docker Compose runtime path."""
    del c
    _require_compose_env_file()
    env_values = _compose_effective_env()
    _validate_compose_env(env_values)
    env_values = _prepare_wsl_ollama_host(env_values)
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
