from __future__ import annotations

import hashlib
import re
from pathlib import Path


MIGRATIONS_DIR = Path(__file__).parents[1] / "core" / "migrations"
BASELINE = MIGRATIONS_DIR / "001_current_schema.sql"
SOURCE_REVISION = "773e534c09710792607d243bc5f56aab85fa5ecc"
SOURCE_MANIFEST_SHA256 = "16b6caab761aa22668bbf12133ed980a301cdd810f469fe497bb8d84807a2cdc"
BASELINE_SHA256 = "9991a949be50b5b4c65f010f168005acc10438e2aced9d47e3a510f9f53c1ed9"
SOURCE_COUNT = 63
SOURCE_BYTES = 97_915
TRANSACTION_WRAPPER = (b"BEGIN;\n", b"COMMIT;\n")
BEGIN_SOURCE = re.compile(
    rb"^-- BEGIN SOURCE: (?P<path>core/migrations/[^ ]+) "
    rb"blob=(?P<blob>[0-9a-f]{40}) source_bytes=(?P<source_size>[0-9]+) "
    rb"source_sha256=(?P<source_sha256>[0-9a-f]{64}) "
    rb"rendered_bytes=(?P<rendered_size>[0-9]+) "
    rb"rendered_sha256=(?P<rendered_sha256>[0-9a-f]{64}) "
    rb"transform=(?P<transform>none|strip_outer_transaction|strip_trailing_whitespace)\n",
    re.MULTILINE,
)


def _source_blocks(raw: bytes) -> list[tuple[str, str, int, str, str, bytes]]:
    blocks: list[tuple[str, str, int, str, str, bytes]] = []
    for match in BEGIN_SOURCE.finditer(raw):
        path = match.group("path").decode()
        blob = match.group("blob").decode()
        source_size = int(match.group("source_size"))
        source_sha256 = match.group("source_sha256").decode()
        rendered_size = int(match.group("rendered_size"))
        rendered_sha256 = match.group("rendered_sha256").decode()
        transform = match.group("transform").decode()
        content_start = match.end()
        content_end = content_start + rendered_size
        rendered = raw[content_start:content_end]
        assert len(rendered) == rendered_size, f"truncated source block: {path}"
        assert hashlib.sha256(rendered).hexdigest() == rendered_sha256

        source = rendered
        if transform == "strip_outer_transaction":
            assert path == "core/migrations/063_runtime_team_manifests.up.sql"
            assert re.findall(rb"^(?:BEGIN|COMMIT);$", rendered, re.MULTILINE) == []
            source = TRANSACTION_WRAPPER[0] + rendered + TRANSACTION_WRAPPER[1]
            assert source_size - len(rendered) == sum(map(len, TRANSACTION_WRAPPER))
            assert source[len(TRANSACTION_WRAPPER[0]) : -len(TRANSACTION_WRAPPER[1])] == rendered
        elif transform == "strip_trailing_whitespace":
            replacements = {
                "core/migrations/003_mission_hierarchy.up.sql": (
                    b"ALTER TABLE teams\n",
                    b"ALTER TABLE teams \n",
                ),
                "core/migrations/007_team_fabric.up.sql": (
                    b"-- Let's add a `directives` JSONB column to teams?\n",
                    b"-- Let's add a `directives` JSONB column to teams? \n",
                ),
            }
            assert path in replacements
            rendered_line, source_line = replacements[path]
            assert rendered.count(rendered_line) == 1
            source = rendered.replace(rendered_line, source_line)
            assert len(source) - len(rendered) == 1
            assert not re.search(rb"[ \t]+$", rendered, re.MULTILINE)
        assert len(source) == source_size
        assert hashlib.sha256(source).hexdigest() == source_sha256

        marker_start = content_end if rendered.endswith(b"\n") else content_end + 1
        if marker_start != content_end:
            assert raw[content_end:marker_start] == b"\n", f"missing separator newline: {path}"
        end_marker = f"-- END SOURCE: {path}\n".encode()
        assert raw[marker_start : marker_start + len(end_marker)] == end_marker
        blocks.append((path, blob, source_size, source_sha256, transform, source))
    return blocks


def test_current_schema_is_the_only_installable_sql_file():
    assert sorted(path.name for path in MIGRATIONS_DIR.glob("*.sql")) == [BASELINE.name]


def test_current_schema_matches_immutable_dev_manifest():
    raw = BASELINE.read_bytes()
    assert hashlib.sha256(raw).hexdigest() == BASELINE_SHA256
    assert f"-- Source revision: {SOURCE_REVISION}\n".encode() in raw
    assert f"-- Source manifest SHA-256: {SOURCE_MANIFEST_SHA256}\n".encode() in raw

    blocks = _source_blocks(raw)
    assert len(blocks) == SOURCE_COUNT
    assert sum(block[2] for block in blocks) == SOURCE_BYTES
    assert blocks[0][0] == "core/migrations/001_init_memory.sql"
    assert blocks[-1][0] == "core/migrations/064_code_context.up.sql"
    assert len({block[0] for block in blocks}) == SOURCE_COUNT
    assert len({block[1] for block in blocks}) == SOURCE_COUNT
    transformed = [block for block in blocks if block[4] != "none"]
    assert [(block[0], block[4]) for block in transformed] == [
        ("core/migrations/003_mission_hierarchy.up.sql", "strip_trailing_whitespace"),
        ("core/migrations/007_team_fabric.up.sql", "strip_trailing_whitespace"),
        ("core/migrations/063_runtime_team_manifests.up.sql", "strip_outer_transaction")
    ]
    assert re.findall(rb"^(?:BEGIN|COMMIT);$", transformed[-1][5], re.MULTILINE) == [
        b"BEGIN;",
        b"COMMIT;",
    ]
    assert {block[0] for block in blocks if not block[5].endswith(b"\n")} == {
        "core/migrations/006_cognitive_registry.up.sql",
        "core/migrations/008_context_engine.up.sql",
        "core/migrations/011_fix_nodes_type_schema.up.sql",
        "core/migrations/012_fix_nodes_specs_schema.up.sql",
        "core/migrations/013_fix_provider_ip.up.sql",
        "core/migrations/014_fix_provider_url_suffix.up.sql",
        "core/migrations/015_fix_provider_ipv4.up.sql",
    }
    transaction_sources = {
        block[0]
        for block in blocks
        if re.search(rb"^(BEGIN|COMMIT);$", block[5], re.MULTILINE)
    }
    assert transaction_sources == {
        "core/migrations/063_runtime_team_manifests.up.sql"
    }

    manifest = "".join(
        f"{path}\t{blob}\t{size}\t{sha256}\n"
        for path, blob, size, sha256, _transform, _content in blocks
    ).encode()
    assert hashlib.sha256(manifest).hexdigest() == SOURCE_MANIFEST_SHA256


def test_current_schema_stops_at_code_context_064():
    raw = BASELINE.read_bytes()
    assert re.findall(rb"^BEGIN;$", raw, re.MULTILINE) == [b"BEGIN;"]
    assert re.findall(rb"^COMMIT;$", raw, re.MULTILINE) == [b"COMMIT;"]
    assert b"065_" not in raw
    assert b"outcome_collaboration" not in raw
    for required in (
        b"CREATE EXTENSION IF NOT EXISTS vector",
        b"CREATE TABLE IF NOT EXISTS execution_contracts",
        b"CREATE TABLE IF NOT EXISTS proof_artifacts",
        b"CREATE TABLE IF NOT EXISTS runtime_team_manifests",
        b"CREATE TABLE IF NOT EXISTS code_context_sources",
        b"CREATE TABLE IF NOT EXISTS code_context_snapshots",
        b"CREATE TABLE IF NOT EXISTS code_context_files",
        b"CREATE TABLE IF NOT EXISTS code_context_symbols",
        b"CREATE TABLE IF NOT EXISTS code_context_edges",
    ):
        assert required in raw


def test_current_schema_has_framework_worker_authority_invariants():
    schema = BASELINE.read_text(encoding="utf-8")
    for table in (
        "worker_run_bindings",
        "worker_event_receipts",
        "worker_approval_requests",
        "worker_control_commands",
    ):
        assert f"CREATE TABLE IF NOT EXISTS {table}" in schema

    binding = schema.split("CREATE TABLE IF NOT EXISTS worker_run_bindings", 1)[1]
    binding = binding.split("CREATE TABLE IF NOT EXISTS worker_event_receipts", 1)[0]
    assert "run_id UUID PRIMARY KEY REFERENCES mission_runs(id)" in binding
    assert "backend_run_id" not in binding
    assert "backend = 'framework_runs'" in binding
    assert "protocol = 'runs_api'" in binding
    assert "last_event_sequence BIGINT NOT NULL DEFAULT 0" in binding
    assert "cursor_version BIGINT NOT NULL DEFAULT 0" in binding
    assert "request_digest ~ '^[0-9a-f]{64}$'" in binding

    receipts = schema.split("CREATE TABLE IF NOT EXISTS worker_event_receipts", 1)[1]
    receipts = receipts.split("CREATE TABLE IF NOT EXISTS worker_approval_requests", 1)[0]
    assert "UNIQUE (run_id, event_id)" in receipts
    assert "UNIQUE (run_id, sequence)" in receipts
    assert "sequence > 0 AND service_version >= 1" in receipts

    approvals = schema.split("CREATE TABLE IF NOT EXISTS worker_approval_requests", 1)[1]
    approvals = approvals.split("CREATE TABLE IF NOT EXISTS worker_control_commands", 1)[0]
    assert "state IN ('pending', 'decided', 'expired', 'withdrawn')" in approvals
    assert "decision IS NULL OR decision IN ('approve', 'deny')" in approvals
    assert "state = 'decided' AND decision IS NOT NULL AND decided_by IS NOT NULL" in approvals

    commands = schema.split("CREATE TABLE IF NOT EXISTS worker_control_commands", 1)[1]
    assert "kind IN ('approve', 'deny', 'stop')" in commands
    assert "state IN ('staged', 'pending', 'acknowledged', 'failed', 'uncertain')" in commands
