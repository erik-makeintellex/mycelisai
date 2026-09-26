"""Provenance and contract pins for the ORGANIZATIONS_EXTENSION block in 001."""
from __future__ import annotations

import hashlib
import re
from pathlib import Path

from ops import db_schema, db_upgrade, lifecycle_first_boot

BASELINE = Path(__file__).parents[1] / "core" / "migrations" / "001_current_schema.sql"
ORG_BLOCK_SHA256 = "e65bcb3f28ba22357dcb3cb63f581773051793f2661d238145fc248ddffaa5a5"


def _org_block(raw: bytes) -> bytes:
    begin, end = db_upgrade.ORG_BEGIN_MARKER.encode(), db_upgrade.ORG_END_MARKER.encode()
    assert raw.count(begin) == 1 and raw.count(end) == 1
    return raw[raw.index(begin) + len(begin):raw.index(end)]


def test_org_block_bytes_are_pinned():
    assert hashlib.sha256(_org_block(BASELINE.read_bytes())).hexdigest() == ORG_BLOCK_SHA256


def test_org_block_follows_c2a_and_precedes_the_single_commit():
    raw = BASELINE.read_bytes()
    c2a_end = raw.index(db_upgrade.C2A_END_MARKER.encode())
    org_begin = raw.index(db_upgrade.ORG_BEGIN_MARKER.encode())
    org_end = raw.index(db_upgrade.ORG_END_MARKER.encode())
    commits = [m.start() for m in re.finditer(rb"^COMMIT;$", raw, re.MULTILINE)]
    assert len(commits) == 1
    assert c2a_end < org_begin < org_end < commits[0]
    # The ROLE seed-retirement block is the only content between ORG and COMMIT.
    assert raw[org_end:].startswith(db_upgrade.ORG_END_MARKER.encode() + b"\n\n" + db_upgrade.ROLE_BEGIN_MARKER.encode())


def test_org_block_is_plain_create_with_authoritative_fixture_scope():
    block = _org_block(BASELINE.read_bytes()).decode()
    assert "IF NOT EXISTS" not in block
    assert block.count("CREATE TABLE organizations (") == 1
    for column in ("id UUID PRIMARY KEY", "tenant_id TEXT NOT NULL DEFAULT 'default'",
                   "name TEXT NOT NULL", "qa_fixture_scope_id TEXT NOT NULL DEFAULT ''",
                   "document JSONB NOT NULL",
                   "CONSTRAINT chk_organizations_document_object CHECK (jsonb_typeof(document) = 'object')",
                   "CREATE INDEX idx_organizations_tenant_name ON organizations(tenant_id, name, id)"):
        assert column in block
    # No identity/authority coupling in S1: organizations stay out of A3/A4 tables.
    assert not re.search(r"REFERENCES\s+(accounts|groups|memberships|users)", block)
    assert not re.search(r"^\s*(DROP|ALTER|DELETE|UPDATE|INSERT|TRUNCATE)\b", block, re.MULTILINE)


def test_compatibility_snapshots_bound_the_org_extension():
    labels = [label for label, _ in db_schema.SCHEMA_COMPATIBILITY_CHECKS]
    c2a = db_schema.C2A_SCHEMA_COMPATIBILITY_CHECKS
    assert db_schema.SCHEMA_COMPATIBILITY_CHECKS[:len(c2a)] == c2a
    assert db_schema.G4_SCHEMA_COMPATIBILITY_CHECKS == c2a[:len(db_schema.G4_SCHEMA_COMPATIBILITY_CHECKS)]
    assert not any("organizations" in sql for _, sql in c2a)
    full_org = db_schema.ORG_SCHEMA_COMPATIBILITY_CHECKS
    assert db_schema.SCHEMA_COMPATIBILITY_CHECKS[:len(full_org)] == full_org
    org = [sql for _, sql in full_org[len(c2a):]]
    assert labels[len(c2a):len(full_org)] == ["organizations table", "organizations not-null columns",
                                 "organizations fixture scope default", "organizations primary key",
                                 "organizations document object", "organizations tenant listing index"]
    assert all("organizations" in sql for sql in org)
    assert "qa_fixture_scope_id" in org[1] and "is_nullable='NO'" in org[1]
    # Same-name weakened constraints must not masquerade as the accepted contract.
    assert "pg_get_constraintdef(oid)='CHECK ((jsonb_typeof(document) = ''object''::text))'" in org[4]
    assert db_upgrade.ORG_ABSENT_SQL == "SELECT 1 WHERE to_regclass('public.organizations') IS NULL;"


def test_clean_first_boot_requires_empty_organizations():
    assert "organizations" in lifecycle_first_boot.CLEAN_FIRST_BOOT_USER_TABLES
    assert "organizations" not in lifecycle_first_boot.CLEAN_FIRST_BOOT_BOOTSTRAP_TABLES
