from __future__ import annotations

import subprocess
import time
from pathlib import Path
from typing import Callable

from .db_schema import SCHEMA_COMPATIBILITY_CHECKS

COMPOSE_LONG_TERM_STORAGE_CHECKS = SCHEMA_COMPATIBILITY_CHECKS

def compose_db_user(env_values: dict[str, str], clean_env_value: Callable[[str], str]) -> str:
    return clean_env_value(env_values.get("DB_USER") or env_values.get("POSTGRES_USER") or "mycelis")


def compose_db_name(env_values: dict[str, str], clean_env_value: Callable[[str], str]) -> str:
    return clean_env_value(env_values.get("DB_NAME") or env_values.get("POSTGRES_DB") or "cortex")


def run_compose_psql(
    sql: str,
    env_values: dict[str, str],
    *,
    compose_command: Callable[..., list[str]],
    compose_runtime_env: Callable[[dict[str, str]], dict[str, str] | None],
    db_user: Callable[[dict[str, str]], str],
    db_name: Callable[[dict[str, str]], str],
) -> subprocess.CompletedProcess[str]:
    return subprocess.run(
        compose_command(
            "exec",
            "-T",
            "postgres",
            "psql",
            "-t",
            "-A",
            "-h",
            "127.0.0.1",
            "-U",
            db_user(env_values),
            "-d",
            db_name(env_values),
            "-c",
            sql,
        ),
        text=True,
        capture_output=True,
        env=compose_runtime_env(env_values),
    )


def compose_query_succeeds(
    sql: str,
    env_values: dict[str, str],
    run_psql: Callable[[str, dict[str, str]], subprocess.CompletedProcess[str]],
) -> bool:
    result = run_psql(sql, env_values)
    return result.returncode == 0 and "1" in result.stdout.split()


def compose_check_results(
    checks: tuple[tuple[str, str], ...],
    env_values: dict[str, str],
    *,
    run_psql: Callable[[str, dict[str, str]], subprocess.CompletedProcess[str]],
    failure_guidance: Callable[..., str],
) -> list[tuple[str, bool]]:
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
    result = run_psql(query, env_values)
    if result.returncode != 0:
        detail = result.stderr.strip() or result.stdout.strip() or "unknown psql error"
        raise SystemExit(
            failure_guidance(
                f"Compose PostgreSQL check failed: {detail}",
                "Run 'uv run inv compose.infra-health' to confirm the data plane is reachable.",
                "Run 'uv run inv compose.logs postgres' to inspect database service logs.",
            )
        )

    parsed: list[tuple[str, bool]] = []
    for raw_line in result.stdout.splitlines():
        line = raw_line.strip()
        if not line or "\t" not in line:
            continue
        label, state = line.split("\t", 1)
        parsed.append((label, state == "ok"))
    return parsed


def compose_host_port(env_values: dict[str, str], key: str, default: str) -> int:
    try:
        return int(env_values.get(key, default))
    except ValueError as exc:
        raise SystemExit(f"Invalid .env.compose {key}: {env_values.get(key)!r} must be an integer port.") from exc


def compose_host_ports(env_values: dict[str, str], api_default: int, interface_default: int) -> tuple[int, int, int, int]:
    return (
        compose_host_port(env_values, "MYCELIS_COMPOSE_POSTGRES_PORT", "5432"),
        compose_host_port(env_values, "MYCELIS_COMPOSE_NATS_PORT", "4222"),
        compose_host_port(env_values, "MYCELIS_COMPOSE_CORE_PORT", str(api_default)),
        compose_host_port(env_values, "MYCELIS_COMPOSE_INTERFACE_PORT", str(interface_default)),
    )


def print_data_plane_connection_guidance(
    env_values: dict[str, str],
    *,
    host_port: Callable[[dict[str, str], str, str], int],
    db_user: Callable[[dict[str, str]], str],
    db_name: Callable[[dict[str, str]], str],
):
    postgres_port = host_port(env_values, "MYCELIS_COMPOSE_POSTGRES_PORT", "5432")
    nats_port = host_port(env_values, "MYCELIS_COMPOSE_NATS_PORT", "4222")
    nats_monitor_port = host_port(env_values, "MYCELIS_COMPOSE_NATS_MONITOR_PORT", "8222")

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
    print(f"    DB_USER={db_user(env_values)}")
    print("    DB_PASSWORD=<from .env; not printed>")
    print(f"    DB_NAME={db_name(env_values)}")
    print(f"    NATS monitor=http://127.0.0.1:{nats_monitor_port}/varz")


def run_compose_migration_file(
    migration: Path,
    env_values: dict[str, str],
    *,
    run_compose: Callable[..., subprocess.CompletedProcess[str]],
    compose_command: Callable[..., list[str]],
    compose_runtime_env: Callable[[dict[str, str]], dict[str, str] | None],
    db_user: Callable[[dict[str, str]], str],
    db_name: Callable[[dict[str, str]], str],
):
    result = run_compose(
        compose_command(
            "exec",
            "-T",
            "postgres",
            "psql",
            "-v",
            "ON_ERROR_STOP=1",
            "-h",
            "127.0.0.1",
            "-U",
            db_user(env_values),
            "-d",
            db_name(env_values),
            "-f",
            f"/migrations/{migration.name}",
        ),
        check=False,
        env=compose_runtime_env(env_values),
    )
    if result.returncode != 0:
        raise SystemExit(f"Compose migration failed: {migration.name}")


def run_compose_migrations(
    *,
    effective_env: Callable[[], dict[str, str]],
    schema_bootstrapped: Callable[[dict[str, str]], bool],
    schema_nonempty: Callable[[dict[str, str]], bool],
    migration_files: Callable[[], list[Path]],
    run_migration_file: Callable[[Path, dict[str, str]], None],
    canonical_schema_name: str,
):
    env_values = effective_env()
    if schema_bootstrapped(env_values):
        print(
            "Compose schema already appears compatible with the current runtime; "
            "skipping current-schema installation."
        )
        return

    if schema_nonempty(env_values):
        raise SystemExit(
            "Compose PostgreSQL has a nonempty public schema that is incompatible with this Mycelis build. "
            "Back up any retained data, then use 'uv run inv compose.down --volumes' only for disposable local data, "
            "or follow an explicitly supported upgrade path. Automatic replay of historical migrations is disabled."
        )

    migrations = migration_files()
    if len(migrations) != 1 or migrations[0].name != canonical_schema_name:
        raise SystemExit(f"Missing canonical schema installer: {canonical_schema_name}")
    run_migration_file(migrations[0], env_values)
    if not schema_bootstrapped(env_values):
        raise SystemExit(
            f"Compose current-schema installation completed but {canonical_schema_name} "
            "did not satisfy the runtime schema contract."
        )


def wait_for_postgres_ready(
    timeout_seconds: int,
    env_values: dict[str, str],
    *,
    run_compose: Callable[..., subprocess.CompletedProcess[str]],
    compose_command: Callable[..., list[str]],
    compose_runtime_env: Callable[[dict[str, str]], dict[str, str] | None],
    db_user: Callable[[dict[str, str]], str],
    db_name: Callable[[dict[str, str]], str],
) -> bool:
    deadline = time.time() + timeout_seconds
    while time.time() < deadline:
        result = run_compose(
            compose_command(
                "exec",
                "-T",
                "postgres",
                "pg_isready",
                "-h",
                "127.0.0.1",
                "-U",
                db_user(env_values),
                "-d",
                db_name(env_values),
            ),
            check=False,
            env=compose_runtime_env(env_values),
        )
        if result.returncode == 0:
            return True
        time.sleep(2)
    return False
