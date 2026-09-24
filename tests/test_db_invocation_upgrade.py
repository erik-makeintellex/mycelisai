from types import SimpleNamespace

import pytest

from ops import db_upgrade


def result(ok=True):
    return SimpleNamespace(returncode=0, stdout="1\n" if ok else "", stderr="")


def test_upgrade_runs_only_bounded_extension(tmp_path, monkeypatch):
    schema = tmp_path / "001_current_schema.sql"
    schema.write_text("HISTORICAL DESTRUCTIVE SQL\n-- BEGIN G4_E10_EXTENSION\nSELECT 'extension';\n-- END G4_E10_EXTENSION\n")
    calls = []
    monkeypatch.setattr(db_upgrade, "BASE_SCHEMA_COMPATIBILITY_CHECKS", (("base", "base-check"),))
    monkeypatch.setattr(db_upgrade, "SCHEMA_COMPATIBILITY_CHECKS", (("new", "new-check"),))
    def run(sql):
        calls.append(sql)
        return result()
    assert db_upgrade.upgrade_retained(schema, run)
    applied = next(sql for sql in calls if sql.startswith("BEGIN;"))
    assert "HISTORICAL" not in applied
    assert applied.endswith("COMMIT;\n")
    assert calls[-1] == "new-check"


def test_unknown_baseline_and_partial_upgrade_do_not_execute(tmp_path, monkeypatch):
    monkeypatch.setattr(db_upgrade, "BASE_SCHEMA_COMPATIBILITY_CHECKS", (("base", "base-check"),))
    for fail_base in (True, False):
        calls = []
        def run(sql):
            calls.append(sql)
            return result(not fail_base and sql == "base-check")
        assert not db_upgrade.upgrade_retained(tmp_path / "absent.sql", run)
        assert not any(sql.startswith("BEGIN;") for sql in calls)


def test_failed_upgrade_stops_before_success(tmp_path, monkeypatch):
    schema = tmp_path / "schema.sql"
    schema.write_text("-- BEGIN G4_E10_EXTENSION\nBROKEN;\n-- END G4_E10_EXTENSION")
    monkeypatch.setattr(db_upgrade, "BASE_SCHEMA_COMPATIBILITY_CHECKS", ())
    def run(sql):
        return SimpleNamespace(returncode=1, stdout="", stderr="failed") if sql.startswith("BEGIN;") else result()
    with pytest.raises(SystemExit, match="rolled back"):
        db_upgrade.upgrade_retained(schema, run)


def test_missing_or_duplicate_extension_boundary_denied(tmp_path):
    path = tmp_path / "schema.sql"
    for text in ("no extension", "-- BEGIN G4_E10_EXTENSION\n-- BEGIN G4_E10_EXTENSION\n-- END G4_E10_EXTENSION"):
        path.write_text(text)
        with pytest.raises(SystemExit, match="unique"):
            db_upgrade.extension_sql(path)
