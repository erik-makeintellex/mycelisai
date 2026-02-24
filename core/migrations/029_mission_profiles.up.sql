-- Migration 029: Mission Profiles
-- Named configuration sets that map agent roles to specific providers
-- and define reactive NATS subscriptions. Multiple profiles can be active
-- (auto_start=true profiles always run; at most one non-auto profile is active).

CREATE TABLE IF NOT EXISTS mission_profiles (
    id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name             TEXT NOT NULL,
    description      TEXT,
    role_providers   JSONB NOT NULL DEFAULT '{}',   -- {"architect":"vllm","coder":"ollama"}
    subscriptions    JSONB NOT NULL DEFAULT '[]',   -- [{"topic":"swarm.team.dev-team.*","condition":null}]
    context_strategy TEXT NOT NULL DEFAULT 'fresh', -- "fresh" | "warm" | "snapshot:<uuid>"
    auto_start       BOOLEAN NOT NULL DEFAULT false,
    is_active        BOOLEAN NOT NULL DEFAULT false,
    tenant_id        TEXT NOT NULL DEFAULT 'default',
    created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at       TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS mission_profiles_name_tenant
    ON mission_profiles(name, tenant_id);

CREATE INDEX IF NOT EXISTS mission_profiles_tenant_active
    ON mission_profiles(tenant_id, is_active);

-- Trigger to keep updated_at current
CREATE OR REPLACE FUNCTION update_mission_profiles_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS mission_profiles_updated_at ON mission_profiles;
CREATE TRIGGER mission_profiles_updated_at
    BEFORE UPDATE ON mission_profiles
    FOR EACH ROW EXECUTE FUNCTION update_mission_profiles_updated_at();
