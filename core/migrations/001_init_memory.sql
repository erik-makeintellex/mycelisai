-- Enable Vector Search for future Intelligence phases
CREATE EXTENSION IF NOT EXISTS vector;

-- 1. The "Log" (Immutable Event History)
-- Optimized for high-volume time-series writes.
CREATE TABLE IF NOT EXISTS log_entries (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    trace_id TEXT NOT NULL,
    timestamp TIMESTAMPTZ NOT NULL,
    level TEXT NOT NULL,       -- INFO, WARN, ERROR
    source TEXT NOT NULL,      -- agent:finance
    intent TEXT NOT NULL,      -- spend_money
    message TEXT,
    context JSONB,             -- The full payload snapshot
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Optimize for Cortex UI Stream retrieval
CREATE INDEX IF NOT EXISTS idx_logs_time ON log_entries(timestamp DESC);
CREATE INDEX IF NOT EXISTS idx_logs_source ON log_entries(source);

-- 2. The "State" (Current Projected Reality)
-- A unified registry of who is alive and what they are doing.
CREATE TABLE IF NOT EXISTS agent_registry (
    agent_id TEXT PRIMARY KEY,
    team_id TEXT NOT NULL,
    status TEXT NOT NULL,      -- IDLE, BUSY, OFFLINE, ERROR
    last_seen TIMESTAMPTZ NOT NULL,
    meta JSONB,                -- Battery, Location, Current Task
    last_error_summary TEXT    -- For self-healing logic
);
