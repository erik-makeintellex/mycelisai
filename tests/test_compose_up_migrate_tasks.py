from __future__ import annotations

from pathlib import Path

import pytest

from ops import compose


def test_run_compose_migrations_executes_canonical_files(monkeypatch):
    commands: list[list[str]] = []
    files = [Path("001_init_memory.sql"), Path("002_extra.up.sql")]

    monkeypatch.setattr(compose.db_tasks, "_migration_files", lambda: files)
    monkeypatch.setattr(compose, "_load_compose_env", lambda: {"DB_USER": "mycelis", "DB_NAME": "cortex"})
    monkeypatch.setattr(compose, "_compose_schema_bootstrapped", lambda env_values=None: False)
    monkeypatch.setattr(
        compose,
        "_run_compose",
        lambda args, check=True, env=None: commands.append(args) or type("Result", (), {"returncode": 0})(),
    )

    compose._run_compose_migrations()

    assert commands == [
        compose._compose_command(
            "exec", "-T", "postgres", "psql", "-v", "ON_ERROR_STOP=1",
            "-h", "127.0.0.1", "-U", "mycelis", "-d", "cortex", "-f", "/migrations/001_init_memory.sql",
        ),
        compose._compose_command(
            "exec", "-T", "postgres", "psql", "-v", "ON_ERROR_STOP=1",
            "-h", "127.0.0.1", "-U", "mycelis", "-d", "cortex", "-f", "/migrations/002_extra.up.sql",
        ),
    ]


def test_run_compose_migrations_uses_configured_db_login(monkeypatch):
    commands: list[list[str]] = []
    files = [Path("001_init_memory.sql")]

    monkeypatch.setattr(compose.db_tasks, "_migration_files", lambda: files)
    monkeypatch.setattr(compose, "_load_compose_env", lambda: {"DB_USER": "owner", "DB_NAME": "ownerdb"})
    monkeypatch.setattr(compose, "_compose_schema_bootstrapped", lambda env_values=None: False)
    monkeypatch.setattr(
        compose,
        "_run_compose",
        lambda args, check=True, env=None: commands.append(args) or type("Result", (), {"returncode": 0})(),
    )

    compose._run_compose_migrations()

    assert "-U" in commands[0]
    assert "owner" in commands[0]
    assert "-d" in commands[0]
    assert "ownerdb" in commands[0]


def test_run_compose_migrations_skips_replay_when_schema_is_compatible(monkeypatch, capsys):
    monkeypatch.setattr(compose, "_load_compose_env", lambda: {"DB_USER": "mycelis", "DB_NAME": "cortex"})
    monkeypatch.setattr(compose, "_compose_schema_bootstrapped", lambda env_values=None: True)
    monkeypatch.setattr(compose, "_run_missing_compose_storage_migrations", lambda env_values: False)
    monkeypatch.setattr(
        compose.db_tasks,
        "_migration_files",
        lambda: pytest.fail("migration files should not be replayed for a compatible schema"),
    )

    compose._run_compose_migrations()

    out = capsys.readouterr().out
    assert "skipping forward migration replay" in out
    assert "compose.down --volumes" in out


def test_run_compose_migrations_applies_missing_late_storage_when_base_schema_is_compatible(monkeypatch, capsys):
    files = [Path("038_conversation_templates.up.sql")]
    commands: list[list[str]] = []

    monkeypatch.setattr(compose, "_load_compose_env", lambda: {"DB_USER": "mycelis", "DB_NAME": "cortex"})
    monkeypatch.setattr(compose, "_compose_schema_bootstrapped", lambda env_values=None: True)
    monkeypatch.setattr(
        compose,
        "_compose_storage_check_results",
        lambda env_values: [
            ("pgvector extension", True),
            ("conversation templates", False),
        ],
    )
    monkeypatch.setattr(compose.db_tasks, "_migration_files", lambda: files)
    monkeypatch.setattr(
        compose,
        "_run_compose",
        lambda args, check=True, env=None: commands.append(args) or type("Result", (), {"returncode": 0})(),
    )

    compose._run_compose_migrations()

    out = capsys.readouterr().out
    assert "Applying missing long-term storage migrations" in out
    assert commands == [
        compose._compose_command(
            "exec", "-T", "postgres", "psql", "-v", "ON_ERROR_STOP=1",
            "-h", "127.0.0.1", "-U", "mycelis", "-d", "cortex", "-f", "/migrations/038_conversation_templates.up.sql",
        )
    ]


def test_compose_schema_bootstrapped_requires_all_runtime_objects(monkeypatch):
    monkeypatch.setattr(compose.db_tasks, "SCHEMA_COMPATIBILITY_CHECKS", [("a", "sql"), ("b", "sql"), ("c", "sql")])
    monkeypatch.setattr(compose, "_compose_check_results", lambda checks, env_values: [("a", True), ("b", True), ("c", False)])

    assert compose._compose_schema_bootstrapped() is False


def test_compose_schema_bootstrapped_accepts_current_runtime_schema(monkeypatch):
    monkeypatch.setattr(compose.db_tasks, "SCHEMA_COMPATIBILITY_CHECKS", [("a", "sql"), ("b", "sql")])
    monkeypatch.setattr(compose, "_compose_check_results", lambda checks, env_values: [("a", True), ("b", True)])

    assert compose._compose_schema_bootstrapped() is True


def test_compose_up_orders_infra_then_migrations_then_app(monkeypatch):
    commands: list[list[str]] = []
    waits: list[tuple[str, int | str]] = []

    monkeypatch.setattr(compose, "_require_compose_env_file", lambda: None)
    monkeypatch.setattr(
        compose,
        "_run_compose",
        lambda args, check=True, env=None: commands.append(args) or type("Result", (), {"returncode": 0})(),
    )
    monkeypatch.setattr(compose, "_run_compose_migrations", lambda: waits.append(("migrate", 0)))
    monkeypatch.setattr(compose, "_wait_for_port", lambda port, label, timeout_seconds=60: waits.append((label, port)) or True)
    monkeypatch.setattr(
        compose,
        "_wait_for_postgres_ready",
        lambda timeout_seconds=90, env_values=None: waits.append(("PostgreSQL ready", timeout_seconds)) or True,
    )
    monkeypatch.setattr(
        compose,
        "_wait_for_http_ok",
        lambda url, label, timeout_seconds=60, headers=None: waits.append((label, url)) or True,
    )
    monkeypatch.setattr(compose, "_prepare_wsl_ollama_host", lambda env_values: env_values)
    monkeypatch.setattr(compose.status, "body", lambda _c=None: waits.append(("status", 0)))

    compose.up.body(None, build=False, wait_timeout=180)

    assert commands == [
        compose._compose_command("up", "-d", "postgres", "nats"),
        compose._compose_command("up", "-d", "core", "interface"),
    ]
    assert waits == [
        ("PostgreSQL", 5432),
        ("PostgreSQL ready", 180),
        ("NATS", 4222),
        ("migrate", 0),
        ("Core API", compose.API_PORT),
        ("Core health", f"http://{compose.API_HOST}:{compose.API_PORT}/healthz"),
        ("Frontend", compose.INTERFACE_PORT),
        ("Frontend", f"http://{compose.INTERFACE_HOST}:{compose.INTERFACE_PORT}/"),
        ("status", 0),
    ]


def test_compose_infra_up_starts_only_data_services(monkeypatch):
    commands: list[list[str]] = []
    waits: list[tuple[str, int]] = []

    monkeypatch.setattr(compose, "_require_compose_env_file", lambda: None)
    monkeypatch.setattr(
        compose,
        "_load_compose_env",
        lambda: {
            "MYCELIS_COMPOSE_POSTGRES_PORT": "15432",
            "MYCELIS_COMPOSE_NATS_PORT": "14222",
            "MYCELIS_COMPOSE_OLLAMA_HOST": "http://host.docker.internal:11434",
        },
    )
    monkeypatch.setattr(compose, "_validate_compose_env", lambda env_values: None)
    monkeypatch.setattr(
        compose,
        "_run_compose",
        lambda args, check=True, env=None: commands.append(args) or type("Result", (), {"returncode": 0})(),
    )
    monkeypatch.setattr(compose, "_wait_for_port", lambda port, label, timeout_seconds=60: waits.append((label, port)) or True)
    monkeypatch.setattr(
        compose,
        "_wait_for_postgres_ready",
        lambda timeout_seconds=90, env_values=None: waits.append(("PostgreSQL ready", timeout_seconds)) or True,
    )
    monkeypatch.setattr(compose, "_run_compose_migrations", lambda: pytest.fail("migrations should be opt-in for infra-up"))

    compose.infra_up.body(None, wait_timeout=120)

    assert commands == [compose._compose_command("up", "-d", "postgres", "nats")]
    assert waits == [
        ("PostgreSQL", 15432),
        ("PostgreSQL ready", 120),
        ("NATS", 14222),
    ]


def test_compose_infra_up_can_run_migrations_when_requested(monkeypatch):
    commands: list[list[str]] = []
    migrated: list[bool] = []

    monkeypatch.setattr(compose, "_require_compose_env_file", lambda: None)
    monkeypatch.setattr(compose, "_load_compose_env", lambda: {})
    monkeypatch.setattr(compose, "_validate_compose_env", lambda env_values: None)
    monkeypatch.setattr(
        compose,
        "_run_compose",
        lambda args, check=True, env=None: commands.append(args) or type("Result", (), {"returncode": 0})(),
    )
    monkeypatch.setattr(compose, "_wait_for_port", lambda port, label, timeout_seconds=60: True)
    monkeypatch.setattr(compose, "_wait_for_postgres_ready", lambda timeout_seconds=90, env_values=None: True)
    monkeypatch.setattr(compose, "_run_compose_migrations", lambda: migrated.append(True))

    compose.infra_up.body(None, wait_timeout=120, migrate=True)

    assert commands == [compose._compose_command("up", "-d", "postgres", "nats")]
    assert migrated == [True]
