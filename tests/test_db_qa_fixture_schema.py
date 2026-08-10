from __future__ import annotations

from ops import db as db_tasks


def test_schema_bootstrapped_requires_qa_fixture_ownership():
    checks = {label: sql for label, sql in db_tasks.SCHEMA_COMPATIBILITY_CHECKS}

    assert "qa_fixture_scopes table" in checks
    assert "qa_fixture_scopes" in checks["qa_fixture_scopes table"]
    assert "qa_fixture_resources table" in checks
    assert "qa_fixture_resources" in checks["qa_fixture_resources table"]
    assert "qa_fixture resource ownership index" in checks
    assert "uq_qa_fixture_resource_claim" in checks["qa_fixture resource ownership index"]
    assert "purged QA fixture claims released" in checks
    assert "NOT EXISTS" in checks["purged QA fixture claims released"]
    assert db_tasks.TARGETED_SCHEMA_MIGRATIONS["qa_fixture_scopes table"] == "058_qa_fixture_ownership.up.sql"
    assert db_tasks.TARGETED_SCHEMA_MIGRATIONS["qa_fixture_resources table"] == "058_qa_fixture_ownership.up.sql"
    assert db_tasks.TARGETED_SCHEMA_MIGRATIONS["qa_fixture resource ownership index"] == "059_qa_fixture_ownership_hardening.up.sql"
    assert db_tasks.TARGETED_SCHEMA_MIGRATIONS["purged QA fixture claims released"] == "060_release_purged_qa_fixture_claims.up.sql"


def test_qa_fixture_hardening_fails_closed_before_unique_index():
    migration = (db_tasks.MIGRATIONS_DIR / "059_qa_fixture_ownership_hardening.up.sql").read_text(
        encoding="utf-8"
    )

    release_at = migration.index("DELETE FROM qa_fixture_resources")
    ambiguity_at = migration.index("HAVING COUNT(*) > 1")
    index_at = migration.index("CREATE UNIQUE INDEX")
    assert release_at < ambiguity_at < index_at
    assert "SET status = 'partial'" in migration
