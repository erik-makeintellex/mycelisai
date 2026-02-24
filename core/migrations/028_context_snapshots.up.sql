-- Migration 028: Context Snapshots
-- Stores point-in-time snapshots of conversation context + provider assignments.
-- Used by mission profiles to cache and restore context when switching profiles.

CREATE TABLE IF NOT EXISTS context_snapshots (
    id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name           TEXT NOT NULL,
    description    TEXT,
    messages       JSONB NOT NULL DEFAULT '[]',   -- recent chat history (ChatMessage[])
    run_state      JSONB NOT NULL DEFAULT '{}',   -- active run_id, status, etc.
    role_providers JSONB NOT NULL DEFAULT '{}',   -- provider assignments at snapshot time {"architect":"ollama"}
    source_profile TEXT,                           -- profile ID that was active when snapshot taken
    tenant_id      TEXT NOT NULL DEFAULT 'default',
    created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS context_snapshots_tenant_created
    ON context_snapshots(tenant_id, created_at DESC);
