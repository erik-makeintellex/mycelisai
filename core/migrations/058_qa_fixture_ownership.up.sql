-- Migration 058: Explicit ownership for automated acceptance fixtures.
--
-- QA cleanup is intentionally separate from operator retention controls. A
-- fixture scope names both its owner and one execution, while resource rows
-- enumerate the exact records or governed workspace paths that execution may
-- remove. Shared NATS storage is never a fixture resource kind.

CREATE TABLE IF NOT EXISTS qa_fixture_scopes (
    id UUID PRIMARY KEY,
    tenant_id TEXT NOT NULL DEFAULT 'default',
    owner_ref TEXT NOT NULL,
    execution_ref TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'open',
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_qa_fixture_scope_owner CHECK (BTRIM(owner_ref) <> ''),
    CONSTRAINT chk_qa_fixture_scope_execution CHECK (BTRIM(execution_ref) <> ''),
    CONSTRAINT chk_qa_fixture_scope_status CHECK (status IN ('open', 'partial', 'purged')),
    CONSTRAINT uq_qa_fixture_scope_execution UNIQUE (tenant_id, owner_ref, execution_ref)
);

CREATE TABLE IF NOT EXISTS qa_fixture_resources (
    id UUID PRIMARY KEY,
    scope_id UUID NOT NULL REFERENCES qa_fixture_scopes(id) ON DELETE CASCADE,
    resource_kind TEXT NOT NULL,
    resource_ref TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_qa_fixture_resource_ref CHECK (BTRIM(resource_ref) <> ''),
    CONSTRAINT chk_qa_fixture_resource_kind CHECK (
        resource_kind IN ('organization', 'group', 'team', 'run', 'outcome', 'artifact', 'workspace_path')
    ),
    CONSTRAINT uq_qa_fixture_resource UNIQUE (scope_id, resource_kind, resource_ref)
);

CREATE INDEX IF NOT EXISTS idx_qa_fixture_scopes_expiry
    ON qa_fixture_scopes (status, expires_at);
CREATE INDEX IF NOT EXISTS idx_qa_fixture_resources_scope
    ON qa_fixture_resources (scope_id, resource_kind);
