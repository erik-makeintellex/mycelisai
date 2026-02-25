-- V7: Conversation turns — full-fidelity agent conversation log.
-- Separate from mission_events (lightweight audit); turns are full-text blobs (10KB+).
-- session_id groups turns for a single chat cycle; run_id is nullable (standing-team chats have no run).

CREATE TABLE IF NOT EXISTS conversation_turns (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    run_id          UUID,                        -- nullable: standing-team chats have no run
    session_id      UUID NOT NULL,               -- groups turns for one chat cycle
    tenant_id       VARCHAR(255) NOT NULL DEFAULT 'default',
    agent_id        VARCHAR(255) NOT NULL,       -- "admin", "council-architect", etc.
    team_id         VARCHAR(255),
    turn_index      INT NOT NULL,                -- order within session
    role            VARCHAR(50) NOT NULL,         -- system|user|assistant|tool_call|tool_result|interjection
    content         TEXT NOT NULL,
    provider_id     VARCHAR(100),
    model_used      VARCHAR(255),
    tool_name       VARCHAR(255),
    tool_args       JSONB,
    parent_turn_id  UUID,                        -- links tool_result → tool_call
    consultation_of VARCHAR(255),                -- who was consulted
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_convturns_run     ON conversation_turns(run_id, created_at);
CREATE INDEX idx_convturns_session ON conversation_turns(session_id, turn_index);
CREATE INDEX idx_convturns_agent   ON conversation_turns(agent_id, created_at);
