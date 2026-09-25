-- 041: Operator templates (MYC-RC-004 slice RC-004G).
-- Clones of file template bundles and starters that an operator created with
-- POST /api/v1/config/templates/{id}/clone. The Helm-mounted templates
-- directory is read-only, so clones live here. `document` is the full bundle
-- (same shape and validation as a starter file); `document_sha256` is the
-- canonical-JSON hash that governed apply binds its confirm token to.
-- Every row is written only after a consumed confirm token; the audit and
-- intent-proof ids tie it to log_entries and intent_proofs.

CREATE TABLE IF NOT EXISTS operator_templates (
    id TEXT PRIMARY KEY,
    tenant_id TEXT NOT NULL DEFAULT 'default',
    name TEXT NOT NULL,
    source_template_id TEXT NOT NULL DEFAULT '',
    document JSONB NOT NULL DEFAULT '{}'::jsonb,
    document_sha256 TEXT NOT NULL DEFAULT '',
    created_by TEXT NOT NULL DEFAULT '',
    created_audit_event_id UUID,
    updated_audit_event_id UUID,
    intent_proof_id UUID,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT operator_templates_document_object_check
        CHECK (jsonb_typeof(document) = 'object')
);

CREATE INDEX IF NOT EXISTS idx_operator_templates_tenant_name
    ON operator_templates(tenant_id, name);
CREATE INDEX IF NOT EXISTS idx_operator_templates_source
    ON operator_templates(source_template_id)
    WHERE source_template_id <> '';
