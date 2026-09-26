"""Real retained-upgrade proof in uniquely named disposable databases only."""
import os
from pathlib import Path
import shutil
import subprocess
from urllib.parse import urlsplit, urlunsplit
from uuid import uuid4

import pytest

from ops import db_upgrade
from ops.db_schema import SCHEMA_COMPATIBILITY_CHECKS

SCHEMA = Path(__file__).parents[1] / "core/migrations/001_current_schema.sql"


@pytest.fixture
def database():
    admin = os.environ.get("MYCELIS_SCHEMA_UPGRADE_TEST_DSN")
    if not admin:
        pytest.skip("MYCELIS_SCHEMA_UPGRADE_TEST_DSN requires disposable PostgreSQL with CREATEDB")
    if not shutil.which("psql"):
        pytest.fail("psql required for real schema upgrade proof")
    name = "c2a_upgrade_" + uuid4().hex
    def execute(dsn, sql):
        return subprocess.run(["psql", "-X", "-qAt", "-v", "ON_ERROR_STOP=1", dsn],
                              input=sql, capture_output=True, text=True, timeout=60)
    created = execute(admin, f'CREATE DATABASE "{name}";')
    assert created.returncode == 0, created.stderr
    parts = urlsplit(admin)
    target = urlunsplit((parts.scheme, parts.netloc, "/" + name, parts.query, parts.fragment))
    try:
        yield lambda sql: execute(target, sql)
    finally:
        dropped = execute(admin, f'DROP DATABASE "{name}";')
        assert dropped.returncode == 0, dropped.stderr


def install_baseline(run, *, g4=True):
    text = SCHEMA.read_text()
    marker = db_upgrade.C2A_BEGIN_MARKER if g4 else db_upgrade.BEGIN_MARKER
    result = run(text.split(marker)[0] + "COMMIT;\n")
    assert result.returncode == 0, result.stderr
    result = run("INSERT INTO runtime_team_manifests (tenant_id,team_id,manifest_digest,manifest) "
                 "VALUES ('retained','retained-team','sha256:retained','{\"id\":\"retained-team\",\"purpose\":\"keep\"}');")
    assert result.returncode == 0, result.stderr
    return run("SELECT row_to_json(m)::text FROM runtime_team_manifests m;").stdout.strip()


@pytest.mark.parametrize("g4", [False, True])
def test_real_retained_upgrade_preserves_rows_and_starts_unowned(database, g4):
    import json
    before = json.loads(install_baseline(database, g4=g4))
    assert db_upgrade.upgrade_retained(SCHEMA, database)
    for label, sql in SCHEMA_COMPATIBILITY_CHECKS:
        result = database(sql)
        assert result.returncode == 0 and "1" in result.stdout.split(), (label, result.stderr)
    after = json.loads(database("SELECT row_to_json(m)::text FROM runtime_team_manifests m;").stdout)
    assert {key: after[key] for key in before} == before
    assert all(after[column] is None for column, _ in db_upgrade.TEAM_OWNERSHIP_COLUMNS)
    assert database("SELECT count(*) FROM role_permissions WHERE permission_key='framework_runs.provision_team';").stdout.strip() == "0"
    assert db_upgrade.upgrade_retained(SCHEMA, database)
    assert json.loads(database("SELECT row_to_json(m)::text FROM runtime_team_manifests m;").stdout) == after
    invalid = database("UPDATE runtime_team_manifests SET owner_account_id=gen_random_uuid();")
    assert invalid.returncode != 0


def test_real_partial_c2a_refused_without_mutating_retained_row(database):
    before = install_baseline(database)
    assert database("ALTER TABLE runtime_team_manifests ADD COLUMN owner_account_id UUID;").returncode == 0
    assert not db_upgrade.upgrade_retained(SCHEMA, database)
    assert database("SELECT count(*) FROM information_schema.columns WHERE table_name='runtime_team_manifests' AND column_name='owner_group_id';").stdout.strip() == "0"
    import json
    after = json.loads(database("SELECT row_to_json(m)::text FROM runtime_team_manifests m;").stdout)
    del after["owner_account_id"]
    assert after == json.loads(before)


def test_real_failed_extension_rolls_back_all_schema_changes(database, tmp_path):
    before = install_baseline(database, g4=False)
    broken = tmp_path / "broken.sql"
    broken.write_text(SCHEMA.read_text().replace(db_upgrade.C2A_END_MARKER,
                      "SELECT 1 / 0;\n" + db_upgrade.C2A_END_MARKER))
    with pytest.raises(SystemExit, match="rolled back"):
        db_upgrade.upgrade_retained(broken, database)
    assert database(db_upgrade.G4_ABSENT_SQL).stdout.strip() == "1"
    assert database(db_upgrade.C2A_ABSENT_SQL).stdout.strip() == "1"
    assert database(db_upgrade.ORG_ABSENT_SQL).stdout.strip() == "1"
    assert database("SELECT row_to_json(m)::text FROM runtime_team_manifests m;").stdout.strip() == before


@pytest.mark.parametrize("constraint", ["chk_runtime_team_provisioned", "chk_runtime_team_revoked",
    "runtime_team_manifests_ownership_provisioned_by_fkey", "runtime_team_manifests_ownership_revoked_by_fkey"])
def test_real_weakened_ownership_contract_is_refused(database, constraint):
    result = database(SCHEMA.read_text())
    assert result.returncode == 0, result.stderr
    assert database(f"ALTER TABLE runtime_team_manifests DROP CONSTRAINT {constraint};").returncode == 0
    # A same-name always-true check must not masquerade as the accepted contract.
    assert database(f"ALTER TABLE runtime_team_manifests ADD CONSTRAINT {constraint} CHECK (true);").returncode == 0
    assert not db_upgrade.upgrade_retained(SCHEMA, database)
