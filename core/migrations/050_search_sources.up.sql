-- 050: Governed Search Sources
-- Stores operator-configured search boundaries and secret references for Soma search.

CREATE TABLE IF NOT EXISTS search_sources (
    id                  TEXT PRIMARY KEY,
    name                TEXT NOT NULL,
    provider            TEXT NOT NULL,
    source_type         TEXT NOT NULL,
    endpoint            TEXT,
    scope_kind          TEXT NOT NULL DEFAULT 'all',
    scope_ref           TEXT,
    boundary            TEXT NOT NULL,
    auth_scheme         TEXT NOT NULL DEFAULT 'none',
    secret_ref          TEXT,
    mode                TEXT NOT NULL DEFAULT 'preview',
    sensitivity_class   TEXT NOT NULL DEFAULT 'public',
    trust_class         TEXT NOT NULL DEFAULT 'bounded_external',
    status              TEXT NOT NULL DEFAULT 'available',
    recovery            TEXT,
    tenant_id           TEXT NOT NULL DEFAULT 'default',
    created_at          TIMESTAMPTZ DEFAULT NOW(),
    updated_at          TIMESTAMPTZ DEFAULT NOW()
);

ALTER TABLE search_sources
    DROP CONSTRAINT IF EXISTS chk_search_sources_scope_kind,
    DROP CONSTRAINT IF EXISTS chk_search_sources_scope_ref;

ALTER TABLE search_sources
    ADD CONSTRAINT chk_search_sources_scope_kind
    CHECK (scope_kind IN ('all', 'group', 'host'));

ALTER TABLE search_sources
    ADD CONSTRAINT chk_search_sources_scope_ref
    CHECK ((scope_kind = 'all' AND COALESCE(scope_ref, '') = '') OR (scope_kind IN ('group', 'host') AND COALESCE(scope_ref, '') <> ''));

CREATE INDEX IF NOT EXISTS idx_search_sources_scope
    ON search_sources(tenant_id, scope_kind, scope_ref);

CREATE INDEX IF NOT EXISTS idx_search_sources_provider
    ON search_sources(provider);
