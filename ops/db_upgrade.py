"""Bounded, additive G4/E10 upgrade of an otherwise compatible retained schema."""
from pathlib import Path

from .db_schema import BASE_SCHEMA_COMPATIBILITY_CHECKS, SCHEMA_COMPATIBILITY_CHECKS

BEGIN_MARKER = "-- BEGIN G4_E10_EXTENSION"
END_MARKER = "-- END G4_E10_EXTENSION"


def extension_sql(path: Path) -> str:
    text = path.read_text(encoding="utf-8")
    if text.count(BEGIN_MARKER) != 1 or text.count(END_MARKER) != 1:
        raise SystemExit("Canonical schema has no unique G4/E10 upgrade boundary.")
    block = text.split(BEGIN_MARKER, 1)[1].split(END_MARKER, 1)[0]
    return "BEGIN;\n" + block + "\nCOMMIT;\n"


def _passes(run_sql, sql: str) -> bool:
    result = run_sql(sql)
    return result.returncode == 0 and "1" in result.stdout.split()


def upgrade_retained(path: Path, run_sql) -> bool:
    """Return False for an unsupported baseline; never replay historical SQL."""
    if not all(_passes(run_sql, sql) for _, sql in BASE_SCHEMA_COMPATIBILITY_CHECKS):
        return False
    # Partial extensions require operator repair; IF NOT EXISTS must not hide drift.
    absent = (
        "SELECT 1 WHERE to_regclass('public.execution_effect_grants') IS NULL "
        "AND to_regclass('public.execution_effect_grant_state') IS NULL "
        "AND to_regclass('public.execution_invocations') IS NULL;"
    )
    if not _passes(run_sql, absent):
        return False
    result = run_sql(extension_sql(path))
    if result.returncode != 0:
        raise SystemExit("G4/E10 additive schema upgrade failed; transaction rolled back.")
    if not all(_passes(run_sql, sql) for _, sql in SCHEMA_COMPATIBILITY_CHECKS):
        raise SystemExit("G4/E10 schema upgrade did not satisfy the runtime contract.")
    print("G4/E10 retained-schema upgrade complete; existing rows preserved.")
    return True
