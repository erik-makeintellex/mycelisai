DELETE FROM agent_catalogue WHERE source = 'built_in' AND profile_key LIKE 'default.%';

DROP INDEX IF EXISTS idx_agent_catalogue_profile_key;

ALTER TABLE agent_catalogue
    DROP COLUMN IF EXISTS usage_policy,
    DROP COLUMN IF EXISTS context_bindings,
    DROP COLUMN IF EXISTS capability_refs,
    DROP COLUMN IF EXISTS locked,
    DROP COLUMN IF EXISTS source,
    DROP COLUMN IF EXISTS description,
    DROP COLUMN IF EXISTS profile_key;
