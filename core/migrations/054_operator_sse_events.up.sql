CREATE TABLE IF NOT EXISTS operator_sse_events (
    sequence BIGSERIAL PRIMARY KEY,
    payload TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_operator_sse_events_created_at
    ON operator_sse_events (created_at);
