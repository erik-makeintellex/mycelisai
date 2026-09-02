from __future__ import annotations

from ops import db as db_tasks
from ops import compose_storage


def test_schema_bootstrap_requires_config_document_activation_state():
    checks = {label: sql for label, sql in db_tasks.SCHEMA_COMPATIBILITY_CHECKS}

    for label, table in (
        ("config_documents table", "config_documents"),
        ("config_document_activations table", "config_document_activations"),
        ("config_document_activation_history table", "config_document_activation_history"),
    ):
        assert table in checks[label]


def test_config_document_activation_history_preserves_rollback_boundary():
    migration = (db_tasks.MIGRATIONS_DIR / "001_current_schema.sql").read_text(
        encoding="utf-8"
    )

    assert "config_document_activations" in migration
    assert "config_document_activation_history" in migration
    assert "from_record_id" in migration
    assert "to_record_id" in migration
    assert "PRIMARY KEY (tenant_id, kind, document_id, scope_kind, scope_ref)" in migration
    assert "CHECK (action IN ('activate', 'rollback'))" in migration


def test_compose_storage_health_includes_config_document_schema():
    checks = {label for label, _sql in compose_storage.COMPOSE_LONG_TERM_STORAGE_CHECKS}
    expected = {
        "config_documents table",
        "config_document_activations table",
        "config_document_activation_history table",
    }
    assert expected <= checks


def test_fixture_ownership_is_part_of_current_schema_contract():
    checks = dict(db_tasks.SCHEMA_COMPATIBILITY_CHECKS)
    compose_checks = dict(compose_storage.COMPOSE_LONG_TERM_STORAGE_CHECKS)
    baseline = (db_tasks.MIGRATIONS_DIR / "001_current_schema.sql").read_text(encoding="utf-8")

    assert "config_document" in checks["config_document fixture ownership"]
    assert "config_document" in compose_checks["config_document fixture ownership"]
    assert "chk_qa_fixture_resource_kind" in baseline
    assert "config_document" in baseline
