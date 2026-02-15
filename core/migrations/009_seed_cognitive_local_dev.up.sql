-- Migration 009: Seed Cognitive Registry for Local Development
-- Dependency: Migration 006 (creates llm_providers + system_config tables)
--
-- WHY: The Meta-Architect, Archivist, and all cognitive profiles require
-- entries in llm_providers and system_config to resolve. Without this seed,
-- the system returns 502 on any cognitive endpoint when the YAML config
-- fails to load or DB is the only config source.

-- 1. Local Dev Ollama Provider (LAN-accessible Ollama instance)
INSERT INTO llm_providers (id, name, driver, base_url, config, is_default)
VALUES (
    'local-ollama-dev',
    'Local Ollama (Dev LAN)',
    'ollama',
    'http://192.168.50.156:11434',
    '{"model_id": "qwen2.5-coder:7b"}'::jsonb,
    FALSE
) ON CONFLICT (id) DO UPDATE SET
    base_url = EXCLUDED.base_url,
    config = EXCLUDED.config;

-- 2. Ensure K8s sovereign entry has model config (patch 006 gap)
UPDATE llm_providers
SET config = '{"model_id": "qwen2.5-coder:7b"}'::jsonb
WHERE id = 'local-sovereign' AND (config IS NULL OR config = '{}'::jsonb);

-- 3. Complete Role→Provider Mappings (006 only had architect, coder, sentry)
INSERT INTO system_config (key, value) VALUES
    ('role.architect', 'local-sovereign'),
    ('role.coder',     'local-sovereign'),
    ('role.sentry',    'local-sovereign'),
    ('role.chat',      'local-sovereign'),
    ('role.creative',  'local-sovereign'),
    ('role.overseer',  'local-sovereign')
ON CONFLICT (key) DO NOTHING;
