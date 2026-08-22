from __future__ import annotations

from types import SimpleNamespace

from ops import db as db_tasks


def test_schema_bootstrapped_requires_team_work_recovery_deadline():
    checks = {label: sql for label, sql in db_tasks.SCHEMA_COMPATIBILITY_CHECKS}

    label = "team_work_items recovery_deadline_at column"
    assert label in checks
    assert "recovery_deadline_at" in checks[label]
    assert db_tasks.TARGETED_SCHEMA_MIGRATIONS[label] == "057_team_work_recovery_deadline.up.sql"


def test_targeted_recovery_deadline_migration_applies_when_column_is_missing(monkeypatch):
    migration = db_tasks.MIGRATIONS_DIR / "057_team_work_recovery_deadline.up.sql"
    applied: list[str] = []
    monkeypatch.setattr(db_tasks, "_migration_files", lambda: [migration])
    monkeypatch.setattr(
        db_tasks,
        "TARGETED_SCHEMA_MIGRATIONS",
        {"team_work_items recovery_deadline_at column": migration.name},
    )
    monkeypatch.setattr(
        db_tasks,
        "_run_psql",
        lambda sql=None, file=None, dbname=None: SimpleNamespace(returncode=0, stdout="", stderr=""),
    )
    monkeypatch.setattr(db_tasks, "_psql", lambda sql=None, file=None, dbname=None: applied.append(file.name) or 0)

    assert db_tasks._apply_missing_targeted_migrations() is True
    assert applied == ["057_team_work_recovery_deadline.up.sql"]
