-- 040: Durable AI Organization homes (MYC-RC-004 slice RC-004-0).
-- Replaces the Core in-memory OrganizationStore so organizations survive
-- restarts. Identity columns are promoted for listing and lookup; the full
-- OrganizationHomePayload (departments, agent-type profiles, AI Engine,
-- Response Style, and output model routing settings) lives in `document`.
-- Those sub-settings are always read and written as one aggregate through a
-- single row lock, so they stay in the document instead of child tables.

CREATE TABLE IF NOT EXISTS organizations (
    id TEXT PRIMARY KEY,
    tenant_id TEXT NOT NULL DEFAULT 'default',
    name TEXT NOT NULL,
    purpose TEXT NOT NULL DEFAULT '',
    template_id TEXT NOT NULL DEFAULT '',
    document JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT organizations_document_object_check
        CHECK (jsonb_typeof(document) = 'object')
);

CREATE INDEX IF NOT EXISTS idx_organizations_tenant_name
    ON organizations(tenant_id, name);
CREATE INDEX IF NOT EXISTS idx_organizations_tenant_updated
    ON organizations(tenant_id, updated_at DESC);
CREATE INDEX IF NOT EXISTS idx_organizations_template
    ON organizations(template_id)
    WHERE template_id <> '';
