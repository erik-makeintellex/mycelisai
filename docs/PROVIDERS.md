# 🧠 Cognitive Providers Support

Mycelis Core uses a flexible routing system allowing you to connect to various AI providers.
Configuration is strictly managed via `core/config/brain.yaml` and Environment Variables.

## Supported Protocols
1. **Ollama** (Local/Private) - Default
2. **OpenAI Compatible** (Universal) - Supports OpenAI, xAI (Grok), Gemini, DeepSeek, OpenRouter, vLLM.

## Configuration Guide (`brain.yaml`)

### 1. The Model Definition
To add a new resource, define it in the `models` list.
```yaml
models:
  - id: "my-model-id"           # Internal Slug (used in Profiles)
    provider: "openai"          # Protocol (matches "openai" or "ollama")
    name: "upstream-model-id"   # The actual API model name (e.g. grok-beta)
    endpoint: "https://..."     # Base API URL
    auth_key_env: "MY_API_KEY"  # Name of the Env Var containing the secret
```

### 2. Examples

#### 🌌 Grok (xAI)
*   **Provider**: `openai`
*   **Endpoint**: `https://api.x.ai/v1`
*   **Env Var**: `XAI_API_KEY`
```yaml
- id: "grok-beta"
  provider: "openai"
  name: "grok-beta"
  endpoint: "https://api.x.ai/v1"
  auth_key_env: "XAI_API_KEY"
```

#### 💎 Gemini (Google)
*   **Provider**: `openai` (via their compatibility layer)
*   **Endpoint**: `https://generativelanguage.googleapis.com/v1beta/openai/`
*   **Env Var**: `GEMINI_API_KEY`
```yaml
- id: "gemini-flash"
  provider: "openai"
  name: "gemini-1.5-flash"
  endpoint: "https://generativelanguage.googleapis.com/v1beta/openai/"
  auth_key_env: "GEMINI_API_KEY"
```

#### 🤖 Claude (Anthropic)
*Note: Direct Anthropic API support is planned. Currently recommended via OpenRouter.*
```yaml
- id: "claude-sonnet"
  provider: "openai" # Using OpenRouter Bridge
  name: "anthropic/claude-3.5-sonnet"
  endpoint: "https://openrouter.ai/api/v1"
  auth_key_env: "OPENROUTER_API_KEY"
```

#### 🦙 Ollama (Local)
```yaml
- id: "local-qwen"
  provider: "ollama"
  name: "qwen2.5:7b"
  endpoint: "http://host.docker.internal:11434"
```

## Security
**NEVER** commit API keys to `brain.yaml`.
Always use `auth_key_env` to reference a variable injected via Kubernetes Secrets or `.env` files.
