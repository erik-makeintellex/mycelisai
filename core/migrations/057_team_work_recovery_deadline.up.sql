-- Migration 057: Queryable recovery deadlines for durable asynchronous work.

ALTER TABLE team_work_items
    ADD COLUMN IF NOT EXISTS recovery_deadline_at TIMESTAMPTZ NULL;

CREATE INDEX IF NOT EXISTS idx_team_work_items_recovery_deadline
    ON team_work_items (recovery_deadline_at)
    WHERE recovery_deadline_at IS NOT NULL;
