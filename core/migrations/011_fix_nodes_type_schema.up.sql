-- Migration 011: Fix Nodes Schema Drift
-- Adds missing 'type' column to nodes table required by heartbeat logic.
ALTER TABLE nodes
ADD COLUMN IF NOT EXISTS type TEXT DEFAULT 'worker';
CREATE INDEX IF NOT EXISTS idx_nodes_type ON nodes(type);