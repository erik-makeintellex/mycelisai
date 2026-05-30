DROP INDEX IF EXISTS trigger_rules_schedule_due;

ALTER TABLE trigger_rules
    DROP COLUMN IF EXISTS recovery_behavior,
    DROP COLUMN IF EXISTS proof_expectations,
    DROP COLUMN IF EXISTS next_run_at,
    DROP COLUMN IF EXISTS schedule_interval_seconds,
    DROP COLUMN IF EXISTS trigger_kind;
