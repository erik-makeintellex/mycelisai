-- 018: Artifacts — structured agent output persistence
-- Every agent output (code, documents, media, data) is stored here.
-- Links to mission → team → agent provenance chain.
-- Content can be inline (text/code/json) or file-referenced (/data/artifacts/).

CREATE TABLE IF NOT EXISTS artifacts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    mission_id UUID REFERENCES missions(id) ON DELETE SET NULL,
    team_id UUID REFERENCES teams(id) ON DELETE SET NULL,
    agent_id TEXT NOT NULL,                          -- agent manifest ID (not UUID — matches AgentManifest.ID)
    trace_id TEXT,                                   -- CTS correlation ID
    artifact_type TEXT NOT NULL,                     -- code, document, image, audio, data, file, chart
    title TEXT NOT NULL,
    content_type TEXT NOT NULL DEFAULT 'text/plain', -- MIME type
    content TEXT,                                    -- inline content (code, markdown, json, small text)
    file_path TEXT,                                  -- relative path under /data/artifacts/ (for binary/large files)
    file_size_bytes BIGINT,
    metadata JSONB DEFAULT '{}',                     -- type-specific: {language: "go", filename: "main.go"} etc.
    trust_score DOUBLE PRECISION,                    -- inherited from CTSEnvelope
    status TEXT NOT NULL DEFAULT 'pending',          -- pending, approved, rejected, archived
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_artifacts_mission ON artifacts(mission_id);
CREATE INDEX IF NOT EXISTS idx_artifacts_team ON artifacts(team_id);
CREATE INDEX IF NOT EXISTS idx_artifacts_agent ON artifacts(agent_id);
CREATE INDEX IF NOT EXISTS idx_artifacts_type ON artifacts(artifact_type);
CREATE INDEX IF NOT EXISTS idx_artifacts_status ON artifacts(status);

-- Working memory: key-value store for fast agent state (hot cache spillover).
-- Agents can read/write ephemeral state here with optional TTL.
-- This supplements the existing working_memory table from 008 by adding
-- agent and mission scoping.
CREATE TABLE IF NOT EXISTS agent_state (
    agent_id TEXT NOT NULL,
    key TEXT NOT NULL,
    value TEXT NOT NULL,
    mission_id UUID REFERENCES missions(id) ON DELETE CASCADE,
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    expires_at TIMESTAMPTZ,                          -- optional TTL for auto-cleanup
    PRIMARY KEY (agent_id, key)
);

CREATE INDEX IF NOT EXISTS idx_agent_state_mission ON agent_state(mission_id);
CREATE INDEX IF NOT EXISTS idx_agent_state_expires ON agent_state(expires_at) WHERE expires_at IS NOT NULL;
