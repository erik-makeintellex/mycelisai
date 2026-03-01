-- V7 Sprint 0: Temporary Working Memory Channels
-- Restart-safe, short-horizon memory for lead agents (admin/council)
-- to preserve in-flight work and interaction continuity across service/provider restarts.

CREATE TABLE IF NOT EXISTS temp_memory_channels (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id       TEXT NOT NULL DEFAULT 'default',
    channel_key     TEXT NOT NULL,                 -- e.g. lead.shared, lead.council-architect, interaction.contract
    owner_agent_id  TEXT NOT NULL,                 -- writer identity
    content         TEXT NOT NULL,                 -- checkpoint payload (markdown/text/json string)
    metadata        JSONB NOT NULL DEFAULT '{}',
    expires_at      TIMESTAMPTZ,                   -- optional TTL
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_temp_memory_channel_tenant_updated
    ON temp_memory_channels(tenant_id, channel_key, updated_at DESC);

CREATE INDEX IF NOT EXISTS idx_temp_memory_expires_at
    ON temp_memory_channels(expires_at);

CREATE OR REPLACE FUNCTION update_temp_memory_channels_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS temp_memory_channels_updated_at ON temp_memory_channels;
CREATE TRIGGER temp_memory_channels_updated_at
    BEFORE UPDATE ON temp_memory_channels
    FOR EACH ROW EXECUTE FUNCTION update_temp_memory_channels_updated_at();

