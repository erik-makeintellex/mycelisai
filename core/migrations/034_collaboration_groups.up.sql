-- Migration 034: Root-admin collaboration groups (DB-backed persistence)
-- Supports Team A/B contracts:
-- - tenant-scoped group storage
-- - policy reference linkage
-- - audit linkage for create/update mutations

CREATE TABLE IF NOT EXISTS collaboration_groups (
    id UUID PRIMARY KEY,
    tenant_id TEXT NOT NULL DEFAULT 'default',
    name TEXT NOT NULL,
    goal_statement TEXT NOT NULL,
    work_mode TEXT NOT NULL,
    allowed_capabilities JSONB NOT NULL DEFAULT '[]'::jsonb,
    member_user_ids JSONB NOT NULL DEFAULT '[]'::jsonb,
    team_ids JSONB NOT NULL DEFAULT '[]'::jsonb,
    coordinator_profile TEXT NOT NULL DEFAULT '',
    approval_policy_ref TEXT NOT NULL DEFAULT '',
    status TEXT NOT NULL DEFAULT 'active',
    expiry TIMESTAMPTZ NULL,
    created_by TEXT NOT NULL,
    created_audit_event_id UUID NULL,
    updated_audit_event_id UUID NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_collaboration_groups_tenant ON collaboration_groups(tenant_id);
CREATE INDEX IF NOT EXISTS idx_collaboration_groups_status ON collaboration_groups(status);
CREATE INDEX IF NOT EXISTS idx_collaboration_groups_name ON collaboration_groups(name);
CREATE UNIQUE INDEX IF NOT EXISTS idx_collaboration_groups_tenant_name_unique
    ON collaboration_groups(tenant_id, name);
