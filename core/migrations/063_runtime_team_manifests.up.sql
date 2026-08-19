BEGIN;

-- Migration 063: Exact runtime-team manifest persistence.
-- Runtime-created teams keep their approved agent/profile snapshots across
-- Core restarts instead of being reconstructed from partial work metadata.

CREATE TABLE IF NOT EXISTS runtime_team_manifests (
    tenant_id TEXT NOT NULL DEFAULT 'default',
    team_id TEXT NOT NULL,
    schema_version TEXT NOT NULL DEFAULT 'v1',
    manifest_digest TEXT NOT NULL,
    manifest JSONB NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (tenant_id, team_id),
    CONSTRAINT chk_runtime_team_id_not_blank
        CHECK (BTRIM(team_id) <> ''),
    CONSTRAINT chk_runtime_team_manifest_object
        CHECK (jsonb_typeof(manifest) = 'object'),
    CONSTRAINT chk_runtime_team_manifest_identity
        CHECK (manifest->>'id' = team_id)
);

CREATE INDEX IF NOT EXISTS idx_runtime_team_manifests_updated
    ON runtime_team_manifests (tenant_id, updated_at DESC);

COMMIT;
