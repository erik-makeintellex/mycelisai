# 🧠 Cognitive Providers Support

Mycelis Core uses a flexible universal adapter system allowing you to connect to various AI providers.
Configuration is managed via `core/config/brain.yaml`.

## Supported Protocols
1.  **Ollama** (Local/Private) - Default dev setup.
2.  **OpenAI** (Universal) - Supports OpenAI, xAI, DeepSeek, OpenRouter, vLLM.
3.  **Anthropic** (Native) - Direct support for Claude.
4.  **Google** (Native) - Direct support for Gemini.

## Configuration Guide (`brain.yaml`)

The V2 Configuration splits **Providers** (Hardware/API) from **Profiles** (Intent/Role).

### 1. Define Providers
Providers are the actual connections to AI services.

```yaml
providers:
  # --- Local Development ---
  local_ollama:
    type: "openai_compatible"
    endpoint: "http://host.docker.internal:11434/v1" # Use docker internal host!
    model_id: "qwen2.5-coder:7b"
    api_key: "ollama" # Dummy key

  # --- Commercial ---
  
  # Method A: OpenAI (or Compatible)
  production_gpt4:
    type: "openai"
    endpoint: "https://api.openai.com/v1"
    model_id: "gpt-4-turbo"
    api_key_env: "OPENAI_API_KEY"

  # Method B: Anthropic Native
  production_claude:
    type: "anthropic"
    model_id: "claude-3-5-sonnet-20240620"
    api_key_env: "ANTHROPIC_API_KEY"
  
  # Method C: Google Gemini Native
  production_gemini:
    type: "google"
    model_id: "gemini-1.5-pro"
    api_key_env: "GEMINI_API_KEY"
```

### 2. Map Profiles
Profiles define *how* the system uses AI. You map a Profile to a Provider ID.

```yaml
profiles:
  sentry: "local_ollama"      # Security Analysis (Fast)
  architect: "production_gpt4" # Complex Reasoning (Smart)
  creative: "production_claude" # Content Generation (Nuanced)
```

## Security
**NEVER** commit API keys to `brain.yaml`.
Always use `api_key_env` to reference a variable injected via Kubernetes Secrets or `.env` files.

- `OPENAI_API_KEY`
- `ANTHROPIC_API_KEY`
- `GEMINI_API_KEY`
