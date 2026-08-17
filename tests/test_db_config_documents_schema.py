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
        assert db_tasks.TARGETED_SCHEMA_MIGRATIONS[label] == "061_config_documents.up.sql"


def test_config_document_activation_history_preserves_rollback_boundary():
    migration = (db_tasks.MIGRATIONS_DIR / "061_config_documents.up.sql").read_text(
        encoding="utf-8"
    )

    assert "config_document_activations" in migration
    assert "config_document_activation_history" in migration
    assert "from_record_id" in migration
    assert "to_record_id" in migration
    assert "PRIMARY KEY (tenant_id, kind, document_id, scope_kind, scope_ref)" in migration
    assert "CHECK (action IN ('activate', 'rollback'))" in migration


def test_compose_targeted_repair_includes_config_document_schema():
    checks = {label for label, _sql in compose_storage.COMPOSE_LONG_TERM_STORAGE_CHECKS}
    expected = {
        "config documents",
        "config document activations",
        "config document activation history",
    }
    assert expected <= checks
    for label in expected:
        assert compose_storage.COMPOSE_STORAGE_MIGRATIONS_BY_CHECK[label] == (
            "061_config_documents.up.sql",
        )
