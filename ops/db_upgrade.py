"""Bounded additive upgrades; never replay historical schema on retained data."""
from pathlib import Path

from .db_schema import (
    BASE_SCHEMA_COMPATIBILITY_CHECKS, G4_SCHEMA_COMPATIBILITY_CHECKS,
    SCHEMA_COMPATIBILITY_CHECKS, TEAM_OWNERSHIP_COLUMNS,
)

BEGIN_MARKER = "-- BEGIN G4_E10_EXTENSION"
END_MARKER = "-- END G4_E10_EXTENSION"
C2A_BEGIN_MARKER = "-- BEGIN C2A_TEAM_OWNERSHIP_EXTENSION"
C2A_END_MARKER = "-- END C2A_TEAM_OWNERSHIP_EXTENSION"
G4_ABSENT_SQL = (
    "SELECT 1 WHERE to_regclass('public.execution_effect_grants') IS NULL "
    "AND to_regclass('public.execution_effect_grant_state') IS NULL "
    "AND to_regclass('public.execution_invocations') IS NULL;"
)
_column_names = ",".join("'" + column + "'" for column, _ in TEAM_OWNERSHIP_COLUMNS)
C2A_ABSENT_SQL = (
    "SELECT 1 WHERE NOT EXISTS (SELECT 1 FROM information_schema.columns "
    "WHERE table_schema='public' AND table_name='runtime_team_manifests' "
    f"AND column_name IN ({_column_names})) AND NOT EXISTS "
    "(SELECT 1 FROM pg_constraint WHERE connamespace='public'::regnamespace "
    "AND conname IN ('uq_groups_account_id','fk_runtime_team_owner',"
    "'chk_runtime_team_provisioned','chk_runtime_team_revoked'));"
)


def extension_sql(path: Path, begin=BEGIN_MARKER, end=END_MARKER) -> str:
    text = path.read_text(encoding="utf-8")
    if text.count(begin) != 1 or text.count(end) != 1 or text.index(begin) >= text.index(end):
        raise SystemExit("Canonical schema has no unique ordered upgrade boundary.")
    block = text.split(begin, 1)[1].split(end, 1)[0]
    return "BEGIN;\n" + block + "\nCOMMIT;\n"


def _passes(run_sql, sql: str) -> bool:
    result = run_sql(sql)
    return result.returncode == 0 and "1" in result.stdout.split()


def _compatible(run_sql, checks) -> bool:
    return all(_passes(run_sql, sql) for _, sql in checks)


def upgrade_retained(path: Path, run_sql) -> bool:
    """Upgrade only complete accepted baselines; refuse partial extensions."""
    if not _compatible(run_sql, BASE_SCHEMA_COMPATIBILITY_CHECKS):
        return False
    if _compatible(run_sql, SCHEMA_COMPATIBILITY_CHECKS):
        return True
    g4_present = _compatible(run_sql, G4_SCHEMA_COMPATIBILITY_CHECKS)
    if not g4_present and not _passes(run_sql, G4_ABSENT_SQL):
        return False
    if not _passes(run_sql, C2A_ABSENT_SQL):
        return False
    # Validate both boundaries before any mutation. One transaction covers both
    # missing extensions, so C2a failure cannot leave a partially upgraded host.
    blocks = []
    if not g4_present:
        blocks.append(extension_sql(path)[len("BEGIN;\n"):-len("COMMIT;\n")])
    blocks.append(extension_sql(path, C2A_BEGIN_MARKER, C2A_END_MARKER)[len("BEGIN;\n"):-len("COMMIT;\n")])
    result = run_sql("BEGIN;\n" + "\n".join(blocks) + "COMMIT;\n")
    if result.returncode != 0:
        raise SystemExit("Additive schema upgrade failed; transaction rolled back.")
    if not _compatible(run_sql, SCHEMA_COMPATIBILITY_CHECKS):
        raise SystemExit("Schema upgrade did not satisfy the runtime contract.")
    print("Retained-schema upgrade complete (G4/E10 + C2a); existing rows preserved.")
    return True
