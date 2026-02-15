-- Migration 005: Hardware Nodes (Bootstrap)

CREATE TABLE IF NOT EXISTS nodes (
    id TEXT PRIMARY KEY, -- "rpi-01", "drone-alpha", etc.
    status TEXT DEFAULT 'pending', -- pending, assigned, error
    capabilities JSONB DEFAULT '[]', -- ["gpio", "camera"]
    last_seen TIMESTAMPTZ DEFAULT NOW(),
    created_at TIMESTAMPTZ DEFAULT NOW()
);
