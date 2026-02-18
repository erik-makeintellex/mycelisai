-- Migration 013: Fix LLM Provider URL and Roles
-- Align provider 'local-ollama-dev' to localhost and map roles to it.
-- 1. Correct the URL for the dev provider (was 192.168...)
UPDATE llm_providers
SET base_url = 'http://localhost:11434'
WHERE id = 'local-ollama-dev';
-- 2. Repoint roles to the working dev provider
INSERT INTO system_config (key, value)
VALUES ('role.architect', 'local-ollama-dev'),
    ('role.coder', 'local-ollama-dev'),
    ('role.sentry', 'local-ollama-dev'),
    ('role.chat', 'local-ollama-dev'),
    ('role.creative', 'local-ollama-dev'),
    ('role.overseer', 'local-ollama-dev') ON CONFLICT (key) DO
UPDATE
SET value = EXCLUDED.value;