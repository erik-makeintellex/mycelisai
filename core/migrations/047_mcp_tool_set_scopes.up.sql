-- 047: MCP tool-set scope layering.
-- Scope modes:
-- - all: available as a shared tool set across the organization.
-- - group: targeted to a collaboration/team/group lane identified by scope_ref.
-- - host: targeted to a deployment/runtime host identified by scope_ref.

ALTER TABLE mcp_tool_sets
    ADD COLUMN IF NOT EXISTS scope_kind TEXT NOT NULL DEFAULT 'all',
    ADD COLUMN IF NOT EXISTS scope_ref TEXT;

UPDATE mcp_tool_sets
SET scope_kind = 'all'
WHERE scope_kind IS NULL OR scope_kind = '';

ALTER TABLE mcp_tool_sets
    DROP CONSTRAINT IF EXISTS chk_mcp_tool_sets_scope_kind,
    DROP CONSTRAINT IF EXISTS chk_mcp_tool_sets_scope_ref;

ALTER TABLE mcp_tool_sets
    ADD CONSTRAINT chk_mcp_tool_sets_scope_kind
    CHECK (scope_kind IN ('all', 'group', 'host'));

ALTER TABLE mcp_tool_sets
    ADD CONSTRAINT chk_mcp_tool_sets_scope_ref
    CHECK (
        (scope_kind = 'all' AND (scope_ref IS NULL OR scope_ref = ''))
        OR (scope_kind IN ('group', 'host') AND scope_ref IS NOT NULL AND scope_ref <> '')
    );

CREATE INDEX IF NOT EXISTS idx_mcp_tool_sets_scope
    ON mcp_tool_sets(scope_kind, scope_ref);

ALTER TABLE mcp_tool_sets
    DROP CONSTRAINT IF EXISTS mcp_tool_sets_name_key;

DROP INDEX IF EXISTS idx_mcp_tool_sets_name;

CREATE UNIQUE INDEX IF NOT EXISTS idx_mcp_tool_sets_unique_scope
    ON mcp_tool_sets(tenant_id, name, scope_kind, COALESCE(scope_ref, ''));

CREATE INDEX IF NOT EXISTS idx_mcp_tool_sets_name
    ON mcp_tool_sets(name);
