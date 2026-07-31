-- Migration 053: Durable, replayable dispatch for approved execution plans.

CREATE TABLE IF NOT EXISTS execution_dispatch_outbox (
    id UUID PRIMARY KEY,
    idempotency_key TEXT NOT NULL UNIQUE,
    dispatch_kind TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'staged',
    run_id TEXT NOT NULL,
    intent_proof_id UUID NOT NULL REFERENCES intent_proofs(id) ON DELETE CASCADE,
    contract_id UUID NULL,
    team_id TEXT NULL,
    work_item_id TEXT NULL,
    source_kind TEXT NOT NULL,
    source_channel TEXT NOT NULL,
    payload_kind TEXT NOT NULL,
    payload JSONB NOT NULL DEFAULT '{}'::jsonb,
    attempt_count INTEGER NOT NULL DEFAULT 0,
    available_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    lease_until TIMESTAMPTZ NULL,
    last_error TEXT NULL,
    recovery JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    published_at TIMESTAMPTZ NULL,
    completed_at TIMESTAMPTZ NULL
);

CREATE INDEX IF NOT EXISTS idx_execution_dispatch_outbox_ready
    ON execution_dispatch_outbox (status, available_at, created_at);

CREATE INDEX IF NOT EXISTS idx_execution_dispatch_outbox_run
    ON execution_dispatch_outbox (run_id, created_at);
