-- Phase 7.7: Agent Memories (persistent knowledge for admin/council)
CREATE TABLE IF NOT EXISTS agent_memories (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    category    VARCHAR(50) NOT NULL DEFAULT 'fact',
    content     TEXT NOT NULL,
    context     TEXT,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Indexes for recall queries
CREATE INDEX IF NOT EXISTS idx_agent_memories_category ON agent_memories(category);
CREATE INDEX IF NOT EXISTS idx_agent_memories_content_trgm ON agent_memories USING gin (content gin_trgm_ops);
