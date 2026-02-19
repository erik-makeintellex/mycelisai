-- Migration 021: Conversation Summaries
-- Stores LLM-generated summaries of chat windows for long-term memory.
-- Vectors go into existing context_vectors table (reuse pgvector infra).

CREATE TABLE IF NOT EXISTS conversation_summaries (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    agent_id TEXT NOT NULL DEFAULT 'admin',
    summary TEXT NOT NULL,
    key_topics TEXT[] DEFAULT '{}',
    user_preferences JSONB DEFAULT '{}',
    personality_notes TEXT,
    data_references JSONB DEFAULT '[]',
    message_count INT DEFAULT 0,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_convsumm_agent ON conversation_summaries(agent_id);
CREATE INDEX IF NOT EXISTS idx_convsumm_created ON conversation_summaries(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_convsumm_topics ON conversation_summaries USING gin(key_topics);
