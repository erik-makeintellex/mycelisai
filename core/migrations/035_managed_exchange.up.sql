CREATE TABLE IF NOT EXISTS exchange_field_registry (
    name             TEXT PRIMARY KEY,
    field_type       TEXT NOT NULL,
    semantic_meaning TEXT NOT NULL,
    indexed          BOOLEAN NOT NULL DEFAULT FALSE,
    visibility       TEXT NOT NULL DEFAULT 'default',
    usage_contexts   JSONB NOT NULL DEFAULT '[]'::jsonb,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS exchange_schema_registry (
    id                    TEXT PRIMARY KEY,
    label                 TEXT NOT NULL,
    description           TEXT NOT NULL,
    required_fields       JSONB NOT NULL DEFAULT '[]'::jsonb,
    optional_fields       JSONB NOT NULL DEFAULT '[]'::jsonb,
    required_capabilities JSONB NOT NULL DEFAULT '[]'::jsonb,
    created_at            TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS exchange_channels (
    id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name             TEXT NOT NULL UNIQUE,
    channel_type     TEXT NOT NULL,
    owner            TEXT NOT NULL,
    participants     JSONB NOT NULL DEFAULT '[]'::jsonb,
    schema_id        TEXT NOT NULL REFERENCES exchange_schema_registry(id),
    retention_policy TEXT NOT NULL,
    visibility       TEXT NOT NULL DEFAULT 'advanced',
    description      TEXT NOT NULL DEFAULT '',
    metadata         JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS exchange_threads (
    id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    channel_id     UUID NOT NULL REFERENCES exchange_channels(id) ON DELETE CASCADE,
    thread_type    TEXT NOT NULL,
    title          TEXT NOT NULL,
    status         TEXT NOT NULL DEFAULT 'active',
    participants   JSONB NOT NULL DEFAULT '[]'::jsonb,
    continuity_key TEXT,
    created_by     TEXT NOT NULL,
    metadata       JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS exchange_items (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    channel_id   UUID NOT NULL REFERENCES exchange_channels(id) ON DELETE CASCADE,
    schema_id    TEXT NOT NULL REFERENCES exchange_schema_registry(id),
    payload      JSONB NOT NULL,
    created_by   TEXT NOT NULL,
    addressed_to TEXT,
    thread_id    UUID REFERENCES exchange_threads(id) ON DELETE SET NULL,
    visibility   TEXT NOT NULL DEFAULT 'default',
    metadata     JSONB NOT NULL DEFAULT '{}'::jsonb,
    summary      TEXT NOT NULL DEFAULT '',
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_exchange_channels_type
    ON exchange_channels(channel_type, visibility, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_exchange_threads_channel_status
    ON exchange_threads(channel_id, status, updated_at DESC);

CREATE INDEX IF NOT EXISTS idx_exchange_threads_continuity
    ON exchange_threads(continuity_key)
    WHERE continuity_key IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_exchange_items_channel_created
    ON exchange_items(channel_id, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_exchange_items_thread_created
    ON exchange_items(thread_id, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_exchange_items_schema_created
    ON exchange_items(schema_id, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_exchange_items_payload
    ON exchange_items USING GIN(payload);

CREATE OR REPLACE FUNCTION update_exchange_threads_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS exchange_threads_updated_at ON exchange_threads;
CREATE TRIGGER exchange_threads_updated_at
    BEFORE UPDATE ON exchange_threads
    FOR EACH ROW EXECUTE FUNCTION update_exchange_threads_updated_at();
