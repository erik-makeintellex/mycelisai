from __future__ import annotations

from ops import db as db_tasks


def test_schema_bootstrapped_requires_team_work_recovery_deadline():
    checks = {label: sql for label, sql in db_tasks.SCHEMA_COMPATIBILITY_CHECKS}

    label = "team_work_items recovery_deadline_at column"
    assert label in checks
    assert "recovery_deadline_at" in checks[label]


def test_current_schema_contains_recovery_deadline():
    baseline = (db_tasks.MIGRATIONS_DIR / "001_current_schema.sql").read_text(encoding="utf-8")

    assert "recovery_deadline_at TIMESTAMPTZ" in baseline
