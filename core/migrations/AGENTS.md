# Core Migrations (`core/migrations`)

> Navigation: [Core AGENTS.md](../AGENTS.md) | [Repo AGENTS.md](../../AGENTS.md)

## Owns / does not own
- `core/migrations/001_current_schema.sql`, the single current-schema baseline, plus the schema-integrity/installer-compatibility tests that pin it (`tests/test_db_*.py`) -> `mycelis-schema` (`.claude/agents/`, tracked roster), the sole writer. No other role edits DDL; request changes through the lead.
- `ops/db*.py` installer behavior is shared with `mycelis-platform-ops` only where installer plumbing (not schema content) is involved.

## Contracts
- Edit the single `001_current_schema.sql` baseline directly; never recreate historical forward/downgrade migration chains (root `AGENTS.md` mirrors this in the schema role file).
- Every change is additive-and-compatible: an empty install, a compatible retained schema (row-preserving no-op), and a partial/incompatible nonempty schema failing before any SQL runs must all stay true in the same change.
- Constraints enforce invariants (FKs with `RESTRICT` where ownership matters, paired nullability checks, partial unique indexes) rather than trusting application code.
- `001_current_schema.sql` is exempt from the line-cap gate as a mechanically generated baseline (`ops/quality.py` `GENERATED_OR_LOCK_PATHS`); do not hand-trim it to satisfy a line count.

## Gates
- Schema integrity tests: `uv run pytest -q tests/test_db_current_schema.py` plus the other `tests/test_db_*.py` files relevant to the change (see `tests/AGENTS.md`).
- A retained additive-upgrade test against the pre-change baseline, and provenance/digest pins in the same change.
- Request the lead run `uv run inv lifecycle.first-boot-proof --isolated --build` and `uv run inv compose.migrate` (or `db.migrate`) for live upgrade proof; schema writers do not hold the shared-stack lease themselves.

## Gotchas
- Schema work is serialized through the lead's first boot; do not assume your change can land independently of the current first-boot queue (`.state/V8_DEV_STATE.md` "Delivery Targets And Teams", the "Budget metering (B1 tail)" row is `QUEUED` for exactly this reason).
- Go PostgreSQL tests elsewhere in `core/` need the current schema applied to their test DSN before they pass; a schema change here can silently break unrelated packages until that DSN is refreshed.
- Never reset or touch retained databases; the only allowed destructive proof is `uv run inv lifecycle.first-boot-proof --isolated`, run by the lead.
