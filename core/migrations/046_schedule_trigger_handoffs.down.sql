DROP INDEX IF EXISTS trigger_executions_contract;
DROP INDEX IF EXISTS trigger_executions_intent_proof;
DROP INDEX IF EXISTS trigger_executions_handoff_key_unique;

ALTER TABLE trigger_executions
    DROP COLUMN IF EXISTS handoff_payload,
    DROP COLUMN IF EXISTS proposal_status,
    DROP COLUMN IF EXISTS contract_id,
    DROP COLUMN IF EXISTS intent_proof_id,
    DROP COLUMN IF EXISTS handoff_key;
