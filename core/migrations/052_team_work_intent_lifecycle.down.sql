ALTER TABLE team_status_events
    DROP COLUMN IF EXISTS work_intent,
    DROP COLUMN IF EXISTS execution_mode;

ALTER TABLE team_work_items
    DROP COLUMN IF EXISTS work_intent,
    DROP COLUMN IF EXISTS execution_mode;
