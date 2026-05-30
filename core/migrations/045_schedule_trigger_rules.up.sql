ALTER TABLE trigger_rules
    ADD COLUMN IF NOT EXISTS trigger_kind TEXT NOT NULL DEFAULT 'event',
    ADD COLUMN IF NOT EXISTS schedule_interval_seconds INT,
    ADD COLUMN IF NOT EXISTS next_run_at TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS proof_expectations TEXT,
    ADD COLUMN IF NOT EXISTS recovery_behavior TEXT;

CREATE INDEX IF NOT EXISTS trigger_rules_schedule_due
    ON trigger_rules(tenant_id, is_active, next_run_at)
    WHERE trigger_kind = 'schedule';
