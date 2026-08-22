-- Migration 052: Preserve approved WorkIntent lifecycle on durable team work.

ALTER TABLE team_work_items
    ADD COLUMN IF NOT EXISTS execution_mode TEXT NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS work_intent JSONB NULL;

ALTER TABLE team_status_events
    ADD COLUMN IF NOT EXISTS execution_mode TEXT NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS work_intent JSONB NULL;
