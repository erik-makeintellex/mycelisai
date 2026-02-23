-- Migration 024: Mission Events (V7 Event Spine — Team A)
--
-- The authoritative audit trail for every action in a mission run.
-- Dual-layer design: this table = persistent record; CTS = real-time transport.
-- DB-first rule: Emit() persists here BEFORE publishing the CTS signal.
-- If NATS is offline, events still persist here (degraded mode, no data loss).

CREATE TABLE IF NOT EXISTS mission_events (
    id            UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    run_id        UUID         NOT NULL REFERENCES mission_runs(id) ON DELETE CASCADE,
    tenant_id     VARCHAR(255) NOT NULL DEFAULT 'default',
    event_type    VARCHAR(100) NOT NULL,   -- tool.invoked, mission.started, artifact.created, etc.
    severity      VARCHAR(20)  NOT NULL DEFAULT 'info',  -- debug|info|warn|error
    source_agent  VARCHAR(255),
    source_team   VARCHAR(255),
    payload       JSONB,
    audit_event_id UUID,                  -- links to log_entries(id) for CE-1 audit events
    emitted_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_mission_events_run      ON mission_events(run_id, emitted_at);
CREATE INDEX IF NOT EXISTS idx_mission_events_type     ON mission_events(event_type);
CREATE INDEX IF NOT EXISTS idx_mission_events_agent    ON mission_events(source_agent)
    WHERE source_agent IS NOT NULL;
