from __future__ import annotations

import builtins
from types import SimpleNamespace

import pytest

from ops import db as db_tasks
from ops import lifecycle


def _missing_dotenv_import(name, globals=None, locals=None, fromlist=(), level=0):
    err = ModuleNotFoundError("No module named 'dotenv'")
    err.name = "dotenv"
    raise err


def test_db_load_env_reports_recovery_guidance_when_dotenv_missing(monkeypatch):
    monkeypatch.setattr(builtins, "__import__", _missing_dotenv_import)

    with pytest.raises(SystemExit, match="uv run inv"):
        db_tasks._load_env()


def test_lifecycle_load_env_reports_recovery_guidance_when_dotenv_missing(monkeypatch):
    monkeypatch.setattr(builtins, "__import__", _missing_dotenv_import)

    with pytest.raises(SystemExit, match="uv run inv"):
        lifecycle._load_env()


def test_require_postgres_reports_bridge_guidance(monkeypatch):
    monkeypatch.setattr(db_tasks, "_dsn", lambda dbname=None: ("127.0.0.1", "5432", "mycelis", "password", dbname or "postgres"))
    monkeypatch.setattr(db_tasks, "_psql", lambda **kwargs: 1)

    with pytest.raises(SystemExit, match="uv run inv k8s.bridge"):
        db_tasks._require_postgres()


def test_create_skips_create_when_database_already_exists(monkeypatch):
    calls: list[tuple[str, str | None]] = []

    monkeypatch.setattr(db_tasks, "_load_env", lambda: None)
    monkeypatch.setattr(db_tasks, "_require_postgres", lambda dbname="postgres": calls.append(("require", dbname)))
    monkeypatch.setattr(
        db_tasks,
        "_run_psql",
        lambda sql=None, file=None, dbname=None: calls.append(("run", sql)) or SimpleNamespace(returncode=0, stdout="1\n", stderr=""),
    )
    monkeypatch.setattr(db_tasks, "_emit_psql_output", lambda result: None)
    monkeypatch.setattr(db_tasks, "_psql", lambda **kwargs: calls.append(("psql", kwargs.get("sql"))) or 0)
    monkeypatch.setenv("DB_NAME", "cortex")

    db_tasks.create.body(None)

    assert ("require", "postgres") in calls
    assert not any(kind == "psql" and sql == "CREATE DATABASE cortex;" for kind, sql in calls)


def test_run_psql_enables_on_error_stop(monkeypatch):
    captured = {}

    def fake_run(cmd, env=None, capture_output=None, text=None):
        captured["cmd"] = cmd
        return SimpleNamespace(returncode=0, stdout="", stderr="")

    monkeypatch.setattr(db_tasks, "_dsn", lambda dbname=None: ("127.0.0.1", "5432", "mycelis", "password", dbname or "cortex"))
    monkeypatch.setattr(db_tasks.subprocess, "run", fake_run)

    db_tasks._run_psql(sql="SELECT 1;")

    assert captured["cmd"][:3] == ["psql", "-v", "ON_ERROR_STOP=1"]


def test_run_psql_leaves_transaction_ownership_to_current_schema(monkeypatch, tmp_path):
    captured = {}
    schema = tmp_path / "001_current_schema.sql"

    def fake_run(cmd, env=None, capture_output=None, text=None):
        captured["cmd"] = cmd
        return SimpleNamespace(returncode=0, stdout="", stderr="")

    monkeypatch.setattr(db_tasks, "_dsn", lambda dbname=None: ("127.0.0.1", "5432", "mycelis", "password", dbname or "cortex"))
    monkeypatch.setattr(db_tasks.subprocess, "run", fake_run)

    db_tasks._run_psql(file=schema)

    assert captured["cmd"][-2:] == ["-f", str(schema)]
    assert "--single-transaction" not in captured["cmd"]


def test_migration_files_only_include_current_schema():
    assert db_tasks._migration_files() == [
        db_tasks.MIGRATIONS_DIR / "001_current_schema.sql"
    ]


def test_reset_fails_fast_when_current_schema_install_errors(monkeypatch):
    calls: list[str] = []

    monkeypatch.setattr(db_tasks, "_load_env", lambda: None)
    monkeypatch.setattr(db_tasks, "_require_postgres", lambda dbname="postgres": None)
    monkeypatch.setenv("DB_NAME", "cortex")
    monkeypatch.setattr(
        db_tasks,
        "_migration_files",
        lambda: [db_tasks.MIGRATIONS_DIR / "001_current_schema.sql"],
    )
    monkeypatch.setattr(db_tasks, "schema_bootstrapped", lambda: False)
    monkeypatch.setattr(db_tasks, "schema_nonempty", lambda: False)

    def fake_run_psql(sql=None, file=None, dbname=None):
        if sql is not None:
            calls.append(sql)
            return 0
        calls.append(file.name)
        return 1 if file else 0

    monkeypatch.setattr(db_tasks, "_psql", fake_run_psql)
    monkeypatch.setattr(db_tasks, "_ensure_database_exists", lambda: None)
    with pytest.raises(SystemExit, match="Current-schema installation failed: 001_current_schema.sql"):
        db_tasks.reset.body(None)

    assert "001_current_schema.sql" in calls


def test_migrate_skips_replay_when_schema_is_already_bootstrapped(monkeypatch, capsys):
    migrate_calls: list[str] = []

    monkeypatch.setattr(db_tasks, "_load_env", lambda: None)
    monkeypatch.setattr(db_tasks, "_ensure_database_exists", lambda: None)
    monkeypatch.setattr(db_tasks, "schema_bootstrapped", lambda: True)
    monkeypatch.setattr(
        db_tasks,
        "_migration_files",
        lambda: [db_tasks.MIGRATIONS_DIR / "001_current_schema.sql"],
    )
    monkeypatch.setattr(
        db_tasks,
        "_psql",
        lambda sql=None, file=None, dbname=None: migrate_calls.append(file.name if file else sql or "") or 0,
    )
    monkeypatch.setenv("DB_NAME", "cortex")

    db_tasks.migrate.body(None)

    captured = capsys.readouterr()
    assert "already appears compatible with the current runtime" in captured.out
    assert "skipping current-schema installation" in captured.out
    assert migrate_calls == []


def test_migrate_fails_closed_for_nonempty_incompatible_schema(monkeypatch):
    monkeypatch.setattr(db_tasks, "_load_env", lambda: None)
    monkeypatch.setattr(db_tasks, "_ensure_database_exists", lambda: None)
    monkeypatch.setattr(db_tasks, "schema_bootstrapped", lambda: False)
    monkeypatch.setattr(db_tasks, "schema_nonempty", lambda: True)
    monkeypatch.setattr(
        db_tasks,
        "_migration_files",
        lambda: pytest.fail("baseline must not run against nonempty incompatible schema"),
    )
    monkeypatch.setenv("DB_NAME", "cortex")

    with pytest.raises(SystemExit, match="nonempty.*db.reset.*upgrade"):
        db_tasks.migrate.body(None)


def test_migrate_installs_current_schema_once_when_empty(monkeypatch):
    applied: list[str] = []
    compatibility = iter([False, True])
    monkeypatch.setattr(db_tasks, "_load_env", lambda: None)
    monkeypatch.setattr(db_tasks, "_ensure_database_exists", lambda: None)
    monkeypatch.setattr(db_tasks, "schema_bootstrapped", lambda: next(compatibility))
    monkeypatch.setattr(db_tasks, "schema_nonempty", lambda: False)
    monkeypatch.setattr(
        db_tasks,
        "_migration_files",
        lambda: [db_tasks.MIGRATIONS_DIR / "001_current_schema.sql"],
    )
    monkeypatch.setattr(
        db_tasks,
        "_psql",
        lambda sql=None, file=None, dbname=None: applied.append(file.name) or 0,
    )

    db_tasks.migrate.body(None)

    assert applied == ["001_current_schema.sql"]


def test_migrate_rejects_incomplete_post_install_schema(monkeypatch):
    compatibility = iter([False, False])
    monkeypatch.setattr(db_tasks, "_load_env", lambda: None)
    monkeypatch.setattr(db_tasks, "_ensure_database_exists", lambda: None)
    monkeypatch.setattr(db_tasks, "schema_bootstrapped", lambda: next(compatibility))
    monkeypatch.setattr(db_tasks, "schema_nonempty", lambda: False)
    monkeypatch.setattr(
        db_tasks,
        "_migration_files",
        lambda: [db_tasks.MIGRATIONS_DIR / "001_current_schema.sql"],
    )
    monkeypatch.setattr(db_tasks, "_psql", lambda **kwargs: 0)

    with pytest.raises(SystemExit, match="did not satisfy the runtime schema contract"):
        db_tasks.migrate.body(None)


def test_reset_terminates_and_recreates_database_before_migrations(monkeypatch):
    calls: list[tuple[str, str | None]] = []

    monkeypatch.setattr(db_tasks, "_load_env", lambda: None)
    monkeypatch.setattr(db_tasks, "_require_postgres", lambda dbname="postgres": calls.append(("require", dbname)))
    monkeypatch.setenv("DB_NAME", "cortex")
    monkeypatch.setattr(
        db_tasks,
        "_psql",
        lambda sql=None, file=None, dbname=None: calls.append((dbname or "", sql or (file.name if file else ""))) or 0,
    )
    monkeypatch.setattr(db_tasks, "_ensure_database_exists", lambda: None)
    monkeypatch.setattr(db_tasks, "_apply_migrations", lambda: calls.append(("apply", "current")))

    db_tasks.reset.body(None)

    assert calls == [
        ("require", "postgres"),
        ("postgres", "SELECT pg_terminate_backend(pid) FROM pg_stat_activity WHERE datname = 'cortex' AND pid <> pg_backend_pid();"),
        ("postgres", "DROP DATABASE IF EXISTS cortex;"),
        ("postgres", "CREATE DATABASE cortex OWNER mycelis;"),
        ("apply", "current"),
    ]


def test_schema_bootstrapped_requires_all_current_runtime_objects(monkeypatch):
    responses = iter(
        [
            SimpleNamespace(returncode=0, stdout="1\n", stderr=""),
            SimpleNamespace(returncode=0, stdout="1\n", stderr=""),
            SimpleNamespace(returncode=0, stdout="1\n", stderr=""),
            SimpleNamespace(returncode=0, stdout="1\n", stderr=""),
            SimpleNamespace(returncode=0, stdout="", stderr=""),
        ]
    )

    monkeypatch.setattr(db_tasks, "_load_env", lambda: None)
    monkeypatch.setattr(db_tasks, "_run_psql", lambda sql=None, file=None, dbname=None: next(responses))

    assert db_tasks.schema_bootstrapped() is False


def test_schema_bootstrapped_requires_group_workspace_folder_column():
    checks = {label: sql for label, sql in db_tasks.SCHEMA_COMPATIBILITY_CHECKS}

    assert "collaboration_groups workspace_folder column" in checks
    assert "collaboration_groups" in checks["collaboration_groups workspace_folder column"]
    assert "workspace_folder" in checks["collaboration_groups workspace_folder column"]


def test_schema_bootstrapped_requires_outcome_ownership_tables():
    checks = {label: sql for label, sql in db_tasks.SCHEMA_COMPATIBILITY_CHECKS}

    assert "outcome_projects table" in checks
    assert "outcome_projects" in checks["outcome_projects table"]
    assert "team_registry_entries table" in checks
    assert "team_registry_entries" in checks["team_registry_entries table"]


def test_schema_bootstrapped_requires_search_source_registry_table():
    checks = {label: sql for label, sql in db_tasks.SCHEMA_COMPATIBILITY_CHECKS}

    assert "search_sources table" in checks
    assert "search_sources" in checks["search_sources table"]


def test_schema_bootstrapped_requires_operator_sse_event_ledger():
    checks = {label: sql for label, sql in db_tasks.SCHEMA_COMPATIBILITY_CHECKS}

    assert "operator_sse_events table" in checks
    assert "operator_sse_events" in checks["operator_sse_events table"]


def test_schema_bootstrapped_requires_team_signal_receipt_ledger():
    checks = {label: sql for label, sql in db_tasks.SCHEMA_COMPATIBILITY_CHECKS}

    assert "team_signal_receipts table" in checks
    assert "team_signal_receipts" in checks["team_signal_receipts table"]


def test_schema_bootstrapped_requires_worker_profile_columns():
    checks = {label: sql for label, sql in db_tasks.SCHEMA_COMPATIBILITY_CHECKS}

    assert "agent_catalogue profile_key column" in checks
    assert "agent_catalogue" in checks["agent_catalogue profile_key column"]


def test_schema_bootstrapped_requires_all_code_context_tables():
    checks = {label: sql for label, sql in db_tasks.SCHEMA_COMPATIBILITY_CHECKS}

    for table in (
        "code_context_sources",
        "code_context_snapshots",
        "code_context_files",
        "code_context_symbols",
        "code_context_edges",
    ):
        assert f"{table} table" in checks
        assert table in checks[f"{table} table"]


def test_schema_bootstrapped_requires_team_work_lifecycle_columns():
    checks = {label: sql for label, sql in db_tasks.SCHEMA_COMPATIBILITY_CHECKS}

    for table in ("team_work_items", "team_status_events"):
        for column in ("work_intent", "execution_mode"):
            label = f"{table} {column} column"
            assert label in checks
            assert table in checks[label]
            assert column in checks[label]


def test_schema_bootstrapped_accepts_current_runtime_schema(monkeypatch):
    monkeypatch.setattr(db_tasks, "_load_env", lambda: None)
    monkeypatch.setattr(
        db_tasks,
        "_run_psql",
        lambda sql=None, file=None, dbname=None: SimpleNamespace(returncode=0, stdout="1\n", stderr=""),
    )

    assert db_tasks.schema_bootstrapped() is True
