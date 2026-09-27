"""M1: one-shot governed-context taxonomy rewrite (diary -> worklog), fresh and retained."""
from __future__ import annotations

import json
import re

from ops import db_schema, db_upgrade
from tests.test_db_team_ownership_upgrade import SCHEMA, database  # noqa: F401 (fixture)

MARKER = db_schema.DEPLOYMENT_CONTEXT_TAXONOMY_MARKER
MARKER_ROW = f"SELECT value FROM system_config WHERE key='{MARKER}';"
GOVERNED = {"knowledge_store": "governed_context_store", "source_kind": "diary_entry",
            "reflection_kind": "diary_entry", "content_domain": "diary",
            "tags": ["diary", "ops", "diary_entry", "worklog"]}
FOREIGN = {"source_kind": "diary_entry", "content_domain": "diary", "tags": ["diary"]}


def _block(raw: bytes) -> str:
    begin, end = db_upgrade.TAXONOMY_BEGIN_MARKER.encode(), db_upgrade.TAXONOMY_END_MARKER.encode()
    assert raw.count(begin) == 1 and raw.count(end) == 1
    return raw[raw.index(begin) + len(begin):raw.index(end)].decode()


def test_taxonomy_block_follows_token_and_precedes_the_ledger():
    raw = SCHEMA.read_bytes()
    token_end = raw.index(db_upgrade.TOKEN_END_MARKER.encode())
    begin = raw.index(db_upgrade.TAXONOMY_BEGIN_MARKER.encode())
    end = raw.index(db_upgrade.TAXONOMY_END_MARKER.encode())
    assert token_end < begin < end
    assert raw[end:].startswith(db_upgrade.TAXONOMY_END_MARKER.encode() + b"\n\n" + db_upgrade.LEDGER_BEGIN_MARKER.encode())


def test_taxonomy_block_is_marker_gated_governed_only_and_additive():
    block = _block(SCHEMA.read_bytes())
    assert block.count("UPDATE artifacts SET metadata") == 1 and block.count("UPDATE context_vectors SET metadata") == 1
    assert block.count(f"AND NOT EXISTS (SELECT 1 FROM system_config WHERE key = '{MARKER}');") == 2
    assert block.count("COALESCE(metadata->>'knowledge_store', '') = 'governed_context_store'") == 2
    assert block.rstrip().endswith(f"INSERT INTO system_config (key, value) VALUES ('{MARKER}', '1')\n"
                                   "    ON CONFLICT (key) DO NOTHING;")
    assert not re.search(r"\b(DROP|ALTER|DELETE|TRUNCATE|CREATE)\b", block)
    assert db_schema.DEPLOYMENT_CONTEXT_TAXONOMY_RENAMES == (("diary_entry", "worklog_entry"), ("diary", "worklog"))


def test_taxonomy_checks_extend_the_token_complete_contract():
    token = db_schema.TOKEN_COMPLETE_SCHEMA_COMPATIBILITY_CHECKS
    checks = db_schema.SCHEMA_COMPATIBILITY_CHECKS
    taxonomy_end = len(token) + len(db_schema.TAXONOMY_SCHEMA_COMPATIBILITY_CHECKS)
    assert checks[:len(token)] == token and checks[len(token):taxonomy_end] == db_schema.TAXONOMY_SCHEMA_COMPATIBILITY_CHECKS
    assert checks[:taxonomy_end] == db_schema.TAXONOMY_COMPLETE_SCHEMA_COMPATIBILITY_CHECKS
    assert any("is_nullable='YES'" in sql for _, sql in db_schema.TAXONOMY_SCHEMA_COMPATIBILITY_CHECKS)


# --- Real PostgreSQL (MYCELIS_SCHEMA_UPGRADE_TEST_DSN, disposable databases only) ---

def _seed(run):
    rows = []
    for table, meta in (("artifacts", GOVERNED), ("artifacts", FOREIGN), ("context_vectors", GOVERNED), ("context_vectors", FOREIGN)):
        literal = json.dumps(meta).replace("'", "''")
        if table == "artifacts":
            sql = ("INSERT INTO artifacts (agent_id, artifact_type, title, content, metadata, status) "
                   f"VALUES ('soma','document','t','x','{literal}'::jsonb,'approved') RETURNING id;")
        else:
            sql = f"INSERT INTO context_vectors (content, embedding, metadata) VALUES ('x', NULL, '{literal}'::jsonb) RETURNING id;"
        result = run(sql)
        assert result.returncode == 0, result.stderr
        rows.append((table, result.stdout.strip().splitlines()[0], meta))
    return rows


def _meta(run, table, row_id):
    return json.loads(run(f"SELECT metadata::text FROM {table} WHERE id='{row_id}';").stdout.strip())


def _assert_contract(run):
    for label, sql in db_schema.SCHEMA_COMPATIBILITY_CHECKS:
        result = run(sql)
        assert result.returncode == 0 and result.stdout.strip() == "1", (label, result.stderr)


def test_real_fresh_install_carries_the_marker(database):
    result = database(SCHEMA.read_text())
    assert result.returncode == 0, result.stderr
    assert database(MARKER_ROW).stdout.strip() == "1"
    _assert_contract(database)


def test_real_retained_upgrade_rewrites_governed_rows_once(database, capsys):
    result = database(SCHEMA.read_text().split(db_upgrade.TAXONOMY_BEGIN_MARKER)[0] + "COMMIT;\n")
    assert result.returncode == 0, result.stderr
    rows = _seed(database)
    assert database(MARKER_ROW).stdout.strip() == ""
    assert db_upgrade.upgrade_retained(SCHEMA, database)
    assert "upgrade complete (ROLE + TAXONOMY + LEDGER)" in capsys.readouterr().out
    assert database(MARKER_ROW).stdout.strip() == "1"
    for table, row_id, seeded in rows:
        meta = _meta(database, table, row_id)
        if seeded is FOREIGN:
            assert meta == FOREIGN, (table, meta)
            continue
        assert meta["source_kind"] == "worklog_entry" and meta["reflection_kind"] == "worklog_entry"
        assert meta["content_domain"] == "worklog" and meta["tags"] == ["worklog", "ops", "worklog_entry"]
    _assert_contract(database)

    # A later run of the block is a no-op: the marker gates every rewrite.
    late = _seed(database)
    block = SCHEMA.read_text().split(db_upgrade.TAXONOMY_BEGIN_MARKER)[1].split(db_upgrade.TAXONOMY_END_MARKER)[0]
    result = database("BEGIN;\n" + block + "COMMIT;\n")
    assert result.returncode == 0, result.stderr
    for table, row_id, seeded in late:
        assert _meta(database, table, row_id) == seeded
    assert db_upgrade.upgrade_retained(SCHEMA, database)
