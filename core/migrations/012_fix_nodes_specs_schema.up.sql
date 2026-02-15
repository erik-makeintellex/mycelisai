-- Migration 012: Fix Nodes Specs Schema
-- Adds missing 'specs' column to nodes table required by bootstrap service.
ALTER TABLE nodes
ADD COLUMN IF NOT EXISTS specs JSONB DEFAULT '{}';