-- Migration 055: Durable acceptance and replay protection for team bus signals.

CREATE TABLE IF NOT EXISTS team_signal_receipts (
    id UUID PRIMARY KEY,
    tenant_id TEXT NOT NULL DEFAULT 'default',
    team_id TEXT NOT NULL,
    work_item_id UUID NOT NULL REFERENCES team_work_items(id) ON DELETE CASCADE,
    direction TEXT NOT NULL,
    signal_key TEXT NOT NULL,
    source_channel TEXT NOT NULL DEFAULT '',
    accepted_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_team_signal_receipt_direction
        CHECK (direction IN ('command', 'result')),
    CONSTRAINT uq_team_signal_receipt
        UNIQUE (tenant_id, team_id, work_item_id, direction, signal_key)
);

CREATE INDEX IF NOT EXISTS idx_team_signal_receipts_work
    ON team_signal_receipts (team_id, work_item_id, accepted_at DESC);
