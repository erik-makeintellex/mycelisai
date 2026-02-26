-- 032: MCP Tool Sets — named bundles of mcp: tool references for agent-scoped binding.
-- Agents reference tool sets via "toolset:<name>" in their Tools[] manifest field.

CREATE TABLE IF NOT EXISTS mcp_tool_sets (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name        TEXT NOT NULL UNIQUE,
    description TEXT,
    tool_refs   JSONB NOT NULL DEFAULT '[]',
    tenant_id   TEXT NOT NULL DEFAULT 'default',
    created_at  TIMESTAMPTZ DEFAULT NOW(),
    updated_at  TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_mcp_tool_sets_name ON mcp_tool_sets(name);
