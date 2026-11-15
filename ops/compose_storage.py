from __future__ import annotations

import subprocess
import time
from pathlib import Path
from typing import Callable


COMPOSE_LONG_TERM_STORAGE_CHECKS = (
    ("pgvector extension", "SELECT 1 FROM pg_extension WHERE extname = 'vector';"),
    ("semantic context vectors", "SELECT 1 FROM information_schema.tables WHERE table_schema = 'public' AND table_name = 'context_vectors';"),
    ("durable agent memory", "SELECT 1 FROM information_schema.tables WHERE table_schema = 'public' AND table_name = 'agent_memories';"),
    ("conversation continuity", "SELECT 1 FROM information_schema.tables WHERE table_schema = 'public' AND table_name = 'conversation_turns';"),
    ("retained artifacts", "SELECT 1 FROM information_schema.tables WHERE table_schema = 'public' AND table_name = 'artifacts';"),
    ("temporary continuity", "SELECT 1 FROM information_schema.tables WHERE table_schema = 'public' AND table_name = 'temp_memory_channels';"),
    ("collaboration groups", "SELECT 1 FROM information_schema.tables WHERE table_schema = 'public' AND table_name = 'collaboration_groups';"),
    ("managed exchange channels", "SELECT 1 FROM information_schema.tables WHERE table_schema = 'public' AND table_name = 'exchange_channels';"),
    ("managed exchange items", "SELECT 1 FROM information_schema.tables WHERE table_schema = 'public' AND table_name = 'exchange_items';"),
    ("conversation templates", "SELECT 1 FROM information_schema.tables WHERE table_schema = 'public' AND table_name = 'conversation_templates';"),
    ("durable organizations", "SELECT 1 FROM information_schema.tables WHERE table_schema = 'public' AND table_name = 'organizations';"),
    ("operator templates", "SELECT 1 FROM information_schema.tables WHERE table_schema = 'public' AND table_name = 'operator_templates';"),
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
    "durable organizations": ("040_organizations.up.sql",),
    "operator templates": ("041_operator_templates.up.sql",),
}


def compose_db_user(env_values: dict[str, str], clean_env_value: Callable[[str], str]) -> str:
    return clean_env_value(env_values.get("DB_USER") or env_values.get("POSTGRES_USER") or "mycelis")


def compose_db_name(env_values: dict[str, str], clean_env_value: Callable[[str], str]) -> str:
    return clean_env_value(env_values.get("DB_NAME") or env_values.get("POSTGRES_DB") or "cortex")


def run_compose_psql(
    sql: str, env_values: dict[str, str], *, compose_command: Callable[..., list[str]],
    compose_runtime_env: Callable[[dict[str, str]], dict[str, str] | None],
    db_user: Callable[[dict[str, str]], str], db_name: Callable[[dict[str, str]], str],
) -> subprocess.CompletedProcess[str]:
    cmd = compose_command(
        "exec", "-T", "postgres", "psql", "-t", "-A",
        "-h", "127.0.0.1", "-U", db_user(env_values), "-d", db_name(env_values), "-c", sql,
    )
    return subprocess.run(cmd, text=True, capture_output=True, env=compose_runtime_env(env_values))


def compose_query_succeeds(
    sql: str, env_values: dict[str, str], run_psql: Callable[[str, dict[str, str]], subprocess.CompletedProcess[str]],
) -> bool:
    result = run_psql(sql, env_values)
    return result.returncode == 0 and "1" in result.stdout.split()


def compose_check_results(
    checks: tuple[tuple[str, str], ...], env_values: dict[str, str], *,
    run_psql: Callable[[str, dict[str, str]], subprocess.CompletedProcess[str]], failure_guidance: Callable[..., str],
) -> list[tuple[str, bool]]:
    values = [
        f"({i}, '{label.replace(chr(39), chr(39) * 2)}', EXISTS({sql.strip().rstrip(';')}))"
        for i, (label, sql) in enumerate(checks, start=1)
    ]
    query = (
        "WITH checks(ord, label, ok) AS (VALUES " + ", ".join(values)
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
        if line and "\t" in line:
            label, state = line.split("\t", 1)
            parsed.append((label, state == "ok"))
    return parsed


def compose_host_port(env_values: dict[str, str], key: str, default: str) -> int:
    try:
        return int(env_values.get(key, default))
    except ValueError as exc:
        raise SystemExit(f"Invalid .env.compose {key}: {env_values.get(key)!r} must be an integer port.") from exc


def print_data_plane_connection_guidance(
    env_values: dict[str, str], *, host_port: Callable[[dict[str, str], str, str], int],
    db_user: Callable[[dict[str, str]], str], db_name: Callable[[dict[str, str]], str],
):
    postgres_port = host_port(env_values, "MYCELIS_COMPOSE_POSTGRES_PORT", "5432")
    nats_port = host_port(env_values, "MYCELIS_COMPOSE_NATS_PORT", "4222")
    nats_monitor_port = host_port(env_values, "MYCELIS_COMPOSE_NATS_MONITOR_PORT", "8222")
    print("\nData service connection settings:")
    print("  Same compose project app containers:\n    DB_HOST=postgres\n    DB_PORT=5432\n    NATS_URL=nats://nats:4222")
    print(f"  Host-native clients:\n    DB_HOST=127.0.0.1\n    DB_PORT={postgres_port}\n    NATS_URL=nats://127.0.0.1:{nats_port}")
    print(f"  Separate Docker Compose app project:\n    DB_HOST=host.docker.internal\n    DB_PORT={postgres_port}\n    NATS_URL=nats://host.docker.internal:{nats_port}")
    print(f"  Credentials:\n    DB_USER={db_user(env_values)}\n    DB_PASSWORD=<from .env; not printed>\n    DB_NAME={db_name(env_values)}")
    print(f"    NATS monitor=http://127.0.0.1:{nats_monitor_port}/varz")


def run_compose_migration_file(
    migration: Path, env_values: dict[str, str], *, run_compose: Callable[..., subprocess.CompletedProcess[str]],
    compose_command: Callable[..., list[str]], compose_runtime_env: Callable[[dict[str, str]], dict[str, str] | None],
    db_user: Callable[[dict[str, str]], str], db_name: Callable[[dict[str, str]], str],
):
    cmd = compose_command(
        "exec", "-T", "postgres", "psql", "-v", "ON_ERROR_STOP=1",
        "-h", "127.0.0.1", "-U", db_user(env_values), "-d", db_name(env_values), "-f", f"/migrations/{migration.name}",
    )
    result = run_compose(cmd, check=False, env=compose_runtime_env(env_values))
    if result.returncode != 0:
        raise SystemExit(f"Compose migration failed: {migration.name}")


def run_missing_compose_storage_migrations(
    env_values: dict[str, str], *, storage_check_results: Callable[[dict[str, str]], list[tuple[str, bool]]],
    migration_files: Callable[[], list[Path]], run_migration_file: Callable[[Path, dict[str, str]], None],
) -> bool:
    missing = [label for label, ok in storage_check_results(env_values) if not ok]
    migration_names: list[str] = []
    for label in missing:
        for name in COMPOSE_STORAGE_MIGRATIONS_BY_CHECK.get(label, ()):
            if name not in migration_names:
                migration_names.append(name)
    if not migration_names:
        return False
    migrations_by_name = {m.name: m for m in migration_files()}
    print("Applying missing long-term storage migrations:")
    for name in migration_names:
        migration = migrations_by_name.get(name)
        if migration is None:
            raise SystemExit(f"Missing migration file required for Compose storage bootstrap: {name}")
        print(f"  - {name}")
        run_migration_file(migration, env_values)
    return True


def run_compose_migrations(
    strict: bool, *, effective_env: Callable[[], dict[str, str]], schema_bootstrapped: Callable[[dict[str, str]], bool],
    run_missing_storage_migrations: Callable[[dict[str, str]], bool], migration_files: Callable[[], list[Path]],
    run_migration_file: Callable[[Path, dict[str, str]], None],
):
    env_values = effective_env()
    if not strict and schema_bootstrapped(env_values):
        print("Compose schema already appears compatible with the current runtime; skipping forward migration replay.")
        print("Use 'uv run inv compose.down --volumes' for a truly fresh compose rebuild when you need to replay the canonical migration stack end-to-end.")
        run_missing_storage_migrations(env_values)
        return
    for migration in migration_files():
        run_migration_file(migration, env_values)


def wait_for_postgres_ready(
    timeout_seconds: int, env_values: dict[str, str], *, run_compose: Callable[..., subprocess.CompletedProcess[str]],
    compose_command: Callable[..., list[str]], compose_runtime_env: Callable[[dict[str, str]], dict[str, str] | None],
    db_user: Callable[[dict[str, str]], str], db_name: Callable[[dict[str, str]], str],
) -> bool:
    deadline = time.time() + timeout_seconds
    while time.time() < deadline:
        cmd = compose_command(
            "exec", "-T", "postgres", "pg_isready", "-h", "127.0.0.1", "-U", db_user(env_values), "-d", db_name(env_values),
        )
        if run_compose(cmd, check=False, env=compose_runtime_env(env_values)).returncode == 0:
            return True
        time.sleep(2)
    return False


def _compose_db_user(env_values: dict[str, str]) -> str:
    from . import compose as _c
    return compose_db_user(env_values, _c._clean_env_value)


def _compose_db_name(env_values: dict[str, str]) -> str:
    from . import compose as _c
    return compose_db_name(env_values, _c._clean_env_value)


def _run_compose_psql(sql: str, env_values: dict[str, str]) -> subprocess.CompletedProcess[str]:
    from . import compose as _c
    return run_compose_psql(
        sql, env_values, compose_command=_c._compose_command,
        compose_runtime_env=_c._compose_runtime_env, db_user=_c._compose_db_user, db_name=_c._compose_db_name,
    )


def _compose_query_succeeds(sql: str, env_values: dict[str, str] | None = None) -> bool:
    from . import compose as _c
    return compose_query_succeeds(sql, env_values or _c._compose_effective_env(), _c._run_compose_psql)


def _compose_check_results(checks: tuple[tuple[str, str], ...], env_values: dict[str, str]) -> list[tuple[str, bool]]:
    from . import compose as _c
    return compose_check_results(checks, env_values, run_psql=_c._run_compose_psql, failure_guidance=_c._failure_guidance)


def _compose_schema_bootstrapped(env_values: dict[str, str] | None = None) -> bool:
    from . import compose as _c
    env_values = env_values or _c._compose_effective_env()
    return all(ok for _label, ok in _c._compose_check_results(_c.db_tasks.SCHEMA_COMPATIBILITY_CHECKS, env_values))


def _compose_storage_check_results(env_values: dict[str, str]) -> list[tuple[str, bool]]:
    from . import compose as _c
    return _c._compose_check_results(_c.COMPOSE_LONG_TERM_STORAGE_CHECKS, env_values)


def _compose_host_port(env_values: dict[str, str], key: str, default: str) -> int:
    return compose_host_port(env_values, key, default)


def _print_data_plane_connection_guidance(env_values: dict[str, str]):
    from . import compose as _c
    print_data_plane_connection_guidance(
        env_values, host_port=_c._compose_host_port, db_user=_c._compose_db_user, db_name=_c._compose_db_name,
    )


def _run_compose_migration_file(migration: Path, env_values: dict[str, str]):
    from . import compose as _c
    run_compose_migration_file(
        migration, env_values, run_compose=_c._run_compose, compose_command=_c._compose_command,
        compose_runtime_env=_c._compose_runtime_env, db_user=_c._compose_db_user, db_name=_c._compose_db_name,
    )


def _run_missing_compose_storage_migrations(env_values: dict[str, str]) -> bool:
    from . import compose as _c
    return run_missing_compose_storage_migrations(
        env_values, storage_check_results=_c._compose_storage_check_results,
        migration_files=_c.db_tasks._migration_files, run_migration_file=_c._run_compose_migration_file,
    )


def _run_compose_migrations(strict: bool = False):
    from . import compose as _c
    run_compose_migrations(
        strict, effective_env=_c._compose_effective_env, schema_bootstrapped=_c._compose_schema_bootstrapped,
        run_missing_storage_migrations=_c._run_missing_compose_storage_migrations,
        migration_files=_c.db_tasks._migration_files, run_migration_file=_c._run_compose_migration_file,
    )


def _wait_for_postgres_ready(timeout_seconds: int = 90, env_values: dict[str, str] | None = None) -> bool:
    from . import compose as _c
    return wait_for_postgres_ready(
        timeout_seconds, env_values or _c._compose_effective_env(),
        run_compose=_c._run_compose, compose_command=_c._compose_command,
        compose_runtime_env=_c._compose_runtime_env, db_user=_c._compose_db_user, db_name=_c._compose_db_name,
    )
