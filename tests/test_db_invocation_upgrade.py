from types import SimpleNamespace

import pytest

from ops import db_upgrade

# Accepted retained baselines: pre-G4, G4, C2a, ORG/ROLE-complete, TOKEN (the retained stack), TAXONOMY, LEDGER.
BASELINES = {"pre-g4": (False,) * 6, "g4": (True,) + (False,) * 5, "c2a": (True,) * 2 + (False,) * 4,
             "org": (True,) * 3 + (False,) * 3, "token": (True,) * 4 + (False,) * 2,
             "taxonomy": (True,) * 5 + (False,), "ledger": (True,) * 6}
BLOCKS = ("-- BEGIN G4_E10_EXTENSION\nSELECT 'g4';\n-- END G4_E10_EXTENSION\n"
          "-- BEGIN C2A_TEAM_OWNERSHIP_EXTENSION\nSELECT 'c2a';\n-- END C2A_TEAM_OWNERSHIP_EXTENSION\n"
          "-- BEGIN ORGANIZATIONS_EXTENSION\nSELECT 'org';\n-- END ORGANIZATIONS_EXTENSION\n"
          "-- BEGIN ROLE_SEED_RETIREMENT_EXTENSION\nSELECT 'role';\n-- END ROLE_SEED_RETIREMENT_EXTENSION\n"
          "-- BEGIN CONFIRM_TOKEN_BINDING_EXTENSION\nSELECT 'token';\n-- END CONFIRM_TOKEN_BINDING_EXTENSION\n"
          "-- BEGIN DEPLOYMENT_CONTEXT_TAXONOMY_EXTENSION\nSELECT 'taxonomy';\n"
          "-- END DEPLOYMENT_CONTEXT_TAXONOMY_EXTENSION\n"
          "-- BEGIN TOKEN_USAGE_LEDGER_EXTENSION\nSELECT 'ledger';\n-- END TOKEN_USAGE_LEDGER_EXTENSION\n")


def result(ok=True):
    return SimpleNamespace(returncode=0, stdout="1\n" if ok else "", stderr="")


@pytest.fixture
def upgrade_fixture(tmp_path, monkeypatch):
    schema = tmp_path / "001_current_schema.sql"
    schema.write_text("HISTORICAL DESTRUCTIVE SQL\n" + BLOCKS + "\nCOMMIT;\n")
    monkeypatch.setattr(db_upgrade, "BASE_SCHEMA_COMPATIBILITY_CHECKS", (("base", "base-check"),))
    monkeypatch.setattr(db_upgrade, "G4_SCHEMA_COMPATIBILITY_CHECKS", (("g4", "g4-check"),))
    monkeypatch.setattr(db_upgrade, "C2A_SCHEMA_COMPATIBILITY_CHECKS", (("c2a", "c2a-check"),))
    monkeypatch.setattr(db_upgrade, "ORG_SCHEMA_COMPATIBILITY_CHECKS", (("org", "org-check"),))
    monkeypatch.setattr(db_upgrade, "TOKEN_SCHEMA_COMPATIBILITY_CHECKS", (("token", "token-check"),))
    monkeypatch.setattr(db_upgrade, "TAXONOMY_SCHEMA_COMPATIBILITY_CHECKS", (("taxonomy", "taxonomy-check"),))
    monkeypatch.setattr(db_upgrade, "LEDGER_SCHEMA_COMPATIBILITY_CHECKS", (("ledger", "ledger-check"),))
    monkeypatch.setattr(db_upgrade, "SCHEMA_COMPATIBILITY_CHECKS", (("new", "new-check"),))
    return schema


@pytest.mark.parametrize("baseline", BASELINES)
def test_upgrade_runs_only_missing_bounded_extensions(upgrade_fixture, baseline):
    g4_present, c2a_present, org_present, token_present, taxonomy_present, ledger_present = BASELINES[baseline]
    calls = []
    applied = False
    def run(sql):
        nonlocal applied
        calls.append(sql)
        if sql.startswith("BEGIN;"):
            applied = True
        states = {"new-check": applied, "g4-check": g4_present, "c2a-check": c2a_present,
                  "org-check": org_present, "token-check": token_present,
                  "taxonomy-check": taxonomy_present, "ledger-check": ledger_present}
        return result(states.get(sql, True))
    assert db_upgrade.upgrade_retained(upgrade_fixture, run)
    transactions = [sql for sql in calls if sql.startswith("BEGIN;")]
    assert len(transactions) == 1
    transaction = transactions[0]
    assert "HISTORICAL" not in transaction
    assert ("SELECT 'g4';" in transaction) == (not g4_present)
    assert ("SELECT 'c2a';" in transaction) == (not c2a_present)
    assert ("SELECT 'org';" in transaction) == (not org_present)
    # ROLE is an idempotent exact-tuple DELETE and always runs last in the transaction.
    assert transaction.index("SELECT 'role';") > max(transaction.find(s) for s in ("SELECT 'g4';", "SELECT 'c2a';", "SELECT 'org';"))
    if not org_present:
        assert transaction.index("SELECT 'org';") > max(transaction.find("SELECT 'g4';"), transaction.find("SELECT 'c2a';"))
    assert transaction.endswith("COMMIT;\n") and transaction.count("COMMIT;") == 1
    assert (db_upgrade.ORG_ABSENT_SQL in calls[:calls.index(transaction)]) == (not org_present)
    assert db_upgrade.ROLE_TABLE_SQL in calls[:calls.index(transaction)]
    # TOKEN is a plain ALTER: it runs only when absent, after ROLE, behind its absent check.
    assert ("SELECT 'token';" in transaction) == (not token_present)
    if not token_present:
        assert transaction.index("SELECT 'token';") > transaction.index("SELECT 'role';")
    assert (db_upgrade.TOKEN_ABSENT_SQL in calls[:calls.index(transaction)]) == (not token_present)
    # TAXONOMY is a marker-gated one-shot rewrite: it runs last, only when its marker is missing.
    assert ("SELECT 'taxonomy';" in transaction) == (not taxonomy_present)
    if not taxonomy_present:
        assert transaction.index("SELECT 'taxonomy';") > transaction.index("SELECT 'role';")
        assert transaction.index("SELECT 'taxonomy';") > transaction.find("SELECT 'token';")
    assert (db_upgrade.TAXONOMY_READY_SQL in calls[:calls.index(transaction)]) == (not taxonomy_present)
    # LEDGER is an additive table: it runs last, only when absent, behind its absent check.
    assert ("SELECT 'ledger';" in transaction) == (not ledger_present)
    if not ledger_present:
        assert transaction.index("SELECT 'ledger';") > transaction.index("SELECT 'role';")
        assert transaction.index("SELECT 'ledger';") > transaction.find("SELECT 'taxonomy';")
    assert (db_upgrade.LEDGER_ABSENT_SQL in calls[:calls.index(transaction)]) == (not ledger_present)
    assert calls[-1] == "new-check"


@pytest.mark.parametrize("failure", ["base-check", db_upgrade.G4_ABSENT_SQL, db_upgrade.C2A_ABSENT_SQL,
                                     db_upgrade.ORG_ABSENT_SQL, db_upgrade.ROLE_TABLE_SQL,
                                     db_upgrade.TOKEN_ABSENT_SQL, db_upgrade.TAXONOMY_READY_SQL,
                                     db_upgrade.LEDGER_ABSENT_SQL])
def test_unknown_baseline_and_partial_upgrade_do_not_execute(upgrade_fixture, failure):
    calls = []
    def run(sql):
        calls.append(sql)
        return result(sql not in (failure, "g4-check", "c2a-check", "org-check", "token-check",
                                  "taxonomy-check", "ledger-check", "new-check"))
    assert not db_upgrade.upgrade_retained(upgrade_fixture, run)
    assert not any(sql.startswith("BEGIN;") for sql in calls)


def test_partial_organizations_on_retained_c2a_stack_is_refused(upgrade_fixture):
    calls = []
    def run(sql):
        calls.append(sql)
        return result(sql not in ("new-check", "org-check", db_upgrade.ORG_ABSENT_SQL))
    assert not db_upgrade.upgrade_retained(upgrade_fixture, run)
    assert not any(sql.startswith("BEGIN;") for sql in calls)


def test_c2a_without_g4_is_not_an_accepted_baseline(upgrade_fixture):
    # C2a checks are cumulative; a host that fails G4 and has C2a residue is refused.
    def run(sql):
        return result(sql not in ("g4-check", "new-check", db_upgrade.C2A_ABSENT_SQL))
    assert not db_upgrade.upgrade_retained(upgrade_fixture, run)


def test_compatible_retained_schema_is_noop(upgrade_fixture):
    calls = []
    def run(sql):
        calls.append(sql)
        return result()
    assert db_upgrade.upgrade_retained(upgrade_fixture, run)
    assert not any(sql.startswith("BEGIN;") for sql in calls)


def test_failed_upgrade_stops_before_success(upgrade_fixture):
    def run(sql):
        if sql.startswith("BEGIN;"):
            return SimpleNamespace(returncode=1, stdout="", stderr="failed")
        return result(sql != "new-check")
    with pytest.raises(SystemExit, match="rolled back"):
        db_upgrade.upgrade_retained(upgrade_fixture, run)


@pytest.mark.parametrize("text", [
    "-- BEGIN G4_E10_EXTENSION\nSELECT 1;\n-- END G4_E10_EXTENSION",
    BLOCKS.split("-- BEGIN ORGANIZATIONS_EXTENSION")[0],
    BLOCKS.replace("-- END ORGANIZATIONS_EXTENSION", "-- END ORGANIZATIONS_EXTENSION\n-- END ORGANIZATIONS_EXTENSION"),
    BLOCKS.split("-- BEGIN ORGANIZATIONS_EXTENSION")[1].split("-- END ORGANIZATIONS_EXTENSION")[1]
    + "-- BEGIN ORGANIZATIONS_EXTENSION\n-- END ORGANIZATIONS_EXTENSION\n" + BLOCKS.split("-- BEGIN ORGANIZATIONS_EXTENSION")[0],
])
def test_missing_or_misordered_boundary_prevents_any_mutation(upgrade_fixture, text):
    upgrade_fixture.write_text(text)
    calls = []
    def run(sql):
        calls.append(sql)
        return result(sql not in ("c2a-check", "new-check"))
    with pytest.raises(SystemExit, match="unique"):
        db_upgrade.upgrade_retained(upgrade_fixture, run)
    assert not any(sql.startswith("BEGIN;") for sql in calls)


def test_missing_duplicate_or_reversed_extension_boundary_denied(tmp_path):
    path = tmp_path / "schema.sql"
    for text in ("no extension", "-- BEGIN G4_E10_EXTENSION\n-- BEGIN G4_E10_EXTENSION\n-- END G4_E10_EXTENSION",
                 "-- END G4_E10_EXTENSION\n-- BEGIN G4_E10_EXTENSION"):
        path.write_text(text)
        with pytest.raises(SystemExit, match="unique"):
            db_upgrade.extension_sql(path)
