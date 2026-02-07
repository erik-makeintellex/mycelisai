-- Migration 006: Cognitive Registry (The Brain Stem)
CREATE TABLE IF NOT EXISTS llm_providers (
    id TEXT PRIMARY KEY,
    -- "ollama-local", "openai-cloud"
    name TEXT NOT NULL,
    driver TEXT NOT NULL,
    -- "ollama", "openai", "anthropic"
    base_url TEXT NOT NULL,
    api_key_env_var TEXT,
    -- Name of env var to read key from (not the key itself!)
    config JSONB DEFAULT '{}',
    -- Model mapping, strict params
    is_default BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMPTZ DEFAULT NOW()
);
CREATE TABLE IF NOT EXISTS system_config (
    key TEXT PRIMARY KEY,
    -- "role.architect", "role.overseer"
    value TEXT NOT NULL,
    -- "ollama-local" (References llm_providers.id)
    updated_at TIMESTAMPTZ DEFAULT NOW()
);
-- Seed Data: The Sovereign Default
INSERT INTO llm_providers (id, name, driver, base_url, is_default)
VALUES (
        'local-sovereign',
        'Local Sovereign (Ollama)',
        'ollama',
        'http://ollama-service:11434',
        -- Internal K8s DNS
        TRUE
    ) ON CONFLICT (id) DO NOTHING;
-- Seed Data: System Roles
INSERT INTO system_config (key, value)
VALUES ('role.architect', 'local-sovereign'),
    ('role.coder', 'local-sovereign'),
    ('role.sentry', 'local-sovereign') ON CONFLICT (key) DO NOTHING;