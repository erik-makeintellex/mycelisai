from ops import compose_storage
from ops import db as db_tasks


def test_runtime_team_manifest_schema_is_required_for_compatibility():
    checks = dict(db_tasks.SCHEMA_COMPATIBILITY_CHECKS)
    assert "runtime_team_manifests table" in checks


def test_compose_storage_health_requires_runtime_team_manifests():
    checks = dict(compose_storage.COMPOSE_LONG_TERM_STORAGE_CHECKS)
    assert "runtime_team_manifests table" in checks


def test_current_schema_contains_runtime_team_manifests():
    baseline = (db_tasks.MIGRATIONS_DIR / "001_current_schema.sql").read_text(encoding="utf-8")

    assert "CREATE TABLE IF NOT EXISTS runtime_team_manifests" in baseline
