-- Migration 010: Mission Activation State
-- Adds status tracking and activation timestamp to missions table
-- for the Unified Host Internalization (Phase 6.0).

ALTER TABLE missions
    ADD COLUMN IF NOT EXISTS status TEXT DEFAULT 'draft',
    ADD COLUMN IF NOT EXISTS activated_at TIMESTAMPTZ;

CREATE INDEX IF NOT EXISTS idx_missions_status ON missions(status);
