"""B1: append-only token usage ledger, pinned block, fresh install and retained upgrade."""
from __future__ import annotations

import hashlib
import re

from ops import db_schema, db_upgrade
from tests.test_db_team_ownership_upgrade import SCHEMA, database  # noqa: F401 (fixture)

LEDGER_BLOCK_SHA256 = "4edbc9d35812b9bf42018f11291d67fda01cbb81769e64940ab2f27c509e88d8"
LEDGER_TABLE = "SELECT 1 WHERE to_regclass('public.token_usage_ledger') IS NOT NULL;"
INSERT = ("INSERT INTO token_usage_ledger (execution_id, execution_kind, provider_id, model_id, budget_class, "
          "total_tokens, usage_reported, outcome) VALUES ('e1', '{kind}', 'p', 'm', 'local_large', 10, true, '{outcome}');")


def _block(raw: bytes) -> bytes:
    begin, end = db_upgrade.LEDGER_BEGIN_MARKER.encode(), db_upgrade.LEDGER_END_MARKER.encode()
    assert raw.count(begin) == 1 and raw.count(end) == 1
    return raw[raw.index(begin) + len(begin):raw.index(end)]


def test_ledger_block_bytes_are_pinned():
    assert hashlib.sha256(_block(SCHEMA.read_bytes())).hexdigest() == LEDGER_BLOCK_SHA256


def test_ledger_block_follows_taxonomy_and_owns_the_single_commit():
    raw = SCHEMA.read_bytes()
    taxonomy_end = raw.index(db_upgrade.TAXONOMY_END_MARKER.encode())
    begin = raw.index(db_upgrade.LEDGER_BEGIN_MARKER.encode())
    end = raw.index(db_upgrade.LEDGER_END_MARKER.encode())
    assert taxonomy_end < begin < end
    assert raw[end:] == db_upgrade.LEDGER_END_MARKER.encode() + b"\n\nCOMMIT;\n"
    assert re.findall(rb"^COMMIT;$", raw, re.MULTILINE) == [b"COMMIT;"]


def test_ledger_block_is_additive_only():
    block = _block(SCHEMA.read_bytes()).decode()
    assert block.count("CREATE TABLE IF NOT EXISTS token_usage_ledger") == 1 and block.count("CREATE INDEX IF NOT EXISTS") == 3
    assert not re.search(r"\b(DROP|ALTER|DELETE|TRUNCATE|UPDATE|INSERT)\b", block)


def test_ledger_checks_extend_the_taxonomy_complete_contract():
    taxonomy = db_schema.TAXONOMY_COMPLETE_SCHEMA_COMPATIBILITY_CHECKS
    checks = db_schema.SCHEMA_COMPATIBILITY_CHECKS
    assert checks[:len(taxonomy)] == taxonomy and checks[len(taxonomy):] == db_schema.LEDGER_SCHEMA_COMPATIBILITY_CHECKS
    labels = " ".join(label for label, _ in db_schema.LEDGER_SCHEMA_COMPATIBILITY_CHECKS)
    for name in ("token_usage_ledger table", "execution_kind", "outcome", "idx_token_ledger_run"):
        assert name in labels


# --- Real PostgreSQL (MYCELIS_SCHEMA_UPGRADE_TEST_DSN, disposable databases only) ---

def _assert_contract(run):
    for label, sql in db_schema.SCHEMA_COMPATIBILITY_CHECKS:
        result = run(sql)
        assert result.returncode == 0 and result.stdout.strip() == "1", (label, result.stderr)


def test_real_fresh_install_has_the_ledger_and_rejects_unknown_values(database):
    result = database(SCHEMA.read_text())
    assert result.returncode == 0, result.stderr
    _assert_contract(database)
    assert database(INSERT.format(kind="agent_turn", outcome="charged")).returncode == 0
    assert database(INSERT.format(kind="unbounded", outcome="charged")).returncode != 0
    assert database(INSERT.format(kind="draft", outcome="completed")).returncode != 0


def test_real_retained_upgrade_adds_the_ledger_once(database, capsys):
    result = database(SCHEMA.read_text().split(db_upgrade.LEDGER_BEGIN_MARKER)[0] + "COMMIT;\n")
    assert result.returncode == 0, result.stderr
    assert database(LEDGER_TABLE).stdout.strip() == ""
    assert db_upgrade.upgrade_retained(SCHEMA, database)
    assert "upgrade complete (ROLE + LEDGER)" in capsys.readouterr().out
    _assert_contract(database)
    assert db_upgrade.upgrade_retained(SCHEMA, database)
    assert "upgrade complete" not in capsys.readouterr().out


def test_real_partial_ledger_is_refused_without_mutation(database):
    result = database(SCHEMA.read_text().split(db_upgrade.LEDGER_BEGIN_MARKER)[0] + "COMMIT;\n")
    assert result.returncode == 0, result.stderr
    assert database("CREATE TABLE token_usage_ledger (id BIGSERIAL PRIMARY KEY);").returncode == 0
    assert not db_upgrade.upgrade_retained(SCHEMA, database)
    columns = database("SELECT count(*) FROM information_schema.columns WHERE table_name='token_usage_ledger';")
    assert columns.stdout.strip() == "1"
