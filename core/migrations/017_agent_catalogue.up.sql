-- 017: Agent Catalogue — persistent agent template definitions
-- These are reusable agent definitions that can be assigned to teams.
-- Distinct from service_manifests (which are team-scoped runtime instances).

CREATE TABLE IF NOT EXISTS agent_catalogue (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL UNIQUE,
    role TEXT NOT NULL,                          -- cognitive, sensory, actuation, ledger
    system_prompt TEXT,
    model TEXT,                                  -- LLM profile name (e.g., "chat", "coder")
    tools JSONB DEFAULT '[]',                    -- MCP tool name bindings
    inputs JSONB DEFAULT '[]',                   -- NATS topic patterns
    outputs JSONB DEFAULT '[]',                  -- NATS topic patterns
    verification_strategy TEXT,                  -- none, semantic, empirical
    verification_rubric JSONB DEFAULT '[]',      -- rubric lines
    validation_command TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_agent_catalogue_role ON agent_catalogue(role);
