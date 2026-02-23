-- Migration 023: Mission Runs (V7 Event Spine — Team A)
--
-- Every mission execution is a "run". A mission definition may have many runs.
-- Runs form a chain: parent → trigger → child (for V7 Trigger Engine, Team B).
-- tenant_id = 'default' for V7 (schema-ready, operationally single-tenant).

CREATE TABLE IF NOT EXISTS mission_runs (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    mission_id    VARCHAR(255) NOT NULL,                            -- references missions.id (label or UUID)
    tenant_id     VARCHAR(255) NOT NULL DEFAULT 'default',
    status        VARCHAR(50)  NOT NULL DEFAULT 'pending',          -- pending|running|completed|failed
    run_depth     INT          NOT NULL DEFAULT 0,                  -- recursion depth for trigger chains
    parent_run_id UUID         REFERENCES mission_runs(id),        -- set by Trigger Engine (Team B)
    trigger_rule_id UUID,                                           -- set by Trigger Engine (Team B)
    started_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    completed_at  TIMESTAMPTZ,
    metadata      JSONB
);

CREATE INDEX IF NOT EXISTS idx_mission_runs_mission   ON mission_runs(mission_id);
CREATE INDEX IF NOT EXISTS idx_mission_runs_status    ON mission_runs(status);
CREATE INDEX IF NOT EXISTS idx_mission_runs_started   ON mission_runs(started_at DESC);
CREATE INDEX IF NOT EXISTS idx_mission_runs_parent    ON mission_runs(parent_run_id)
    WHERE parent_run_id IS NOT NULL;
