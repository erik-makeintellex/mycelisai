-- V7: Inception Recipes — structured prompt patterns for Soma to store and recall.
-- When Soma successfully completes a complex task (blueprint creation, analysis, team setup),
-- it can distill a structured recipe capturing: how to ask for it, key parameters, and outcome shape.
-- Recipes are dual-persisted: RDBMS (structured query) + pgvector (semantic recall).

CREATE TABLE IF NOT EXISTS inception_recipes (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id       VARCHAR(255) NOT NULL DEFAULT 'default',
    category        VARCHAR(100) NOT NULL,       -- 'blueprint', 'analysis', 'team_setup', 'deployment', 'research', etc.
    title           VARCHAR(500) NOT NULL,       -- "How to create a microservices deployment blueprint"
    intent_pattern  TEXT NOT NULL,               -- The structured way to phrase this request
    parameters      JSONB DEFAULT '{}',          -- Key parameters: {"service_count": "number of services", ...}
    example_prompt  TEXT,                        -- A concrete example prompt that worked well
    outcome_shape   TEXT,                        -- What the successful outcome looks like
    source_run_id   UUID,                        -- The run that generated this recipe (nullable)
    source_session_id UUID,                      -- The session that generated this (nullable)
    agent_id        VARCHAR(255) NOT NULL DEFAULT 'admin',
    tags            TEXT[] DEFAULT '{}',          -- Searchable tags for filtering
    quality_score   FLOAT DEFAULT 0.0,           -- Feedback quality (0.0 = unrated, 1.0 = excellent)
    usage_count     INT DEFAULT 0,               -- Times this recipe was retrieved and used
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_inception_category ON inception_recipes(category);
CREATE INDEX idx_inception_agent    ON inception_recipes(agent_id, created_at DESC);
CREATE INDEX idx_inception_tags     ON inception_recipes USING gin(tags);
CREATE INDEX idx_inception_title_trgm ON inception_recipes USING gin(title gin_trgm_ops);
