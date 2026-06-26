-- Migration 048: Durable outcome ownership spine
-- Gives Soma-facing deliverables a project owner record without replacing groups.

CREATE TABLE IF NOT EXISTS outcome_projects (
    id UUID PRIMARY KEY,
    tenant_id TEXT NOT NULL DEFAULT 'default',
    outcome_id TEXT NOT NULL,
    title TEXT NOT NULL,
    purpose TEXT NOT NULL DEFAULT '',
    execution_mode TEXT NOT NULL DEFAULT 'project',
    workspace_folder TEXT NOT NULL DEFAULT '',
    status TEXT NOT NULL DEFAULT 'active',
    run_id TEXT NULL,
    intent_proof_id TEXT NULL,
    contract_id TEXT NULL,
    proof_id TEXT NULL,
    work_item_refs JSONB NOT NULL DEFAULT '[]'::jsonb,
    output_refs JSONB NOT NULL DEFAULT '[]'::jsonb,
    proof_refs JSONB NOT NULL DEFAULT '[]'::jsonb,
    recovery_refs JSONB NOT NULL DEFAULT '[]'::jsonb,
    retention_policy TEXT NOT NULL DEFAULT 'retained',
    version TEXT NOT NULL DEFAULT 'v1',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_outcome_projects_updated ON outcome_projects(tenant_id, updated_at DESC);
CREATE INDEX IF NOT EXISTS idx_outcome_projects_run ON outcome_projects(run_id) WHERE run_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_outcome_projects_intent_proof ON outcome_projects(intent_proof_id) WHERE intent_proof_id IS NOT NULL;
CREATE UNIQUE INDEX IF NOT EXISTS idx_outcome_projects_tenant_outcome
    ON outcome_projects(tenant_id, outcome_id);

CREATE TABLE IF NOT EXISTS team_registry_entries (
    id UUID PRIMARY KEY,
    tenant_id TEXT NOT NULL DEFAULT 'default',
    project_id UUID NOT NULL REFERENCES outcome_projects(id) ON DELETE CASCADE,
    group_id TEXT NOT NULL DEFAULT '',
    role TEXT NOT NULL,
    team_id TEXT NOT NULL DEFAULT '',
    agent_id TEXT NOT NULL DEFAULT '',
    assignment_reason TEXT NOT NULL DEFAULT '',
    temporary BOOLEAN NOT NULL DEFAULT TRUE,
    expires_at TIMESTAMPTZ NULL,
    status TEXT NOT NULL DEFAULT 'active',
    version TEXT NOT NULL DEFAULT 'v1',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_team_registry_entries_project ON team_registry_entries(project_id, created_at);
CREATE INDEX IF NOT EXISTS idx_team_registry_entries_team ON team_registry_entries(team_id) WHERE team_id <> '';
CREATE INDEX IF NOT EXISTS idx_team_registry_entries_status ON team_registry_entries(status);
