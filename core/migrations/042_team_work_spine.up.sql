-- Migration 042: Durable team-work spine
-- Persists Soma/Council team work without implying created teams are already executing.

CREATE TABLE IF NOT EXISTS team_work_items (
    id UUID PRIMARY KEY,
    tenant_id TEXT NOT NULL DEFAULT 'default',
    team_id TEXT NOT NULL,
    run_id UUID NULL,
    intent_proof_id UUID NULL,
    contract_id TEXT NULL,
    proof_id TEXT NULL,
    objective TEXT NOT NULL,
    scope JSONB NOT NULL DEFAULT '[]'::jsonb,
    owner TEXT NOT NULL DEFAULT '',
    execution_shape TEXT NOT NULL,
    expected_outputs JSONB NOT NULL DEFAULT '[]'::jsonb,
    expected_proof JSONB NOT NULL DEFAULT '[]'::jsonb,
    capability_requirements JSONB NOT NULL DEFAULT '[]'::jsonb,
    governance_posture TEXT NOT NULL DEFAULT '',
    state TEXT NOT NULL,
    last_event JSONB NULL,
    needs_operator BOOLEAN NOT NULL DEFAULT FALSE,
    degradation_state TEXT NOT NULL DEFAULT '',
    recovery_options JSONB NOT NULL DEFAULT '[]'::jsonb,
    output_refs JSONB NOT NULL DEFAULT '[]'::jsonb,
    proof_refs JSONB NOT NULL DEFAULT '[]'::jsonb,
    audit_refs JSONB NOT NULL DEFAULT '[]'::jsonb,
    version TEXT NOT NULL DEFAULT 'v1',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_team_work_create_team_state
        CHECK (execution_shape <> 'create_team' OR state IN ('new', 'briefed')),
    CONSTRAINT chk_team_work_execution_state
        CHECK (execution_shape = 'create_team' OR state IN ('queued', 'running', 'needs_operator', 'reviewing', 'output_ready', 'degraded', 'paused', 'archived'))
);

CREATE INDEX IF NOT EXISTS idx_team_work_items_team ON team_work_items(team_id, updated_at DESC);
CREATE INDEX IF NOT EXISTS idx_team_work_items_state ON team_work_items(state);
CREATE INDEX IF NOT EXISTS idx_team_work_items_run ON team_work_items(run_id) WHERE run_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_team_work_items_intent_proof ON team_work_items(intent_proof_id) WHERE intent_proof_id IS NOT NULL;

CREATE TABLE IF NOT EXISTS team_interactions (
    id UUID PRIMARY KEY,
    tenant_id TEXT NOT NULL DEFAULT 'default',
    team_id TEXT NOT NULL,
    work_item_id UUID NOT NULL REFERENCES team_work_items(id) ON DELETE CASCADE,
    run_id UUID NULL,
    intent_proof_id UUID NULL,
    contract_id TEXT NULL,
    proof_id TEXT NULL,
    source_kind TEXT NOT NULL,
    source_channel TEXT NOT NULL,
    actor_ref TEXT NOT NULL DEFAULT '',
    verb TEXT NOT NULL,
    summary TEXT NOT NULL,
    payload_kind TEXT NOT NULL,
    payload_ref TEXT NOT NULL DEFAULT '',
    payload JSONB NULL,
    approval_ref TEXT NOT NULL DEFAULT '',
    audit_refs JSONB NOT NULL DEFAULT '[]'::jsonb,
    timestamp TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    version TEXT NOT NULL DEFAULT 'v1'
);

CREATE INDEX IF NOT EXISTS idx_team_interactions_team ON team_interactions(team_id, timestamp DESC);
CREATE INDEX IF NOT EXISTS idx_team_interactions_work ON team_interactions(work_item_id, timestamp DESC);
CREATE INDEX IF NOT EXISTS idx_team_interactions_run ON team_interactions(run_id) WHERE run_id IS NOT NULL;

CREATE TABLE IF NOT EXISTS team_status_events (
    id UUID PRIMARY KEY,
    tenant_id TEXT NOT NULL DEFAULT 'default',
    team_id TEXT NOT NULL,
    work_item_id UUID NOT NULL REFERENCES team_work_items(id) ON DELETE CASCADE,
    run_id UUID NULL,
    intent_proof_id UUID NULL,
    contract_id TEXT NULL,
    proof_id TEXT NULL,
    state TEXT NOT NULL,
    headline TEXT NOT NULL,
    details TEXT NOT NULL DEFAULT '',
    confidence_posture TEXT NOT NULL DEFAULT '',
    blocked_by JSONB NOT NULL DEFAULT '[]'::jsonb,
    next_action TEXT NOT NULL DEFAULT '',
    source_kind TEXT NOT NULL DEFAULT '',
    source_channel TEXT NOT NULL DEFAULT '',
    payload_kind TEXT NOT NULL DEFAULT '',
    audit_refs JSONB NOT NULL DEFAULT '[]'::jsonb,
    timestamp TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    version TEXT NOT NULL DEFAULT 'v1'
);

CREATE INDEX IF NOT EXISTS idx_team_status_events_team ON team_status_events(team_id, timestamp DESC);
CREATE INDEX IF NOT EXISTS idx_team_status_events_work ON team_status_events(work_item_id, timestamp DESC);
CREATE INDEX IF NOT EXISTS idx_team_status_events_state ON team_status_events(state);
