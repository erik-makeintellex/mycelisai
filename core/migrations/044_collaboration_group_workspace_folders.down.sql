DROP INDEX IF EXISTS idx_collaboration_groups_workspace_folder;

ALTER TABLE collaboration_groups
    DROP COLUMN IF EXISTS workspace_folder;
