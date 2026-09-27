"""A2b: confirm tokens record purpose, binding digest and minter; legacy rows stay NULL."""
from __future__ import annotations

import hashlib
import json
import re
from types import SimpleNamespace

import pytest

from ops import compose, db_schema, db_upgrade
from tests.test_db_team_ownership_upgrade import SCHEMA, database  # noqa: F401 (fixture)

TOKEN_BLOCK_SHA256 = "3a3353c5128b2da172c08f25431b177c83c07f2ec19decbb9fb3c13405adf6d8"
FROZEN_BLOCK = (
    "\n-- Confirm tokens record purpose, bound-subject digest and minting principal at\n"
    "-- mint. Legacy rows keep NULLs and are refused by Core (never backfilled).\n"
    "ALTER TABLE confirm_tokens ADD COLUMN purpose TEXT, ADD COLUMN binding_digest TEXT, ADD COLUMN minted_by TEXT;\n"
    "ALTER TABLE confirm_tokens ADD CONSTRAINT chk_confirm_tokens_purpose CHECK (purpose IS NULL OR purpose IN "
    "('chat_action', 'mission_blueprint', 'group_mutation', 'invocation'));\n"
)
# Accepted retained baselines, named by the first extension each one lacks.
BASELINES = {"pre-g4": db_upgrade.BEGIN_MARKER, "g4": db_upgrade.C2A_BEGIN_MARKER,
             "c2a": db_upgrade.ORG_BEGIN_MARKER, "org": db_upgrade.ROLE_BEGIN_MARKER,
             "role": db_upgrade.TOKEN_BEGIN_MARKER}
PROOF_ID = "22222222-2222-4222-8222-222222222222"
LEGACY_TOKEN = "33333333-3333-4333-8333-333333333333"
LEGACY_ROWS = (
    "INSERT INTO intent_proofs (id, template_id, resolved_intent) VALUES "
    f"('{PROOF_ID}', 'chat-to-proposal', 'chat_action');"
    "INSERT INTO confirm_tokens (token, intent_proof_id, template_id, expires_at) VALUES "
    f"('{LEGACY_TOKEN}', '{PROOF_ID}', 'chat-to-proposal', NOW() + interval '15 minutes');"
)
TOKEN_ROW = "SELECT row_to_json(t)::text FROM confirm_tokens t;"
TOKEN_CHECKS = db_schema.TOKEN_SCHEMA_COMPATIBILITY_CHECKS


def _token_block(raw: bytes) -> bytes:
    begin, end = db_upgrade.TOKEN_BEGIN_MARKER.encode(), db_upgrade.TOKEN_END_MARKER.encode()
    assert raw.count(begin) == 1 and raw.count(end) == 1
    return raw[raw.index(begin) + len(begin):raw.index(end)]


def test_token_block_bytes_are_pinned_and_match_the_frozen_packet():
    block = _token_block(SCHEMA.read_bytes())
    assert hashlib.sha256(block).hexdigest() == TOKEN_BLOCK_SHA256
    assert block.decode() == FROZEN_BLOCK


def test_token_block_follows_role_and_precedes_the_single_commit():
    raw = SCHEMA.read_bytes()
    role_end = raw.index(db_upgrade.ROLE_END_MARKER.encode())
    token_begin = raw.index(db_upgrade.TOKEN_BEGIN_MARKER.encode())
    token_end = raw.index(db_upgrade.TOKEN_END_MARKER.encode())
    assert role_end < token_begin < token_end
    # The DEPLOYMENT_CONTEXT_TAXONOMY block follows TOKEN and owns the single COMMIT tail.
    assert raw[token_end:].startswith(db_upgrade.TOKEN_END_MARKER.encode() + b"\n\n"
                                      + db_upgrade.TAXONOMY_BEGIN_MARKER.encode())
    assert re.findall(rb"^COMMIT;$", raw, re.MULTILINE) == [b"COMMIT;"]


def test_token_block_is_only_two_plain_confirm_token_alters():
    block = _token_block(SCHEMA.read_bytes()).decode()
    statements = [line for line in block.splitlines() if line and not line.startswith("--")]
    assert len(statements) == 2 and all(s.startswith("ALTER TABLE confirm_tokens ADD ") for s in statements)
    assert block.count(";") == 2
    # Plain ADD refuses partial prior state; no backfill, default or destructive DDL.
    assert not re.search(r"\b(UPDATE|DROP|DELETE|DEFAULT|INSERT|TRUNCATE|CREATE)\b|IF NOT EXISTS", block)
    purposes = re.findall(r"'([a-z_]+)'", statements[1])
    assert tuple(purposes) == db_schema.CONFIRM_TOKEN_PURPOSES
    for column in db_schema.CONFIRM_TOKEN_BINDING_COLUMNS:
        assert f"ADD COLUMN {column} TEXT" in statements[0]


def test_compatibility_snapshot_ends_with_exact_token_checks():
    role = db_schema.ROLE_SCHEMA_COMPATIBILITY_CHECKS
    checks = db_schema.TOKEN_COMPLETE_SCHEMA_COMPATIBILITY_CHECKS
    assert db_schema.SCHEMA_COMPATIBILITY_CHECKS[:len(checks)] == checks
    assert checks[:len(role)] == role and checks[len(role):] == TOKEN_CHECKS
    assert not any("chk_confirm_tokens_purpose" in sql or "table_name='confirm_tokens' AND column_name" in sql
                   for _, sql in role)
    assert [label for label, _ in TOKEN_CHECKS] == [
        "confirm_tokens purpose column", "confirm_tokens binding_digest column",
        "confirm_tokens minted_by column", "confirm_tokens purpose check"]
    assert all("is_nullable='YES' AND column_default IS NULL" in sql for _, sql in TOKEN_CHECKS[:3])
    assert "pg_get_constraintdef(oid)='CHECK (((purpose IS NULL) OR (purpose = ANY (ARRAY[" in TOKEN_CHECKS[3][1]
    assert "chk_confirm_tokens_purpose" in db_upgrade.TOKEN_ABSENT_SQL
    assert all(f"'{c}'" in db_upgrade.TOKEN_ABSENT_SQL for c in db_schema.CONFIRM_TOKEN_BINDING_COLUMNS)


def _stubbed(*, token_present=False, token_absent=True):
    calls, applied = [], []
    token_sql = {sql for _, sql in TOKEN_CHECKS}
    def run(sql):
        calls.append(sql)
        if sql.startswith("BEGIN;"):
            applied.append(sql)
            return SimpleNamespace(returncode=0, stdout="", stderr="")
        ok = True
        if sql in token_sql:
            ok = token_present or bool(applied)
        elif sql == db_upgrade.TOKEN_ABSENT_SQL:
            ok = token_absent
        elif sql in (db_upgrade.G4_ABSENT_SQL, db_upgrade.C2A_ABSENT_SQL, db_upgrade.ORG_ABSENT_SQL):
            ok = False  # ROLE-complete: every earlier extension is present
        return SimpleNamespace(returncode=0, stdout="1\n" if ok else "", stderr="")
    return run, calls


def test_role_complete_host_runs_role_token_and_taxonomy_only(capsys):
    run, calls = _stubbed()
    assert db_upgrade.upgrade_retained(SCHEMA, run)
    transactions = [sql for sql in calls if sql.startswith("BEGIN;")]
    assert len(transactions) == 1
    body = transactions[0]
    assert body.index("DELETE FROM system_config") < body.index("ALTER TABLE confirm_tokens") < body.index("COMMIT;")
    assert db_upgrade.TOKEN_ABSENT_SQL in calls[:calls.index(body)]
    for marker in ("CREATE TABLE organizations", "execution_effect_grants", "fk_runtime_team_owner"):
        assert marker not in body
    assert "upgrade complete (ROLE + TOKEN + TAXONOMY + LEDGER)" in capsys.readouterr().out


def test_partial_token_binding_is_refused_before_any_sql():
    run, calls = _stubbed(token_absent=False)
    assert not db_upgrade.upgrade_retained(SCHEMA, run)
    assert not any(sql.startswith("BEGIN;") for sql in calls)


# --- Real PostgreSQL (MYCELIS_SCHEMA_UPGRADE_TEST_DSN, disposable databases only) ---

def install(run, baseline):
    result = run(SCHEMA.read_text().split(BASELINES[baseline])[0] + "COMMIT;\n")
    assert result.returncode == 0, result.stderr
    assert run(LEGACY_ROWS).returncode == 0
    return json.loads(run(TOKEN_ROW).stdout)


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


def insert_token(run, token, purpose):
    return run("INSERT INTO confirm_tokens (token, intent_proof_id, template_id, expires_at, purpose, "
               f"binding_digest, minted_by) VALUES ('{token}', '{PROOF_ID}', 't', NOW(), {purpose}, 'd', 'u');")


def test_real_fresh_install_has_columns_and_enforced_purpose(database):
    result = database(SCHEMA.read_text())
    assert result.returncode == 0, result.stderr
    assert_contract(database)
    assert database(db_upgrade.TOKEN_ABSENT_SQL).stdout.strip() == ""
    assert database(LEGACY_ROWS).returncode == 0  # an insert naming no binding keeps NULLs
    for index, purpose in enumerate(db_schema.CONFIRM_TOKEN_PURPOSES):
        assert insert_token(database, f"44444444-4444-4444-8444-44444444444{index}", f"'{purpose}'").returncode == 0
    for bad in ("'CHAT_ACTION'", "''", "'schedule'"):
        result = insert_token(database, "55555555-5555-4555-8555-555555555555", bad)
        assert result.returncode != 0 and "chk_confirm_tokens_purpose" in result.stderr


@pytest.mark.parametrize("baseline", BASELINES)
def test_real_retained_baselines_converge_with_legacy_rows_null(database, baseline):
    before = install(database, baseline)
    assert db_upgrade.upgrade_retained(SCHEMA, database)
    assert_contract(database)
    after = json.loads(database(TOKEN_ROW).stdout)
    assert {key: after[key] for key in before} == before
    assert all(after[column] is None for column in db_schema.CONFIRM_TOKEN_BINDING_COLUMNS)
    run, calls = recording(database)
    assert db_upgrade.upgrade_retained(SCHEMA, run)  # compatible schema is a no-op
    assert not any(sql.startswith("BEGIN;") for sql in calls)
    assert insert_token(database, "66666666-6666-4666-8666-666666666666", "'nope'").returncode != 0


@pytest.mark.parametrize("partial", [
    "ALTER TABLE confirm_tokens ADD COLUMN purpose TEXT;",
    "ALTER TABLE confirm_tokens ADD COLUMN minted_by TEXT;",
    "ALTER TABLE confirm_tokens ADD CONSTRAINT chk_confirm_tokens_purpose CHECK (true);",
    "ALTER TABLE confirm_tokens ADD COLUMN purpose TEXT, ADD COLUMN binding_digest TEXT, ADD COLUMN minted_by TEXT;",
])
@pytest.mark.parametrize("baseline", ["c2a", "role"])
def test_real_partial_prior_binding_is_refused_without_mutation(database, baseline, partial):
    install(database, baseline)
    assert database(partial).returncode == 0
    snapshot = database(TOKEN_ROW).stdout
    marker = "SELECT count(*) FROM system_config WHERE key='schema.role_seed_retirement';"
    role_marker = database(marker).stdout
    run, calls = recording(database)
    assert not db_upgrade.upgrade_retained(SCHEMA, run)
    assert not any(sql.startswith("BEGIN;") for sql in calls)
    assert database(TOKEN_ROW).stdout == snapshot and database(marker).stdout == role_marker


def test_real_weakened_purpose_check_is_refused(database):
    result = database(SCHEMA.read_text())
    assert result.returncode == 0, result.stderr
    assert database("ALTER TABLE confirm_tokens DROP CONSTRAINT chk_confirm_tokens_purpose;"
                    "ALTER TABLE confirm_tokens ADD CONSTRAINT chk_confirm_tokens_purpose "
                    "CHECK (purpose IS NULL OR purpose IN ('chat_action', 'anything'));").returncode == 0
    run, calls = recording(database)
    assert not db_upgrade.upgrade_retained(SCHEMA, run)
    assert not any(sql.startswith("BEGIN;") for sql in calls)


def test_real_failed_token_block_rolls_back_role_too(database, tmp_path):
    before = install(database, "org")
    broken = tmp_path / "broken.sql"
    broken.write_text(SCHEMA.read_text().replace(db_upgrade.TOKEN_END_MARKER,
                      "SELECT 1 / 0;\n" + db_upgrade.TOKEN_END_MARKER))
    with pytest.raises(SystemExit, match="rolled back"):
        db_upgrade.upgrade_retained(broken, database)
    assert database(db_upgrade.TOKEN_ABSENT_SQL).stdout.strip() == "1"
    assert database("SELECT count(*) FROM system_config WHERE key='schema.role_seed_retirement';").stdout.strip() == "0"
    assert json.loads(database(TOKEN_ROW).stdout) == before


def test_real_compose_migrate_binds_tokens_on_role_complete_stack(database, monkeypatch, capsys):
    before = install(database, "role")
    monkeypatch.setattr(compose, "_compose_effective_env", lambda env_values=None: {})
    monkeypatch.setattr(compose, "_run_compose_psql", lambda sql, env_values: database(sql))
    monkeypatch.setattr(compose, "_run_compose_migration_file",
                        lambda *_: pytest.fail("compose.migrate replayed the installer on retained data"))
    compose._run_compose_migrations()
    assert "Retained-schema upgrade complete (ROLE + TOKEN + TAXONOMY + LEDGER)" in capsys.readouterr().out
    assert_contract(database)
    after = json.loads(database(TOKEN_ROW).stdout)
    assert {key: after[key] for key in before} == before and after["purpose"] is None
    compose._run_compose_migrations()
    assert "already appears compatible" in capsys.readouterr().out
