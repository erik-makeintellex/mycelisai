CREATE TABLE IF NOT EXISTS capability_manifests (
    id                    TEXT PRIMARY KEY,
    version               TEXT NOT NULL DEFAULT 'capability_manifest.v1',
    display_name          TEXT NOT NULL,
    kind                  TEXT NOT NULL,
    source                TEXT NOT NULL,
    status                TEXT NOT NULL,
    risk_class            TEXT NOT NULL,
    description           TEXT NOT NULL DEFAULT '',
    tool_refs             JSONB NOT NULL DEFAULT '[]'::jsonb,
    default_allowed_roles JSONB NOT NULL DEFAULT '[]'::jsonb,
    audit_required        BOOLEAN NOT NULL DEFAULT FALSE,
    approval_required     BOOLEAN NOT NULL DEFAULT FALSE,
    metadata              JSONB NOT NULL DEFAULT '{}'::jsonb,
    derived_at            TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at            TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_capability_manifests_kind
    ON capability_manifests(kind, status);

CREATE INDEX IF NOT EXISTS idx_capability_manifests_source
    ON capability_manifests(source, status);

