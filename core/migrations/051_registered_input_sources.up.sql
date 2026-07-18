-- 051: Registered Input Sources
-- Stores governed external/service/device ingress sources and their durable buffers.

CREATE TABLE IF NOT EXISTS input_sources (
    id                      TEXT PRIMARY KEY,
    name                    TEXT NOT NULL,
    source_type             TEXT NOT NULL,
    adapter_kind            TEXT NOT NULL,
    scope_kind              TEXT NOT NULL DEFAULT 'all',
    scope_ref               TEXT,
    target_outcome_id       TEXT,
    target_group_id         TEXT,
    target_host_id          TEXT,
    auth_scheme             TEXT NOT NULL DEFAULT 'none',
    secret_ref              TEXT,
    allowed_ingress_subject TEXT NOT NULL,
    payload_schema_ref      TEXT,
    buffer_mode             TEXT NOT NULL DEFAULT 'append_log',
    buffer_policy           JSONB NOT NULL DEFAULT '{}'::jsonb,
    sensitivity_class       TEXT NOT NULL DEFAULT 'governed',
    trust_class             TEXT NOT NULL DEFAULT 'bounded_external',
    status                  TEXT NOT NULL DEFAULT 'available',
    recovery                TEXT,
    tenant_id               TEXT NOT NULL DEFAULT 'default',
    created_at              TIMESTAMPTZ DEFAULT NOW(),
    updated_at              TIMESTAMPTZ DEFAULT NOW()
);

ALTER TABLE input_sources
    DROP CONSTRAINT IF EXISTS chk_input_sources_scope_kind,
    DROP CONSTRAINT IF EXISTS chk_input_sources_scope_ref,
    DROP CONSTRAINT IF EXISTS chk_input_sources_auth_secret,
    DROP CONSTRAINT IF EXISTS chk_input_sources_buffer_mode,
    DROP CONSTRAINT IF EXISTS chk_input_sources_ingress_subject;

ALTER TABLE input_sources
    ADD CONSTRAINT chk_input_sources_scope_kind
    CHECK (scope_kind IN ('all', 'group', 'host'));

ALTER TABLE input_sources
    ADD CONSTRAINT chk_input_sources_scope_ref
    CHECK ((scope_kind = 'all' AND COALESCE(scope_ref, '') = '') OR (scope_kind IN ('group', 'host') AND COALESCE(scope_ref, '') <> ''));

ALTER TABLE input_sources
    ADD CONSTRAINT chk_input_sources_auth_secret
    CHECK ((auth_scheme = 'none' AND COALESCE(secret_ref, '') = '') OR (auth_scheme <> 'none' AND COALESCE(secret_ref, '') <> ''));

ALTER TABLE input_sources
    ADD CONSTRAINT chk_input_sources_buffer_mode
    CHECK (buffer_mode IN ('append_log', 'latest_state', 'append_with_latest', 'windowed_rollup'));

ALTER TABLE input_sources
    ADD CONSTRAINT chk_input_sources_ingress_subject
    CHECK (allowed_ingress_subject LIKE 'swarm.global.input.%' AND allowed_ingress_subject NOT LIKE '%>%');

CREATE TABLE IF NOT EXISTS input_source_events (
    event_id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    source_id         TEXT NOT NULL REFERENCES input_sources(id) ON DELETE CASCADE,
    channel_key       TEXT NOT NULL DEFAULT 'default',
    payload           JSONB NOT NULL,
    payload_hash      TEXT,
    source_timestamp  TIMESTAMPTZ,
    received_at       TIMESTAMPTZ DEFAULT NOW(),
    run_id            TEXT,
    team_id           TEXT,
    agent_id          TEXT,
    source_kind       TEXT NOT NULL DEFAULT 'web_api',
    source_channel    TEXT NOT NULL,
    payload_kind      TEXT NOT NULL DEFAULT 'event',
    tenant_id         TEXT NOT NULL DEFAULT 'default'
);

CREATE TABLE IF NOT EXISTS input_source_latest (
    source_id         TEXT NOT NULL REFERENCES input_sources(id) ON DELETE CASCADE,
    channel_key       TEXT NOT NULL DEFAULT 'default',
    event_id          UUID REFERENCES input_source_events(event_id) ON DELETE SET NULL,
    payload           JSONB NOT NULL,
    received_at       TIMESTAMPTZ DEFAULT NOW(),
    source_timestamp  TIMESTAMPTZ,
    tenant_id         TEXT NOT NULL DEFAULT 'default',
    PRIMARY KEY (source_id, channel_key)
);

CREATE TABLE IF NOT EXISTS input_source_windows (
    source_id     TEXT NOT NULL REFERENCES input_sources(id) ON DELETE CASCADE,
    channel_key   TEXT NOT NULL DEFAULT 'default',
    window_key    TEXT NOT NULL,
    summary       TEXT NOT NULL,
    payload       JSONB NOT NULL DEFAULT '{}'::jsonb,
    count         INTEGER NOT NULL DEFAULT 0,
    started_at    TIMESTAMPTZ NOT NULL,
    ended_at      TIMESTAMPTZ NOT NULL,
    tenant_id     TEXT NOT NULL DEFAULT 'default',
    PRIMARY KEY (source_id, channel_key, window_key)
);

CREATE INDEX IF NOT EXISTS idx_input_sources_scope
    ON input_sources(tenant_id, scope_kind, scope_ref);

CREATE INDEX IF NOT EXISTS idx_input_sources_status
    ON input_sources(status);

CREATE INDEX IF NOT EXISTS idx_input_source_events_source_received
    ON input_source_events(source_id, channel_key, received_at DESC);

CREATE INDEX IF NOT EXISTS idx_input_source_events_team
    ON input_source_events(team_id, received_at DESC);

CREATE INDEX IF NOT EXISTS idx_input_source_windows_source
    ON input_source_windows(source_id, channel_key, ended_at DESC);
