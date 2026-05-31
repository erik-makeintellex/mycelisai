ALTER TABLE trigger_executions
    ADD COLUMN IF NOT EXISTS handoff_key TEXT,
    ADD COLUMN IF NOT EXISTS intent_proof_id UUID REFERENCES intent_proofs(id) ON DELETE SET NULL,
    ADD COLUMN IF NOT EXISTS contract_id UUID REFERENCES execution_contracts(id) ON DELETE SET NULL,
    ADD COLUMN IF NOT EXISTS proposal_status TEXT NOT NULL DEFAULT 'recorded',
    ADD COLUMN IF NOT EXISTS handoff_payload JSONB NOT NULL DEFAULT '{}'::jsonb;

CREATE UNIQUE INDEX IF NOT EXISTS trigger_executions_handoff_key_unique
    ON trigger_executions(rule_id, handoff_key)
    WHERE handoff_key IS NOT NULL;

CREATE INDEX IF NOT EXISTS trigger_executions_intent_proof
    ON trigger_executions(intent_proof_id)
    WHERE intent_proof_id IS NOT NULL;

CREATE INDEX IF NOT EXISTS trigger_executions_contract
    ON trigger_executions(contract_id)
    WHERE contract_id IS NOT NULL;
