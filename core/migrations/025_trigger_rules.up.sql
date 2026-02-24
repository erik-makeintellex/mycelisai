-- V7 Team B: Trigger Rules — declarative IF/THEN rules evaluated on event ingest.
-- Depends on: mission_runs (023), mission_events (024).

CREATE TABLE trigger_rules (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id           TEXT NOT NULL DEFAULT 'default',
    name                TEXT NOT NULL,
    description         TEXT,
    event_pattern       TEXT NOT NULL,                       -- e.g. "mission.completed", "tool.completed"
    condition           JSONB NOT NULL DEFAULT '{}',         -- optional filter on event payload
    target_mission_id   TEXT NOT NULL,                       -- mission to launch when rule fires
    mode                TEXT NOT NULL DEFAULT 'propose',     -- "propose" | "auto_execute"
    cooldown_seconds    INT NOT NULL DEFAULT 60,             -- min seconds between consecutive firings
    max_depth           INT NOT NULL DEFAULT 5,              -- recursion guard (max trigger chain depth)
    max_active_runs     INT NOT NULL DEFAULT 3,              -- concurrency guard per target mission
    is_active           BOOLEAN NOT NULL DEFAULT true,
    last_fired_at       TIMESTAMPTZ,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX trigger_rules_tenant_active ON trigger_rules(tenant_id, is_active);
CREATE INDEX trigger_rules_event_pattern ON trigger_rules(event_pattern) WHERE is_active = true;
