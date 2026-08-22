DROP INDEX IF EXISTS uq_qa_fixture_resource_claim;

UPDATE qa_fixture_scopes SET status='partial' WHERE status='purging';

ALTER TABLE qa_fixture_scopes
    DROP CONSTRAINT IF EXISTS chk_qa_fixture_scope_status;

ALTER TABLE qa_fixture_scopes
    ADD CONSTRAINT chk_qa_fixture_scope_status
    CHECK (status IN ('open', 'partial', 'purged'));
