DROP INDEX IF EXISTS idx_mcp_tool_sets_scope;
DROP INDEX IF EXISTS idx_mcp_tool_sets_unique_scope;

ALTER TABLE mcp_tool_sets
    DROP CONSTRAINT IF EXISTS chk_mcp_tool_sets_scope_ref,
    DROP CONSTRAINT IF EXISTS chk_mcp_tool_sets_scope_kind,
    DROP COLUMN IF EXISTS scope_ref,
    DROP COLUMN IF EXISTS scope_kind;

ALTER TABLE mcp_tool_sets
    ADD CONSTRAINT mcp_tool_sets_name_key UNIQUE (name);

CREATE INDEX IF NOT EXISTS idx_mcp_tool_sets_name ON mcp_tool_sets(name);
