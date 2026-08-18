-- 062 rollback: Refuse to discard live ConfigDocument fixture ownership.

DO $$
BEGIN
    IF EXISTS (
        SELECT 1 FROM qa_fixture_resources WHERE resource_kind = 'config_document'
    ) THEN
        RAISE EXCEPTION
            'cannot roll back migration 062 while config_document fixture claims exist';
    END IF;
END $$;

ALTER TABLE qa_fixture_resources
    DROP CONSTRAINT IF EXISTS chk_qa_fixture_resource_kind;

ALTER TABLE qa_fixture_resources
    ADD CONSTRAINT chk_qa_fixture_resource_kind CHECK (
        resource_kind IN (
            'organization', 'group', 'team', 'run', 'outcome', 'artifact',
            'workspace_path'
        )
    );
