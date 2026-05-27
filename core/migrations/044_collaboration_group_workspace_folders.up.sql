-- Migration 044: Dedicated workspace folders for collaboration groups
-- Permanent user/Soma groups get a stable workspace-relative output root so
-- generated packages, media, and retained artifacts do not collapse into the
-- general output folders.

ALTER TABLE collaboration_groups
    ADD COLUMN IF NOT EXISTS workspace_folder TEXT NOT NULL DEFAULT '';

UPDATE collaboration_groups
SET workspace_folder = 'groups/' || LOWER(
    TRIM(BOTH '-' FROM REGEXP_REPLACE(
        COALESCE(NULLIF(team_ids->>0, ''), NULLIF(name, ''), id::text),
        '[^[:alnum:]]+',
        '-',
        'g'
    ))
)
WHERE workspace_folder = '';

CREATE INDEX IF NOT EXISTS idx_collaboration_groups_workspace_folder
    ON collaboration_groups(tenant_id, workspace_folder);
