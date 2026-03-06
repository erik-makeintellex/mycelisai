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
