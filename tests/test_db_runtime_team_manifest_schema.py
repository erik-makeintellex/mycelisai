from ops import compose_storage
from ops import db as db_tasks


def test_runtime_team_manifest_schema_has_targeted_repair():
    checks = dict(db_tasks.SCHEMA_COMPATIBILITY_CHECKS)
    assert "runtime_team_manifests table" in checks
    assert (
        db_tasks.TARGETED_SCHEMA_MIGRATIONS["runtime_team_manifests table"]
        == "063_runtime_team_manifests.up.sql"
    )


def test_compose_storage_repairs_runtime_team_manifests():
    checks = dict(compose_storage.COMPOSE_LONG_TERM_STORAGE_CHECKS)
    assert "runtime team manifests" in checks
    assert compose_storage.COMPOSE_STORAGE_MIGRATIONS_BY_CHECK["runtime team manifests"] == (
        "063_runtime_team_manifests.up.sql",
    )
