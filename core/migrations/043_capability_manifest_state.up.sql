-- Migration 043: Durable CapabilityManifestState runtime fields
--
-- Extends the first manifest registry table into the V8.2 trust object shape
-- used by Runtime/Capability and operator capability views.

ALTER TABLE capability_manifests
    ADD COLUMN IF NOT EXISTS capability_id TEXT,
    ADD COLUMN IF NOT EXISTS manifest_version TEXT NOT NULL DEFAULT 'capability_manifest.v1',
    ADD COLUMN IF NOT EXISTS purpose TEXT NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS approval_posture TEXT NOT NULL DEFAULT 'not_required',
    ADD COLUMN IF NOT EXISTS allowed_roles JSONB NOT NULL DEFAULT '[]'::jsonb,
    ADD COLUMN IF NOT EXISTS input_schema_ref TEXT NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS output_schema_ref TEXT NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS health TEXT NOT NULL DEFAULT 'unknown',
    ADD COLUMN IF NOT EXISTS last_probe_status TEXT NOT NULL DEFAULT 'unknown',
    ADD COLUMN IF NOT EXISTS last_probe_at TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS failure_posture TEXT NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS recovery_posture TEXT NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS audit_policy TEXT NOT NULL DEFAULT 'not_required',
    ADD COLUMN IF NOT EXISTS secret_ref_policy TEXT NOT NULL DEFAULT 'no_raw_secrets',
    ADD COLUMN IF NOT EXISTS owner TEXT NOT NULL DEFAULT 'runtime_capability';

UPDATE capability_manifests
SET capability_id = id
WHERE capability_id IS NULL OR capability_id = '';

ALTER TABLE capability_manifests
    ALTER COLUMN capability_id SET NOT NULL;

CREATE INDEX IF NOT EXISTS idx_capability_manifests_health
    ON capability_manifests(health, last_probe_status);
