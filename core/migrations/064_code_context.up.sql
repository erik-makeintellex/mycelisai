-- 064: Native Mycelis code context source registration and deterministic index storage.

CREATE TABLE IF NOT EXISTS code_context_sources (
    id                  TEXT NOT NULL,
    tenant_id           TEXT NOT NULL DEFAULT 'default',
    name                TEXT NOT NULL,
    source_type         TEXT NOT NULL,
    root_path           TEXT NOT NULL,
    scope_kind          TEXT NOT NULL DEFAULT 'workspace',
    scope_ref           TEXT,
    include_globs       JSONB NOT NULL DEFAULT '[]'::jsonb,
    exclude_globs       JSONB NOT NULL DEFAULT '[]'::jsonb,
    languages           JSONB NOT NULL DEFAULT '[]'::jsonb,
    config_record_id    UUID NULL REFERENCES config_documents(record_id),
    config_digest       TEXT,
    extraction_version  TEXT NOT NULL DEFAULT 'code-context-fixture-v1',
    sensitivity_class   TEXT NOT NULL DEFAULT 'restricted',
    trust_class         TEXT NOT NULL DEFAULT 'trusted_internal',
    status              TEXT NOT NULL DEFAULT 'available',
    recovery            TEXT,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (tenant_id, id),
    CONSTRAINT code_context_sources_type_check
        CHECK (source_type IN ('repository', 'local_code_folder')),
    CONSTRAINT code_context_sources_scope_check
        CHECK (scope_kind IN ('all', 'group', 'host', 'operator', 'workspace', 'organization', 'built_in')),
    CONSTRAINT code_context_sources_scope_ref_check
        CHECK ((scope_kind IN ('all', 'built_in') AND COALESCE(scope_ref, '') = '') OR scope_kind NOT IN ('all', 'built_in')),
    CONSTRAINT code_context_sources_extraction_check
        CHECK (extraction_version = 'code-context-fixture-v1')
);

CREATE INDEX IF NOT EXISTS idx_code_context_sources_scope
    ON code_context_sources(tenant_id, scope_kind, scope_ref);

CREATE TABLE IF NOT EXISTS code_context_snapshots (
    id                  TEXT NOT NULL,
    tenant_id           TEXT NOT NULL DEFAULT 'default',
    source_id           TEXT NOT NULL,
    snapshot_ref        TEXT NOT NULL,
    commit_or_digest    TEXT NOT NULL,
    extraction_version  TEXT NOT NULL,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (tenant_id, id),
    FOREIGN KEY (tenant_id, source_id) REFERENCES code_context_sources(tenant_id, id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_code_context_snapshots_source
    ON code_context_snapshots(tenant_id, source_id, created_at DESC);

CREATE TABLE IF NOT EXISTS code_context_files (
    snapshot_id TEXT NOT NULL,
    tenant_id   TEXT NOT NULL DEFAULT 'default',
    path        TEXT NOT NULL,
    language    TEXT NOT NULL,
    digest      TEXT NOT NULL,
    bytes       BIGINT NOT NULL DEFAULT 0,
    PRIMARY KEY (snapshot_id, tenant_id, path),
    FOREIGN KEY (tenant_id, snapshot_id) REFERENCES code_context_snapshots(tenant_id, id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_code_context_files_path
    ON code_context_files(tenant_id, path);

CREATE TABLE IF NOT EXISTS code_context_symbols (
    snapshot_id TEXT NOT NULL,
    tenant_id   TEXT NOT NULL DEFAULT 'default',
    symbol_id   TEXT NOT NULL,
    name        TEXT NOT NULL,
    kind        TEXT NOT NULL,
    path        TEXT NOT NULL,
    start_line  INTEGER NOT NULL,
    end_line    INTEGER NOT NULL,
    PRIMARY KEY (snapshot_id, tenant_id, symbol_id),
    FOREIGN KEY (snapshot_id, tenant_id, path) REFERENCES code_context_files(snapshot_id, tenant_id, path) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_code_context_symbols_name
    ON code_context_symbols(tenant_id, name);

CREATE TABLE IF NOT EXISTS code_context_edges (
    id          BIGSERIAL PRIMARY KEY,
    snapshot_id TEXT NOT NULL,
    tenant_id   TEXT NOT NULL DEFAULT 'default',
    from_ref    TEXT NOT NULL,
    to_ref      TEXT NOT NULL,
    kind        TEXT NOT NULL,
    provenance  TEXT NOT NULL,
    FOREIGN KEY (tenant_id, snapshot_id) REFERENCES code_context_snapshots(tenant_id, id) ON DELETE CASCADE,
    CONSTRAINT code_context_edges_provenance_check
        CHECK (provenance IN ('extracted', 'inferred'))
);

CREATE INDEX IF NOT EXISTS idx_code_context_edges_refs
    ON code_context_edges(tenant_id, snapshot_id, from_ref, to_ref);
