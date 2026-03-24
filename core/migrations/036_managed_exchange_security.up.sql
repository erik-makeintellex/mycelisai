CREATE TABLE IF NOT EXISTS exchange_capability_registry (
    id                    TEXT PRIMARY KEY,
    label                 TEXT NOT NULL,
    source                TEXT NOT NULL,
    risk_class            TEXT NOT NULL,
    default_allowed_roles JSONB NOT NULL DEFAULT '[]'::jsonb,
    audit_required        BOOLEAN NOT NULL DEFAULT FALSE,
    approval_required     BOOLEAN NOT NULL DEFAULT FALSE,
    description           TEXT NOT NULL DEFAULT '',
    created_at            TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

ALTER TABLE exchange_channels
    ADD COLUMN IF NOT EXISTS reviewers JSONB NOT NULL DEFAULT '[]'::jsonb,
    ADD COLUMN IF NOT EXISTS sensitivity_class TEXT NOT NULL DEFAULT 'role_scoped';

ALTER TABLE exchange_threads
    ADD COLUMN IF NOT EXISTS allowed_reviewers JSONB NOT NULL DEFAULT '[]'::jsonb,
    ADD COLUMN IF NOT EXISTS escalation_rights JSONB NOT NULL DEFAULT '[]'::jsonb;

ALTER TABLE exchange_items
    ADD COLUMN IF NOT EXISTS sensitivity_class TEXT NOT NULL DEFAULT 'role_scoped',
    ADD COLUMN IF NOT EXISTS source_role TEXT NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS source_team TEXT,
    ADD COLUMN IF NOT EXISTS target_role TEXT,
    ADD COLUMN IF NOT EXISTS target_team TEXT,
    ADD COLUMN IF NOT EXISTS allowed_consumers JSONB NOT NULL DEFAULT '[]'::jsonb,
    ADD COLUMN IF NOT EXISTS capability_id TEXT REFERENCES exchange_capability_registry(id),
    ADD COLUMN IF NOT EXISTS trust_class TEXT NOT NULL DEFAULT 'trusted_internal',
    ADD COLUMN IF NOT EXISTS review_required BOOLEAN NOT NULL DEFAULT FALSE;

CREATE INDEX IF NOT EXISTS idx_exchange_channels_sensitivity
    ON exchange_channels(sensitivity_class, visibility, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_exchange_items_security
    ON exchange_items(sensitivity_class, trust_class, review_required, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_exchange_items_capability
    ON exchange_items(capability_id, created_at DESC);
