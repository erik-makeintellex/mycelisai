from __future__ import annotations

import subprocess
import time
from pathlib import Path
from typing import Callable


COMPOSE_LONG_TERM_STORAGE_CHECKS = (
    ("pgvector extension", "SELECT 1 FROM pg_extension WHERE extname = 'vector';"),
    (
        "semantic context vectors",
        "SELECT 1 FROM information_schema.tables WHERE table_schema = 'public' AND table_name = 'context_vectors';",
    ),
    (
        "durable agent memory",
        "SELECT 1 FROM information_schema.tables WHERE table_schema = 'public' AND table_name = 'agent_memories';",
    ),
    ("conversation continuity", "SELECT 1 FROM information_schema.tables WHERE table_schema = 'public' AND table_name = 'conversation_turns';"),
    ("retained artifacts", "SELECT 1 FROM information_schema.tables WHERE table_schema = 'public' AND table_name = 'artifacts';"),
    ("temporary continuity", "SELECT 1 FROM information_schema.tables WHERE table_schema = 'public' AND table_name = 'temp_memory_channels';"),
    ("collaboration groups", "SELECT 1 FROM information_schema.tables WHERE table_schema = 'public' AND table_name = 'collaboration_groups';"),
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
    print("    DB_PASSWORD=<from .env.compose; not printed>")
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


def run_missing_compose_storage_migrations(
    env_values: dict[str, str],
    *,
    storage_check_results: Callable[[dict[str, str]], list[tuple[str, bool]]],
    migration_files: Callable[[], list[Path]],
    run_migration_file: Callable[[Path, dict[str, str]], None],
) -> bool:
    missing = [label for label, ok in storage_check_results(env_values) if not ok]
    migration_names: list[str] = []
    for label in missing:
        for migration_name in COMPOSE_STORAGE_MIGRATIONS_BY_CHECK.get(label, ()):
            if migration_name not in migration_names:
                migration_names.append(migration_name)

    if not migration_names:
        return False

    migrations_by_name = {migration.name: migration for migration in migration_files()}
    print("Applying missing long-term storage migrations:")
    for migration_name in migration_names:
        migration = migrations_by_name.get(migration_name)
        if migration is None:
            raise SystemExit(f"Missing migration file required for Compose storage bootstrap: {migration_name}")
        print(f"  - {migration_name}")
        run_migration_file(migration, env_values)
    return True


def run_compose_migrations(
    strict: bool,
    *,
    effective_env: Callable[[], dict[str, str]],
    schema_bootstrapped: Callable[[dict[str, str]], bool],
    run_missing_storage_migrations: Callable[[dict[str, str]], bool],
    migration_files: Callable[[], list[Path]],
    run_migration_file: Callable[[Path, dict[str, str]], None],
):
    env_values = effective_env()
    if not strict and schema_bootstrapped(env_values):
        print(
            "Compose schema already appears compatible with the current runtime; "
            "skipping forward migration replay."
        )
        print(
            "Use 'uv run inv compose.down --volumes' for a truly fresh compose rebuild "
            "when you need to replay the canonical migration stack end-to-end."
        )
        run_missing_storage_migrations(env_values)
        return

    for migration in migration_files():
        run_migration_file(migration, env_values)


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
