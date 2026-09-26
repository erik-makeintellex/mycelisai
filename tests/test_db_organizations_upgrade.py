"""Real ORG retained-upgrade proof in uniquely named disposable databases only."""
import json

import pytest

from ops import compose, db_upgrade
from ops.db_schema import SCHEMA_COMPATIBILITY_CHECKS
from tests.test_db_team_ownership_upgrade import SCHEMA, database  # noqa: F401 (fixture)

# Accepted starting schemas: pre-G4, G4-complete, C2a-complete (the retained stack).
BASELINE_MARKERS = {"pre-g4": db_upgrade.BEGIN_MARKER, "g4": db_upgrade.C2A_BEGIN_MARKER,
                    "c2a": db_upgrade.ORG_BEGIN_MARKER}
MANIFEST_ROW = "SELECT row_to_json(m)::text FROM runtime_team_manifests m;"
ORG_ROW = ("INSERT INTO organizations (id,name,document,qa_fixture_scope_id) VALUES "
           "('11111111-1111-4111-8111-111111111111','Retained Org','{\"id\":\"x\"}','scope-a');")


def install(run, baseline):
    result = run(SCHEMA.read_text().split(BASELINE_MARKERS[baseline])[0] + "COMMIT;\n")
    assert result.returncode == 0, result.stderr
    result = run("INSERT INTO runtime_team_manifests (tenant_id,team_id,manifest_digest,manifest) "
                 "VALUES ('retained','retained-team','sha256:retained','{\"id\":\"retained-team\"}');")
    assert result.returncode == 0, result.stderr
    return json.loads(run(MANIFEST_ROW).stdout)


def recording(run):
    calls = []
    def wrapped(sql):
        calls.append(sql)
        return run(sql)
    return wrapped, calls


def assert_contract(run):
    for label, sql in SCHEMA_COMPATIBILITY_CHECKS:
        result = run(sql)
        assert result.returncode == 0 and result.stdout.strip() == "1", (label, result.stderr)


@pytest.mark.parametrize("baseline", BASELINE_MARKERS)
def test_real_accepted_baselines_upgrade_preserving_rows(database, baseline):
    before = install(database, baseline)
    assert database(db_upgrade.ORG_ABSENT_SQL).stdout.strip() == "1"
    assert db_upgrade.upgrade_retained(SCHEMA, database)
    assert_contract(database)
    after = json.loads(database(MANIFEST_ROW).stdout)
    assert {key: after[key] for key in before} == before
    assert database("SELECT count(*) FROM organizations;").stdout.strip() == "0"
    run, calls = recording(database)
    assert db_upgrade.upgrade_retained(SCHEMA, run)  # compatible schema is a no-op
    assert not any(sql.startswith("BEGIN;") for sql in calls)
    # Constraints, not application code, enforce the durable row invariants.
    assert database(ORG_ROW).returncode == 0
    row = database("SELECT tenant_id||'|'||qa_fixture_scope_id||'|'||purpose FROM organizations;")
    assert row.stdout.strip() == "default|scope-a|"
    assert database(ORG_ROW).returncode != 0  # duplicate id is rejected, never upserted
    for bad in ("'[]'", "'\"s\"'", "NULL"):
        assert database(f"INSERT INTO organizations (id,name,document) VALUES (gen_random_uuid(),'b',{bad});").returncode != 0
    assert database("INSERT INTO organizations (id,name,document,qa_fixture_scope_id) "
                    "VALUES (gen_random_uuid(),'b','{}',NULL);").returncode != 0


@pytest.mark.parametrize("baseline", ["pre-g4", "c2a"])
def test_real_partial_organizations_table_refused_before_any_sql(database, baseline):
    before = install(database, baseline)
    assert database("CREATE TABLE organizations (id UUID PRIMARY KEY, name TEXT);").returncode == 0
    run, calls = recording(database)
    assert not db_upgrade.upgrade_retained(SCHEMA, run)
    assert not any(sql.startswith("BEGIN;") for sql in calls)
    columns = "SELECT count(*) FROM information_schema.columns WHERE table_schema='public' AND table_name='organizations';"
    assert database(columns).stdout.strip() == "2"
    if baseline == "pre-g4":
        assert database(db_upgrade.G4_ABSENT_SQL).stdout.strip() == "1"
    assert json.loads(database(MANIFEST_ROW).stdout) == before


def test_real_failed_org_extension_rolls_back_all_schema_changes(database, tmp_path):
    before = install(database, "g4")
    broken = tmp_path / "broken.sql"
    broken.write_text(SCHEMA.read_text().replace(db_upgrade.ORG_END_MARKER,
                      "SELECT 1 / 0;\n" + db_upgrade.ORG_END_MARKER))
    with pytest.raises(SystemExit, match="rolled back"):
        db_upgrade.upgrade_retained(broken, database)
    assert database(db_upgrade.C2A_ABSENT_SQL).stdout.strip() == "1"
    assert database(db_upgrade.ORG_ABSENT_SQL).stdout.strip() == "1"
    assert json.loads(database(MANIFEST_ROW).stdout) == before


def test_real_weakened_org_contract_is_refused(database):
    result = database(SCHEMA.read_text())
    assert result.returncode == 0, result.stderr
    assert_contract(database)
    assert database("ALTER TABLE organizations DROP CONSTRAINT chk_organizations_document_object;").returncode == 0
    assert database("ALTER TABLE organizations ADD CONSTRAINT chk_organizations_document_object CHECK (true);").returncode == 0
    run, calls = recording(database)
    assert not db_upgrade.upgrade_retained(SCHEMA, run)
    assert not any(sql.startswith("BEGIN;") for sql in calls)


def test_real_compose_migrate_upgrades_retained_c2a_stack(database, monkeypatch, capsys):
    """Drive the compose.migrate decision chain against a C2a-complete retained schema."""
    before = install(database, "c2a")
    # Only the transport is replaced: compose's `exec postgres psql -tA ON_ERROR_STOP`
    # becomes psql against the disposable database; every decision stays compose's own.
    monkeypatch.setattr(compose, "_compose_effective_env", lambda env_values=None: {})
    monkeypatch.setattr(compose, "_run_compose_psql", lambda sql, env_values: database(sql))
    monkeypatch.setattr(compose, "_run_compose_migration_file",
                        lambda *_: pytest.fail("compose.migrate replayed the installer on retained data"))
    compose._run_compose_migrations()
    assert "Retained-schema upgrade complete (ORG + ROLE)" in capsys.readouterr().out
    assert_contract(database)
    after = json.loads(database(MANIFEST_ROW).stdout)
    assert {key: after[key] for key in before} == before
    compose._run_compose_migrations()
    assert "already appears compatible" in capsys.readouterr().out
