-- Migration 015: Fix Provider IPv4
-- Force 127.0.0.1 to avoid localhost IPv6 resolution issues.
UPDATE llm_providers
SET base_url = 'http://127.0.0.1:11434/v1'
WHERE id = 'local-ollama-dev';