DROP INDEX IF EXISTS idx_capability_manifests_health;

ALTER TABLE capability_manifests
    DROP COLUMN IF EXISTS owner,
    DROP COLUMN IF EXISTS secret_ref_policy,
    DROP COLUMN IF EXISTS audit_policy,
    DROP COLUMN IF EXISTS recovery_posture,
    DROP COLUMN IF EXISTS failure_posture,
    DROP COLUMN IF EXISTS last_probe_at,
    DROP COLUMN IF EXISTS last_probe_status,
    DROP COLUMN IF EXISTS health,
    DROP COLUMN IF EXISTS output_schema_ref,
    DROP COLUMN IF EXISTS input_schema_ref,
    DROP COLUMN IF EXISTS allowed_roles,
    DROP COLUMN IF EXISTS approval_posture,
    DROP COLUMN IF EXISTS purpose,
    DROP COLUMN IF EXISTS manifest_version,
    DROP COLUMN IF EXISTS capability_id;
