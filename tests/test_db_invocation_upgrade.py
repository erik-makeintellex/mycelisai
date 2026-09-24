from types import SimpleNamespace

import pytest

from ops import db_upgrade


def result(ok=True):
    return SimpleNamespace(returncode=0, stdout="1\n" if ok else "", stderr="")


@pytest.fixture
def upgrade_fixture(tmp_path, monkeypatch):
    schema = tmp_path / "001_current_schema.sql"
    schema.write_text("HISTORICAL DESTRUCTIVE SQL\n-- BEGIN G4_E10_EXTENSION\nSELECT 'g4';\n-- END G4_E10_EXTENSION\n"
                      "-- BEGIN C2A_TEAM_OWNERSHIP_EXTENSION\nSELECT 'c2a';\n-- END C2A_TEAM_OWNERSHIP_EXTENSION\n")
    monkeypatch.setattr(db_upgrade, "BASE_SCHEMA_COMPATIBILITY_CHECKS", (("base", "base-check"),))
    monkeypatch.setattr(db_upgrade, "G4_SCHEMA_COMPATIBILITY_CHECKS", (("g4", "g4-check"),))
    monkeypatch.setattr(db_upgrade, "SCHEMA_COMPATIBILITY_CHECKS", (("new", "new-check"),))
    return schema


@pytest.mark.parametrize("g4_present", [False, True])
def test_upgrade_runs_only_missing_bounded_extensions(upgrade_fixture, g4_present):
    calls = []
    applied = False
    def run(sql):
        nonlocal applied
        calls.append(sql)
        if sql.startswith("BEGIN;"):
            applied = True
        return result(applied if sql == "new-check" else g4_present if sql == "g4-check" else True)
    assert db_upgrade.upgrade_retained(upgrade_fixture, run)
    transactions = [sql for sql in calls if sql.startswith("BEGIN;")]
    assert len(transactions) == 1
    transaction = transactions[0]
    assert "HISTORICAL" not in transaction
    assert ("SELECT 'g4';" in transaction) == (not g4_present)
    assert "SELECT 'c2a';" in transaction
    assert transaction.endswith("COMMIT;\n")
    assert calls[-1] == "new-check"


@pytest.mark.parametrize("failure", ["base-check", db_upgrade.G4_ABSENT_SQL, db_upgrade.C2A_ABSENT_SQL])
def test_unknown_baseline_and_partial_upgrade_do_not_execute(upgrade_fixture, failure):
    calls = []
    def run(sql):
        calls.append(sql)
        return result(sql not in (failure, "g4-check", "new-check"))
    assert not db_upgrade.upgrade_retained(upgrade_fixture, run)
    assert not any(sql.startswith("BEGIN;") for sql in calls)


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
        return result(sql not in ("g4-check", "new-check"))
    with pytest.raises(SystemExit, match="rolled back"):
        db_upgrade.upgrade_retained(upgrade_fixture, run)


def test_missing_c2a_boundary_prevents_g4_mutation(upgrade_fixture):
    upgrade_fixture.write_text("-- BEGIN G4_E10_EXTENSION\nSELECT 1;\n-- END G4_E10_EXTENSION")
    calls = []
    def run(sql):
        calls.append(sql)
        return result(sql not in ("g4-check", "new-check"))
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
