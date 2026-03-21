from __future__ import annotations

import builtins
from pathlib import Path
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


def test_migration_files_only_include_forward_steps(monkeypatch):
    fake_files = [
        db_tasks.MIGRATIONS_DIR / "001_init_memory.sql",
        db_tasks.MIGRATIONS_DIR / "020_cascade_service_manifests.down.sql",
        db_tasks.MIGRATIONS_DIR / "020_cascade_service_manifests.up.sql",
        db_tasks.MIGRATIONS_DIR / "031_inception_recipes.down.sql",
        db_tasks.MIGRATIONS_DIR / "031_inception_recipes.up.sql",
    ]
    original_glob = Path.glob

    def fake_glob(self, pattern):
        if self == db_tasks.MIGRATIONS_DIR:
            return fake_files
        return original_glob(self, pattern)

    monkeypatch.setattr(Path, "glob", fake_glob)

    selected = db_tasks._migration_files()

    assert [path.name for path in selected] == [
        "001_init_memory.sql",
        "020_cascade_service_manifests.up.sql",
        "031_inception_recipes.up.sql",
    ]


def test_reset_fails_fast_when_a_migration_errors(monkeypatch):
    calls: list[str] = []

    monkeypatch.setattr(db_tasks, "_load_env", lambda: None)
    monkeypatch.setattr(db_tasks, "_require_postgres", lambda dbname="postgres": None)
    monkeypatch.setenv("DB_NAME", "cortex")
    monkeypatch.setattr(
        db_tasks,
        "_migration_files",
        lambda: [db_tasks.MIGRATIONS_DIR / "001_init_memory.sql", db_tasks.MIGRATIONS_DIR / "019_agent_memories.up.sql"],
    )

    def fake_run_psql(sql=None, file=None, dbname=None):
        if sql is not None:
            calls.append(sql)
            return 0
        calls.append(file.name)
        return 1 if file.name == "019_agent_memories.up.sql" else 0

    monkeypatch.setattr(db_tasks, "_psql", fake_run_psql)
    monkeypatch.setattr(db_tasks, "_ensure_database_exists", lambda: None)

    with pytest.raises(SystemExit, match="Migration failed: 019_agent_memories.up.sql"):
        db_tasks.reset.body(None)

    assert "019_agent_memories.up.sql" in calls
