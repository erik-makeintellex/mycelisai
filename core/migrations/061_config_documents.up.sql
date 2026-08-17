-- 061: File-authoritative configuration revisions and scoped activation history.

CREATE TABLE IF NOT EXISTS config_documents (
    record_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id TEXT NOT NULL DEFAULT 'default',
    document_id TEXT NOT NULL,
    api_version TEXT NOT NULL,
    kind TEXT NOT NULL,
    name TEXT NOT NULL,
    version TEXT NOT NULL,
    owner_id TEXT NOT NULL,
    scope_kind TEXT NOT NULL,
    scope_ref TEXT NOT NULL DEFAULT '',
    enabled BOOLEAN NOT NULL DEFAULT TRUE,
    source_kind TEXT NOT NULL,
    source_ref TEXT NOT NULL DEFAULT '',
    secret_refs JSONB NOT NULL DEFAULT '[]'::jsonb,
    governance JSONB NOT NULL DEFAULT '{}'::jsonb,
    spec JSONB NOT NULL,
    digest TEXT NOT NULL,
    validation_state TEXT NOT NULL DEFAULT 'valid',
    created_by TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT config_documents_scope_kind_check
        CHECK (scope_kind IN ('operator', 'workspace', 'organization', 'built_in')),
    CONSTRAINT config_documents_source_kind_check
        CHECK (source_kind IN ('built_in', 'file', 'soma', 'api')),
    CONSTRAINT config_documents_validation_state_check
        CHECK (validation_state IN ('valid', 'invalid')),
    CONSTRAINT config_documents_revision_unique
        UNIQUE (tenant_id, document_id, version, digest)
);

CREATE INDEX IF NOT EXISTS idx_config_documents_lookup
    ON config_documents(tenant_id, kind, scope_kind, scope_ref, document_id, created_at DESC);

CREATE TABLE IF NOT EXISTS config_document_activations (
    tenant_id TEXT NOT NULL DEFAULT 'default',
    kind TEXT NOT NULL,
    document_id TEXT NOT NULL,
    scope_kind TEXT NOT NULL,
    scope_ref TEXT NOT NULL DEFAULT '',
    config_document_record_id UUID NOT NULL REFERENCES config_documents(record_id),
    activated_by TEXT NOT NULL,
    activated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (tenant_id, kind, document_id, scope_kind, scope_ref)
);

ALTER TABLE config_document_activations ADD COLUMN IF NOT EXISTS kind TEXT;
UPDATE config_document_activations activation
SET kind = document.kind
FROM config_documents document
WHERE activation.config_document_record_id = document.record_id
  AND activation.kind IS NULL;
ALTER TABLE config_document_activations ALTER COLUMN kind SET NOT NULL;
ALTER TABLE config_document_activations DROP CONSTRAINT IF EXISTS config_document_activations_pkey;
ALTER TABLE config_document_activations
    ADD PRIMARY KEY (tenant_id, kind, document_id, scope_kind, scope_ref);

CREATE TABLE IF NOT EXISTS config_document_activation_history (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id TEXT NOT NULL DEFAULT 'default',
    kind TEXT NOT NULL,
    document_id TEXT NOT NULL,
    scope_kind TEXT NOT NULL,
    scope_ref TEXT NOT NULL DEFAULT '',
    from_record_id UUID NULL REFERENCES config_documents(record_id),
    to_record_id UUID NOT NULL REFERENCES config_documents(record_id),
    action TEXT NOT NULL,
    actor_id TEXT NOT NULL,
    audit_event_id UUID NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT config_document_activation_action_check
        CHECK (action IN ('activate', 'rollback'))
);

ALTER TABLE config_document_activation_history ADD COLUMN IF NOT EXISTS kind TEXT;
UPDATE config_document_activation_history history
SET kind = document.kind
FROM config_documents document
WHERE history.to_record_id = document.record_id
  AND history.kind IS NULL;
ALTER TABLE config_document_activation_history ALTER COLUMN kind SET NOT NULL;

DROP INDEX IF EXISTS idx_config_document_activation_history;
CREATE INDEX IF NOT EXISTS idx_config_document_activation_history
    ON config_document_activation_history(tenant_id, kind, document_id, scope_kind, scope_ref, created_at DESC);
