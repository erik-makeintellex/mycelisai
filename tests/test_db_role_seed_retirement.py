"""C4: retire the exact legacy role.* provider seeds once per database (fresh and retained)."""
from __future__ import annotations

import hashlib
import re
from types import SimpleNamespace

import pytest

from ops import compose, db_schema, db_upgrade
from tests.test_db_team_ownership_upgrade import SCHEMA, database  # noqa: F401 (fixture)

ROLE_BLOCK_SHA256 = "ed1913d106f1f9340f2d96d0df611606b986608f64e780dc49b173e287692cba"
ROLE_ROWS = "SELECT string_agg(key||'='||value, ',' ORDER BY key) FROM system_config WHERE key LIKE 'role.%';"
MARKER = db_schema.ROLE_SEED_RETIREMENT_MARKER
MARKER_CHECK = db_schema.ROLE_SCHEMA_COMPATIBILITY_CHECKS[-1][1]
MARKER_ROW = f"SELECT value FROM system_config WHERE key='{MARKER}';"


def _role_block(raw: bytes) -> bytes:
    begin, end = db_upgrade.ROLE_BEGIN_MARKER.encode(), db_upgrade.ROLE_END_MARKER.encode()
    assert raw.count(begin) == 1 and raw.count(end) == 1
    return raw[raw.index(begin) + len(begin):raw.index(end)]


def test_role_block_bytes_are_pinned():
    assert hashlib.sha256(_role_block(SCHEMA.read_bytes())).hexdigest() == ROLE_BLOCK_SHA256


def test_role_block_follows_org_and_precedes_the_single_commit():
    raw = SCHEMA.read_bytes()
    org_end = raw.index(db_upgrade.ORG_END_MARKER.encode())
    role_begin = raw.index(db_upgrade.ROLE_BEGIN_MARKER.encode())
    role_end = raw.index(db_upgrade.ROLE_END_MARKER.encode())
    assert org_end < role_begin < role_end
    # The CONFIRM_TOKEN_BINDING block follows ROLE and owns the single COMMIT tail.
    assert raw[role_end:].startswith(db_upgrade.ROLE_END_MARKER.encode() + b"\n\n"
                                     + db_upgrade.TOKEN_BEGIN_MARKER.encode())
    # 013 (ON CONFLICT DO UPDATE) is the last historical role write; the block runs after it.
    assert raw[:role_begin].rindex(b"INSERT INTO system_config") < raw.index(b"-- END SOURCE: core/migrations/013_")


def test_role_block_deletes_only_exact_legacy_seed_tuples_once():
    block = _role_block(SCHEMA.read_bytes()).decode()
    statements = [line for line in block.splitlines() if line and not line.startswith("--")]
    assert statements[0] == "DELETE FROM system_config WHERE (key, value) IN ("
    assert "".join(statements).count(";") == 2 and block.count("DELETE") == 1
    # The DELETE is gated on the marker, and the marker is written in the same transaction.
    assert f"AND NOT EXISTS (SELECT 1 FROM system_config WHERE key = '{MARKER}');" in block
    assert block.rstrip().endswith(f"INSERT INTO system_config (key, value) VALUES ('{MARKER}', '1')\n"
                                   "    ON CONFLICT (key) DO NOTHING;")
    assert block.index("DELETE") < block.index("INSERT")
    assert not re.search(r"\b(DROP|ALTER|UPDATE|TRUNCATE|CREATE|LIKE)\b", block)
    tuples = set(re.findall(r"\('(role\.[a-z]+)', '([a-z-]+)'\)", block))
    assert tuples == set(db_schema.LEGACY_ROLE_SEEDS)
    assert len(db_schema.LEGACY_ROLE_SEEDS) == 12
    assert {key for key, _ in tuples} == {f"role.{r}" for r in
                                          ("architect", "chat", "coder", "creative", "overseer", "sentry")}
    assert {value for _, value in tuples} == {"local-ollama-dev", "local-sovereign"}


def test_compatibility_check_is_the_one_shot_marker_after_org_snapshot():
    org = db_schema.ORG_SCHEMA_COMPATIBILITY_CHECKS
    checks = db_schema.ROLE_SCHEMA_COMPATIBILITY_CHECKS
    assert db_schema.SCHEMA_COMPATIBILITY_CHECKS[:len(checks)] == checks
    assert checks[:len(org)] == org and len(checks) == len(org) + 1
    assert checks[-1][0] == "legacy role seeds retired"
    assert MARKER_CHECK == f"SELECT 1 FROM system_config WHERE key='{MARKER}';"
    # Seed-equal role rows never fail the contract once retired: they are operator choices.
    assert "role." not in MARKER_CHECK and not any("system_config" in sql for _, sql in org)


def _stubbed(marker_present: bool, *, system_config=True):
    calls, applied = [], []
    def run(sql):
        calls.append(sql)
        if sql.startswith("BEGIN;"):
            applied.append(sql)
            return SimpleNamespace(returncode=0, stdout="", stderr="")
        ok = True
        if sql == MARKER_CHECK:
            ok = marker_present or bool(applied)
        elif sql == db_upgrade.ROLE_TABLE_SQL:
            ok = system_config
        elif sql in (db_upgrade.G4_ABSENT_SQL, db_upgrade.C2A_ABSENT_SQL, db_upgrade.ORG_ABSENT_SQL):
            ok = False  # ORG-complete: every earlier extension is present
        return SimpleNamespace(returncode=0, stdout="1\n" if ok else "", stderr="")
    return run, calls


def test_org_complete_host_without_marker_runs_only_the_role_block(capsys):
    run, calls = _stubbed(False)
    assert db_upgrade.upgrade_retained(SCHEMA, run)
    transactions = [sql for sql in calls if sql.startswith("BEGIN;")]
    assert len(transactions) == 1
    body = transactions[0]
    assert "DELETE FROM system_config WHERE (key, value) IN (" in body
    assert body.index("DELETE") < body.index(f"VALUES ('{MARKER}', '1')") < body.index("COMMIT;")
    for marker in ("CREATE TABLE organizations", "execution_effect_grants", "fk_runtime_team_owner"):
        assert marker not in body
    assert "upgrade complete (ROLE)" in capsys.readouterr().out


def test_org_complete_host_with_marker_is_noop():
    run, calls = _stubbed(True)
    assert db_upgrade.upgrade_retained(SCHEMA, run)
    assert not any(sql.startswith("BEGIN;") for sql in calls)


def test_missing_system_config_is_refused_before_any_sql():
    run, calls = _stubbed(False, system_config=False)
    assert not db_upgrade.upgrade_retained(SCHEMA, run)
    assert not any(sql.startswith("BEGIN;") for sql in calls)


# --- Real PostgreSQL (MYCELIS_SCHEMA_UPGRADE_TEST_DSN, disposable databases only) ---

def install_org_complete(run, extra_sql=""):
    result = run(SCHEMA.read_text().split(db_upgrade.ROLE_BEGIN_MARKER)[0] + extra_sql + "COMMIT;\n")
    assert result.returncode == 0, result.stderr
    return run(ROLE_ROWS).stdout.strip()


def assert_contract(run):
    for label, sql in db_schema.SCHEMA_COMPATIBILITY_CHECKS:
        result = run(sql)
        assert result.returncode == 0 and result.stdout.strip() == "1", (label, result.stderr)


def recording(run):
    calls = []
    def wrapped(sql):
        calls.append(sql)
        return run(sql)
    return wrapped, calls


def test_real_fresh_baseline_has_marker_and_no_role_rows(database):
    result = database(SCHEMA.read_text())
    assert result.returncode == 0, result.stderr
    assert database("SELECT count(*) FROM system_config WHERE key LIKE 'role.%';").stdout.strip() == "0"
    assert database(MARKER_ROW).stdout.strip() == "1"
    assert_contract(database)


def test_real_first_retained_upgrade_removes_seeds_and_sets_marker(database, capsys):
    before = install_org_complete(database, "INSERT INTO system_config (key,value) VALUES ('ui.theme','dark');\n")
    assert before == ",".join(f"role.{r}=local-ollama-dev" for r in
                              ("architect", "chat", "coder", "creative", "overseer", "sentry"))
    assert database(MARKER_ROW).stdout.strip() == ""
    providers = database("SELECT string_agg(id, ',' ORDER BY id) FROM llm_providers;").stdout
    assert db_upgrade.upgrade_retained(SCHEMA, database)
    assert "upgrade complete (ROLE + TOKEN + TAXONOMY + LEDGER)" in capsys.readouterr().out
    assert database(ROLE_ROWS).stdout.strip() == ""
    assert database(MARKER_ROW).stdout.strip() == "1"
    assert database("SELECT value FROM system_config WHERE key='ui.theme';").stdout.strip() == "dark"
    assert database("SELECT string_agg(id, ',' ORDER BY id) FROM llm_providers;").stdout == providers
    assert_contract(database)


def test_real_second_migrate_keeps_operator_seed_equal_choice(database):
    install_org_complete(database)
    assert db_upgrade.upgrade_retained(SCHEMA, database)  # first: retire seeds, set marker
    assert database("INSERT INTO system_config (key,value) VALUES ('role.chat','local-ollama-dev');").returncode == 0
    run, calls = recording(database)
    assert db_upgrade.upgrade_retained(SCHEMA, run)
    assert not any(sql.startswith("BEGIN;") for sql in calls)
    # Even replaying the block itself cannot retire the choice once the marker exists.
    assert database("BEGIN;\n" + _role_block(SCHEMA.read_bytes()).decode() + "COMMIT;\n").returncode == 0
    assert database(ROLE_ROWS).stdout.strip() == "role.chat=local-ollama-dev"
    assert_contract(database)


def test_real_retained_customized_rows_are_kept(database):
    custom = ("UPDATE system_config SET value='vllm' WHERE key='role.coder';"
              "UPDATE system_config SET value='Local-Ollama-Dev' WHERE key='role.chat';"
              "UPDATE system_config SET value='local-sovereign-2' WHERE key='role.sentry';"
              "INSERT INTO system_config (key,value) VALUES ('role.research','local-ollama-dev');\n")
    install_org_complete(database, custom)
    assert db_upgrade.upgrade_retained(SCHEMA, database)
    assert database(ROLE_ROWS).stdout.strip() == (
        "role.chat=Local-Ollama-Dev,role.coder=vllm,role.research=local-ollama-dev,role.sentry=local-sovereign-2")
    assert_contract(database)


def test_real_retained_cleaned_stack_only_gains_marker(database):
    install_org_complete(database, "DELETE FROM system_config WHERE key LIKE 'role.%';"
                                   "INSERT INTO system_config (key,value) VALUES ('role.coder','vllm');\n")
    assert db_upgrade.upgrade_retained(SCHEMA, database)
    assert database(ROLE_ROWS).stdout.strip() == "role.coder=vllm"
    assert database(MARKER_ROW).stdout.strip() == "1"
    run, calls = recording(database)
    assert db_upgrade.upgrade_retained(SCHEMA, run)
    assert not any(sql.startswith("BEGIN;") for sql in calls)


def test_real_failed_role_block_rolls_back(database, tmp_path):
    before = install_org_complete(database)
    broken = tmp_path / "broken.sql"
    broken.write_text(SCHEMA.read_text().replace(db_upgrade.ROLE_END_MARKER,
                      "SELECT 1 / 0;\n" + db_upgrade.ROLE_END_MARKER))
    with pytest.raises(SystemExit, match="rolled back"):
        db_upgrade.upgrade_retained(broken, database)
    assert database(ROLE_ROWS).stdout.strip() == before
    assert database(MARKER_ROW).stdout.strip() == ""


def test_real_compose_migrate_retires_seeds_on_org_complete_stack(database, monkeypatch, capsys):
    install_org_complete(database)
    monkeypatch.setattr(compose, "_compose_effective_env", lambda env_values=None: {})
    monkeypatch.setattr(compose, "_run_compose_psql", lambda sql, env_values: database(sql))
    monkeypatch.setattr(compose, "_run_compose_migration_file",
                        lambda *_: pytest.fail("compose.migrate replayed the installer on retained data"))
    compose._run_compose_migrations()
    assert "Retained-schema upgrade complete (ROLE + TOKEN + TAXONOMY + LEDGER)" in capsys.readouterr().out
    assert database(ROLE_ROWS).stdout.strip() == ""
    compose._run_compose_migrations()
    assert "already appears compatible" in capsys.readouterr().out
