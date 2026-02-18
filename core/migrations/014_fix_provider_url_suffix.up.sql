-- Migration 014: Fix LLM Provider URL Suffix
-- Append /v1 to local-ollama-dev URL for OpenAI compatibility.
UPDATE llm_providers
SET base_url = 'http://localhost:11434/v1'
WHERE id = 'local-ollama-dev';