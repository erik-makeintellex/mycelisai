"""Bounded additive upgrades; never replay historical schema on retained data."""
from pathlib import Path

from .db_schema import (
    BASE_SCHEMA_COMPATIBILITY_CHECKS, C2A_SCHEMA_COMPATIBILITY_CHECKS,
    G4_SCHEMA_COMPATIBILITY_CHECKS, ORG_SCHEMA_COMPATIBILITY_CHECKS,
    SCHEMA_COMPATIBILITY_CHECKS, TEAM_OWNERSHIP_COLUMNS, TOKEN_SCHEMA_COMPATIBILITY_CHECKS,
    CONFIRM_TOKEN_BINDING_COLUMNS,
)

BEGIN_MARKER = "-- BEGIN G4_E10_EXTENSION"
END_MARKER = "-- END G4_E10_EXTENSION"
C2A_BEGIN_MARKER = "-- BEGIN C2A_TEAM_OWNERSHIP_EXTENSION"
C2A_END_MARKER = "-- END C2A_TEAM_OWNERSHIP_EXTENSION"
ORG_BEGIN_MARKER = "-- BEGIN ORGANIZATIONS_EXTENSION"
ORG_END_MARKER = "-- END ORGANIZATIONS_EXTENSION"
ROLE_BEGIN_MARKER = "-- BEGIN ROLE_SEED_RETIREMENT_EXTENSION"
ROLE_END_MARKER = "-- END ROLE_SEED_RETIREMENT_EXTENSION"
TOKEN_BEGIN_MARKER = "-- BEGIN CONFIRM_TOKEN_BINDING_EXTENSION"
TOKEN_END_MARKER = "-- END CONFIRM_TOKEN_BINDING_EXTENSION"
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
# Any prior organizations relation (partial, stale, or foreign) is refused.
ORG_ABSENT_SQL = "SELECT 1 WHERE to_regclass('public.organizations') IS NULL;"
# The ROLE DELETE only needs the 006 system_config table; it never touches other rows.
ROLE_TABLE_SQL = "SELECT 1 WHERE to_regclass('public.system_config') IS NOT NULL;"
_token_columns = ",".join("'" + column + "'" for column in CONFIRM_TOKEN_BINDING_COLUMNS)
# Any prior binding column or purpose constraint (partial or foreign) is refused.
TOKEN_ABSENT_SQL = (
    "SELECT 1 WHERE NOT EXISTS (SELECT 1 FROM information_schema.columns "
    "WHERE table_schema='public' AND table_name='confirm_tokens' "
    f"AND column_name IN ({_token_columns})) AND NOT EXISTS (SELECT 1 FROM pg_constraint "
    "WHERE conrelid=to_regclass('public.confirm_tokens') AND conname='chk_confirm_tokens_purpose');"
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


def _extension_blocks(path: Path) -> dict[str, str]:
    """Parse every boundary, in canonical order, before any SQL runs."""
    text = path.read_text(encoding="utf-8")
    markers = (("G4/E10", BEGIN_MARKER, END_MARKER), ("C2a", C2A_BEGIN_MARKER, C2A_END_MARKER),
               ("ORG", ORG_BEGIN_MARKER, ORG_END_MARKER), ("ROLE", ROLE_BEGIN_MARKER, ROLE_END_MARKER),
               ("TOKEN", TOKEN_BEGIN_MARKER, TOKEN_END_MARKER))
    positions = [text.find(marker) for _, begin, end in markers for marker in (begin, end)]
    if -1 in positions or positions != sorted(positions):
        raise SystemExit("Canonical schema has no unique ordered upgrade boundary.")
    return {name: extension_sql(path, begin, end)[len("BEGIN;\n"):-len("COMMIT;\n")]
            for name, begin, end in markers}


def upgrade_retained(path: Path, run_sql) -> bool:
    """Upgrade pre-G4, G4, C2a, ORG or ROLE-complete baselines; refuse partial extensions."""
    if not _compatible(run_sql, BASE_SCHEMA_COMPATIBILITY_CHECKS):
        return False
    if _compatible(run_sql, SCHEMA_COMPATIBILITY_CHECKS):
        return True
    g4_present = _compatible(run_sql, G4_SCHEMA_COMPATIBILITY_CHECKS)
    if not g4_present and not _passes(run_sql, G4_ABSENT_SQL):
        return False
    c2a_present = g4_present and _compatible(run_sql, C2A_SCHEMA_COMPATIBILITY_CHECKS)
    if not c2a_present and not _passes(run_sql, C2A_ABSENT_SQL):
        return False
    org_present = c2a_present and _compatible(run_sql, ORG_SCHEMA_COMPATIBILITY_CHECKS)
    if not org_present and not _passes(run_sql, ORG_ABSENT_SQL):
        return False
    if not _passes(run_sql, ROLE_TABLE_SQL):
        return False
    token_present = org_present and _compatible(run_sql, TOKEN_SCHEMA_COMPATIBILITY_CHECKS)
    if not token_present and not _passes(run_sql, TOKEN_ABSENT_SQL):
        return False
    # Every boundary is validated before any mutation. One transaction covers
    # all missing extensions, so a later failure cannot leave a partial host.
    blocks = _extension_blocks(path)
    # ROLE is an idempotent exact-tuple DELETE, so it always joins the transaction.
    missing = [name for name, present in (("G4/E10", g4_present), ("C2a", c2a_present),
                                          ("ORG", org_present), ("ROLE", False),
                                          ("TOKEN", token_present)) if not present]
    result = run_sql("BEGIN;\n" + "\n".join(blocks[name] for name in missing) + "COMMIT;\n")
    if result.returncode != 0:
        raise SystemExit("Additive schema upgrade failed; transaction rolled back.")
    if not _compatible(run_sql, SCHEMA_COMPATIBILITY_CHECKS):
        raise SystemExit("Schema upgrade did not satisfy the runtime contract.")
    print(f"Retained-schema upgrade complete ({' + '.join(missing)}); existing rows preserved except exact legacy role seeds.")
    return True
