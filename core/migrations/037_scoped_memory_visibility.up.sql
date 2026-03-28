-- V8.2: scoped durable memory metadata for team-aware recall

ALTER TABLE agent_memories
    ADD COLUMN IF NOT EXISTS tenant_id TEXT NOT NULL DEFAULT 'default',
    ADD COLUMN IF NOT EXISTS team_id TEXT,
    ADD COLUMN IF NOT EXISTS agent_id TEXT NOT NULL DEFAULT 'admin',
    ADD COLUMN IF NOT EXISTS run_id TEXT,
    ADD COLUMN IF NOT EXISTS visibility TEXT NOT NULL DEFAULT 'team';

CREATE INDEX IF NOT EXISTS idx_agent_memories_scope_created
    ON agent_memories(tenant_id, team_id, agent_id, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_agent_memories_visibility
    ON agent_memories(visibility);
