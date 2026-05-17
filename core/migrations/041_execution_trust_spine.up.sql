-- Migration 041: Durable execution trust spine
--
-- execution_contracts records the governed promise created before mutation.
-- proof_artifacts records the bounded evidence produced by a confirm-action
-- attempt. Both tables are intentionally minimal but include provenance fields
-- needed for future confidence review and recovery lineage.

CREATE TABLE IF NOT EXISTS execution_contracts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id VARCHAR(255) NOT NULL DEFAULT 'default',
    intent_proof_id UUID REFERENCES intent_proofs(id) ON DELETE SET NULL,
    run_id UUID REFERENCES mission_runs(id) ON DELETE SET NULL,
    template_id TEXT NOT NULL,
    resolved_intent TEXT NOT NULL DEFAULT '',
    execution_shape TEXT NOT NULL DEFAULT 'guided_proposal',
    status TEXT NOT NULL DEFAULT 'proposed',
    execution_status TEXT NOT NULL DEFAULT 'proposed',
    validation_source TEXT NOT NULL DEFAULT 'intent_proof',
    evidence_strength TEXT NOT NULL DEFAULT 'intent_only',
    proof_quality TEXT NOT NULL DEFAULT 'proposed',
    latest_proof_artifact_id UUID,
    audit_event_id UUID,
    output_refs JSONB NOT NULL DEFAULT '[]'::jsonb,
    audit_refs JSONB NOT NULL DEFAULT '[]'::jsonb,
    review_lineage JSONB NOT NULL DEFAULT '[]'::jsonb,
    degradation JSONB NOT NULL DEFAULT '{}'::jsonb,
    recovery JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    confirmed_at TIMESTAMPTZ,
    completed_at TIMESTAMPTZ,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_execution_contracts_intent_proof
    ON execution_contracts(intent_proof_id)
    WHERE intent_proof_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_execution_contracts_run ON execution_contracts(run_id)
    WHERE run_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_execution_contracts_status ON execution_contracts(status, updated_at DESC);

CREATE TABLE IF NOT EXISTS proof_artifacts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id VARCHAR(255) NOT NULL DEFAULT 'default',
    contract_id UUID REFERENCES execution_contracts(id) ON DELETE SET NULL,
    intent_proof_id UUID REFERENCES intent_proofs(id) ON DELETE SET NULL,
    run_id UUID REFERENCES mission_runs(id) ON DELETE SET NULL,
    artifact_kind TEXT NOT NULL DEFAULT 'confirm_action',
    status TEXT NOT NULL,
    proof_class TEXT NOT NULL DEFAULT 'run_and_audit',
    validation_source TEXT NOT NULL DEFAULT 'confirm_action',
    evidence_strength TEXT NOT NULL DEFAULT 'run_audit',
    proof_quality TEXT NOT NULL,
    output_refs JSONB NOT NULL DEFAULT '[]'::jsonb,
    audit_refs JSONB NOT NULL DEFAULT '[]'::jsonb,
    review_lineage JSONB NOT NULL DEFAULT '[]'::jsonb,
    degradation JSONB NOT NULL DEFAULT '{}'::jsonb,
    recovery JSONB NOT NULL DEFAULT '{}'::jsonb,
    payload JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_proof_artifacts_contract ON proof_artifacts(contract_id, created_at DESC)
    WHERE contract_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_proof_artifacts_intent_proof ON proof_artifacts(intent_proof_id, created_at DESC)
    WHERE intent_proof_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_proof_artifacts_run ON proof_artifacts(run_id, created_at DESC)
    WHERE run_id IS NOT NULL;
