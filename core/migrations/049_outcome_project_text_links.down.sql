DROP INDEX IF EXISTS idx_outcome_projects_run;
DROP INDEX IF EXISTS idx_outcome_projects_intent_proof;

ALTER TABLE outcome_projects
    ALTER COLUMN run_id TYPE UUID USING NULLIF(run_id, '')::uuid,
    ALTER COLUMN intent_proof_id TYPE UUID USING NULLIF(intent_proof_id, '')::uuid;

CREATE INDEX IF NOT EXISTS idx_outcome_projects_run
    ON outcome_projects(run_id)
    WHERE run_id IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_outcome_projects_intent_proof
    ON outcome_projects(intent_proof_id)
    WHERE intent_proof_id IS NOT NULL;
