DROP INDEX IF EXISTS idx_team_work_items_recovery_deadline;

ALTER TABLE team_work_items
    DROP COLUMN IF EXISTS recovery_deadline_at;
