-- Migration 049: Preserve user-facing runtime/proof IDs on OutcomeProject.
-- OutcomeProject is a durable ownership surface and must retain text IDs from
-- runtime flows that are not always UUID-backed rows.

ALTER TABLE outcome_projects
    ALTER COLUMN run_id TYPE TEXT USING COALESCE(run_id::text, ''),
    ALTER COLUMN intent_proof_id TYPE TEXT USING COALESCE(intent_proof_id::text, '');

DROP INDEX IF EXISTS idx_outcome_projects_run;
DROP INDEX IF EXISTS idx_outcome_projects_intent_proof;

CREATE INDEX IF NOT EXISTS idx_outcome_projects_run
    ON outcome_projects(run_id)
    WHERE run_id IS NOT NULL AND run_id <> '';

CREATE INDEX IF NOT EXISTS idx_outcome_projects_intent_proof
    ON outcome_projects(intent_proof_id)
    WHERE intent_proof_id IS NOT NULL AND intent_proof_id <> '';
