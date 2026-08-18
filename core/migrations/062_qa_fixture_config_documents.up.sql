-- 062: Allow exact ConfigDocument revisions to participate in QA cleanup.

ALTER TABLE qa_fixture_resources
    DROP CONSTRAINT IF EXISTS chk_qa_fixture_resource_kind;

ALTER TABLE qa_fixture_resources
    ADD CONSTRAINT chk_qa_fixture_resource_kind CHECK (
        resource_kind IN (
            'organization', 'group', 'team', 'run', 'outcome', 'artifact',
            'config_document', 'workspace_path'
        )
    );
