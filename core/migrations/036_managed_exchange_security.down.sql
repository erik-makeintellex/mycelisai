DROP INDEX IF EXISTS idx_exchange_items_capability;
DROP INDEX IF EXISTS idx_exchange_items_security;
DROP INDEX IF EXISTS idx_exchange_channels_sensitivity;

ALTER TABLE exchange_items
    DROP COLUMN IF EXISTS review_required,
    DROP COLUMN IF EXISTS trust_class,
    DROP COLUMN IF EXISTS capability_id,
    DROP COLUMN IF EXISTS allowed_consumers,
    DROP COLUMN IF EXISTS target_team,
    DROP COLUMN IF EXISTS target_role,
    DROP COLUMN IF EXISTS source_team,
    DROP COLUMN IF EXISTS source_role,
    DROP COLUMN IF EXISTS sensitivity_class;

ALTER TABLE exchange_threads
    DROP COLUMN IF EXISTS escalation_rights,
    DROP COLUMN IF EXISTS allowed_reviewers;

ALTER TABLE exchange_channels
    DROP COLUMN IF EXISTS sensitivity_class,
    DROP COLUMN IF EXISTS reviewers;

DROP TABLE IF EXISTS exchange_capability_registry;
